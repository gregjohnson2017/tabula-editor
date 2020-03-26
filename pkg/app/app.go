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
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// performance debugging metrics
var imageTotalNs, bbTotalNs, bTotalNs, mlTotalNs, iterations int64

// Application holds state for the tabula application
type Application struct {
	cfg         *config.Config
	comps       []ui.Component
	currHover   ui.Component
	framerate   *gfx.FPSmanager
	lastHover   ui.Component
	moved       bool
	postEvtActs chan func()
	running     bool
	win         *sdl.Window
}

// New returns a newly instantiated application state struct
func New(win *sdl.Window, cfg *config.Config) *Application {
	var fileName string
	var err error
	if len(os.Args) == 2 {
		fileName = os.Args[1]
	} else {
		if fileName, err = util.OpenFileDialog(win); err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
	}
	err = util.SetupMenuBar(win)
	errCheck(err)

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
	buttonAreaOpen := &sdl.Rect{
		X: 0,
		Y: 30,
		W: 125,
		H: 20,
	}
	buttonAreaCenter := &sdl.Rect{
		X: 125,
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
	openButton, err := menu.NewButton(buttonAreaOpen, cfg, "Open File", func() {
		// TODO fix spam click crash bug
		newFileName, err := util.OpenFileDialog(win)
		if err != nil {
			fmt.Printf("No file chosen\n")
			return
		}
		go func() {
			actionComms <- func() {
				err = iv.LoadFromFile(newFileName)
				errCheck(err)
			}
		}()
	})
	centerButton, err := menu.NewButton(buttonAreaCenter, cfg, "Center Image", func() {
		go func() {
			actionComms <- func() {
				iv.CenterImage()
			}
		}()
	})
	centerButton.SetHighlightBackgroundColor([4]float32{1.0, 0.0, 0.0, 1.0})
	centerButton.SetDefaultTextColor([4]float32{0.0, 0.0, 1.0, 1.0})

	catMenuList := menu.NewMenuList(cfg, false)
	toolsMenuList := menu.NewMenuList(cfg, false)

	menuBar := menu.NewMenuList(cfg, true)
	menuItems := []menu.Definition{
		{"cat", catMenuList, func() { fmt.Println("cat") }},
		{"Tools", toolsMenuList, func() {}},
	}
	if err = menuBar.SetChildren(0, 0, menuItems); err != nil {
		panic(err)
	}

	kittenMenuList := menu.NewMenuList(cfg, false)
	catSubmenuItems := []menu.Definition{
		{"kitty", &menu.MenuList{}, func() { fmt.Println("kitty") }},
		{"kitten", kittenMenuList, func() { fmt.Println("kitten") }},
	}
	if err = catMenuList.SetChildren(0, menuBar.GetBoundary().H, catSubmenuItems); err != nil {
		panic(err)
	}

	kittenSubmenuItems := []menu.Definition{
		{"Mooney", &menu.MenuList{}, func() { fmt.Println("Mooney") }},
		{"Buttercup", &menu.MenuList{}, func() { fmt.Println("Buttercup") }},
		{"Sunny", &menu.MenuList{}, func() { fmt.Println("Sunny") }},
	}
	if err = kittenMenuList.SetChildren(catMenuList.GetBoundary().W, menuBar.GetBoundary().H+catMenuList.GetBoundary().H/2, kittenSubmenuItems); err != nil {
		panic(err)
	}

	toolsSubmenuItems := []menu.Definition{
		{"No tool", &menu.MenuList{}, func() {
			// clear the image view tool
			go func() { toolComms <- image.EmptyTool{} }()
		}},
		{"Pixel selection tool", &menu.MenuList{}, func() {
			// set the image view tool to the pixel selection tool
			go func() { toolComms <- image.PixelSelectionTool{} }()
		}},
	}
	if err = toolsMenuList.SetChildren(0, menuBar.GetBoundary().H, toolsSubmenuItems); err != nil {
		panic(err)
	}

	var framerate = &gfx.FPSmanager{}
	gfx.InitFramerate(framerate)
	if gfx.SetFramerate(framerate, 144) != true {
		panic(fmt.Errorf("could not set framerate: %v", sdl.GetError()))
	}

	return &Application{
		running:     false,
		comps:       []ui.Component{iv, bottomBar, openButton, centerButton, menuBar},
		cfg:         cfg,
		postEvtActs: actionComms,
		framerate:   framerate,
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
		case *menu.MenuList:
			mlTotalNs += ns
		}
	}
	iterations++

	app.win.GLSwap()
	gfx.FramerateDelay(app.framerate)
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
	fmt.Printf("image.View avg:\t%v\n", avg)
	avgNs = int64(float64(bbTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("BottomBar avg:\t%v\n", avg)
	avgNs = int64(float64(bTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("Button avg:\t%v\n", avg)
	avgNs = int64(float64(mlTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("MenuList avg:\t%v\n", avg)

	// free ui.Component assets
	for _, comp := range app.comps {
		comp.Destroy()
	}

	app.win.Destroy()
	sdl.Quit()
	img.Quit()
}

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}