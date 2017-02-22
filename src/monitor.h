#if ! defined(__MONITOR_H__)

#define __MONITOR_H__
#define _DEFAULT_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <stdbool.h>
#include <string.h>
#include <unistd.h>
#include <dirent.h>
#include <limits.h>
#include <time.h>
#include <sys/stat.h>
#include <signal.h>

#define SLASH '/'

#define DIRS_MAX 12

#define MONITOR_NONE 0
#define MONITOR_ADD 1 
#define MONITOR_DEL 2 
#define MONITOR_MOD 3

typedef struct file_t file_t;
struct file_t {
	char *path;
	int size;
	int mtime;
	int changed;
	file_t *next;
};

typedef int (*callback)(void *data);

callback monitor_add_callback;
callback monitor_del_callback;
callback monitor_mod_callback;

typedef int (*fn_callback_set)(int type, callback);
typedef int (*fn_watch_add)(void *self, const char * path);
typedef int (*fn_init)(void *self, bool recursive);
typedef int (*fn_watch)(void *self, int poll_interval);
typedef int (*fn_error)(char *string);
typedef int (*fn_authenticate)(void *self);
typedef int (*fn_remote_del)(void *self, char *file);
typedef int (*fn_remote_add)(void *self, char *file);

/* External functions */
int monitor_watch(void *self, int poll_interval);
int monitor_mainloop(void *self, int poll_interval);
int monitor_watch_add(void *self, const char *path);
int monitor_callback_set(int type, callback);
int monitor_init(void *self, bool);
int error(char *);

typedef struct monitor_t monitor_t;
struct monitor_t {
	monitor_t *self;
	char *hostname;
	int sock;
	char *username;
	char *password;
	char *directories[DIRS_MAX];
	int _d_idx, _w_pos;
	char *state_file;
	fn_init init;
	fn_watch_add watch_add;	
	fn_watch watch;
	fn_watch mainloop;
	fn_callback_set callback_set;
	fn_error error;
	fn_authenticate authenticate;
	fn_remote_add remote_add;
	fn_remote_del remote_del;
	callback add_callback;
	callback del_callback;
	callback mod_callback;	
};

monitor_t *monitor_new(void);

/* Internal functions */
	
void file_list_free(file_t *list);
file_t *file_list_add(file_t *list, const char *path, struct stat *st);
file_t *file_exists(file_t *list, const char *filename);
int file_lists_compare(monitor_t *monitor, file_t *first, file_t *second);
const char *directory_next(monitor_t *mon);
void _list_append(file_t *one, file_t *two);
file_t *scan_recursive(const char *path);
file_t * monitor_files_get(monitor_t *mon, file_t *list);
file_t *_monitor_compare_lists(void *self, file_t *one, file_t *two);
int _check_add_files(monitor_t *mon, file_t *first, file_t *second);
int _check_del_files(monitor_t *mon, file_t *first, file_t *second);
int _check_mod_files(monitor_t *mon, file_t *first, file_t *second);

#endif
