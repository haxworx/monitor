PROGRAM=monitor
CFLAGS=-g -ggdb3 -Wall -pedantic -std=c99
SRC_DIR=src
SCRIPTS_DIR=scripts
FILES= main.c monitor.c
OBJECTS=objects

default: directories scripts monitor

directories:
	-mkdir $(OBJECTS)

monitor: monitor.o main.o
	$(CC) $(CFLAGS) $(OBJECTS)/*.o -o $(PROGRAM)

monitor.o: $(SRC_DIR)/monitor.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/monitor.c -o $(OBJECTS)/$@

main.o: $(SRC_DIR)/main.c
	$(CC) -c $(CFLAGS) $(SRC_DIR)/main.c -o $(OBJECTS)/$@

scripts: 
	chmod +x $(SCRIPTS)/*.sh 

clean:
	-rm $(PROGRAM) 
	-rm $(OBJECTS)/*.o
