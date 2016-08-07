#include "monitor.h"
#include "scripts.h"

#define PROGRAM_NAME "monitor"

void usage(void)
{
	printf("%s <path>\n", PROGRAM_NAME);
	exit(EXIT_FAILURE);
}

void print_info(char *directory)
{
	printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n");
	printf("See: http://haxlab.org\n\n");
	printf("Monitoring: %s\n\n", directory);
}

int main(int argc, char **argv)
{
	time_t interval = 3;
	bool recursive = true;
	char *directory = NULL;
	
	if (argc < 2) usage();

	directory = argv[1];

	monitor_t *m = monitor_new();
	if (!m) 
		m->error("monitor_new()");

	m->init(recursive);

	m->watch_add(directory);
	
	m->callback_set(MONITOR_ADD, do_add);
	m->callback_set(MONITOR_DEL, do_del);
	m->callback_set(MONITOR_MOD, do_mod);

	print_info(directory);
	/* CTRL+C or SIGTERM to exit gracefully */
	m->mainloop(interval);

	free(m);
		
	return EXIT_SUCCESS;
}

