package app

import (
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/comms"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/image"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/menu"
	"github.com/gregjohnson2017/tabula-editor/pkg/perf"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/kroppt/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

// Application holds state for the tabula application
type Application struct {
	cfg         *config.Config
	comps       []ui.Component
	currHover   ui.Component
	lastHover   ui.Component
	moved       bool
	postEvtActs chan func()
	running     bool
	ticker      *time.Ticker
	win         *sdl.Window
}

// New returns a newly instantiated application state struct
func New(fileName, project string, win *sdl.Window, cfg *config.Config) *Application {
	var err error

	imageViewArea := sdl.Rect{
		X: 0,
		Y: 0,
		W: cfg.ScreenWidth,
		H: cfg.ScreenHeight - cfg.BottomBarHeight,
	}
	bottomBarArea := &sdl.Rect{
		X: 0,
		Y: cfg.ScreenHeight - cfg.BottomBarHeight,
		W: cfg.ScreenWidth,
		H: cfg.BottomBarHeight,
	}

	bottomBarComms := make(chan comms.Image)
	toolComms := make(chan image.Tool)
	actionComms := make(chan func())

	iv, err := image.NewView(imageViewArea, bottomBarComms, toolComms, cfg)
	if err != nil {
		log.Fatal(err)
	}

	if fileName != "" {
		tex, err := gfx.NewTextureFromFile(fileName)
		if err != nil {
			log.Fatal(err)
		}
		iv.AddLayer(tex)
	}
	if project != "" {
		if err = iv.LoadProject(project); err != nil {
			log.Fatal(err)
		}
	}

	bottomBar, err := NewBottomBar(bottomBarArea, bottomBarComms, cfg)
	if err != nil {
		log.Fatal(err)
	}

	menuBar, err := menu.NewBar(cfg, []menu.Definition{
		{
			Text: "File",
			Children: []menu.Definition{
				{
					Text: "Open as Layer",
					Action: func() {
						newFileName, err := util.OpenFileDialog(win)
						if err != nil {
							log.Warn(err)
							return
						}
						go func() {
							actionComms <- func() {
								tex, err := gfx.NewTextureFromFile(newFileName)
								if err != nil {
									log.Fatal(err)
								}
								iv.AddLayer(tex)
							}
						}()
					},
				},
				{
					Text: "Export Canvas",
					Action: func() {
						newFileName, err := util.SaveFileDialog(win)
						if err != nil {
							log.Warn(err)
							return
						}
						go func() {
							actionComms <- func() {
								if err := iv.WriteToFile(newFileName); err != nil {
									log.Fatal(err)
								}
							}
						}()
					},
				},
				{
					Text: "Save Project",
					Action: func() {
						go func() {
							newFileName, err := util.SaveFileDialog(win)
							if err != nil {
								log.Warn(err)
								return
							}
							actionComms <- func() {
								err = iv.SaveProject(newFileName)
								if err != nil {
									log.Fatal(err)
								}
							}
						}()
					},
				},
				{
					Text: "Load Project",
					Action: func() {
						go func() {
							newFileName, err := util.OpenFileDialog(win)
							if err != nil {
								log.Warn(err)
								return
							}
							actionComms <- func() {
								err = iv.LoadProject(newFileName)
								if err != nil {
									log.Fatal(err)
								}
							}
						}()
					},
				},
				{
					Text:   "kitten",
					Action: func() { log.Info("kitten") },
					Children: []menu.Definition{
						{
							Text:   "Mooney",
							Action: func() { log.Info("Mooney") },
						},
						{
							Text:   "Buttercup",
							Action: func() { log.Info("Buttercup") },
						},
						{
							Text:   "Sunny",
							Action: func() { log.Info("Sunny") },
						},
					},
				},
			},
		},
		{
			Text: "Tools",
			Children: []menu.Definition{
				{
					Text: "None",
					Action: func() {
						// clear the image view tool
						go func() { toolComms <- image.EmptyTool{} }()
					},
				},
				{
					Text: "Pixel selector",
					Action: func() {
						// set the image view tool to the pixel selection tool
						go func() { toolComms <- &image.PixelSelectionTool{} }()
					},
				},
				{
					Text: "Pixel color changer",
					Action: func() {
						// set the image view tool to the pixel selection tool
						go func() { toolComms <- &image.PixelColorTool{} }()
					},
				},
			},
		},
		{
			Text: "Image",
			Children: []menu.Definition{
				{
					Text: "Center Canvas",
					Action: func() {
						go func() {
							actionComms <- func() {
								iv.CenterCanvas()
							}
						}()
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	frametime := time.Second / time.Duration(cfg.FramesPerSecond)
	ticker := time.NewTicker(frametime)
	log.Debugf("set framerate %v with frametime %v", cfg.FramesPerSecond, frametime)

	return &Application{
		running:     false,
		comps:       []ui.Component{iv, bottomBar, menuBar},
		cfg:         cfg,
		postEvtActs: actionComms,
		ticker:      ticker,
		win:         win,
	}
}

// Start sets up the state for running
func (app *Application) Start() {
	app.running = true
	app.win.Show()
}

// HandleSdlEvent checks the type of a given SDL event and runs the method associated with that event
func (app *Application) HandleSdlEvent(e sdl.Event) {
	switch evt := e.(type) {
	case *sdl.QuitEvent:
		app.handleQuitEvent(evt)
	case *sdl.MouseButtonEvent:
		app.handleMouseButtonEvent(evt)
	case *sdl.MouseMotionEvent:
		app.handleMouseMotionEvent(evt)
	case *sdl.MouseWheelEvent:
		app.handleMouseWheelEvent(evt)
	case *sdl.WindowEvent:
		app.handleWindowEvent(evt)
	case *sdl.SysWMEvent:
		app.handleSysWMEvent(evt)
	}
}

// PostEventActions performs any necessary actions following event polling
func (app *Application) PostEventActions() {
	// handle all events in pipe
	hasEvents := true
	for hasEvents {
		select {
		case action := <-app.postEvtActs:
			log.Debug("handling post event action")
			action()
		default:
			// no more in pipe
			hasEvents = false
		}
	}
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	for _, comp := range app.comps {
		comp.Render()
	}

	app.win.GLSwap()
	perf.EndFrame()

	for glErr := gl.GetError(); glErr != gl.NO_ERROR; glErr = gl.GetError() {
		log.Warnf("OpenGL error: %v", glErr)
	}
	// wait until frametime has passed
	<-app.ticker.C
}

func (app *Application) handleQuitEvent(evt *sdl.QuitEvent) {
	app.running = false
}

func (app *Application) handleMouseButtonEvent(evt *sdl.MouseButtonEvent) {
	sw := util.Start()
	defer sw.StopRecordAverage("app.handleMouseButtonEvent")
	for i := range app.comps {
		comp := app.comps[len(app.comps)-i-1]
		if comp.InBoundary(sdl.Point{X: evt.X, Y: evt.Y}) {
			comp.OnClick(evt)
			log.Debugln("mouse button event on", comp.String())
			break
		}
	}
}

func (app *Application) handleMouseMotionEvent(evt *sdl.MouseMotionEvent) {
	sw := util.Start()
	defer sw.StopRecordAverage("app.handleMouseMotionEvent")
	// search top down through components until exhausted or one absorbs the event
	for i := range app.comps {
		comp := app.comps[len(app.comps)-i-1]
		if comp.InBoundary(sdl.Point{X: evt.X, Y: evt.Y}) {
			if app.currHover != comp {
				// entered a new component
				comp.OnEnter()
				app.lastHover = app.currHover
				app.currHover = comp
				app.moved = true
			}
			if comp.OnMotion(evt) {
				break
			}
		}
	}
	if app.lastHover != nil && app.moved {
		app.lastHover.OnLeave()
		app.moved = false
	}
}

func (app *Application) handleMouseWheelEvent(evt *sdl.MouseWheelEvent) {
	sw := util.Start()
	defer sw.StopRecordAverage("app.handleMouseWheelEvent")
	for i := range app.comps {
		comp := app.comps[len(app.comps)-i-1]
		x, y, _ := sdl.GetMouseState()
		if comp.InBoundary(sdl.Point{X: x, Y: y}) {
			if comp.OnScroll(evt) {
				break
			}
		}
	}
}

func (app *Application) handleWindowEvent(evt *sdl.WindowEvent) {
	sw := util.Start()
	defer sw.StopRecordAverage("app.handleWindowEvent")
	if evt.Event == sdl.WINDOWEVENT_LEAVE || evt.Event == sdl.WINDOWEVENT_FOCUS_LOST || evt.Event == sdl.WINDOWEVENT_MINIMIZED {
		log.Debug("window focus lost")
		if app.currHover != nil {
			app.currHover.OnLeave()
			app.lastHover = app.currHover
			app.currHover = nil
			app.moved = false
		}
	} else if evt.Event == sdl.WINDOWEVENT_RESIZED {
		diffx := evt.Data1 - app.cfg.ScreenWidth
		diffy := evt.Data2 - app.cfg.ScreenHeight
		app.cfg.ScreenWidth = evt.Data1
		app.cfg.ScreenHeight = evt.Data2
		for _, comp := range app.comps {
			comp.OnResize(diffx, diffy)
		}
	}
}

func (app *Application) handleSysWMEvent(evt *sdl.SysWMEvent) {
	/*
		var ma MenuAction
		ma = getMenuAction(evt)
		switch ma {
		case MenuExit:
			app.running = false
		}
	*/
}

// Running returns whether the application is still running
func (app *Application) Running() bool {
	return app.running
}

// Quit cleans up resources
func (app *Application) Quit() {
	// free ui.Component assets
	for _, comp := range app.comps {
		log.Debugln("destroying", comp.String())
		comp.Destroy()
	}

	if err := app.win.Destroy(); err != nil {
		log.Fatal(err)
	}
	sdl.Quit()
}
