#include "monitor.h"

#define CMD_LEN 1024

int do_add(void *data)
{
	file_t *f = data;
	char cmd[CMD_LEN];
	snprintf(cmd, sizeof(cmd), "scripts/add.sh '%s'", f->path);
	return system(cmd);
}

int do_del(void *data)
{
	file_t *f = data;
	char cmd[CMD_LEN];
	snprintf(cmd, sizeof(cmd), "scripts/del.sh '%s'", f->path);
	return system(cmd);
}

int do_mod(void *data)
{
	file_t *f = data;
	char cmd[CMD_LEN];
	snprintf(cmd, sizeof(cmd), "scripts/mod.sh '%s'", f->path);
	return system(cmd);
}
void usage(void)
{
	printf("monitor <path>\n");
	exit(EXIT_FAILURE);
}

int main(int argc, char **argv)
{
	time_t interval = 3;
	bool recursive = true;
	char *directory = NULL;
	
	if (argc < 2) usage();

	directory = argv[1];

	if (directory[strlen(directory) - 1] == '/'
		&& strlen(directory) > 1) {
		directory[strlen(directory) - 1] = '\0';
	}

	monitor_t *m = monitor_new();
	if (!m) 
		m->error("monitor_new()");

	m->init(recursive);

	m->watch_add(directory);
	
	m->callback_set(MONITOR_ADD, do_add);
	m->callback_set(MONITOR_DEL, do_del);
	m->callback_set(MONITOR_MOD, do_mod);

	// daemon(1, 0);

	printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n");
	printf("See: http://haxlab.org\n\n");
	printf("Monitoring: %s\n\n", directory);

	m->mainloop(interval);

	free(m);
		
	return EXIT_SUCCESS;
}

