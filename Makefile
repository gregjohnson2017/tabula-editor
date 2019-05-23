LDLIBS = -lSDL2 -lSDL2_image -lSDL2_gfx -lSDL2_ttf
main: main.o
	gcc -o test main.c $(LDLIBS)
debug: main.o
	gcc -g -o test main.c $(LDLIBS)
clean:
	rm main.o test