package main

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/app"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/veandco/go-sdl2/sdl"
)

func initWindow(title string, width, height int32) (*sdl.Window, error) {
	if sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl") != true {
		return nil, fmt.Errorf("failed to set opengl render driver hint")
	}
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return nil, err
	}
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	//sdl.EventState(sdl.SYSWMEVENT, sdl.ENABLE)

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
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
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
	cfg := config.New(960, 720, 30)
	var err error
	win, err := initWindow("Tabula Editor", cfg.ScreenWidth, cfg.ScreenHeight)
	errCheck(err)

	app := app.New(win, cfg)
	app.Start()

	for app.Running() {
		for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
			app.HandleSdlEvent(evt)
		}

		app.PostEventActions()
	}

	app.Quit()
}

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}
