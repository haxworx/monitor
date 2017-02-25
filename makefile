PROGRAM=dropsy
CFLAGS=-g -ggdb3 -O0 -Wall -pedantic -std=c99
SRC_DIR=src
SCRIPTS_DIR=scripts
FILES= main.c monitor.c
OBJECTS=monitor.o network.o scripts.o main.o
VERSION=0.0.0.0.1

default: scripts monitor

monitor: monitor.o network.o scripts.o main.o
	$(CC) $(CFLAGS) *.o -o $(PROGRAM)

monitor.o: $(SRC_DIR)/monitor.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/monitor.c -o $@
network.o: $(SRC_DIR)/network.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/network.c -o $@
scripts.o: $(SRC_DIR)/scripts.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/scripts.c -o $@

main.o: $(SRC_DIR)/main.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/main.c -o $@

scripts: 
	chmod +x $(SCRIPTS)/*.sh 

clean:
	-rm $(PROGRAM) 
	-rm $(OBJECTS)
