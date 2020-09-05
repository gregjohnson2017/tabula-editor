package image

import (
	"bytes"
	"compress/zlib"
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

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
	selProg     gfx.Program
	program     gfx.Program
	projName    string
}

func (iv *View) AddLayer(tex gfx.Texture) error {
	layer, err := NewLayer(sdl.Point{X: 0, Y: 0}, tex)
	if err != nil {
		return err
	}
	iv.layers = append(iv.layers, layer)
	return nil
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
	iv.canvasLayer, err = NewLayer(sdl.Point{X: iv.canvas.X, Y: iv.canvas.Y}, canvasTex)
	if err != nil {
		return nil, err
	}
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

	outlineVsh, err := gfx.NewShader(gfx.VshPassthrough, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	outlineFsh, err := gfx.NewShader(gfx.OutlineFsh, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}
	outlineGeo, err := gfx.NewShader(gfx.OutlineGeometry, gl.GEOMETRY_SHADER_ARB)
	if err != nil {
		return nil, err
	}
	if iv.selProg, err = gfx.NewProgram(outlineVsh, outlineFsh, outlineGeo); err != nil {
		return nil, err
	}

	iv.checkerProg.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
	iv.program.UploadUniform("area", float32(iv.view.W), float32(iv.view.H))
	iv.selProg.UploadUniform("view", float32(iv.view.X), float32(iv.view.Y), float32(iv.view.W), float32(iv.view.H))

	iv.activeTool = &EmptyTool{}

	iv.CenterCanvas()
	iv.projName = "New Project"

	// xWorkers := iv.canvasLayer.GetSelTex().GetWidth() / 10
	// yWorkers := iv.canvasLayer.GetSelTex().GetHeight() / 10

	// ssboData := make([]int32, xWorkers*yWorkers)
	// var ssbo uint32
	// gl.GenBuffers(1, &ssbo)
	// gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, ssbo)
	// gl.BufferData(gl.SHADER_STORAGE_BUFFER, 4*len(ssboData), gl.Ptr(&ssboData[0]), gl.STATIC_DRAW)
	// gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0, ssbo)
	// gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, 0)
	// var x, y, z int32
	// gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 0, &x)
	// gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 1, &y)
	// gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 2, &z)
	// log.Infof("x,y,z =  %v,%v,%v\n", x, y, z)

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

	for _, layer := range iv.layers {
		if layer == iv.canvasLayer {
			layer.Render(iv.view, iv.checkerProg)
		} else {
			layer.Render(iv.view, iv.program)
		}
		layer.RenderSelection(iv.view, iv.selProg)
	}

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
	sw := util.Start()
	iv.program.UploadUniform("area", float32(iv.canvas.W), float32(iv.canvas.H))
	// gl viewport 0, 0 is bottom left
	gl.Viewport(0, 0, iv.canvas.W, iv.canvas.H)

	for _, layer := range iv.layers {
		layer.Render(ui.RectToFRect(iv.canvas), iv.program)
	}

	iv.updateView()
	sw.Stop("RenderCanvas")
}

const maxZoom = 8

// updateView updates the view rectangle according to the zoom multiplier,
// while maintaining the current pan
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
	iv.selProg.UploadUniform("view", float32(iv.view.X), float32(iv.view.Y), float32(iv.view.W), float32(iv.view.H))
}

// CenterCanvas updates the view so the canvas is in the center of the window
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

// setPixel sets the currently hovered texel of the selected layer
// to the specified color
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
	iv.selectLayer()
	iv.activeTool.OnClick(evt, iv)
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
			iv.updateView()
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

// selectLayer sets the currently selected layer to nil, and sets the layer
// that the mouse is currently hovering over, if any.
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
	if iv.selLayer != nil {
		p.X -= iv.selLayer.area.X
		p.Y -= iv.selLayer.area.Y
		return iv.selLayer.SelectTexel(p)
	}
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

// WriteToFile uses an OpenGL Frame Buffer Object to render the data in the canvas
// to a texture, and then write the data in that texture to the specified file
func (iv *View) WriteToFile(fileName string) error {
	sw := util.Start()
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

	sw.Stop("WriteToFile")
	return nil
}

type Project struct {
	ProjName string
	Mult     int32
	Canvas   sdl.Rect
	View     sdl.FRect
	Layers   []*Layer
}

const ErrInvalidFormat log.ConstErr = "invalid project file (not .tabula)"

// SaveProject saves the relevant project data at the specified file location
// in a compressed format. The fileName must end with '.tabula'
func (iv *View) SaveProject(fileName string) error {
	sw := util.Start()
	var ext string
	if ext = filepath.Ext(fileName); ext != ".tabula" {
		return fmt.Errorf("%w: %v", ErrInvalidFormat, fileName)
	}
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	proj := Project{
		ProjName: strings.TrimSuffix(filepath.Base(fileName), ext),
		Mult:     iv.mult,
		Canvas:   iv.canvas,
		View:     iv.view,
		Layers:   iv.layers,
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(proj); err != nil {
		return err
	}

	zw := zlib.NewWriter(out)
	if _, err = zw.Write(buf.Bytes()); err != nil {
		return err
	}
	defer zw.Close()

	iv.projName = proj.ProjName
	sw.Stop("SaveProject")
	return nil
}

// LoadProject loads the project data at the specified file location,
// decompresses and decodes the data and populates the relevant fields in
// the image view. The fileName must end with '.tabula'
func (iv *View) LoadProject(fileName string) error {
	sw := util.Start()
	var err error
	var in *os.File
	if in, err = os.Open(fileName); err != nil {
		return err
	}
	defer in.Close()
	if ext := filepath.Ext(fileName); ext != ".tabula" {
		return fmt.Errorf("%w: %v", ErrInvalidFormat, fileName)
	}

	zr, err := zlib.NewReader(in)
	if err != nil {
		return fmt.Errorf("zlib reader error: %w", err)
	}
	defer zr.Close()

	var proj Project
	dec := gob.NewDecoder(zr)
	if err = dec.Decode(&proj); err != nil {
		return fmt.Errorf("gob decoder error: %w", err)
	}

	iv.layers = proj.Layers
	iv.canvasLayer = proj.Layers[0]
	iv.mult = proj.Mult
	iv.view = proj.View
	iv.canvas = proj.Canvas
	iv.projName = proj.ProjName

	iv.updateView()
	sw.Stop("LoadProject")
	return nil
}
