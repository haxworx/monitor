#ifndef __NETWORK_H__
#define __NETWORK_H__
#include <string.h>
#include <sys/socket.h>
#include <netdb.h>
#include <netinet/in.h>

#define DEFAULT_PORT 12345

char *strdup(const char *s);
int Connect(const char *hostname, int port);

int authenticate(void *self);
int remote_file_add(void *self, char *file);
int remote_file_del(void *self, char *file);

#endif
