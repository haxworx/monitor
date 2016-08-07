#if ! defined(__MONITOR_H__)

#define __MONITOR_H__
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

#define MONITOR_NONE 0
#define MONITOR_ADD 1 
#define MONITOR_DEL 2 
#define MONITOR_MOD 3

typedef struct file_t file_t;
struct file_t {
	char *path;
	struct stat stats;
	int changed;
	file_t *next;
};

typedef int (*callback)(void *data);

callback monitor_add_callback;
callback monitor_del_callback;
callback monitor_mod_callback;

typedef int (*fn_callback_set)(int type, callback);
typedef int (*fn_watch_add)(const char * path);
typedef int (*fn_init)(bool recursive);
typedef int (*fn_watch)(int poll_interval);
typedef int (*fn_error)(char *string);

/* External functions */
int monitor_watch(int poll_interval);
int monitor_mainloop(int poll_interval);
int monitor_watch_add(const char *path);
int monitor_callback_set(int type, callback);
int monitor_init(bool);
int error(char *);

typedef struct monitor_t monitor_t;
struct monitor_t {
	fn_init init;
	fn_watch_add watch_add;	
	fn_watch watch;
	fn_watch mainloop;
	fn_callback_set callback_set;
	fn_error error;
	callback add_callback;
	callback del_callback;
	callback mod_callback;	
};

monitor_t *monitor_new(void);

/* Internal functions */
	
void file_list_free(file_t *list);
file_t *file_list_add(file_t *list, const char *path, struct stat *st);
file_t *file_exists(file_t *list, const char *filename);
int _check_add_files(file_t *first, file_t *second);
int _check_del_files(file_t *first, file_t *second);
int _check_mod_files(file_t *first, file_t *second);
void file_lists_compare(file_t *first, file_t *second);
const char *directory_next(void);
void _list_append(file_t *one, file_t *two);
file_t *scan_recursive(const char *path);
file_t * monitor_files_get(file_t *list);
file_t *_monitor_compare_lists(file_t *one, file_t *two);

#endif
