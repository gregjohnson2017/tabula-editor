package main

import (
	"fmt"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func inBounds(area *sdl.Rect, x int32, y int32) bool {
	if x < area.X || x >= area.X+area.W || y < area.Y || y >= area.Y+area.H {
		return false
	}
	return true
}

func initWindow(title string, width, height int32) (*sdl.Window, error) {
	if sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl") != true {
		return nil, fmt.Errorf("failed to set opengl render driver hint")
	}
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return nil, err
	}
	// other libraries
	if img.Init(img.INIT_PNG) != img.INIT_PNG {
		return nil, fmt.Errorf("could not initialize PNG")
	}
	if err = ttf.Init(); err != nil {
		return nil, err
	}
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 6)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)

	var window *sdl.Window
	if window, err = sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_HIDDEN|sdl.WINDOW_OPENGL); err != nil {
		return nil, err
	}
	window.SetResizable(true)
	// creates context AND makes current
	if _, err = window.GLCreateContext(); err != nil {
		return nil, err
	}
	if err = sdl.GLSetSwapInterval(1); err != nil {
		return nil, err
	}

	// INIT OPENGL
	if err = gl.Init(); err != nil {
		return nil, err
	}
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	// enable anti-aliasing
	// gl.Enable(gl.LINE_SMOOTH)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)

	// version := gl.GoStr(gl.GetString(gl.VERSION))
	// log.Println("OpenGL version", version)
	return window, nil
}

func main() {
	var screenWidth, screenHeight int32 = 960, 720
	var err error
	var win *sdl.Window
	if win, err = initWindow("Tabula Editor", screenWidth, screenHeight); err != nil {
		panic(err)
	}

	var fileName string
	if len(os.Args) == 2 {
		fileName = os.Args[1]
	} else {
		if fileName, err = openFileDialog(win); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}

	win.Show()

	// var framerate = &gfx.FPSmanager{}
	// gfx.InitFramerate(framerate)
	// if gfx.SetFramerate(framerate, ctx.Conf.framerate) != true {
	// 	panic(fmt.Errorf("could not set framerate: %v", sdl.GetError()))
	// }

	imageViewArea := &sdl.Rect{
		X: 0,
		Y: 0,
		W: screenWidth,
		H: screenHeight, // - ctx.Conf.bottomBarHeight,
	}
	// bottomBarArea := &sdl.Rect{
	// 	X: 0,
	// 	Y: ctx.Conf.screenHeight - ctx.Conf.bottomBarHeight,
	// 	W: ctx.Conf.screenWidth,
	// 	H: ctx.Conf.bottomBarHeight,
	// }
	// buttonAreaOpen := &sdl.Rect{
	// 	X: 0,
	// 	Y: 0,
	// 	W: 125,
	// 	H: 20,
	// }
	// buttonAreaCenter := &sdl.Rect{
	// 	X: 125,
	// 	Y: 0,
	// 	W: 125,
	// 	H: 20,
	// }
	comms := make(chan imageComm)
	// fileComm := make(chan func())

	iv, err := NewImageView(imageViewArea, fileName, comms)
	if err != nil {
		panic(err)
	}
	// bottomBar, err := NewBottomBar(bottomBarArea, comms, ctx, "NotoMono-Regular.ttf", 24)
	// if err != nil {
	// 	panic(err)
	// }
	// openButton, err := NewButton(buttonAreaOpen, ctx, "Open File", "NotoMono-Regular.ttf", 14, func() {
	// 	newFileName, err := openFileDialog(ctx.Win)
	// 	if err != nil {
	// 		fmt.Printf("No file chosen\n")
	// 		return
	// 	}
	// 	go func() {
	// 		fileComm <- func() {
	// 			iv.Destroy()
	// 			if err = iv.loadFromFile(newFileName); err != nil {
	// 				panic(err)
	// 			}
	// 			iv.mult = 1.0
	// 			iv.sel = set.NewSet()
	// 		}
	// 	}()
	// })
	// centerButton, err := NewButton(buttonAreaCenter, ctx, "Center Image", "NotoMono-Regular.ttf", 14, func() {
	// 	go func() {
	// 		fileComm <- func() {
	// 			iv.centerImage()
	// 		}
	// 	}()
	// })
	// centerButton.SetHighlightBackgroundColor(&sdl.Color{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF})
	// comps := []UIComponent{iv, bottomBar, openButton, centerButton}
	comps := []UIComponent{iv} // TODO: empty for OpenGL testing
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
					if inBounds(comp.GetBoundary(), evt.X, screenHeight-evt.Y) {
						comp.OnClick(evt)
						break
					}
				}
			case *sdl.MouseMotionEvent:
				// search top down through components until exhausted or one absorbs the event
				for i := range comps {
					comp := comps[len(comps)-i-1]
					if inBounds(comp.GetBoundary(), evt.X, screenHeight-evt.Y) {
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
					if inBounds(comp.GetBoundary(), x, screenHeight-y) {
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

		// handle all events in pipe
		// hasEvents := true
		// for hasEvents {
		// 	select {
		// 	case closure := <-fileComm:
		// 		closure()
		// 	default:
		// 		// no more in pipe
		// 		hasEvents = false
		// 	}
		// }
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		for _, comp := range comps {
			if err = comp.Render(); err != nil {
				panic(err)
			}
		}

		win.GLSwap()
		// TODO wait remainder of frame-time
		// gfx.FramerateDelay(framerate)
	}

	// free UIComponent SDL assets
	for _, comp := range comps {
		comp.Destroy()
	}

	gl.UseProgram(0)
	win.Destroy()
	sdl.Quit()
	img.Quit()
	ttf.Quit()
}
