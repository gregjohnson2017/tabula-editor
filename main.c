#include "SDL2/SDL.h"
#include "SDL2/SDL_image.h"

int init() {
	if (SDL_Init(SDL_INIT_VIDEO) < 0) {
		printf("SDL init error: %s\n", SDL_GetError());
		return -1;
	}
	return 0;
}

SDL_Window* create_window(char *name, int width, int height) {
	SDL_Window *window = SDL_CreateWindow(name, SDL_WINDOWPOS_UNDEFINED, SDL_WINDOWPOS_UNDEFINED, width, height, SDL_WINDOW_SHOWN);
	if (window == NULL) {
		printf("SDL create window error: %s\n", SDL_GetError());
		return NULL;
	}
	return window;
}

int main(int argc, char **argv) {
	init();
	SDL_Window *window = create_window("test", 640, 480);
	SDL_Surface *surface = SDL_GetWindowSurface(window);

	SDL_FillRect(surface, NULL, SDL_MapRGB(surface->format, 0x00, 0xFF, 0xFF));
	SDL_UpdateWindowSurface(window);
	SDL_Delay(2000);

	// clean up
	SDL_DestroyWindow(window);
	SDL_Quit();
	return 0;
}
