#include <sea/list.h>
#include <sea/file.h>
#include <unistd.h>
#include <stdbool.h>
#include <pthread.h>

typedef enum {
   MONITOR_EVENT_CALLBACK_FILE_ADD,
   MONITOR_EVENT_CALLBACK_FILE_DEL,
   MONITOR_EVENT_CALLBACK_FILE_MOD,
   MONITOR_EVENT_CALLBACK_DIR_ADD,
   MONITOR_EVENT_CALLBACK_DIR_DEL,
   MONITOR_EVENT_CALLBACK_DIR_MOD,
} monitor_event_t;

typedef void (*callback_fn)(const char *path, monitor_event_t type, void *data);

typedef struct monitor_t {
   char *path;
   list_t *prev_list;
   int         enabled;

   pthread_t   thread;
   callback_fn file_added_cb;
   callback_fn file_deleted_cb;
   callback_fn file_modified_cb;

   callback_fn dir_added_cb;
   callback_fn dir_deleted_cb;
   callback_fn dir_modified_cb;

   void       *file_added_data;
   void       *file_deleted_data;
   void       *file_modified_data;

   void       *dir_added_data;
   void       *dir_deleted_data;
   void       *dir_modified_data;
} monitor_t;

typedef struct file_info_t
{
   stat_t st; 
   char *path;
} file_info_t;

static int
_path_scan_cb(const char *path, stat_t *st, void *data)
{
   file_info_t *entry;
   list_t *l = data;

   entry = malloc(sizeof(file_info_t));
   entry->path = strdup(path);
   entry->st = *st; 

   l = list_add(l, entry);

   return 0;
}

static void
_file_list_free(list_t *next_list)
{
   list_t *next, *node = next_list;
   file_info_t *file;

   while (node)
     {
        next = node->next;
        file = node->data;
        free(file->path);
        free(file);
        file = NULL;
        free(node);
        node = NULL;
        node = next;
     }
}

static int
_cmp_cb(void *p1, void *p2)
{
   file_info_t *file, *file2;

   file = p1; file2 = p2;

   return strcmp(file->path, file2->path);
}

static void
_monitor_engine_fallback(monitor_t *monitor)
{
   list_t *l, *l2;
   file_info_t *file, *file2;
   const char *directory = monitor->path;

   list_t *next_list = list_new();

   file = malloc(sizeof(file_info_t));
   file->path = strdup(directory);
   file->st = *file_path_stat(directory);

   next_list = list_add(next_list, file);

   file_path_walk(monitor->path, _path_scan_cb, next_list); 

   l = monitor->prev_list;
   while (l) 
     {
        file = l->data;
        bool exists = false;
        l2 = next_list;
        while (l2)
          {
             file2 = l2->data;
             if (file->st.inode == file2->st.inode)
               {
                  if (!strcmp(file->path, file2->path))
                    exists = true;

                  if (file->st.mtime != file2->st.mtime)
                    {
                       if (S_ISDIR(file->st.mode))
                         {
                            if (monitor->dir_modified_cb)
                              {
                                 monitor->dir_modified_cb(file->path, MONITOR_EVENT_CALLBACK_DIR_MOD, monitor->dir_modified_data);
                              }
                         }
                       else
                         {
                            if (monitor->file_modified_cb)
                              {
                                 monitor->file_modified_cb(file->path, MONITOR_EVENT_CALLBACK_FILE_MOD, monitor->file_modified_data);
                              }
                         }
                    }
               }
             l2 = l2->next;
          }
     
        if (!exists)
          {
             if (S_ISDIR(file->st.mode))
               {
                  if (monitor->dir_deleted_cb)
                    {
                       monitor->dir_deleted_cb(file->path, MONITOR_EVENT_CALLBACK_DIR_DEL, monitor->dir_deleted_data);
                    }
               }
             else
               {
                  if (monitor->file_deleted_cb)
                    {
                       monitor->file_deleted_cb(file->path, MONITOR_EVENT_CALLBACK_FILE_DEL, monitor->file_deleted_data);
                    }
               }
          }

        l = l->next;
     }

   l = next_list;
   while (l && monitor->prev_list)
     {
        file = l->data;
        bool exists = false;
        l2 = monitor->prev_list;
        while (l2)
          {
             file2 = l2->data;
             if ((file->st.inode == file2->st.inode) &&
                  (!strcmp(file->path, file2->path)))
               {
                  exists = true;
                  break;
               }

             l2 = l2->next;
          }
       if (!exists)
         {
            if (S_ISDIR(file->st.mode))
              {
                 if (monitor->dir_added_cb)
                   monitor->dir_added_cb(file->path, MONITOR_EVENT_CALLBACK_DIR_ADD, monitor->dir_added_data);
              }
            else
              {
                 if (monitor->file_added_cb)
                   monitor->file_added_cb(file->path, MONITOR_EVENT_CALLBACK_FILE_ADD, monitor->file_added_data);
              }
          }
        l = l->next;
     }

   _file_list_free(monitor->prev_list);

   monitor->prev_list = next_list;
}

void
monitor_watch(monitor_t *monitor)
{
   if (!monitor->path) return;

   while (monitor->enabled)
     {
        _monitor_engine_fallback(monitor);
        sleep(2);
     }

   _file_list_free(monitor->prev_list);
}

static void *_monitor_watch_thread_cb(void *arg)
{
   monitor_t *monitor = arg;

   monitor_watch(monitor);

   return ((void *) 0);
}
void
monitor_path_set(monitor_t *monitor, const char *directory)
{
   monitor->path = strdup(directory);
}

int
monitor_bg_start(monitor_t *monitor)
{
  int error = pthread_create(&monitor->thread, NULL, _monitor_watch_thread_cb, monitor);
  if (!error)
    monitor->enabled = true;

  return error;
}

monitor_t *
monitor_new(void)
{
   monitor_t *monitor;

   monitor = calloc(1, sizeof(monitor_t));
   monitor->prev_list = list_new();


   return monitor;
}

void monitor_stop(monitor_t *monitor)
{
   void *ret = NULL;

   monitor->enabled = false;

   pthread_join(monitor->thread, ret);
}

static void
monitor_event_callback_set(monitor_t *monitor, monitor_event_t type, callback_fn func, void *data)
{
   switch (type) {
      case MONITOR_EVENT_CALLBACK_FILE_ADD:
         monitor->file_added_cb = func;
         monitor->file_added_data = data;
      break;

      case MONITOR_EVENT_CALLBACK_FILE_MOD:
         monitor->file_modified_cb = func;
         monitor->file_modified_data = data;
      break;

      case MONITOR_EVENT_CALLBACK_FILE_DEL:
         monitor->file_deleted_cb = func;
         monitor->file_deleted_data = data;
      break;

      case MONITOR_EVENT_CALLBACK_DIR_ADD:
         monitor->dir_added_cb = func;
         monitor->dir_added_data = data;
      break;

      case MONITOR_EVENT_CALLBACK_DIR_MOD:
         monitor->dir_modified_cb = func;
         monitor->dir_modified_data = data;
      break;

      case MONITOR_EVENT_CALLBACK_DIR_DEL:
         monitor->dir_deleted_cb = func;
         monitor->dir_modified_data = data;
      break;
   };
}

static void
_monitor_change_cb(const char *path, monitor_event_t type, void *data)
{
   printf("change %s and type %d\n", path, type);
}

int main(void)
{
   list_t *l;
   file_info_t *file; 
   const char *path = "/home/netstar";

   monitor_t *monitor = monitor_new();

   monitor_path_set(monitor, path);

   monitor_event_callback_set(monitor, MONITOR_EVENT_CALLBACK_FILE_ADD, _monitor_change_cb, NULL);
   monitor_event_callback_set(monitor, MONITOR_EVENT_CALLBACK_FILE_DEL, _monitor_change_cb, NULL);
   monitor_event_callback_set(monitor, MONITOR_EVENT_CALLBACK_FILE_MOD, _monitor_change_cb, NULL);
   monitor_event_callback_set(monitor, MONITOR_EVENT_CALLBACK_DIR_ADD, _monitor_change_cb, NULL);
   monitor_event_callback_set(monitor, MONITOR_EVENT_CALLBACK_DIR_DEL, _monitor_change_cb, NULL);
   monitor_event_callback_set(monitor, MONITOR_EVENT_CALLBACK_DIR_MOD, _monitor_change_cb, NULL);

   monitor_bg_start(monitor);

   for (int i = 0;i < 20; i++) {
   sleep(1);
   printf("other code!\n");
   }

   monitor_stop(monitor);

   return EXIT_SUCCESS;
}
