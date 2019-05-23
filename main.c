#include "SDL2/SDL.h"
#include "SDL2/SDL_image.h"
#include "SDL2/SDL2_framerate.h"
#include "SDL2/SDL_ttf.h"

const int SCREEN_WIDTH = 640;
const int SCREEN_HEIGHT = 480;
const int BOTTOM_BAR_HEIGHT = 30;

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

char *getIntString(char *before, Uint32 n, char *after) {
	char intstr[11];
	snprintf(intstr, 10, "%d", n);
	int outsize = strlen(before) + strlen(intstr) + strlen(after) + 1;
	char *out = (char *) malloc(outsize);
	sprintf(out, "%s%s%s", before, intstr, after);
	return out;
}

SDL_Texture* create_solid_color_texture(SDL_Renderer *renderer, int width, int height, Uint32 r, Uint32 g, Uint32 b, Uint32 a) {
	SDL_Surface *surface = SDL_CreateRGBSurfaceWithFormat(0, width, height, 32, SDL_PIXELFORMAT_RGBA32);
	SDL_FillRect(surface, NULL, SDL_MapRGBA(surface->format, r, g, b, a));

	if (surface == NULL) {
		printf("SDL_CreateRGBSurface error: %s\n", SDL_GetError());
		return NULL;
	}
	SDL_Texture *texture = SDL_CreateTextureFromSurface(renderer, surface);
	if (texture == NULL) {
		printf("SDL_CreateTextureFromSurface error: %s\n", SDL_GetError());
		return NULL;
	}
	SDL_FreeSurface(surface);
	return texture;
}

int main(int argc, char **argv) {
	init();
	SDL_Window *window = create_window("test", SCREEN_WIDTH, SCREEN_HEIGHT);
	SDL_Renderer *renderer = create_renderer(window);
	SDL_Texture *texture = load_texture(renderer, "monkaW.png");
	FPSmanager framerate = {0};
	SDL_initFramerate(&framerate);
	if(SDL_setFramerate(&framerate, 144) < 0) {
		printf("SDL_setFramerate error: %s\n", SDL_GetError());
	}

	SDL_Rect bottom_bar;
	bottom_bar.x = 0;
	bottom_bar.y = SCREEN_HEIGHT - BOTTOM_BAR_HEIGHT;
	bottom_bar.w = SCREEN_WIDTH;
	bottom_bar.h = BOTTOM_BAR_HEIGHT;
	SDL_Texture *bottom_bar_texture = create_solid_color_texture(renderer, SCREEN_WIDTH, BOTTOM_BAR_HEIGHT, 0x80, 0x80, 0x80, 0xFF);

	SDL_Rect canvas;
	canvas.x = 0;
	canvas.y = 0;
	canvas.w = SCREEN_WIDTH;
	canvas.h = SCREEN_HEIGHT - BOTTOM_BAR_HEIGHT;

	int running = 1;
	SDL_Event e;
	Uint32 time;
	Uint32 lastTime = SDL_GetTicks();
	while (running) {
		while (SDL_PollEvent(&e) != 0) {
			switch (e.type) {
			case SDL_QUIT:
				running = 0;
				break;
			case SDL_MOUSEMOTION:
				if (e.motion.state == SDL_BUTTON_RMASK) {
					canvas.x += e.motion.xrel;
					canvas.y += e.motion.yrel;
				}
				break;
			}
		}
		SDL_RenderClear(renderer);
		SDL_RenderSetViewport(renderer, &canvas);
		SDL_RenderCopy(renderer, texture, NULL, NULL);
		SDL_RenderSetViewport(renderer, &bottom_bar);
		SDL_RenderCopy(renderer, bottom_bar_texture, NULL, NULL);
		SDL_RenderPresent(renderer);
		SDL_framerateDelay(&framerate);
		time = SDL_GetTicks();
		char *str = getIntString("frametime: ", time - lastTime, " ms");
		SDL_SetWindowTitle(window, str);
		free(str);
		lastTime = time;
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
