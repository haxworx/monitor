#include "monitor.h"
#include "scripts.h"

#define PROGRAM_NAME "monitor"

void usage(void)
{
	printf("%s username@hostname:directory\n", PROGRAM_NAME);
	exit(EXIT_FAILURE);
}

void print_info(char *directory)
{
	printf("(c) Copyright 2016. Al Poole <netstar@gmail.com>.\n");
	printf("See: http://haxlab.org\n\n");
	printf("Monitoring: %s\n\n", directory);
}

void set_arguments(monitor_t *mon, char *cmd_string)
{
	char *user_start = cmd_string;
	char *user_end = strchr(user_start, '@');
	if (!user_end) usage();
	*user_end = '\0';
	mon->username = strdup(user_start);

	char *host_start = user_end + 1;
	char *host_end = strchr(host_start, ':');
	if (!host_end) usage();
	*host_end = '\0';
	mon->hostname = strdup(host_start);

	char *directory = host_end + 1;
	mon->watch_add(mon, directory);
}
    
int main(int argc, char **argv)
{
	time_t interval = 3;
	bool recursive = true;
	
	if (argc < 2) usage();

	char *cmd_string = argv[1];

	monitor_t *m = monitor_new();
	if (!m) 
		m->error("monitor_new()");

	m->init(m->self, recursive);
	set_arguments(m, cmd_string);

	//m->watch_add(m->self, m->directory);
	
	/* Optional scripts to execute on changes 
	 *
	   m->callback_set(MONITOR_ADD, do_add);
           m->callback_set(MONITOR_DEL, do_del);
	   m->callback_set(MONITOR_MOD, do_mod);
	*/

	print_info(m->directories[0]);

	m->authenticate(m->self);
	/* Get here we've authenticated */

	/* CTRL+C or SIGTERM to exit gracefully */
	m->mainloop(m->self, interval);

	free(m);
		
	return EXIT_SUCCESS;
}

