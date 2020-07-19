package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/app"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/perf"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
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

func getWindowRefreshRateRange() (low int32, high int32, _ error) {
	n, err := sdl.GetNumVideoDisplays()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get number of video displays: %w", err)
	}
	if n >= 1 {
		mode, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get display mode for display 0: %w", err)
		}
		hz := mode.RefreshRate
		low, high = hz, hz
	}
	for i := 1; i < n; i++ {
		mode, err := sdl.GetCurrentDisplayMode(i)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to get display mode for display %v: %w", i, err)
		}
		hz := mode.RefreshRate
		if hz < low {
			low = hz
		}
		if hz > high {
			high = hz
		}
	}
	return low, high, nil
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
	var file string
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
		fmt.Fprintf(flag.CommandLine.Output(), "  %v [OPTIONS]\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(flag.CommandLine.Output(), "  Specify a file name with -file to open immediately.")
		fmt.Fprintln(flag.CommandLine.Output(), "  Otherwise, an open file dialog will be used, if supported.")
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions:")
		flag.PrintDefaults()
	}
	flag.BoolVar(&color, "color", true, "colorize the output logs")
	flag.BoolVar(&debug, "debug", false, "show debug logging")
	flag.StringVar(&file, "file", "", "name of the file to open without prompt")
	flag.IntVar(&fps, "fps", 144, "the frames per second to render at")
	flag.IntVar(&height, "height", 720, "the initial height of the window")
	flag.BoolVar(&info, "info", true, "show info logging")
	flag.BoolVar(&perform, "perf", false, "show performormance logging")
	flag.BoolVar(&quiet, "quiet", false, "hide all output, overrides other logging options")
	flag.BoolVar(&warn, "warn", true, "show warning logging")
	flag.IntVar(&width, "width", 960, "the initial width of the window")
	flag.Parse()

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
	log.Debugf("file option: \"%v\"", file)
	log.Debugf("enabled loggers: %v", strings.Join(loggers, ", "))
	log.Debugf("output colorized: %v", color)

	if fps <= 0 {
		log.Fatal("fps must be >= 0")
	}
	if width <= 0 {
		log.Fatal("width must be >= 0")
	}
	if height <= 0 {
		log.Fatal("height must be >= 0")
	}

	cfg := config.New(int32(width), int32(height), 30, fps)
	win, err := initWindow("Tabula Editor", cfg.ScreenWidth, cfg.ScreenHeight)
	if err != nil {
		log.Fatal(err)
	}

	lowHz, highHz, err := getWindowRefreshRateRange()
	if err != nil {
		log.Fatal(err)
	}
	if fps > int(highHz) || fps < int(lowHz) {
		log.Warnf("framerate %v not within set refresh rate range %v-%v", fps, lowHz, highHz)
	}

	app := app.New(file, win, cfg)
	app.Start()

	for app.Running() {
		sw := util.Start()
		for evt := sdl.PollEvent(); evt != nil; evt = sdl.PollEvent() {
			app.HandleSdlEvent(evt)
		}

		app.PostEventActions()
		nanos := sw.StopGetNano()
		perf.RecordAverageTime("main.frametime", nanos)
		frametime := time.Second / time.Duration(cfg.FramesPerSecond)
		limit := int64(float32(frametime) * 1.1)
		if nanos > limit {
			log.Perff("degredation: took %v out of expected %v", time.Duration(nanos), frametime)
		}
	}

	app.Quit()
}
