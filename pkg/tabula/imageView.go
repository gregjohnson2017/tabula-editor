package tabula

import (
	"fmt"
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"

	set "github.com/kroppt/Int32Set"
)

var _ UIComponent = UIComponent(&ImageView{})

// ImageView defines an interactable image viewing pane
type ImageView struct {
	area           *sdl.Rect
	canvas         *sdl.Rect
	origW, origH   int32
	cfg            *Config
	mouseLoc       coord
	mousePix       coord
	dragPoint      coord
	dragging       bool
	mult           float64
	programID      uint32
	selProgramID   uint32
	textureID      uint32
	glSquare       []float32
	vaoID, vboID   uint32
	selVao, selVbo uint32
	bbComms        chan<- imageComm
	fileName       string
	fullPath       string
	toolComms      <-chan ImageTool
	activeTool     ImageTool
	selection      set.Set
}

type imageComm struct {
	fileName string
	mousePix coord
	mult     float64
}

func (iv *ImageView) loadFromFile(fileName string) error {
	surf, err := loadImage(fileName)
	if err != nil {
		return err
	}
	iv.origW = surf.W
	iv.origH = surf.H
	uniformID := gl.GetUniformLocation(iv.selProgramID, &[]byte("origDims\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform2f(uniformID, float32(iv.origW), float32(iv.origH))
	gl.UseProgram(0)

	format := int32(gl.RGBA)
	gl.DeleteTextures(1, &iv.textureID)
	gl.GenTextures(1, &iv.textureID)
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	// copy pixels to texture
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, surf.W, surf.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&surf.Pixels()[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	surf.Free()

	iv.canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: surf.W,
		H: surf.H,
	}
	iv.centerImage()
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

// NewImageView returns a pointer to a new ImageView struct that implements UIComponent
func NewImageView(area *sdl.Rect, fileName string, bbComms chan<- imageComm, toolComms <-chan ImageTool, cfg *Config) (*ImageView, error) {
	var err error
	var iv = &ImageView{}
	iv.cfg = cfg
	iv.area = area
	iv.bbComms = bbComms
	iv.toolComms = toolComms
	if err = iv.loadFromFile(fileName); err != nil {
		return nil, err
	}

	if iv.programID, err = CreateShaderProgram(vertexShaderSource, checkerShaderFragment); err != nil {
		return nil, err
	}

	if iv.selProgramID, err = CreateShaderProgram(outlineVsh, outlineFsh); err != nil {
		return nil, err
	}

	uniformID := gl.GetUniformLocation(iv.programID, &[]byte("area\x00")[0])
	gl.UseProgram(iv.programID)
	gl.Uniform2f(uniformID, float32(iv.area.W), float32(iv.area.H))
	gl.UseProgram(0)

	uniformID = gl.GetUniformLocation(iv.selProgramID, &[]byte("origDims\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform2f(uniformID, float32(iv.origW), float32(iv.origH))
	gl.UseProgram(0)

	uniformID = gl.GetUniformLocation(iv.selProgramID, &[]byte("mult\x00")[0])
	gl.UseProgram(iv.selProgramID)
	gl.Uniform1f(uniformID, float32(iv.mult))
	gl.UseProgram(0)

	gl.GenBuffers(1, &iv.vboID)
	gl.GenVertexArrays(1, &iv.vaoID)
	configureVAO(iv.vaoID, iv.vboID, []int32{2, 2})

	gl.GenBuffers(1, &iv.selVbo)
	gl.GenVertexArrays(1, &iv.selVao)
	configureVAO(iv.selVao, iv.selVbo, []int32{2})

	iv.selection = set.NewSet()
	iv.activeTool = EmptyTool{}

	return iv, nil
}

// Destroy frees all assets acquired by the UIComponent
func (iv *ImageView) Destroy() {
	gl.DeleteTextures(1, &iv.textureID)
	gl.DeleteBuffers(1, &iv.vboID)
	gl.DeleteVertexArrays(1, &iv.vaoID)
	gl.DeleteBuffers(1, &iv.selVbo)
	gl.DeleteVertexArrays(1, &iv.selVao)
	gl.DeleteProgram(iv.programID)
	gl.DeleteProgram(iv.selProgramID)
}

// InBoundary returns whether a point is in this UIComponent's bounds
func (iv *ImageView) InBoundary(pt sdl.Point) bool {
	return inBounds(iv.area, pt.X, pt.Y)
}

// GetBoundary returns the clickable region of the UIComponent
func (iv *ImageView) GetBoundary() *sdl.Rect {
	return iv.area
}

// Render draws the UIComponent
func (iv *ImageView) Render() {
	go func() {
		iv.bbComms <- imageComm{fileName: iv.fileName, mousePix: iv.mousePix, mult: iv.mult}
	}()

	// TODO optimize this (ex: move elsewhere, update as changes come in, use a better algorithm)
	// make array of 2d-vertex pairs representing texel selection outlines
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
		iv.activeTool = tool
	default:
	}
}

func (iv *ImageView) zoomIn() {
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

func (iv *ImageView) zoomOut() {
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

func (iv *ImageView) centerImage() {
	iv.canvas.X = int32(float64(iv.area.W)/2.0 - float64(iv.canvas.W)/2.0)
	iv.canvas.Y = int32(float64(iv.area.H)/2.0 - float64(iv.canvas.H)/2.0)
}

func (iv *ImageView) setPixel(x, y int32, color []byte) {
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	gl.TextureSubImage2D(iv.textureID, 0, x, y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&color[0]))
	// TODO update mipmap textures only when needed
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (iv *ImageView) updateMousePos(x, y int32) {
	iv.mouseLoc.x = x
	iv.mouseLoc.y = y
	relx := float64(iv.mouseLoc.x - iv.canvas.X)
	rely := float64(iv.mouseLoc.y - iv.canvas.Y)
	iv.mousePix.x = int32(math.Floor(relx / iv.mult))
	iv.mousePix.y = int32(math.Floor(rely / iv.mult))
}

// OnEnter is called when the cursor enters the UIComponent's region
func (iv *ImageView) OnEnter() {}

// OnLeave is called when the cursor leaves the UIComponent's region
func (iv *ImageView) OnLeave() {
	iv.dragging = false
}

// OnMotion is called when the cursor moves within the UIComponent's region
func (iv *ImageView) OnMotion(evt *sdl.MouseMotionEvent) bool {
	if !iv.dragging {
		iv.updateMousePos(evt.X, evt.Y)
	}
	if !iv.dragging && !inBounds(iv.canvas, evt.X, evt.Y) {
		return false
	}
	if evt.State == sdl.ButtonRMask() && iv.dragging {
		iv.canvas.X += evt.X - iv.dragPoint.x
		iv.canvas.Y += evt.Y - iv.dragPoint.y
		iv.dragPoint.x = evt.X
		iv.dragPoint.y = evt.Y
	}
	iv.activeTool.OnMotion(evt, iv)
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (iv *ImageView) OnScroll(evt *sdl.MouseWheelEvent) bool {
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

// SelectPixel adds the given x, y pixel to the
func (iv *ImageView) SelectPixel(x, y int32) error {
	if x < 0 || y < 0 || x > iv.area.W || y > iv.area.H {
		return fmt.Errorf("x and y coordinates (%v, %v) are out of range", x, y)
	}
	iv.selection.Add(iv.mousePix.x + iv.mousePix.y*iv.origW)
	return nil
}

// OnClick is called when the user clicks within the UIComponent's region
func (iv *ImageView) OnClick(evt *sdl.MouseButtonEvent) bool {
	iv.updateMousePos(evt.X, evt.Y)
	if !inBounds(iv.canvas, evt.X, evt.Y) {
		return true
	}
	if evt.Button == sdl.BUTTON_RIGHT {
		if evt.State == sdl.PRESSED {
			iv.dragging = true
		} else if evt.State == sdl.RELEASED {
			iv.dragging = false
		}
		iv.dragPoint.x = evt.X
		iv.dragPoint.y = evt.Y
	}
	iv.activeTool.OnClick(evt, iv)
	return true
}

// OnResize is called when the user resizes the window
func (iv *ImageView) OnResize(x, y int32) {
	iv.area.W += x
	iv.area.H += y

	uniformID := gl.GetUniformLocation(iv.programID, &[]byte("area\x00")[0])
	gl.UseProgram(iv.programID)
	gl.Uniform2f(uniformID, float32(iv.area.W), float32(iv.area.H))
	gl.UseProgram(0)

	iv.centerImage()
}

// String returns the name of the component type
func (iv *ImageView) String() string {
	return "ImageView"
}
