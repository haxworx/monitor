#include <sea/list.h>
#include <sea/file.h>
#include <sys/types.h>
#include <sys/event.h>
#include <sys/stat.h>
#include <unistd.h>
#include <stdbool.h>

typedef struct monitor_t {
   list_t *prev_list;
   char *path;
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
_monitor_file_lists_compare(monitor_t *monitor)
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
                           printf("dir %s modified\n", file->path);
                         }
                       else
                         {
                            printf("file %s modified\n", file->path);
                         }
                    }
               }
             l2 = l2->next;
          }
     
        if (!exists)
          {
             if (S_ISDIR(file->st.mode))
               {
                  printf("dir %s deleted\n", file->path);
               }
             else
               {
                  printf("file %s deleted\n", file->path);
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
                 printf("dir %s created\n", file->path);
              }
            else
              {
                 printf("file %s created\n", file->path);
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

   for (;;)
     {
        _monitor_file_lists_compare(monitor);
        sleep(2);
     }
}

monitor_t *
monitor_new(const char *directory)
{
   monitor_t *monitor;

   if (!file_path_is_directory(directory))
     return NULL;

   monitor = malloc(sizeof(monitor_t));
   monitor->path = strdup(directory);
   monitor->prev_list = list_new();

   monitor_watch(monitor);

   return monitor;
}


int main(void)
{
   list_t *l;
   file_info_t *file; 
   const char *directory = "/home/netstar";

   monitor_t *monitor = monitor_new(directory);

}
