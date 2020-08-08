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
	"unsafe"

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
	area           *sdl.Rect
	canvas         *sdl.Rect
	origW, origH   int32
	cfg            *config.Config
	mouseLoc       sdl.Point
	mousePix       sdl.Point
	dragPoint      sdl.Point
	dragging       bool
	mult           float64
	programID      uint32
	selProgramID   uint32
	textureID      uint32
	vaoID, vboID   uint32
	selVao, selVbo uint32
	bbComms        chan<- comms.Image
	fileName       string
	fullPath       string
	toolComms      <-chan Tool
	activeTool     Tool
	selection      set.Set
}

func (iv *View) LoadFromFile(fileName string) error {
	width, height, data, err := loadImage(fileName)
	if err != nil {
		return err
	}
	iv.origW = int32(width)
	iv.origH = int32(height)
	uniformID := gl.GetUniformLocation(iv.selProgramID, &[]byte("origDims\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform2f(uniformID, float32(iv.origW), float32(iv.origH))
	gl.UseProgram(0)

	format := int32(gl.RGBA)
	gl.DeleteTextures(1, &iv.textureID)
	gl.GenTextures(1, &iv.textureID)
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	// copy pixels to texture
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, iv.origW, iv.origH, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	iv.canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: iv.origW,
		H: iv.origH,
	}
	iv.CenterImage()
	iv.mult = 1.0
	uniformID = gl.GetUniformLocation(iv.selProgramID, &[]byte("mult\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform1f(uniformID, float32(iv.mult))
	gl.UseProgram(0)

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

	if iv.programID, err = gfx.CreateShaderProgram(gfx.VertexShaderSource, gfx.CheckerShaderFragment); err != nil {
		return nil, err
	}

	if iv.selProgramID, err = gfx.CreateShaderProgram(gfx.OutlineVsh, gfx.OutlineFsh); err != nil {
		return nil, err
	}

	if err = iv.LoadFromFile(fileName); err != nil {
		return nil, err
	}

	uniformID := gl.GetUniformLocation(iv.programID, &[]byte("area\x00")[0])
	gl.UseProgram(iv.programID)
	gl.Uniform2f(uniformID, float32(iv.area.W), float32(iv.area.H))
	gl.UseProgram(0)

	gl.GenBuffers(1, &iv.vboID)
	gl.GenVertexArrays(1, &iv.vaoID)
	gfx.ConfigureVAO(iv.vaoID, iv.vboID, []int32{2, 2})

	gl.GenBuffers(1, &iv.selVbo)
	gl.GenVertexArrays(1, &iv.selVao)
	gfx.ConfigureVAO(iv.selVao, iv.selVbo, []int32{2})

	iv.selection = set.NewSet()
	iv.activeTool = EmptyTool{}

	return iv, nil
}

// Destroy frees all assets acquired by the ui.Component
func (iv *View) Destroy() {
	gl.DeleteTextures(1, &iv.textureID)
	gl.DeleteBuffers(1, &iv.vboID)
	gl.DeleteVertexArrays(1, &iv.vaoID)
	gl.DeleteBuffers(1, &iv.selVbo)
	gl.DeleteVertexArrays(1, &iv.selVao)
	gl.DeleteProgram(iv.programID)
	gl.DeleteProgram(iv.selProgramID)
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
		gl.BindBuffer(gl.ARRAY_BUFFER, iv.selVbo)
		gl.BufferData(gl.ARRAY_BUFFER, 4*len(lines), gl.Ptr(&lines[0]), gl.STATIC_DRAW)
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
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

	gl.BindBuffer(gl.ARRAY_BUFFER, iv.vboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(triangles), gl.Ptr(&triangles[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// gl viewport x,y is bottom left
	gl.Viewport(iv.area.X, iv.cfg.ScreenHeight-iv.area.H-iv.area.Y, iv.area.W, iv.area.H)
	// draw image
	gl.UseProgram(iv.programID)

	gl.BindVertexArray(iv.vaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangles)/4))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindVertexArray(0)

	// draw selection outlines
	gl.Viewport(iv.area.X+iv.canvas.X, iv.cfg.ScreenHeight-iv.area.Y-iv.canvas.Y-iv.canvas.H, iv.canvas.W, iv.canvas.H)
	gl.UseProgram(iv.selProgramID)

	gl.BindVertexArray(iv.selVao)
	gl.EnableVertexAttribArray(0)
	gl.DrawArrays(gl.LINES, 0, int32(len(lines)/2))
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)

	gl.UseProgram(0)

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
	uniformID := gl.GetUniformLocation(iv.selProgramID, &[]byte("mult\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform1f(uniformID, float32(iv.mult))
	gl.UseProgram(0)
}

func (iv *View) zoomOut() {
	iv.mult /= 2.0
	iv.canvas.W = int32(float64(iv.origW) * iv.mult)
	iv.canvas.H = int32(float64(iv.origH) * iv.mult)
	iv.canvas.X = int32(math.Round(float64(iv.canvas.X)/2.0 + float64(iv.area.W)/4.0)) //iv.mouseLoc.x/2
	iv.canvas.Y = int32(math.Round(float64(iv.canvas.Y)/2.0 + float64(iv.area.H)/4.0)) //iv.mouseLoc.y/2
	uniformID := gl.GetUniformLocation(iv.selProgramID, &[]byte("mult\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform1f(uniformID, float32(iv.mult))
	gl.UseProgram(0)
}

func (iv *View) CenterImage() {
	iv.canvas.X = int32(float64(iv.area.W)/2.0 - float64(iv.canvas.W)/2.0)
	iv.canvas.Y = int32(float64(iv.area.H)/2.0 - float64(iv.canvas.H)/2.0)
}

func (iv *View) setPixel(x, y int32, color []byte) {
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	gl.TextureSubImage2D(iv.textureID, 0, x, y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&color[0]))
	// TODO update mipmap textures only when needed
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)
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
	}
	if !iv.dragging && !ui.InBounds(*iv.canvas, sdl.Point{X: evt.X, Y: evt.Y}) {
		return false
	}
	if evt.State == sdl.ButtonRMask() && iv.dragging {
		iv.canvas.X += evt.X - iv.dragPoint.X
		iv.canvas.Y += evt.Y - iv.dragPoint.Y
		iv.dragPoint.X = evt.X
		iv.dragPoint.Y = evt.Y
	}
	iv.activeTool.OnMotion(evt, iv)
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
func (iv *View) SelectPixel(x, y int32) error {
	if x < 0 || y < 0 || x > iv.origW || y > iv.origH {
		return fmt.Errorf("SelectPixel(%v, %v): %w", x, y, ErrCoordOutOfRange)
	}
	iv.selection.Add(iv.mousePix.X + iv.mousePix.Y*iv.origW)
	return nil
}

// OnClick is called when the user clicks within the ui.Component's region
func (iv *View) OnClick(evt *sdl.MouseButtonEvent) bool {
	iv.updateMousePos(evt.X, evt.Y)
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
	iv.activeTool.OnClick(evt, iv)
	return true
}

// OnResize is called when the user resizes the window
func (iv *View) OnResize(x, y int32) {
	iv.area.W += x
	iv.area.H += y

	uniformID := gl.GetUniformLocation(iv.programID, &[]byte("area\x00")[0])
	gl.UseProgram(iv.programID)
	gl.Uniform2f(uniformID, float32(iv.area.W), float32(iv.area.H))
	gl.UseProgram(0)

	iv.CenterImage()
}

// String returns the name of the component type
func (iv *View) String() string {
	return "image.View"
}

func loadImage(fileName string) (width, height int, data []byte, err error) {
	in, err := os.Open(fileName)
	if err != nil {
		return 0, 0, nil, err
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return 0, 0, nil, err
	}
	// TODO load from underlying arrays directly and correctly format in OpenGL
	// switch img.(type) {
	// case *image.Alpha:
	// case *image.Alpha16:
	// case *image.CMYK:
	// case *image.Gray:
	// case *image.Gray16:
	// case *image.NRGBA:
	// case *image.NRGBA64:
	// case *image.Paletted:
	// case *image.RGBA:
	// case *image.RGBA64:
	// case *image.YCbCr, *image.NYCbCrA, *image.Uniform:
	// 	// no Pix array
	// }
	width = img.Bounds().Dx()
	height = img.Bounds().Dy()
	data = make([]byte, 0, width*height*4)
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			col := color.NRGBAModel.Convert(img.At(i, j))
			nrgba := col.(color.NRGBA)
			r, g, b, a := nrgba.R, nrgba.G, nrgba.B, nrgba.A
			data = append(data, r, g, b, a)
		}
	}
	return width, height, data, nil
}

// ErrWriteFormat indicates that an unsupported image format was trying to be written to
const ErrWriteFormat log.ConstErr = "unsupported image format"

// WriteToFile writes the image data stored in the OpenGL texture to a file specified by fileName
func (iv *View) WriteToFile(fileName string) error {
	var texWidth, texHeight int32
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_WIDTH, &texWidth)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_HEIGHT, &texHeight)
	// TODO do this in batches to avoid memory limitations
	var data = make([]byte, texWidth*texHeight*4)
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	gl.BindTexture(gl.TEXTURE_2D, 0)

	img := image.NewNRGBA(image.Rect(0, 0, int(texWidth), int(texHeight)))
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
