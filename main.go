package main

import (
	"fmt"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/go-gl/gl/v2.1/gl"

	"github.com/veandco/go-sdl2/img"
)

type config struct {
	screenWidth     int32
	screenHeight    int32
	bottomBarHeight int32
}

func inBounds(area *Rect, x int32, y int32) bool {
	if x < area.X || x >= area.X+area.W || y < area.Y || y >= area.Y+area.H {
		return false
	}
	return true
}

func initLibs() error {
	gtk.Init(nil)
	// other libraries
	if img.Init(img.INIT_PNG) != img.INIT_PNG { // TODO is there a gtk library for this?
		return fmt.Errorf("could not initialize PNG")
	}

	return nil
}

func errCheck(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var err error
	errCheck(initLibs())

	builder, err := gtk.BuilderNewFromFile("layout.ui")
	errCheck(err)
	winObj, err := builder.GetObject("main_window")
	errCheck(err)
	win, ok := winObj.(*gtk.Window)
	if !ok {
		panic(fmt.Errorf("obj is not a window"))
	}
	running := true
	win.Connect("destroy", func() {
		running = false
	})
	winWidth, winHeight := win.GetSize()
	win.AddEvents(int(gdk.SCROLL_MASK | gdk.POINTER_MOTION_MASK | gdk.BUTTON_PRESS_MASK | gdk.BUTTON_RELEASE_MASK))

	glareaObj, err := builder.GetObject("gl_drawing_area")
	errCheck(err)
	glarea, ok := glareaObj.(*gtk.GLArea)
	if !ok {
		panic(fmt.Errorf("obj is not a glarea"))
	}
	glarea.SetRequiredVersion(4, 6)

	fileName := "happyhug.png"
	if len(os.Args) == 2 {
		fileName = os.Args[1]
	}

	bottomBarHeight := int32(30)
	imageViewArea := &Rect{
		X: 0,
		Y: bottomBarHeight,
		W: int32(winWidth),
		H: int32(winHeight) - bottomBarHeight,
	}
	bottomBarArea := &Rect{
		X: 0,
		Y: 0,
		W: int32(winWidth),
		H: bottomBarHeight,
	}
	bottomBarComms := make(chan imageComm)
	cfg := &config{int32(winWidth), int32(winHeight), bottomBarHeight}
	var iv *ImageView
	var bb *BottomBar
	var comps []UIComponent
	realize := func(glarea *gtk.GLArea) {
		glarea.MakeCurrent()
		// INIT OPENGL
		errCheck(gl.Init())
		gl.ClearColor(1.0, 1.0, 1.0, 1.0)
		gl.Enable(gl.MULTISAMPLE)
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)
		// version := gl.GoStr(gl.GetString(gl.VERSION))
		// fmt.Printf("OpenGL version %s\n", version)
		var err error
		iv, err = NewImageView(imageViewArea, fileName, bottomBarComms, cfg)
		errCheck(err)
		bb, err = NewBottomBar(bottomBarArea, bottomBarComms, cfg)
		errCheck(err)
		comps = append(comps, iv, bb)
	}

	var iterations int64
	var imageTotalNs int64
	var bbTotalNs int64
	signals := map[string]interface{}{
		"init": realize,
		"draw": func(glarea *gtk.GLArea) bool {
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
			for _, comp := range comps {
				sw := start()
				errCheck(comp.Render())
				ns := sw.stopGetNano()
				switch comp.(type) {
				case *ImageView:
					imageTotalNs += ns
				case *BottomBar:
					bbTotalNs += ns
				}
			}
			iterations++
			return true
		},
		"quit": func(glarea *gtk.GLArea) {
			for _, comp := range comps {
				comp.Destroy()
			}
		},
		"window_resize": func(glarea *gtk.GLArea, width, height int) {
			diffx := int32(width - winWidth)
			diffy := int32(height - winHeight)
			winWidth = width
			winHeight = height
			for _, comp := range comps {
				comp.OnResize(diffx, diffy)
			}
			glarea.QueueRender()
		},
		"mouse_motion": func(win *gtk.Window, evt *gdk.Event) bool {
			motionEvt := &gdk.EventMotion{Event: evt}
			x, y := motionEvt.MotionVal()
			for _, comp := range comps {
				comp.OnMotion(int32(x), int32(y), motionEvt.State())
			}
			glarea.QueueRender()
			return true
		},
		"mouse_click": func(win *gtk.Window, evt *gdk.Event) bool {
			clickEvt := &gdk.EventButton{Event: evt}
			x, y := clickEvt.MotionVal()
			for _, comp := range comps {
				comp.OnClick(int32(x), int32(y), clickEvt)
			}
			glarea.QueueRender()
			return true
		},
		"mouse_scroll": func(win *gtk.Window, evt *gdk.Event) bool {
			scrollEvt := &gdk.EventScroll{Event: evt}
			for _, comp := range comps {
				comp.OnScroll(int32(scrollEvt.DeltaY()))
			}
			glarea.QueueRender()
			return true
		},
	}
	builder.ConnectSignals(signals)

	win.ShowAll()
	for running {
		gtk.MainIterationDo(false) // do not block
	}

	// print statistics
	fmt.Printf("Average render times:\n")
	fmt.Printf("ImageView: %v ns\n", float64(imageTotalNs)/float64(iterations))
	fmt.Printf("BottomBar: %v ns\n", float64(bbTotalNs)/float64(iterations))

	// cleanup
	glarea.Destroy()
	win.Destroy()
	img.Quit()
}
