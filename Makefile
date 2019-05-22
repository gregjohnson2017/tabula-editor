LDLIBS = -lSDL2
main: main.o
	gcc -o test main.c $(LDLIBS)
