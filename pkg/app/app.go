package app

import (
	"fmt"
	"os"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/comms"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/image"
	"github.com/gregjohnson2017/tabula-editor/pkg/menu"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/veandco/go-sdl2/sdl"
)

// performance debugging metrics
var imageTotalNs, bbTotalNs, bTotalNs, mlTotalNs, iterations int64

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
func New(fileName string, win *sdl.Window, cfg *config.Config) *Application {
	var err error
	if fileName == "" {
		if fileName, err = util.OpenFileDialog(win); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}

	imageViewArea := &sdl.Rect{
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
	buttonAreaCenter := &sdl.Rect{
		X: 0,
		Y: 30,
		W: 125,
		H: 20,
	}

	bottomBarComms := make(chan comms.Image)
	toolComms := make(chan image.Tool)
	actionComms := make(chan func())

	iv, err := image.NewView(imageViewArea, fileName, bottomBarComms, toolComms, cfg)
	errCheck(err)
	bottomBar, err := NewBottomBar(bottomBarArea, bottomBarComms, cfg)
	errCheck(err)
	centerButton, err := menu.NewButton(buttonAreaCenter, cfg, "Center Image", func() {
		go func() {
			actionComms <- func() {
				iv.CenterImage()
			}
		}()
	})
	errCheck(err)
	centerButton.SetHighlightBackgroundColor([4]float32{1.0, 0.0, 0.0, 1.0})
	centerButton.SetDefaultTextColor([4]float32{0.0, 0.0, 1.0, 1.0})

	menuBar, err := menu.NewBar(cfg, []menu.Definition{
		{
			Text: "File",
			Children: []menu.Definition{
				{
					Text: "Open",
					Action: func() {
						newFileName, err := util.OpenFileDialog(win)
						if err != nil {
							fmt.Printf("%v\n", err)
							return
						}
						go func() {
							actionComms <- func() {
								err = iv.LoadFromFile(newFileName)
								errCheck(err)
							}
						}()
					},
				},
				{
					Text: "Save As",
					Action: func() {
						newFileName, err := util.SaveFileDialog(win)
						if err != nil {
							fmt.Printf("%v\n", err)
							return
						}
						go func() {
							actionComms <- func() {
								err := iv.WriteToFile(newFileName)
								errCheck(err)
								err = iv.LoadFromFile(newFileName)
								errCheck(err)
							}
						}()
					},
				},
				{
					Text:   "kitten",
					Action: func() { fmt.Println("kitten") },
					Children: []menu.Definition{
						{
							Text:   "Mooney",
							Action: func() { fmt.Println("Mooney") },
						},
						{
							Text:   "Buttercup",
							Action: func() { fmt.Println("Buttercup") },
						},
						{
							Text:   "Sunny",
							Action: func() { fmt.Println("Sunny") },
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
						go func() { toolComms <- image.PixelSelectionTool{} }()
					},
				},
				{
					Text: "Pixel color changer",
					Action: func() {
						// set the image view tool to the pixel selection tool
						go func() { toolComms <- image.PixelColorTool{} }()
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	frametime := time.Second / time.Duration(cfg.FramesPerSecond)
	ticker := time.NewTicker(frametime)

	return &Application{
		running:     false,
		comps:       []ui.Component{iv, bottomBar, centerButton, menuBar},
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
		case closure := <-app.postEvtActs:
			closure()
		default:
			// no more in pipe
			hasEvents = false
		}
	}
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	for _, comp := range app.comps {
		sw := util.Start()
		comp.Render()
		ns := sw.StopGetNano()
		switch comp.(type) {
		case *image.View:
			imageTotalNs += ns
		case *BottomBar:
			bbTotalNs += ns
		case *menu.Button:
			bTotalNs += ns
		case *menu.Bar:
			mlTotalNs += ns
		}
	}
	iterations++

	app.win.GLSwap()
	<-app.ticker.C
}

func (app *Application) handleQuitEvent(evt *sdl.QuitEvent) {
	app.running = false
}

func (app *Application) handleMouseButtonEvent(evt *sdl.MouseButtonEvent) {
	for i := range app.comps {
		comp := app.comps[len(app.comps)-i-1]
		if comp.InBoundary(sdl.Point{X: evt.X, Y: evt.Y}) {
			comp.OnClick(evt)
			break
		}
	}
}

func (app *Application) handleMouseMotionEvent(evt *sdl.MouseMotionEvent) {
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
	if evt.Event == sdl.WINDOWEVENT_LEAVE || evt.Event == sdl.WINDOWEVENT_FOCUS_LOST || evt.Event == sdl.WINDOWEVENT_MINIMIZED {
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
	avgNs := int64(float64(imageTotalNs) / float64(iterations))
	avg := time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("image.View avg:\t %v\n", avg)
	avgNs = int64(float64(bbTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("BottomBar avg:\t %v\n", avg)
	avgNs = int64(float64(bTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("menu.Button avg: %v\n", avg)
	avgNs = int64(float64(mlTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("menu.Bar avg:\t %v\n", avg)

	// free ui.Component assets
	for _, comp := range app.comps {
		comp.Destroy()
	}

	err := app.win.Destroy()
	errCheck(err)
	sdl.Quit()
}

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}
