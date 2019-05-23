#include "SDL2/SDL.h"
#include "SDL2/SDL_image.h"
#include "SDL2/SDL2_framerate.h"
#include "SDL2/SDL_ttf.h"
#include <string.h>

static const int SCREEN_WIDTH = 640;
static const int SCREEN_HEIGHT = 480;
static const int BOTTOM_BAR_HEIGHT = 30;
static const char *font_name = "NotoMono-Regular.ttf";
static const int font_size = 24;
static TTF_Font *font = NULL;

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
	// initialize TTF
	if (TTF_Init() == -1) {
		printf("TTF init error: %s\n", TTF_GetError());
		return -1;
	}
	font = TTF_OpenFont(font_name, font_size);
	if (font == NULL) {
		printf("TTF_OpenFont error opening \"%s\" with size %d\n", font_name, font_size);
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

void render_text(SDL_Renderer *renderer, char *text, int relx, int rely, int right) {
	SDL_Color color = {255, 255, 255, 0};
	SDL_Surface *message_surface = TTF_RenderText_Blended(font, text, color);
	if (message_surface == NULL) {
		printf("TTF_RenderText_Solid error: %s\n", TTF_GetError());
		return;
	}
	SDL_Texture *message_texture = SDL_CreateTextureFromSurface(renderer, message_surface);
	if (message_texture == NULL) {
		printf("SDL_CreateTextureFromSurface error: %s\n", SDL_GetError());
		SDL_FreeSurface(message_surface);
		return;
	}
	int w, h;
	if (TTF_SizeText(font, text, &w, &h) == -1) {
		printf("TTF_SizeText error: %s\n", TTF_GetError());
		SDL_FreeSurface(message_surface);
		SDL_DestroyTexture(message_texture);
		return;
	}
	SDL_Rect rect;
	rect.x = relx;
	if (right) rect.x -= w;
	rect.y = rely;
	rect.w = w;
	rect.h = h;
	SDL_RenderCopy(renderer, message_texture, NULL, &rect);
	SDL_FreeSurface(message_surface);
	SDL_DestroyTexture(message_texture);
}

char *getUInt32String(char *before, Uint32 n, char *after) {
	char intstr[12];
	snprintf(intstr, 11, "%d", n);
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
	if (init() == -1) {
		printf("Initialization failed... exiting\n");
		return -1;
	}
	SDL_Window *window = create_window("test", SCREEN_WIDTH, SCREEN_HEIGHT);
	SDL_Renderer *renderer = create_renderer(window);
	SDL_Texture *texture = load_texture(renderer, "monkaW.png");
	FPSmanager framerate = {0};
	SDL_initFramerate(&framerate);
	if (SDL_setFramerate(&framerate, 144) < 0) {
		printf("SDL_setFramerate error: %s\n", SDL_GetError());
		return -1;
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
	int rmouse_down = 0;
	struct {
		int x;
		int y;
	} rmouse_point;
	rmouse_point.x = 0;
	rmouse_point.y = 0;
	while (running) {
		while (SDL_PollEvent(&e) != 0) {
			switch (e.type) {
			case SDL_QUIT:
				running = 0;
				break;
			case SDL_MOUSEBUTTONDOWN:
				if (e.button.button == SDL_BUTTON_RIGHT && e.button.y < bottom_bar.y) {
					rmouse_down = 1;
					rmouse_point.x = e.button.x;
					rmouse_point.y = e.button.y;
				}
				break;
			case SDL_MOUSEBUTTONUP:
				if (e.button.button == SDL_BUTTON_RIGHT) {
					rmouse_down = 0;
				}
				break;
			case SDL_MOUSEMOTION:
				if (e.motion.state == SDL_BUTTON_RMASK && rmouse_down) {
					canvas.x += e.motion.x - rmouse_point.x;
					rmouse_point.x = e.motion.x;
					canvas.y += e.motion.y - rmouse_point.y;
					rmouse_point.y = e.motion.y;
				}
				break;
			}
		}
		SDL_RenderClear(renderer);
		SDL_RenderSetViewport(renderer, &canvas);
		SDL_RenderCopy(renderer, texture, NULL, NULL);
		SDL_RenderSetViewport(renderer, &bottom_bar);
		SDL_RenderCopy(renderer, bottom_bar_texture, NULL, NULL);
		SDL_framerateDelay(&framerate);
		time = SDL_GetTicks();
		int fps = (int) 1.0/(float)((time - lastTime) / 1000.0);
		char *fps_str = getUInt32String("FPS: ", fps, "");
		// SDL_SetWindowTitle(window, str);
		char *temp = getUInt32String("(", rmouse_point.x, ", ");
		char *coord_str = getUInt32String(temp, rmouse_point.y, ")");
		render_text(renderer, coord_str, SCREEN_WIDTH, 0, 1);
		render_text(renderer, fps_str, 0, 0, 0);
		free(temp);
		free(fps_str);
		free(coord_str);
		lastTime = time;
		SDL_RenderPresent(renderer);

	}

	// clean up
	printf("exiting\n");
	SDL_DestroyTexture(texture);
	SDL_DestroyRenderer(renderer);
	SDL_DestroyWindow(window);
	IMG_Quit();
	TTF_CloseFont(font);
	TTF_Quit();
	SDL_Quit();
	return 0;
}
