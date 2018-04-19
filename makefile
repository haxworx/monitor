default:
	$(CC) -g -ggdb3 -O0 -I/usr/local/include -L/usr/local/lib -lsea monitor.c -o monitor
