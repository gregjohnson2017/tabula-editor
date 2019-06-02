package main

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type config struct {
	screenWidth     int32
	screenHeight    int32
	bottomBarHeight int32
	fontName        string
	fontSize        int32
	font            *ttf.Font
	framerate       uint32
}

func initConfig() *config {
	c := config{
		screenWidth:     640,
		screenHeight:    480,
		bottomBarHeight: 30,
		fontName:        "NotoMono-Regular.ttf",
		fontSize:        24,
		framerate:       144,
	}
	return &c
}

func initLibraries(conf *config) error {
	if sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl") != true {
		return fmt.Errorf("failed to set opengl render driver hint")
	}
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
		return err
	}
	if img.Init(img.INIT_PNG) != img.INIT_PNG {
		return fmt.Errorf("could not initialize PNG")
	}
	if err = ttf.Init(); err != nil {
		return err
	}
	if conf.font, err = ttf.OpenFont(conf.fontName, int(conf.fontSize)); err != nil {
		return err
	}
	return err
}

func quit(conf *config) {
	sdl.Quit()
	img.Quit()
	conf.font.Close()
	ttf.Quit()
}

func inBounds(area *sdl.Rect, x int32, y int32) bool {
	if x < area.X || x >= area.X+area.W || y < area.Y || y >= area.Y+area.H {
		return false
	}
	return true
}

func initialize(title string) (*context, error) {
	conf := initConfig()
	var err error
	if err = initLibraries(conf); err != nil {
		return nil, err
	}
	var win *sdl.Window
	if win, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, conf.screenWidth, conf.screenHeight, 0); err != nil {
		return nil, err
	}
	var rend *sdl.Renderer
	if rend, err = sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED); err != nil {
		return nil, err
	}
	if err = rend.SetDrawColor(0xFF, 0xFF, 0xFF, 0xFF); err != nil {
		return nil, err
	}
	var info sdl.RendererInfo
	if info, err = rend.GetInfo(); err != nil {
		return nil, err
	}
	if info.MaxTextureWidth == 0 {
		info.MaxTextureWidth = math.MaxInt32
	}
	if info.MaxTextureHeight == 0 {
		info.MaxTextureHeight = math.MaxInt32
	}
	return &context{
		Win:      win,
		Rend:     rend,
		RendInfo: &info,
		Conf:     conf,
	}, err
}

func main() {
	var err error
	var ctx *context
	if ctx, err = initialize("Tabula Editor"); err != nil {
		panic(err)
	}

	var framerate = &gfx.FPSmanager{}
	gfx.InitFramerate(framerate)
	if gfx.SetFramerate(framerate, ctx.Conf.framerate) != true {
		panic(fmt.Errorf("could not set framerate: %v", sdl.GetError()))
	}

	imageViewArea := &sdl.Rect{
		X: 0,
		Y: 0,
		W: ctx.Conf.screenWidth,
		H: ctx.Conf.screenHeight - ctx.Conf.bottomBarHeight,
	}
	bottomBarArea := &sdl.Rect{
		X: 0,
		Y: ctx.Conf.screenHeight - ctx.Conf.bottomBarHeight,
		W: ctx.Conf.screenWidth,
		H: ctx.Conf.bottomBarHeight,
	}
	mouseComms := make(chan coord)

	iv, err := NewImageView(imageViewArea, "monkaDetect.png", mouseComms, ctx)
	if err != nil {
		panic(err)
	}
	bb, err := NewBottomBar(bottomBarArea, mouseComms, ctx, &sdl.Color{0x80, 0x80, 0x80, 0xFF})
	if err != nil {
		panic(err)
	}
	comps := []UIComponent{iv, bb}

	var lastHover UIComponent
	var currHover UIComponent
	running := true
	for running {
		var e sdl.Event
		for e = sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch evt := e.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				for i := range comps {
					comp := comps[len(comps)-i-1]
					if inBounds(comp.GetBoundary(), evt.X, evt.Y) {
						comp.OnClick(evt)
						break
					}
				}
			case *sdl.MouseMotionEvent:
				for i := range comps {
					comp := comps[len(comps)-i-1]
					if inBounds(comp.GetBoundary(), evt.X, evt.Y) {
						if lastHover != comp && currHover != comp {
							comp.OnEnter(evt)
							currHover = comp
						} else if lastHover == comp {
							currHover = lastHover
						}
						if comp.OnMotion(evt) {
							break
						}
					}
				}
				if lastHover != nil && lastHover != currHover {
					lastHover.OnLeave(evt)
					lastHover = nil
				}
			case *sdl.MouseWheelEvent:
				for i := range comps {
					comp := comps[len(comps)-i-1]
					x, y, _ := sdl.GetMouseState()
					if inBounds(comp.GetBoundary(), x, y) {
						if comp.OnScroll(evt) {
							break
						}
					}
				}
			}
		}
		if currHover != nil {
			lastHover = currHover
			currHover = nil
		}

		if err = ctx.Rend.Clear(); err != nil {
			panic(err)
		}
		for _, comp := range comps {
			if err = comp.Render(ctx.Rend); err != nil {
				panic(err)
			}
		}

		// wait remainder of frame-time before presenting
		gfx.FramerateDelay(framerate)
		ctx.Rend.Present()
	}

	// free UIComponent SDL assets
	for _, comp := range comps {
		comp.Destroy()
	}

	quit(ctx.Conf)
}
