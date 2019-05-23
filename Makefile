LDLIBS = -lSDL2 -lSDL2_image
main: main.o
	gcc -o test main.c $(LDLIBS)
