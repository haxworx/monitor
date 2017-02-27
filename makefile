PROGRAM=dropsy
CFLAGS=-g -ggdb3 -O0 -Wall -pedantic -std=c99 
PKGS=openssl 
FLAGS=$(shell pkg-config --libs --cflags $(PKGS))
INCLUDES=$(shell pkg-config --cflags $(PKGS))
SRC_DIR=src
SCRIPTS_DIR=scripts
OBJECTS=monitor.o network.o scripts.o system.o main.o
VERSION=0.0.0.0.1

default: scripts monitor

monitor: monitor.o network.o scripts.o system.o main.o
	$(CC) $(CFLAGS) *.o $(FLAGS) -o $(PROGRAM)

monitor.o: $(SRC_DIR)/monitor.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/monitor.c -o $@
network.o: $(SRC_DIR)/network.c
	$(CC) -c $(CFLAGS) $(INCLUDES) $(SRC_DIR)/network.c -o $@
system.o:  $(SRC_DIR)/system.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/system.c -o $@
scripts.o: $(SRC_DIR)/scripts.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/scripts.c -o $@

main.o: $(SRC_DIR)/main.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/main.c -o $@

scripts: 
	chmod +x $(SCRIPTS)/*.sh 

clean:
	-rm $(PROGRAM) 
	-rm $(OBJECTS)
