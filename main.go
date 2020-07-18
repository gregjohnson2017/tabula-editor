package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/app"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/perf"
	"github.com/veandco/go-sdl2/sdl"
)

// ErrRenderDriver indicates that SDL failed to enable the OpenGL render driver
const ErrRenderDriver log.ConstErr = "failed to set opengl render driver hint"

func initWindow(title string, width, height int32) (*sdl.Window, error) {
	if !sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl") {
		return nil, ErrRenderDriver
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

	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Debug("OpenGL version", version)
	return window, nil
}

func main() {
	var fps int
	var width int
	var height int
	var color bool
	var info bool
	var warn bool
	var debug bool
	var perform bool
	var quiet bool
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
	flag.BoolVar(&color, "color", true, "colorize the output logs")
	flag.BoolVar(&info, "info", true, "show info logging")
	flag.BoolVar(&warn, "warn", true, "show warning logging")
	flag.BoolVar(&debug, "debug", false, "show debug logging")
	flag.BoolVar(&perform, "perf", false, "show performormance logging")
	flag.BoolVar(&quiet, "quiet", false, "hide all output, overrides other logging options")
	flag.Parse()
	fileName := flag.Arg(0)

	var loggers []string
	// loggers are discarded by default
	if !quiet {
		if info {
			log.SetInfoOutput(os.Stderr)
			loggers = append(loggers, "info")
		}
		if warn {
			log.SetWarnOutput(os.Stderr)
			loggers = append(loggers, "warn")
		}
		if debug {
			log.SetDebugOutput(os.Stderr)
			loggers = append(loggers, "debug")
		}
		if perform {
			log.SetPerfOutput(os.Stderr)
			loggers = append(loggers, "perform")
			perf.SetMetricsEnabled(true)
			defer perf.LogMetrics()
		}
		log.SetFatalOutput(os.Stderr)
		loggers = append(loggers, "fatal")
	}
	log.SetColorized(color)

	log.Debugf("args: [ %v ]", strings.Join(flag.Args(), ", "))
	log.Debugf("fileName argument: \"%v\"", fileName)
	log.Debugf("enabled loggers: %v", strings.Join(loggers, ", "))
	log.Debugf("output colorized: %v", color)

	cfg := config.New(int32(width), int32(height), 30, fps)
	win, err := initWindow("Tabula Editor", cfg.ScreenWidth, cfg.ScreenHeight)
	if err != nil {
		log.Fatal(err)
	}

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
