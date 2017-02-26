#include "system.h"

int
system_cpu_count(void)
{
	int generic = 4;
#if defined(__WIN32__)
	const char *cpu_count = getenv("NUMBER_OF_PROCESSORS");
	if (!cpu_count) return genetic;
	else return atoi(cpu_count);
#else
	return generic;
#endif
}
