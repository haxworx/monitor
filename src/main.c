#include "monitor.h"

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

int main(int argc, char **argv)
{
	time_t interval = 3;
	
	if (argc < 2) usage();

	char *cmd_string = argv[1];

	monitor_t *mon = monitor_new();
	if (!mon) 
		mon->error("monitor_new()");

	if (!mon->init(mon->self, cmd_string)) {
		usage();	
	}

	print_info(mon->directories[0]);

	mon->authenticate(mon->self);
	/* Get here we've authenticated */

	/* CTRL+C or SIGTERM to exit gracefully */
	mon->mainloop(mon->self, interval);

	free(mon);
	
	printf("done!\n");

	return EXIT_SUCCESS;
}

