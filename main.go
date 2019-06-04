package main

import (
	"fmt"
	"math"
	"os"

	"github.com/jcmuller/gozenity"
	set "github.com/kroppt/IntSet"
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
	framerate       uint32
}

func initConfig() *config {
	c := config{
		screenWidth:     960,
		screenHeight:    720,
		bottomBarHeight: 30,
		fontName:        "NotoMono-Regular.ttf",
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
	return err
}

func quit(conf *config) {
	sdl.Quit()
	img.Quit()
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
	win.SetResizable(true)
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
	var fileName string
	if len(os.Args) == 2 {
		fileName = os.Args[1]
	} else {
		files, err := gozenity.FileSelection("Choose a picture to open", nil)
		if err != nil {
			panic(err)
		}
		fileName = files[0]
	}

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
	buttonAreaOpen := &sdl.Rect{
		X: 0,
		Y: 0,
		W: 125,
		H: 20,
	}
	buttonAreaCenter := &sdl.Rect{
		X: 125,
		Y: 0,
		W: 125,
		H: 20,
	}
	comms := make(chan imageComm)
	fileComm := make(chan func())

	iv, err := NewImageView(imageViewArea, fileName, comms, ctx)
	if err != nil {
		panic(err)
	}
	bottomBar, err := NewBottomBar(bottomBarArea, comms, ctx, nil)
	if err != nil {
		panic(err)
	}
	openButton, err := NewButton(buttonAreaOpen, ctx, nil, nil, "Open File", func() {
		files, err := gozenity.FileSelection("Choose a picture to open", nil)
		if err != nil {
			fmt.Printf("No file chosen\n")
			return
		}
		newFileName := files[0]
		go func() {
			fileComm <- func() {
				iv.Destroy()
				if err = iv.loadFromFile(newFileName); err != nil {
					panic(err)
				}
				iv.mult = 1.0
				iv.sel = set.NewSet()
			}
		}()
	})
	centerButton, err := NewButton(buttonAreaCenter, ctx, nil, nil, "Center Image", func() {
		go func() {
			fileComm <- func() {
				iv.centerImage()
			}
		}()
	})
	comps := []UIComponent{iv, bottomBar, openButton, centerButton}

	var lastHover UIComponent
	var currHover UIComponent
	var moved bool
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
				// search top down through components until exhausted or one absorbs the event
				for i := range comps {
					comp := comps[len(comps)-i-1]
					if inBounds(comp.GetBoundary(), evt.X, evt.Y) {
						if currHover != comp {
							// entered a new component
							comp.OnEnter()
							lastHover = currHover
							currHover = comp
							moved = true
						}
						if comp.OnMotion(evt) {
							break
						}
					}
				}
				if lastHover != nil && moved {
					lastHover.OnLeave()
					moved = false
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
			case *sdl.WindowEvent:
				if evt.Event == sdl.WINDOWEVENT_LEAVE || evt.Event == sdl.WINDOWEVENT_FOCUS_LOST || evt.Event == sdl.WINDOWEVENT_MINIMIZED {
					if currHover != nil {
						currHover.OnLeave()
						lastHover = currHover
						currHover = nil
						moved = false
					}
				} else if evt.Event == sdl.WINDOWEVENT_RESIZED {
					for _, comp := range comps {
						comp.OnResize(evt.Data1, evt.Data2)
					}
				}
			}
		}

		// TODO: handle all events in pipe
		hasEvents := true
		for hasEvents {
			select {
			case closure := <-fileComm:
				closure()
			default:
				// no more in pipe
				hasEvents = false
			}
		}

		if err = ctx.Rend.Clear(); err != nil {
			panic(err)
		}
		for _, comp := range comps {
			if err = comp.Render(); err != nil {
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
