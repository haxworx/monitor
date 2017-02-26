#include "monitor.h"
#include "scripts.h"

void usage(void)
{
	printf("%s username@hostname:path/to/directory\n", PROGRAM_NAME);
	exit(EXIT_FAILURE);
}

void print_info(char *directory)
{
	printf("(c) Copyright 2016-17. Al Poole <netstar@gmail.com>.\n");
	printf("See: http://haxlab.org\n");
	printf("Monitoring: %s\n\n", directory);
}

void set_arguments(monitor_t *mon, char *cmd_string)
{
	char buf[PATH_MAX];
	char *user_start = cmd_string;
	char *user_end = strchr(user_start, '@');
	if (!user_end) usage();
	*user_end = '\0';
	mon->username = strdup(user_start);

	char *host_start = user_end + 1;
	if (!host_start) usage();
	char *host_end = strchr(host_start, ':');
	if (!host_end) usage();
	*host_end = '\0';
	mon->hostname = strdup(host_start);
	char *directory = host_end + 1;
	realpath(directory, buf);
	mon->watch_add(mon->self, buf);
}
    
int main(int argc, char **argv)
{
	time_t interval = 3;
	bool recursive = true;
	
	if (argc < 2) usage();

	char *cmd_string = argv[1];

	monitor_t *mon = monitor_new();
	if (!mon) 
		mon->error("monitor_new()");

	mon->init(mon->self, recursive);

	set_arguments(mon, cmd_string);

	print_info(mon->directories[0]);

	mon->authenticate(mon->self);
	/* Get here we've authenticated */

	/* CTRL+C or SIGTERM to exit gracefully */
	mon->mainloop(mon->self, interval);

	free(mon);
	
	printf("done!\n");

	return EXIT_SUCCESS;
}

