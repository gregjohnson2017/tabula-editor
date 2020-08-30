package image

import (
	"image/color"
	"math"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/comms"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/veandco/go-sdl2/sdl"
)

var _ ui.Component = ui.Component(&View{})

// View defines an interactable image viewing pane
type View struct {
	cfg        *config.Config
	area       sdl.Rect
	view       sdl.Rect
	pan        sdl.Point
	mousePix   sdl.Point
	mult       int32
	activeTool Tool
	layers     []*Layer
	selLayer   *Layer
	mouseLoc   sdl.Point
	dragLoc    sdl.Point
	dragging   bool
	bbComms    chan<- comms.Image
	toolComms  <-chan Tool
	program    gfx.Program
}

func (iv *View) AddLayer(tex gfx.Texture) {
	iv.layers = append(iv.layers, NewLayer(sdl.Point{X: 0, Y: 0}, tex))
}

// NewView returns a pointer to a new View struct that implements ui.Component
func NewView(area sdl.Rect, bbComms chan<- comms.Image, toolComms <-chan Tool, cfg *config.Config) (*View, error) {
	var iv = &View{}
	iv.cfg = cfg
	iv.area = area
	iv.view = area
	iv.bbComms = bbComms
	iv.toolComms = toolComms
	iv.mult = 0

	v1, err := gfx.NewShader(gfx.VertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f1, err := gfx.NewShader(gfx.CheckerShaderFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	if iv.program, err = gfx.NewProgram(v1, f1); err != nil {
		return nil, err
	}

	iv.program.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
	iv.activeTool = &EmptyTool{}

	iv.zoom()

	return iv, nil
}

// Destroy frees all assets acquired by the ui.Component
func (iv *View) Destroy() {
	iv.program.Destroy()
	for _, layer := range iv.layers {
		layer.Destroy()
	}
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (iv *View) InBoundary(pt sdl.Point) bool {
	return ui.InBounds(iv.area, pt)
}

// Render draws the ui.Component
func (iv *View) Render() {
	sw := util.Start()
	go func() {
		iv.bbComms <- comms.Image{FileName: "layer", MousePix: iv.mousePix, Mult: iv.mult}
	}()

	// TODO selection outline

	// gl viewport 0, 0 is bottom left
	gl.Viewport(iv.area.X, iv.cfg.BottomBarHeight, iv.area.W, iv.area.H)

	iv.program.Bind()
	for _, layer := range iv.layers {
		layer.Render(iv.view)
	}
	iv.program.Unbind()

	select {
	case tool := <-iv.toolComms:
		log.Debugln("image.View switching tool to", tool.String())
		iv.activeTool = tool
	default:
	}
	sw.StopRecordAverage(iv.String() + ".Render")
}

const maxZoom = 8

func (iv *View) zoom() {
	frac := float32(math.Pow(2, float64(-iv.mult)))
	newView := sdl.Rect{}
	newView.W = int32(float32(iv.area.W) * frac)
	newView.H = int32(float32(iv.area.H) * frac)
	newView.X = iv.pan.X + (iv.area.W-newView.W)/2 - iv.area.W/2
	newView.Y = iv.pan.Y + (iv.area.H-newView.H)/2 - iv.area.H/2
	iv.view = newView
	iv.program.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
}

func (iv *View) zoomIn() {
	iv.mult++
	if iv.mult > maxZoom {
		iv.mult = maxZoom
	}
	iv.zoom()
}

func (iv *View) zoomOut() {
	iv.mult--
	iv.zoom()
}

func (iv *View) CenterView() {
	iv.view = iv.area
	iv.pan.X = 0
	iv.pan.Y = 0
	iv.zoom()
	iv.program.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
}

func (iv *View) setPixel(p sdl.Point, col color.RGBA) error {
	if iv.selLayer != nil {
		p.X -= iv.selLayer.area.X
		p.Y -= iv.selLayer.area.Y
		return iv.selLayer.texture.SetPixel(p, col)
	}
	return nil
}

// x and y is in the SDL window coordinate space.
func (iv *View) updateMousePos(x, y int32) {
	iv.mousePix = iv.getMousePix(x, y)
}

// x and y is in the SDL window coordinate space.
func (iv *View) getMousePix(x, y int32) sdl.Point {
	return sdl.Point{
		X: iv.view.X + x*iv.view.W/iv.area.W,
		Y: iv.view.Y + y*iv.view.H/iv.area.H,
	}
}

// OnEnter is called when the cursor enters the ui.Component's region
func (iv *View) OnEnter() {}

// OnLeave is called when the cursor leaves the ui.Component's region
func (iv *View) OnLeave() {
	iv.dragging = false
}

// OnMotion is called when the cursor moves within the ui.Component's region
func (iv *View) OnMotion(evt *sdl.MouseMotionEvent) bool {
	if !iv.dragging {
		iv.updateMousePos(evt.X, evt.Y)
		iv.activeTool.OnMotion(evt, iv)
		if iv.selLayer == nil {
			return true
		}
		return ui.InBounds(iv.selLayer.area, sdl.Point{X: evt.X, Y: evt.Y})
	}
	if iv.selLayer == nil {
		return true
	}
	if evt.State == sdl.ButtonRMask() {
		newImgPix := iv.getMousePix(evt.X, evt.Y)
		oldImgPix := iv.getMousePix(iv.dragLoc.X, iv.dragLoc.Y)
		diff := sdl.Point{
			X: newImgPix.X - oldImgPix.X,
			Y: newImgPix.Y - oldImgPix.Y,
		}
		iv.selLayer.area.X += diff.X
		iv.selLayer.area.Y += diff.Y
		iv.dragLoc.X = evt.X
		iv.dragLoc.Y = evt.Y
	}
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (iv *View) OnScroll(evt *sdl.MouseWheelEvent) bool {
	if iv.dragging {
		return true
	}
	if evt.Y > 0 {
		iv.zoomIn()
	} else if evt.Y < 0 {
		iv.zoomOut()
	}
	return true
}

func (iv *View) selectLayer() {
	iv.selLayer = nil
	for i := len(iv.layers) - 1; i >= 0; i-- {
		layer := iv.layers[i]
		if ui.InBounds(layer.area, iv.mousePix) {
			iv.selLayer = layer
			return
		}
	}
}

// ErrCoordOutOfRange indicates that given coordinates are out of range
const ErrCoordOutOfRange log.ConstErr = "coordinates out of range"

// SelectPixel adds the given x, y pixel to the
func (iv *View) SelectPixel(p sdl.Point) error {
	if iv.selLayer == nil {
		return nil
	}
	if !ui.InBounds(iv.selLayer.area, p) {
		return nil
	}
	// TODO
	return nil
}

// OnClick is called when the user clicks within the ui.Component's region
func (iv *View) OnClick(evt *sdl.MouseButtonEvent) bool {
	iv.updateMousePos(evt.X, evt.Y)
	iv.activeTool.OnClick(evt, iv)
	iv.selectLayer()
	if evt.Button == sdl.BUTTON_RIGHT {
		if evt.State == sdl.PRESSED {
			if iv.selLayer == nil {
				// no layer was clicked on
				return true
			}
			iv.dragging = true
		} else if evt.State == sdl.RELEASED {
			iv.dragging = false
		}
		iv.dragLoc.X = evt.X
		iv.dragLoc.Y = evt.Y
	}
	return true
}

// OnResize is called when the user resizes the window
func (iv *View) OnResize(x, y int32) {
	iv.area.W += x
	iv.area.H += y
	iv.zoom()
}

// String returns the name of the component type
func (iv *View) String() string {
	return "image.View"
}

// ErrWriteFormat indicates that an unsupported image format was trying to be written to
const ErrWriteFormat log.ConstErr = "unsupported image format"

// WriteToFile writes the image data stored in the OpenGL texture to a file specified by fileName
func (iv *View) WriteToFile(fileName string) error {
	// TODO after canvas figured out
	return nil
}
