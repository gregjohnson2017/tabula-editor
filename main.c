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

SDL_Renderer* create_renderer(SDL_Window *window) {
	SDL_Renderer *renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_ACCELERATED);
	if (renderer == NULL) {
		printf("SDL_CreateRenderer error: %s\n", SDL_GetError());
		return NULL;
	}
	// default pixel?
	SDL_SetRenderDrawColor(renderer, 0xFF, 0xFF, 0xFF, 0xFF);
	return renderer;
}

SDL_Texture* load_texture(SDL_Renderer *renderer, char *path) {
	SDL_Surface *loaded_surface = IMG_Load(path);
	if (loaded_surface == NULL) {
		printf("IMG_load error loading \"%s\": %s\n", path, IMG_GetError());
		return NULL;
	}
	SDL_Texture *texture = SDL_CreateTextureFromSurface(renderer, loaded_surface);
	if (texture == NULL) {
		printf("SDL_CreateTextureFromSurface error: %s\n", SDL_GetError());
		return NULL;
	}
	return texture;
}

int main(int argc, char **argv) {
	init();
	SDL_Window *window = create_window("test", 640, 480);
	SDL_Renderer *renderer = create_renderer(window);
	SDL_Texture *texture = load_texture(renderer, "monkaW.png");

	int running = 1;
	SDL_Event e;
	while (running) {
		while (SDL_PollEvent(&e) != 0) {
			if (e.type == SDL_QUIT) running = 0;
		}
		SDL_RenderClear(renderer);
		SDL_RenderCopy(renderer, texture, NULL, NULL);
		SDL_RenderPresent(renderer);
	}

	// clean up
	printf("exiting\n");
	SDL_DestroyTexture(texture);
	SDL_DestroyRenderer(renderer);
	SDL_DestroyWindow(window);
	IMG_Quit();
	SDL_Quit();
	return 0;
}
