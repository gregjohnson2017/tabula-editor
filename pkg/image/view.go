package image

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"

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
	cfg         *config.Config
	area        sdl.Rect
	canvas      sdl.Rect
	view        sdl.FRect
	mousePix    sdl.Point
	mult        int32
	activeTool  Tool
	layers      []*Layer
	selLayer    *Layer
	canvasLayer *Layer
	dragLoc     sdl.Point
	panLoc      sdl.Point
	dragging    bool
	panning     bool
	bbComms     chan<- comms.Image
	toolComms   <-chan Tool
	checkerProg gfx.Program
	program     gfx.Program
	projName    string
}

func (iv *View) AddLayer(tex gfx.Texture) {
	iv.layers = append(iv.layers, NewLayer(sdl.Point{X: 0, Y: 0}, tex))
}

// NewView returns a pointer to a new View struct that implements ui.Component
func NewView(area sdl.Rect, bbComms chan<- comms.Image, toolComms <-chan Tool, cfg *config.Config) (*View, error) {
	var iv = &View{}
	iv.cfg = cfg
	iv.area = area
	iv.view = ui.RectToFRect(area)
	iv.bbComms = bbComms
	iv.toolComms = toolComms
	iv.mult = 0

	iv.canvas = sdl.Rect{
		X: -50,
		Y: -50,
		W: 100,
		H: 100,
	}

	var data = make([]byte, iv.canvas.W*iv.canvas.H*4)
	canvasTex, err := gfx.NewTexture(iv.canvas.W, iv.canvas.H, data, gl.RGBA, 4)
	if err != nil {
		return nil, err
	}
	canvasTex.SetParameter(gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	canvasTex.SetParameter(gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	iv.canvasLayer = NewLayer(sdl.Point{X: iv.canvas.X, Y: iv.canvas.Y}, canvasTex)
	iv.layers = append(iv.layers, iv.canvasLayer)

	v1, err := gfx.NewShader(gfx.VertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f1, err := gfx.NewShader(gfx.CheckerShaderFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	if iv.checkerProg, err = gfx.NewProgram(v1, f1); err != nil {
		return nil, err
	}

	f2, err := gfx.NewShader(gfx.FragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	if iv.program, err = gfx.NewProgram(v1, f2); err != nil {
		return nil, err
	}

	iv.checkerProg.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
	iv.program.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))

	iv.activeTool = &EmptyTool{}

	iv.CenterCanvas()
	iv.projName = "New Project"

	return iv, nil
}

// Destroy frees all assets acquired by the ui.Component
func (iv *View) Destroy() {
	iv.checkerProg.Destroy()
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
		iv.bbComms <- comms.Image{FileName: iv.projName, MousePix: iv.mousePix, Mult: iv.mult}
	}()

	// TODO selection outline

	// gl viewport 0, 0 is bottom left
	gl.Viewport(iv.area.X, iv.cfg.BottomBarHeight, iv.area.W, iv.area.H)

	iv.program.Bind()
	for _, layer := range iv.layers {
		if layer == iv.canvasLayer {
			iv.checkerProg.Bind()
			iv.canvasLayer.Render(iv.view)
			iv.program.Bind()
		} else {
			layer.Render(iv.view)
		}
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

// RenderCanvas draws what is on the canvas or area, whichever is larger
func (iv *View) RenderCanvas() {

	iv.program.UploadUniform("area", float32(iv.canvas.W), float32(iv.canvas.H))
	// gl viewport 0, 0 is bottom left
	gl.Viewport(0, 0, iv.canvas.W, iv.canvas.H)

	iv.program.Bind()
	for _, layer := range iv.layers {
		layer.Render(sdl.FRect{
			X: float32(iv.canvas.X),
			Y: float32(iv.canvas.Y),
			W: float32(iv.canvas.W),
			H: float32(iv.canvas.H),
		})
	}
	iv.program.Unbind()

	iv.updateView()
}

const maxZoom = 8

func (iv *View) updateView() {
	frac := float32(math.Pow(2, float64(-iv.mult)))
	newView := sdl.FRect{}
	newView.W = float32(iv.area.W) * frac
	newView.H = float32(iv.area.H) * frac
	newView.X = (iv.view.W-newView.W)/2 + iv.view.X
	newView.Y = (iv.view.H-newView.H)/2 + iv.view.Y
	iv.view = newView
	iv.checkerProg.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
	iv.program.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
}

func (iv *View) CenterCanvas() {
	iv.view = sdl.FRect{
		X: float32(iv.canvas.X) - (float32(iv.area.W)/2 - float32(iv.canvas.W)/2),
		Y: float32(iv.canvas.Y) - (float32(iv.area.H)/2 - float32(iv.canvas.H)/2),
		W: float32(iv.area.W),
		H: float32(iv.area.H),
	}
	iv.updateView()
	iv.checkerProg.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
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
		X: int32(math.Floor(float64(iv.view.X + float32(x)*iv.view.W/float32(iv.area.W)))),
		Y: int32(math.Floor(float64(iv.view.Y + float32(y)*iv.view.H/float32(iv.area.H)))),
	}
}

// OnEnter is called when the cursor enters the ui.Component's region
func (iv *View) OnEnter() {}

// OnLeave is called when the cursor leaves the ui.Component's region
func (iv *View) OnLeave() {
	iv.dragging = false
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
	} else if evt.Button == sdl.BUTTON_MIDDLE {
		if evt.State == sdl.PRESSED {
			iv.panning = true
		} else if evt.State == sdl.RELEASED {
			iv.panning = false
		}
		iv.panLoc.X = evt.X
		iv.panLoc.Y = evt.Y
	}
	return true
}

// OnMotion is called when the cursor moves within the ui.Component's region
func (iv *View) OnMotion(evt *sdl.MouseMotionEvent) bool {
	if !iv.dragging && !iv.panning {
		iv.updateMousePos(evt.X, evt.Y)
		iv.activeTool.OnMotion(evt, iv)
		if iv.selLayer == nil {
			return true
		}
		return ui.InBounds(iv.selLayer.area, sdl.Point{X: evt.X, Y: evt.Y})
	}
	if evt.State == sdl.ButtonRMask() {
		// do not allow the canvas to be dragged
		if iv.selLayer == nil || iv.selLayer == iv.canvasLayer {
			return true
		}
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
	} else if evt.State == sdl.ButtonMMask() {
		if iv.panning {
			iv.view.X += float32(iv.panLoc.X-evt.X) * float32(iv.view.W) / float32(iv.area.W)
			iv.view.Y += float32(iv.panLoc.Y-evt.Y) * float32(iv.view.W) / float32(iv.area.W)
			iv.panLoc.X = evt.X
			iv.panLoc.Y = evt.Y
		}
	}
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (iv *View) OnScroll(evt *sdl.MouseWheelEvent) bool {
	if iv.dragging {
		return true
	}
	if evt.Y > 0 {
		iv.mult++
		if iv.mult > maxZoom {
			iv.mult = maxZoom
		}
		iv.updateView()
	} else if evt.Y < 0 {
		iv.mult--
		iv.updateView()
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

// OnResize is called when the user resizes the window
func (iv *View) OnResize(x, y int32) {
	iv.area.W += x
	iv.area.H += y
	iv.updateView()
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
	w, h := iv.canvas.W, iv.canvas.H

	fb, err := gfx.NewFrameBuffer(iv.canvas.W, iv.canvas.H)
	if err != nil {
		return err
	}
	fb.Bind()
	iv.RenderCanvas()
	fb.Unbind()
	data := fb.GetTexture().GetData()
	img := image.NewNRGBA(image.Rect(0, 0, int(w), int(h)))
	// flip resulting data vertically
	for j := 0; j < int(h)/2; j++ {
		for i := 0; i < int(w)*4; i++ {
			a := j*int(w)*4 + i
			b := (int(h)-j-1)*int(w)*4 + i
			data[a], data[b] = data[b], data[a]
		}
	}
	copy(img.Pix, data)
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	switch ext := filepath.Ext(fileName); ext {
	case ".png":
		err = png.Encode(out, img)
		if err != nil {
			return err
		}
	case ".jpg", ".jpeg", ".jpe", ".jfif":
		// TODO add dialog to choose quality
		var opt jpeg.Options
		opt.Quality = 100
		err = jpeg.Encode(out, img, &opt)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("writing to file extension %v: %w", ext, ErrWriteFormat)
	}
	return nil
}

const ErrInvalidFormat log.ConstErr = "invalid project file (not .tabula)"

func (iv *View) SaveProject(fileName string) error {
	if ext := filepath.Ext(fileName); ext != ".tabula" {
		return fmt.Errorf("%w: %v", ErrInvalidFormat, fileName)
	}
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)

	if err = enc.Encode(iv.mult); err != nil {
		return err
	}
	if err = enc.Encode(iv.view); err != nil {
		return err
	}
	if err = enc.Encode(iv.canvas); err != nil {
		return err
	}
	// canvas layer written first
	if err = iv.canvasLayer.EncodeLayer(enc); err != nil {
		return err
	}
	// write number of other layers
	if err = enc.Encode(len(iv.layers) - 1); err != nil {
		return err
	}
	// then every other layer
	for _, layer := range iv.layers {
		if layer != iv.canvasLayer {
			if err = layer.EncodeLayer(enc); err != nil {
				return err
			}
		}
	}

	_, err = out.Write(network.Bytes())
	if err != nil {
		return err
	}
	iv.projName = filepath.Base(fileName)
	return nil
}

func (iv *View) LoadProject(fileName string) error {
	var err error
	var in *os.File
	if in, err = os.Open(fileName); err != nil {
		return err
	}

	defer in.Close()
	if ext := filepath.Ext(fileName); ext != ".tabula" {
		return fmt.Errorf("%w: %v", ErrInvalidFormat, fileName)
	}

	var network bytes.Buffer
	reader := bufio.NewReader(in)
	if _, err = network.ReadFrom(reader); err != nil {
		return err
	}
	dec := gob.NewDecoder(&network)

	if err = dec.Decode(&iv.mult); err != nil {
		return err
	}
	if err = dec.Decode(&iv.view); err != nil {
		return err
	}
	if err = dec.Decode(&iv.canvas); err != nil {
		return err
	}
	if iv.canvasLayer, err = DecodeLayer(dec); err != nil {
		return err
	}

	// reset layers list
	iv.layers = append([]*Layer{}, iv.canvasLayer)

	var nLayers int
	if err := dec.Decode(&nLayers); err != nil {
		return err
	}

	for i := 0; i < nLayers; i++ {
		var layer *Layer
		if layer, err = DecodeLayer(dec); err != nil {
			return err
		}
		iv.layers = append(iv.layers, layer)
	}

	iv.updateView()
	iv.projName = filepath.Base(fileName)
	return nil
}
