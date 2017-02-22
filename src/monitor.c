#include "monitor.h"
#include "network.h"
#define PROGRAM_NAME "monitor"

file_t *list_prev = NULL, *list_now = NULL;
bool _was_initialized = false;
bool _is_recursive = true;
bool quit = false;

char *_get_state_file_name(void);

int error(char *str)
{
	fprintf(stdout, "Error: %s\n", str);
	exit(1 << 7);
}

monitor_t *monitor_new(void)
{
	monitor_t *m = calloc(1, sizeof(monitor_t));
	if (!m) return NULL;
	m->error = &error;
	m->callback_set = &monitor_callback_set;
	m->watch_add = &monitor_watch_add;
	m->init = &monitor_init;
	m->watch = &monitor_watch;
	m->mainloop = &monitor_mainloop;
	m->self = m;
	m->authenticate = &authenticate;
	m->remote_add = remote_file_add;
	m->remote_del = remote_file_del;
	m->state_file = _get_state_file_name();

	return m;
}

int monitor_mainloop(void *self, int interval) 
{
	monitor_t *mon = self;
	if (mon->_d_idx == 0) exit(1 << 0);
	/*
	if (! monitor_add_callback || ! monitor_del_callback
		|| !monitor_mod_callback)
		error("callbacks not initialised!");
	*/

	list_prev = monitor_files_get(self, list_prev);	

        while (monitor_watch(self, interval) && !quit);

	return 1;
}                


int monitor_callback_set(int type, callback func)
{
        switch (type) {
        case MONITOR_ADD:
                monitor_add_callback = func;
                break;
        case MONITOR_DEL:
                monitor_del_callback = func;
                break;
        case MONITOR_MOD:
                monitor_mod_callback = func;
        };

	return 1;
}

void 
file_list_free(file_t *list)
{
	file_t *c = list;

	while (c) {
		file_t *next = c->next;
		free(c->path);
		free(c);
		c = next;
	}
}

file_t * 
file_list_add(file_t *list, const char *path, struct stat *st)
{
	file_t *c = list;

	while (c->next)
		c = c->next;

	if (! c->next) {
		c->next = calloc(1, sizeof(file_t));
		if (!c->next)
			error("calloc()");

		c = c->next;
		c->next = NULL;
		if (st)	
			c->stats = *st;	

		c->path = strdup(path);
		c->changed = 0; 
	}

	return list;
}

file_t *
file_exists(file_t *list, const char *filename)
{
	file_t *f = list->next;

	while (f) {
		if (!strcmp(f->path, filename))
			return f;
		f = f->next;
	}
	
	return NULL;
}

int
_check_add_files(monitor_t *mon, file_t *first, file_t *second)
{
	file_t *f = second->next;
	int changes = 0;

	while (f) {
		file_t *exists = file_exists(first, f->path);
		if (!exists) {
			f->changed = MONITOR_ADD;
			mon->remote_add(mon->self, f->path);
#if defined(DEBUG)		
			printf("add file : %s\n", f->path);
#endif
			++changes;
		}
		f = f->next;
	}

	return changes;
}

int
_check_del_files(monitor_t *mon, file_t *first, file_t *second)
{
	file_t *f = first->next;
	int changes = 0;

	while (f) {
		file_t *exists = file_exists(second, f->path);
		if (!exists) {
			f->changed = MONITOR_DEL;
			mon->remote_del(mon->self, f->path);
			printf("del file : %s\n", f->path);
			changes++;
		}
		f = f->next;
	}

	return changes;
}

int
_check_mod_files(monitor_t* mon, file_t *first, file_t *second)
{
	file_t *f = second->next;
	int changes = 0;

	while (f) {
		file_t *exists = file_exists(first, f->path);
		if (exists) {
			if (f->stats.st_mtime != exists->stats.st_mtime) {
				f->changed = MONITOR_MOD;
				mon->remote_add(mon->self, f->path);
				printf("mod file : %s\n", f->path);
				changes++;
			}
		}
		f = f->next;
	}
	
	return changes;
}

int
file_lists_compare(monitor_t *monitor, file_t *first, file_t *second)
{
	int modifications = 0;
	
	modifications += _check_del_files(monitor, first, second);

	modifications += _check_add_files(monitor, first, second);

	modifications += _check_mod_files(monitor, first, second);

	return modifications;
}

const char *
directory_next(monitor_t *mon)
{
	if (mon->directories[mon->_w_pos] == NULL) {
		mon->_w_pos = 0; 
		return NULL;
	}

	return mon->directories[mon->_w_pos++];
}

void
 _list_append(file_t *one, file_t *two)
{
	file_t *c = one;
	while (c->next)
		c = c->next;

	if (two->next)
		c->next = two->next;
	else
		c->next = NULL;
}

file_t * 
scan_recursive(const char *path)
{
	DIR *dir = opendir(path);
	if (!dir) return NULL;
	struct dirent *dh = NULL;
	char *directories[8192] = { NULL };
	int i; 

	file_t *list = calloc(1, sizeof(file_t));
	if (!list)
		error("calloc()");

	list->next = NULL;

	for (i = 0; i < sizeof(directories) / sizeof(char *); i++) {
		directories[i] = NULL;
	}

	i = 0;

	while ((dh = readdir(dir)) != NULL) {
		if (dh->d_name[0] == '.') continue;
		
		char buf[PATH_MAX];
		snprintf(buf, sizeof(buf), "%s%c%s", path, SLASH, dh->d_name);
		struct stat fstat;
		if (stat(buf, &fstat) < 0) continue;

		if (S_ISDIR(fstat.st_mode)) {
			directories[i++] = strdup(buf);			
			continue;
		} else {
			list = file_list_add(list, buf, &fstat);
		}
	}

	closedir(dir);

	i = 0;

	if (!_is_recursive) return list;
	
	/* 
	 * We could monitor EVERY file recursively but
	 * that is probably not a good idea!
	*/

	while (directories[i] != NULL) {
		file_t *next = NULL;
		next = scan_recursive(directories[i]);
		_list_append(list, next);
		free(directories[i++]);
	} 

	return list;
}

file_t *
monitor_files_get(monitor_t *mon, file_t *list)
{
	const char *path;

	while ((path = directory_next(mon)) != NULL) {
		list = scan_recursive(path);
	}

	return list;
}

char *
_get_state_file_name(void)
{
	char buf[PATH_MAX];

	const char *home = getenv("HOME");
	snprintf(buf, sizeof(buf), "%s/.%s", home, PROGRAM_NAME);
	struct stat st;

	if (stat(buf, &st) < 0)
		mkdir(buf, 0755);

	const char *cwd = getwd(NULL);

	char hashname[strlen(cwd) * 2 + 1];
       
	memset(hashname, 0, strlen(cwd) * 2 + 1);	

	for (int i = 0; i < strlen(cwd); i++)
		snprintf(hashname, sizeof(hashname), "%s%x", hashname, cwd[i]);

        char path[PATH_MAX];	
	snprintf(path, sizeof(path), "%s/%s", buf, hashname);
	return strdup(path);
}

void
file_list_save_state(char *path, file_t *current_files)
{
	file_t *cursor = current_files->next;
//	printf("path is %s\n", path);
	while (cursor) {
		printf("%s\n", cursor->path);
		cursor = cursor->next;
	}

}

file_t *
_monitor_compare_lists(void *self, file_t *one, file_t *two)
{
	monitor_t *m = self;
	int changes = file_lists_compare(m, one, two);
	file_list_free(one);
	one = two;
        
	if (changes) {
		file_list_save_state(m->state_file, one);
	}

	return one;
}

int 
monitor_watch(void *self, int poll)
{
	if (!_was_initialized) return 0;

	list_now = monitor_files_get(self, list_now);
	list_prev = _monitor_compare_lists(self, list_prev, list_now);  
	sleep(poll);

	return 1;
}
	
int
monitor_watch_add(void *self, const char *path)
{
	monitor_t *mon = self;
	if (mon->_d_idx >= DIRS_MAX) 
		error("watch_add(): dirs limit reached!");	

	struct stat dstat;
	if (stat(path, &dstat) < 0)
		error("watch_add(): directory exists? check permissions.");
	
	if (!S_ISDIR(dstat.st_mode))
		error("watch_add(): not a directory.");
	
	mon->directories[mon->_d_idx++] = strdup(path);

	return 1;
}

void exit_safe(int sig)
{
        if (sig != SIGINT && sig != SIGTERM) return;
        quit = true;
}

int
monitor_init(void *self, bool recursive)
{
	monitor_t *mon = self;
	if (!recursive) 
		_is_recursive = false;

	mon->directories[mon->_d_idx] = NULL;
	mon->directories[DIRS_MAX - 1] = NULL;


        signal(SIGINT, exit_safe);
        signal(SIGTERM, exit_safe);
	
	_was_initialized = true;

	return 1;
}

