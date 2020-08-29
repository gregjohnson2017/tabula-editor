package image

import (
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
	set "github.com/kroppt/Int32Set"
	"github.com/veandco/go-sdl2/sdl"
)

var _ ui.Component = ui.Component(&View{})

// View defines an interactable image viewing pane
type View struct {
	area         *sdl.Rect
	canvas       *sdl.Rect
	origW, origH int32
	cfg          *config.Config
	mouseLoc     sdl.Point
	mousePix     sdl.Point
	dragPoint    sdl.Point
	dragging     bool
	mult         float64
	program      gfx.Program
	selProgram   gfx.Program
	texture      gfx.Texture
	imgBuf       *gfx.BufferArray
	selBuf       *gfx.BufferArray
	bbComms      chan<- comms.Image
	fileName     string
	fullPath     string
	toolComms    <-chan Tool
	activeTool   Tool
	selection    set.Set
}

func (iv *View) LoadFromFile(fileName string) error {
	iv.texture.Delete() // clear old texture data before loading new

	tex, err := gfx.NewTextureFromFile(fileName)
	if err != nil {
		return err
	}
	iv.texture = tex
	iv.origW = int32(tex.GetWidth())
	iv.origH = int32(tex.GetHeight())

	iv.selProgram.UploadUniform("origDims", float32(iv.origW), float32(iv.origH))

	iv.canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: iv.origW,
		H: iv.origH,
	}
	iv.CenterImage()
	iv.mult = 1.0
	iv.selProgram.UploadUniform("mult", float32(iv.mult))

	parts := strings.Split(fileName, string(os.PathSeparator))
	iv.fileName = parts[len(parts)-1]
	iv.fullPath = fileName

	iv.selection = set.NewSet()
	return nil
}

// NewView returns a pointer to a new View struct that implements ui.Component
func NewView(area *sdl.Rect, fileName string, bbComms chan<- comms.Image, toolComms <-chan Tool, cfg *config.Config) (*View, error) {
	var err error
	var iv = &View{}
	iv.cfg = cfg
	iv.area = area
	iv.bbComms = bbComms
	iv.toolComms = toolComms

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

	v2, err := gfx.NewShader(gfx.OutlineVsh, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f2, err := gfx.NewShader(gfx.OutlineFsh, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	if iv.selProgram, err = gfx.NewProgram(v2, f2); err != nil {
		return nil, err
	}

	if err = iv.LoadFromFile(fileName); err != nil {
		return nil, err
	}

	iv.program.UploadUniform("area", float32(iv.area.W), float32(iv.area.H))

	// layout: (x,y s,t)
	iv.imgBuf = gfx.NewBufferArray(gl.TRIANGLES, []int32{2, 2})
	// layout: (x,y)
	iv.selBuf = gfx.NewBufferArray(gl.LINES, []int32{2})

	iv.selection = set.NewSet()
	iv.activeTool = EmptyTool{}

	return iv, nil
}

// Destroy frees all assets acquired by the ui.Component
func (iv *View) Destroy() {
	iv.texture.Delete()
	iv.imgBuf.Destroy()
	iv.selBuf.Destroy()
	iv.program.Delete()
	iv.selProgram.Delete()
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (iv *View) InBoundary(pt sdl.Point) bool {
	return ui.InBounds(*iv.area, pt)
}

// Render draws the ui.Component
func (iv *View) Render() {
	sw := util.Start()
	go func() {
		iv.bbComms <- comms.Image{FileName: iv.fileName, MousePix: iv.mousePix, Mult: iv.mult}
	}()

	// TODO optimize this (ex: move elsewhere, update as changes come in, use a better algorithm)
	// make array of 2d-vertex pairs representing texel selection outlines
	swl := util.Start()
	lines := []float32{}
	iv.selection.Range(func(i int32) bool {
		// i is every y*width+x index
		texelX := float32(i % iv.origW)
		texelY := float32((float32(i) - texelX) / float32(iv.origW))
		tlx, tly := texelX, texelY
		trx, try := (texelX + 1), texelY
		blx, bly := texelX, (texelY + 1)
		brx, bry := (texelX + 1), (texelY + 1)
		// left edge
		if !iv.selection.Contains(i-1) || i%iv.origW == 0 {
			lines = append(lines, tlx, tly, blx, bly)
		}
		// top edge
		if !iv.selection.Contains(i - iv.origW) {
			lines = append(lines, tlx, tly, trx, try)
		}
		// right edge
		if !iv.selection.Contains(i+1) || (i+1)%iv.origW == 0 {
			lines = append(lines, trx, try, brx, bry)
		}
		// bottom edge
		if !iv.selection.Contains(i + iv.origW) {
			lines = append(lines, blx, bly, brx, bry)
		}
		return true
	})
	swl.StopRecordAverage(iv.String() + ".SelLines")

	if len(lines) > 0 {
		err := iv.selBuf.Load(lines, gl.STATIC_DRAW)
		if err != nil {
			log.Warnf("failed to load image selection lines: %v", err)
		}
	}

	// update triangles that represent the position and scale of the image (these are SDL/window coordinates, converted to -1,1 gl space coordinates in the vertex shader)
	tlx, tly := float32(iv.canvas.X), float32(iv.canvas.Y)
	trx, try := float32(iv.canvas.X+iv.canvas.W), float32(iv.canvas.Y)
	blx, bly := float32(iv.canvas.X), float32(iv.canvas.H+iv.canvas.Y)
	brx, bry := float32(iv.canvas.X+iv.canvas.W), float32(iv.canvas.H+iv.canvas.Y)
	triangles := []float32{
		blx, bly, 0.0, 1.0, // bottom-left
		tlx, tly, 0.0, 0.0, // top-left
		trx, try, 1.0, 0.0, // top-right

		blx, bly, 0.0, 1.0, // bottom-left
		trx, try, 1.0, 0.0, // top-right
		brx, bry, 1.0, 1.0, // bottom-right
	}

	err := iv.imgBuf.Load(triangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load image triangles: %v", err)
	}

	// gl viewport x,y is bottom left
	gl.Viewport(iv.area.X, iv.cfg.ScreenHeight-iv.area.H-iv.area.Y, iv.area.W, iv.area.H)
	// draw image
	iv.program.Bind()
	iv.texture.Bind()
	iv.imgBuf.Draw()
	iv.texture.Unbind()
	iv.program.Unbind()

	// draw selection outlines
	gl.Viewport(iv.area.X+iv.canvas.X, iv.cfg.ScreenHeight-iv.area.Y-iv.canvas.Y-iv.canvas.H, iv.canvas.W, iv.canvas.H)

	iv.selProgram.Bind()
	iv.selBuf.Draw()
	iv.selProgram.Unbind()

	select {
	case tool := <-iv.toolComms:
		log.Debugln("image.View switching tool to", tool.String())
		iv.activeTool = tool
	default:
	}
	sw.StopRecordAverage(iv.String() + ".Render")
}

func (iv *View) zoomIn() {
	iv.mult *= 2.0
	iv.canvas.W = int32(float64(iv.origW) * iv.mult)
	iv.canvas.H = int32(float64(iv.origH) * iv.mult)
	iv.canvas.X = 2*iv.canvas.X - int32(math.Round(float64(iv.area.W)/2.0)) //iv.mouseLoc.x
	iv.canvas.Y = 2*iv.canvas.Y - int32(math.Round(float64(iv.area.H)/2.0)) //iv.mouseLoc.y
	iv.selProgram.UploadUniform("mult", float32(iv.mult))
}

func (iv *View) zoomOut() {
	iv.mult /= 2.0
	iv.canvas.W = int32(float64(iv.origW) * iv.mult)
	iv.canvas.H = int32(float64(iv.origH) * iv.mult)
	iv.canvas.X = int32(math.Round(float64(iv.canvas.X)/2.0 + float64(iv.area.W)/4.0)) //iv.mouseLoc.x/2
	iv.canvas.Y = int32(math.Round(float64(iv.canvas.Y)/2.0 + float64(iv.area.H)/4.0)) //iv.mouseLoc.y/2
	iv.selProgram.UploadUniform("mult", float32(iv.mult))
}

func (iv *View) CenterImage() {
	iv.canvas.X = int32(float64(iv.area.W)/2.0 - float64(iv.canvas.W)/2.0)
	iv.canvas.Y = int32(float64(iv.area.H)/2.0 - float64(iv.canvas.H)/2.0)
}

func (iv *View) setPixel(p sdl.Point, col color.RGBA) error {
	return iv.texture.SetPixel(p, col)
}

func (iv *View) updateMousePos(x, y int32) {
	iv.mouseLoc.X = x
	iv.mouseLoc.Y = y
	relx := float64(iv.mouseLoc.X - iv.canvas.X)
	rely := float64(iv.mouseLoc.Y - iv.canvas.Y)
	iv.mousePix.X = int32(math.Floor(relx / iv.mult))
	iv.mousePix.Y = int32(math.Floor(rely / iv.mult))
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
		return ui.InBounds(*iv.canvas, sdl.Point{X: evt.X, Y: evt.Y})
	}
	if evt.State == sdl.ButtonRMask() {
		iv.canvas.X += evt.X - iv.dragPoint.X
		iv.canvas.Y += evt.Y - iv.dragPoint.Y
		iv.dragPoint.X = evt.X
		iv.dragPoint.Y = evt.Y
	}
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (iv *View) OnScroll(evt *sdl.MouseWheelEvent) bool {
	if iv.dragging {
		return true
	}
	if evt.Y > 0 {
		if int32(iv.mult*float64(iv.origW)*2.0) > iv.canvas.W && int32(iv.mult*float64(iv.origH)*2.0) > iv.canvas.H && iv.mult < 256 {
			iv.zoomIn()
		}
	} else if evt.Y < 0 {
		if int32(iv.mult*float64(iv.origW)/2.0) > 0 && int32(iv.mult*float64(iv.origH)/2.0) > 0 {
			iv.zoomOut()
		}
	}
	return true
}

// ErrCoordOutOfRange indicates that given coordinates are out of range
const ErrCoordOutOfRange log.ConstErr = "coordinates out of range"

// SelectPixel adds the given x, y pixel to the
func (iv *View) SelectPixel(p sdl.Point) error {
	if p.X < 0 || p.Y < 0 || p.X >= iv.origW || p.Y >= iv.origH {
		return fmt.Errorf("SelectPixel(%v, %v): %w", p.X, p.Y, ErrCoordOutOfRange)
	}
	iv.selection.Add(p.X + p.Y*iv.origW)
	return nil
}

// OnClick is called when the user clicks within the ui.Component's region
func (iv *View) OnClick(evt *sdl.MouseButtonEvent) bool {
	iv.updateMousePos(evt.X, evt.Y)
	iv.activeTool.OnClick(evt, iv)
	if !ui.InBounds(*iv.canvas, sdl.Point{X: evt.X, Y: evt.Y}) {
		return true
	}
	if evt.Button == sdl.BUTTON_RIGHT {
		if evt.State == sdl.PRESSED {
			iv.dragging = true
		} else if evt.State == sdl.RELEASED {
			iv.dragging = false
		}
		iv.dragPoint.X = evt.X
		iv.dragPoint.Y = evt.Y
	}
	return true
}

// OnResize is called when the user resizes the window
func (iv *View) OnResize(x, y int32) {
	iv.area.W += x
	iv.area.H += y
	iv.program.UploadUniform("area", float32(iv.area.W), float32(iv.area.H))

	iv.CenterImage()
}

// String returns the name of the component type
func (iv *View) String() string {
	return "image.View"
}

// ErrWriteFormat indicates that an unsupported image format was trying to be written to
const ErrWriteFormat log.ConstErr = "unsupported image format"

// WriteToFile writes the image data stored in the OpenGL texture to a file specified by fileName
func (iv *View) WriteToFile(fileName string) error {
	data := iv.texture.GetTextureData()

	img := image.NewNRGBA(image.Rect(0, 0, int(iv.texture.GetWidth()),
		int(iv.texture.GetHeight())))
	copy(img.Pix, data)
	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

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
