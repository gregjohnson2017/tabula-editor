package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/app"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/veandco/go-sdl2/sdl"
)

func initWindow(title string, width, height int32) (*sdl.Window, error) {
	if !sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl") {
		return nil, fmt.Errorf("failed to set opengl render driver hint")
	}
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return nil, err
	}
	if err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3); err != nil {
		return nil, err
	}
	if err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3); err != nil {
		return nil, err
	}
	if err = sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1); err != nil {
		return nil, err
	}
	if err = sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE); err != nil {
		return nil, err
	}
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
	var fps int
	var width int
	var height int
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
		fmt.Fprintf(flag.CommandLine.Output(), "  %v [options] [filename]\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "  Opens filename in the editor as an image if provided.")
		fmt.Fprintln(flag.CommandLine.Output(), "  Otherwise, an open file dialog will be used, if supported.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions:")
		flag.PrintDefaults()
	}
	flag.IntVar(&fps, "fps", 144, "the frames per second to render at")
	flag.IntVar(&width, "width", 960, "the initial width of the window")
	flag.IntVar(&height, "height", 720, "the initial height of the window")
	flag.Parse()
	fileName := flag.Arg(0)

	cfg := config.New(int32(width), int32(height), 30, fps)
	var err error
	win, err := initWindow("Tabula Editor", cfg.ScreenWidth, cfg.ScreenHeight)
	errCheck(err)

	app := app.New(fileName, win, cfg)
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
