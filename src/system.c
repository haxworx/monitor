#include "system.h"

#if defined(__FreeBSD__) || defined(__DragonFly__)
# include <sys/types.h>
# include <sys/sysctl.h>
#endif

#if defined(__OpenBSD__)
# include <sys/param.h>
# include <sys/sysctl.h>
#endif

int
system_cpu_count(void)
{
	int generic = 4;
#if defined(__WIN32__)
	const char *cpu_count = getenv("NUMBER_OF_PROCESSORS");
	if (!cpu_count) return genetic;
	else return atoi(cpu_count);
#elif defined(__FreeBSD__) || defined(__DragonFly_) || defined(__OpenBSD__)
	int cores = 0;
	size_t len;
	int mib[2] = { CTL_HW, HW_NCPU };

	len = sizeof(cores);
	if (sysctl(mib, 2, &cores, &len, NULL, 0) == -1) return generic;
	return cores;
#else
	return generic;
#endif
}
