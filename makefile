PROGRAM=monitor
CFLAGS=-g -ggdb3 -Wall -pedantic -std=c99
SRC_DIR=src
SCRIPTS_DIR=scripts
FILES= main.c monitor.c
OBJECTS=objects

default: scripts monitor

monitor: monitor.o scripts.o main.o
	$(CC) $(CFLAGS) *.o -o $(PROGRAM)

monitor.o: $(SRC_DIR)/monitor.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/monitor.c -o $@

scripts.o: $(SRC_DIR)/scripts.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/scripts.c -o $@

main.o: $(SRC_DIR)/main.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/main.c -o $@

scripts: 
	chmod +x $(SCRIPTS)/*.sh 

clean:
	-rm $(PROGRAM) 
	-rm *.o
