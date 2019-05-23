#include "SDL2/SDL.h"
#include "SDL2/SDL_image.h"

int init() {
	// initialize SDL
	if (SDL_Init(SDL_INIT_VIDEO) < 0) {
		printf("SDL init error: %s\n", SDL_GetError());
		return -1;
	}
	// initialize PNG loading
	int flags = IMG_INIT_PNG;
	if ((IMG_Init(flags) & flags) == 0) {
		printf("SDL_image init error: %s\n", IMG_GetError());
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

SDL_Surface* load_surface(char *path) {
	SDL_Surface *loaded_surface = IMG_Load(path);
	if (loaded_surface == NULL) {
		printf("IMG_load error loading \"%s\": %s\n", path, IMG_GetError());
		return NULL;
	}
	return loaded_surface;
}

SDL_Surface* optimize_surface(SDL_Surface *surface, SDL_Surface *screen) {
	SDL_Surface *optimized_surface = SDL_ConvertSurface(surface, screen->format, 0);
	if (optimized_surface == NULL) {
		printf("optimize_surface error: %s\n", SDL_GetError());
		return NULL;
	}
	return optimized_surface;
}

int main(int argc, char **argv) {
	init();
	SDL_Window *window = create_window("test", 640, 480);
	SDL_Surface *screen = SDL_GetWindowSurface(window);
	SDL_Surface *image = optimize_surface(load_surface("monkaW.png"), screen);

	SDL_FillRect(screen, NULL, SDL_MapRGB(screen->format, 0x00, 0xFF, 0xFF));
	SDL_BlitSurface(image, NULL, screen, NULL);

	int running = 1;
	SDL_Event e;
	while (running) {
		while (SDL_PollEvent(&e) != 0) {
			if (e.type == SDL_QUIT) running = 0;
		}
		SDL_UpdateWindowSurface(window);
		SDL_Delay(200);
	}

	// clean up
	printf("exiting\n");
	SDL_FreeSurface(screen);
	SDL_FreeSurface(image);
	SDL_DestroyWindow(window);
	SDL_Quit();
	return 0;
}
