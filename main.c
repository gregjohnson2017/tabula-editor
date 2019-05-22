#include "SDL2/SDL.h"
#include "SDL2/SDL_image.h"
int main(int argc, char **argv) {
	SDL_Window *window = NULL;
	SDL_Surface *surface = NULL;
	if (SDL_Init(SDL_INIT_VIDEO) < 0) {
		printf("SDL init error: %s\n", SDL_GetError());
		goto exit;
	}

	window = SDL_CreateWindow("SDL Test", SDL_WINDOWPOS_UNDEFINED, SDL_WINDOWPOS_UNDEFINED, 640, 480, SDL_WINDOW_SHOWN);
	if (window == NULL) {
		printf("SDL create window error: %s\n", SDL_GetError());
		goto exit;
	}
	surface = SDL_GetWindowSurface(window);
	SDL_FillRect(surface, NULL, SDL_MapRGB(surface->format, 0x00, 0xFF, 0xFF));
	SDL_UpdateWindowSurface(window);
	SDL_Delay(2000);

exit:
	SDL_DestroyWindow(window);
	SDL_Quit();
	return 0;
}
