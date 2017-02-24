#include "network.h"
#include "monitor.h"

#define REMOTE_URI "/any"
#define BUF_MAX 4096

char *
_file_from_path(char *path)
{
	if (!path) return NULL;

	char *t = strrchr(path, '/');
	if (t) {
		t++;
		return t;
	}
	return path;
}

char *
_directory_from_path(char *path)
{
	if (!path) return NULL;
	char *t = strrchr(path, '/');
	if (*t) {
		t++;
		return t;
	}
	return path;
}

int
authenticate(void *self)
{
	monitor_t *mon = self;

	char echo_to_screen[BUF_MAX];
	snprintf(echo_to_screen, sizeof(echo_to_screen), "%s@%s's password: ",
			mon->username, mon->hostname);

	char *guess = getpass(echo_to_screen);
	if (strlen(guess) == 0) exit(1 << 2);

	mon->password = strdup(guess);

	mon->sock = Connect(mon->hostname, DEFAULT_PORT);
	if (!mon->sock)
		mon->error("Could not connect!");

	char post[BUF_MAX];
	char *fmt =
		"POST %s HTTP/1.1\r\n"
		"Host: %s\r\n"
		"Username: %s\r\n"
		"Password: %s\r\n"
		"Action: AUTH\r\n"
		"\r\n";

	snprintf(post, sizeof(post), fmt, REMOTE_URI, mon->hostname,
			mon->username, mon->password);

	write(mon->sock, post, strlen(post));

	char buf[BUF_MAX] = { 0 };
	ssize_t len = 0;

	len = read(mon->sock, buf, sizeof(buf));
	buf[len] = '\0';

	int status = 1000;
#define AUTH_STATUS_STR "status: "
	char *status_position = strstr(buf, AUTH_STATUS_STR);
	if (status_position) {
		status_position = strchr(status_position, ':');
		status_position++;
		char *status_start = status_position;
		while (*status_position != '\r') 
			status_position++;

		if (*status_position == '\r') {
			*status_position = '\0';
			status = atoi(status_start);
		}
	}

	if (status != 0)
		mon->error("Invalid username or password");

	close(mon->sock);

	return 0;
}

int 
remote_file_del(void *self, char *file)
{
	monitor_t *mon = self;
        char path[PATH_MAX] = { 0 };
        snprintf(path, sizeof(path), "%s", file);

        mon->sock = Connect(mon->hostname, DEFAULT_PORT);
        if (!mon->sock) {
                fprintf(stderr, "Could not connect()\n");
                return false;
        }

        int content_length = 0;

        char *file_from_path = _file_from_path(path);

        char dirname[PATH_MAX] = { 0 };
        snprintf(dirname, sizeof(dirname), "%s", mon->directories[0]);
        char *dir_from_path = _directory_from_path(dirname);

        char post[BUF_MAX] = { 0 };
        char *fmt =
            "POST %s HTTP/1.1\r\n"
            "Host: %s\r\n"
            "Content-Length: %d\r\n"
            "Username: %s\r\n"
            "Password: %s\r\n"
            "Directory: %s\r\n" "Filename: %s\r\n" "Action: DEL\r\n\r\n";

        snprintf(post, sizeof(post), fmt, REMOTE_URI, mon->hostname,
                 content_length, mon->username, mon->password, dir_from_path,
                 file_from_path);
        write(mon->sock, post, strlen(post));

        close(mon->sock);

        return true;
}

void 
_untrim(char *str)
{
	char *s = str;
	while (*s) {
		if (*s == '_')
			*s = ' ';
		s++;
	}
}
int remote_file_add(void *self, char *file)
{
	monitor_t *mon = self;
        char path[PATH_MAX] = { 0 };
        snprintf(path, sizeof(path), file);
	_untrim(path);
        char dirname[PATH_MAX] = { 0 };
        snprintf(dirname, sizeof(dirname), "%s", mon->directories[0]);

        mon->sock = Connect(mon->hostname, DEFAULT_PORT);
        if (!mon->sock) {
                fprintf(stderr, "Could not connect()\n");
                return false;
        }
        struct stat fstats;
        if (strlen(path) == 0) {
                return false;
        }

        if (stat(path, &fstats) < 0) {
                return false;
        }

        FILE *f = fopen(path, "rb");
        if (f == NULL) {
                mon->error
                    ("Unable to open filename this should not happen!");
                     
        }
#define CHUNK 1024

        char buffer[CHUNK + 1] = { 0 };

        int content_length = fstats.st_size;

        char *file_from_path = _file_from_path(path);
        char *dir_from_path = _directory_from_path(dirname);

        char post[BUF_MAX] = { 0 };
        char *fmt =
            "POST %s HTTP/1.1\r\n"
            "Host: %s\r\n"
            "Content-Length: %d\r\n"
            "Username: %s\r\n"
            "Password: %s\r\n"
            "Filename: %s\r\n" "Directory: %s\r\n" "Action: ADD\r\n\r\n";

        snprintf(post, sizeof(post), fmt, REMOTE_URI, mon->hostname,
                 content_length, mon->username, mon->password, file_from_path,
                 dir_from_path);
        write(mon->sock, post, strlen(post));
        int total = 0;

        int size = content_length;
        while (size) {
                while (1) {
                        int count = fread(buffer, 1, CHUNK, f);
                        //ssize_t bytes = send(sock, buffer, count, 0);
                        // this is exactly the same with 0 flag
                        ssize_t bytes = write(mon->sock, buffer, count);
                        if (bytes < count) {
                                return false;
                        }
                        if (bytes == 0) {
                                break;
                        } else if (bytes < 0) {
                                return false;
                        } else {
                                size -= bytes;
                                total += bytes;
                        }
                }
        }

        if (size != 0) {
                mon->error("wtf?");
        }

        close(mon->sock);
        fclose(f);
        return true;
}


int 
Connect(const char *hostname, int port)
{
	int sock;
	struct hostent *host;
	struct sockaddr_in inaddr;

	sock = socket(AF_INET, SOCK_STREAM, 0);
	if (sock < 0) return 0;

	host = gethostbyname(hostname);
	if (host == NULL) return 0;

	inaddr.sin_family = AF_INET;
	inaddr.sin_port = htons(port);
	inaddr.sin_addr = *((struct in_addr *)host->h_addr_list[0]);
	memset(&inaddr.sin_zero, 0, 8);

	int status = connect(sock, (struct sockaddr *) &inaddr,
				sizeof(struct sockaddr));

	if (status == 0) {
		return sock;
	}

	return 0;
}


