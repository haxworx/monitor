#include <sea/list.h>
#include <sea/file.h>
#include <unistd.h>
#include <stdbool.h>
#include <pthread.h>

typedef enum
{
   NOTIFY_EVENT_CALLBACK_FILE_ADD,
   NOTIFY_EVENT_CALLBACK_FILE_DEL,
   NOTIFY_EVENT_CALLBACK_FILE_MOD,
   NOTIFY_EVENT_CALLBACK_DIR_ADD,
   NOTIFY_EVENT_CALLBACK_DIR_DEL,
   NOTIFY_EVENT_CALLBACK_DIR_MOD,
} notify_event_t;

typedef void (*notify_callback_fn)(const char *path, notify_event_t type, void *data);

typedef struct notify_t
{
   char                *path;
   list_t              *prev_list;
   int                  enabled;

   pthread_t            thread;
   notify_callback_fn file_added_cb;
   notify_callback_fn file_deleted_cb;
   notify_callback_fn file_modified_cb;

   notify_callback_fn dir_added_cb;
   notify_callback_fn dir_deleted_cb;
   notify_callback_fn dir_modified_cb;

   void                *file_added_data;
   void                *file_deleted_data;
   void                *file_modified_data;

   void                *dir_added_data;
   void                *dir_deleted_data;
   void                *dir_modified_data;
} notify_t;

typedef struct file_info_t
{
   stat_t st;
   char  *path;
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
_notify_file_list_free(list_t *next_list)
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
_notify_engine_fallback(notify_t *notify)
{
   list_t *l, *l2;
   file_info_t *file, *file2;

   list_t *next_list = list_new();

   file = malloc(sizeof(file_info_t));
   file->path = strdup(notify->path);
   stat_t *tmp = file_path_stat(notify->path);
   file->st = *tmp;
   free(tmp);

   next_list = list_add(next_list, file);

   file_path_walk(notify->path, _path_scan_cb, next_list);

   l = notify->prev_list;
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
                            if (notify->dir_modified_cb)
                              {
                                 notify->dir_modified_cb(file->path, NOTIFY_EVENT_CALLBACK_DIR_MOD, notify->dir_modified_data);
                              }
                         }
                       else
                         {
                            if (notify->file_modified_cb)
                              {
                                 notify->file_modified_cb(file->path, NOTIFY_EVENT_CALLBACK_FILE_MOD, notify->file_modified_data);
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
                  if (notify->dir_deleted_cb)
                    {
                       notify->dir_deleted_cb(file->path, NOTIFY_EVENT_CALLBACK_DIR_DEL, notify->dir_deleted_data);
                    }
               }
             else
               {
                  if (notify->file_deleted_cb)
                    {
                       notify->file_deleted_cb(file->path, NOTIFY_EVENT_CALLBACK_FILE_DEL, notify->file_deleted_data);
                    }
               }
          }

        l = l->next;
     }

   l = next_list;
   while (l && notify->prev_list)
     {
        file = l->data;
        bool exists = false;
        l2 = notify->prev_list;
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
                  if (notify->dir_added_cb)
                    notify->dir_added_cb(file->path, NOTIFY_EVENT_CALLBACK_DIR_ADD, notify->dir_added_data);
               }
             else
               {
                  if (notify->file_added_cb)
                    notify->file_added_cb(file->path, NOTIFY_EVENT_CALLBACK_FILE_ADD, notify->file_added_data);
               }
          }
        l = l->next;
     }

   _notify_file_list_free(notify->prev_list);

   notify->prev_list = next_list;
}

static void
_notify_watch(notify_t *notify)
{
   if (!notify->path) return;

   while (notify->enabled)
     {
        _notify_engine_fallback(notify);
        sleep(2);
     }

   _notify_file_list_free(notify->prev_list);
}

void
notify_stop_wait(notify_t *notify)
{
   void *ret = NULL;

   notify->enabled = false;

   pthread_join(notify->thread, ret);
}

static void *
_notify_watch_thread_cb(void *arg)
{
   notify_t *notify = arg;

   _notify_watch(notify);

   return (void *)0;
}

int
notify_background_run(notify_t *notify)
{
   int error;

   if (!file_path_exists(notify->path))
     return 1;

   error = pthread_create(&notify->thread, NULL, _notify_watch_thread_cb, notify);
   if (!error)
     notify->enabled = true;

   return error;
}

void
notify_free(notify_t *notify)
{
   free(notify->path);
   free(notify);
}

notify_t *
notify_new(void)
{
   notify_t *notify;

   notify = calloc(1, sizeof(notify_t));
   notify->prev_list = list_new();

   return notify;
}

void
notify_path_set(notify_t *notify, const char *directory)
{
   notify->path = strdup(directory);
}

static void
notify_event_callback_set(notify_t *notify, notify_event_t type, notify_callback_fn func, void *data)
{
   switch (type)
     {
      case NOTIFY_EVENT_CALLBACK_FILE_ADD:
        notify->file_added_cb = func;
        notify->file_added_data = data;
        break;

      case NOTIFY_EVENT_CALLBACK_FILE_MOD:
        notify->file_modified_cb = func;
        notify->file_modified_data = data;
        break;

      case NOTIFY_EVENT_CALLBACK_FILE_DEL:
        notify->file_deleted_cb = func;
        notify->file_deleted_data = data;
        break;

      case NOTIFY_EVENT_CALLBACK_DIR_ADD:
        notify->dir_added_cb = func;
        notify->dir_added_data = data;
        break;

      case NOTIFY_EVENT_CALLBACK_DIR_MOD:
        notify->dir_modified_cb = func;
        notify->dir_modified_data = data;
        break;

      case NOTIFY_EVENT_CALLBACK_DIR_DEL:
        notify->dir_deleted_cb = func;
        notify->dir_modified_data = data;
        break;
     }
}

static void
_change_cb(const char *path, notify_event_t type, void *data)
{
   printf("change %s and type %d\n", path, type);
}

int
main(void)
{
   const char *path = "/home/netstar";

   notify_t *notify = notify_new();

   notify_path_set(notify, path);

   notify_event_callback_set(notify, NOTIFY_EVENT_CALLBACK_FILE_ADD, _change_cb, NULL);
   notify_event_callback_set(notify, NOTIFY_EVENT_CALLBACK_FILE_DEL, _change_cb, NULL);
   notify_event_callback_set(notify, NOTIFY_EVENT_CALLBACK_FILE_MOD, _change_cb, NULL);
   notify_event_callback_set(notify, NOTIFY_EVENT_CALLBACK_DIR_ADD, _change_cb, NULL);
   notify_event_callback_set(notify, NOTIFY_EVENT_CALLBACK_DIR_DEL, _change_cb, NULL);
   notify_event_callback_set(notify, NOTIFY_EVENT_CALLBACK_DIR_MOD, _change_cb, NULL);

   int res = notify_background_run(notify);
   if (res)
     {
        fprintf(stderr, "failed to put notify into the background!");
        exit(EXIT_FAILURE);
     }

   for (int i = 0; i < 20; i++) {
        sleep(1);
        printf("other code!\n");
     }

   notify_stop_wait(notify);

   notify_free(notify);

   return EXIT_SUCCESS;
}

