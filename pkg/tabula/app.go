package tabula

import (
	"fmt"
	"os"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// performance debugging metrics
var imageTotalNs, bbTotalNs, bTotalNs, mlTotalNs, iterations int64

// Config represents the window configuration for the application
type Config struct {
	BottomBarHeight int32
	ScreenHeight    int32
	ScreenWidth     int32
}

// Application holds state for the tabula application
type Application struct {
	cfg         Config
	comps       []UIComponent
	currHover   UIComponent
	framerate   *gfx.FPSmanager
	lastHover   UIComponent
	moved       bool
	postEvtActs chan func()
	running     bool
	win         *sdl.Window
}

// NewApplication returns a newly instantiated application state struct
func NewApplication(win *sdl.Window, cfg Config) *Application {
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

	bottomBarComms := make(chan imageComm)
	actionComms := make(chan func())

	iv, err := NewImageView(imageViewArea, fileName, bottomBarComms, &cfg)
	errCheck(err)
	bottomBar, err := NewBottomBar(bottomBarArea, bottomBarComms, &cfg)
	errCheck(err)
	openButton, err := NewButton(buttonAreaOpen, &cfg, "Open File", func() {
		// TODO fix spam click crash bug
		newFileName, err := util.OpenFileDialog(win)
		if err != nil {
			fmt.Printf("No file chosen\n")
			return
		}
		go func() {
			actionComms <- func() {
				err = iv.loadFromFile(newFileName)
				errCheck(err)
			}
		}()
	})
	centerButton, err := NewButton(buttonAreaCenter, &cfg, "Center Image", func() {
		go func() {
			actionComms <- func() {
				iv.centerImage()
			}
		}()
	})
	centerButton.SetHighlightBackgroundColor([4]float32{1.0, 0.0, 0.0, 1.0})
	centerButton.SetDefaultTextColor([4]float32{0.0, 0.0, 1.0, 1.0})

	catmenuList := NewMenuList(&cfg, false)

	menuBar := NewMenuList(&cfg, true)
	menuItems := []struct {
		str string
		ml  *MenuList
		act func()
	}{
		{"cat", catmenuList, func() { fmt.Println("cat") }},
		{"dog", &MenuList{}, func() { fmt.Println("dog") }},
		{"wolf", &MenuList{}, func() { fmt.Println("wolf") }},
		{"giraffe", &MenuList{}, func() { fmt.Println("giraffe") }},
		{"elephant", &MenuList{}, func() { fmt.Println("elephant") }},
		{"lynx", &MenuList{}, func() { fmt.Println("lynx") }},
		{"zebra", &MenuList{}, func() { fmt.Println("zebra") }},
	}
	if err = menuBar.SetChildren(0, 0, menuItems); err != nil {
		panic(err)
	}

	submenuItems := []struct {
		str string
		ml  *MenuList
		act func()
	}{
		{"kitty", &MenuList{}, func() { fmt.Println("kitty") }},
		{"kitten", &MenuList{}, func() { fmt.Println("kitten") }},
	}
	if err = catmenuList.SetChildren(0, menuBar.area.H, submenuItems); err != nil {
		panic(err)
	}

	var framerate = &gfx.FPSmanager{}
	gfx.InitFramerate(framerate)
	if gfx.SetFramerate(framerate, 144) != true {
		panic(fmt.Errorf("could not set framerate: %v", sdl.GetError()))
	}

	return &Application{
		running:     false,
		comps:       []UIComponent{iv, bottomBar, openButton, centerButton, menuBar},
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
		case *ImageView:
			imageTotalNs += ns
		case *BottomBar:
			bbTotalNs += ns
		case *Button:
			bTotalNs += ns
		case *MenuList:
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
	fmt.Printf("ImageView avg: %v ns, %v\n", avgNs, avg)
	avgNs = int64(float64(bbTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("BottomBar avg: %v ns, %v\n", avgNs, avg)
	avgNs = int64(float64(bTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("Button avg: %v ns, %v\n", avgNs, avg)
	avgNs = int64(float64(mlTotalNs) / float64(iterations))
	avg = time.Duration(int64(time.Nanosecond) * avgNs)
	fmt.Printf("MenuList avg: %v ns, %v\n", avgNs, avg)

	// free UIComponent SDL assets
	for _, comp := range app.comps {
		comp.Destroy()
	}

	app.win.Destroy()
	sdl.Quit()
	img.Quit()
}

func inBounds(area *sdl.Rect, x int32, y int32) bool {
	if x < area.X || x >= area.X+area.W || y < area.Y || y >= area.Y+area.H {
		return false
	}
	return true
}

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}
