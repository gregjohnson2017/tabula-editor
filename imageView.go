package main

import (
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&ImageView{})

// ImageView defines an interactable image viewing pane
type ImageView struct {
	area      *sdl.Rect
	canvas    *sdl.Rect
	surf      *sdl.Surface
	cfg       *config
	mouseLoc  coord
	mousePix  coord
	dragPoint coord
	dragging  bool
	mult      float64
	programID uint32
	textureID uint32
	glSquare  []float32
	vaoID     uint32
	vboID     uint32
	comms     chan<- imageComm
	fileName  string
	fullPath  string
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
	iv.surf = surf
	iv.canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: surf.W,
		H: surf.H,
	}
	iv.centerImage()

	parts := strings.Split(fileName, string(os.PathSeparator))
	iv.fileName = parts[len(parts)-1]
	iv.fullPath = fileName
	return nil
}

// NewImageView returns a pointer to a new ImageView struct that implements UIComponent
func NewImageView(area *sdl.Rect, fileName string, comms chan<- imageComm, cfg *config) (*ImageView, error) {
	var err error
	var iv = &ImageView{}
	iv.cfg = cfg
	iv.area = area
	iv.comms = comms
	iv.mult = 1.0
	if err = iv.loadFromFile(fileName); err != nil {
		return nil, err
	}

	if iv.programID, err = CreateShaderProgram(vertexShaderSource, checkerShaderFragment); err != nil {
		return nil, err
	}

	uniformID := gl.GetUniformLocation(iv.programID, &[]byte("area\x00")[0])
	gl.UseProgram(iv.programID)
	gl.Uniform2f(uniformID, float32(iv.area.W), float32(iv.area.H))
	gl.UseProgram(0)

	format := int32(gl.RGBA)
	gl.GenTextures(1, &iv.textureID)
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	// copy pixels to texture
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, iv.surf.W, iv.surf.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&iv.surf.Pixels()[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.GenBuffers(1, &iv.vboID)
	gl.GenVertexArrays(1, &iv.vaoID)
	configureVAO(iv.vaoID, iv.vboID, []int32{2, 2})

	return iv, nil
}

// Destroy frees all assets acquired by the UIComponent
func (iv *ImageView) Destroy() {
	iv.surf.Free()
	gl.DeleteTextures(1, &iv.textureID)
	gl.DeleteBuffers(1, &iv.vboID)
	gl.DeleteVertexArrays(1, &iv.vaoID)
}

// GetBoundary returns the clickable region of the UIComponent
func (iv *ImageView) GetBoundary() *sdl.Rect {
	return iv.area
}

// Render draws the UIComponent
func (iv *ImageView) Render() error {
	go func() {
		iv.comms <- imageComm{fileName: iv.fileName, mousePix: iv.mousePix, mult: iv.mult}
	}()

	// update buffered data
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

	gl.Viewport(iv.area.X, iv.area.Y+iv.cfg.bottomBarHeight, iv.area.W, iv.area.H)
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

	gl.UseProgram(0)
	return nil
}

func (iv *ImageView) zoomIn() {
	iv.mult *= 2.0
	iv.canvas.W = int32(float64(iv.surf.W) * iv.mult)
	iv.canvas.H = int32(float64(iv.surf.H) * iv.mult)
	iv.canvas.X = 2*iv.canvas.X - int32(math.Round(float64(iv.area.W)/2.0)) //iv.mouseLoc.x
	iv.canvas.Y = 2*iv.canvas.Y - int32(math.Round(float64(iv.area.H)/2.0)) //iv.mouseLoc.y
}

func (iv *ImageView) zoomOut() {
	iv.mult /= 2.0
	iv.canvas.W = int32(float64(iv.surf.W) * iv.mult)
	iv.canvas.H = int32(float64(iv.surf.H) * iv.mult)
	iv.canvas.X = int32(math.Round(float64(iv.canvas.X)/2.0 + float64(iv.area.W)/4.0)) //iv.mouseLoc.x/2
	iv.canvas.Y = int32(math.Round(float64(iv.canvas.Y)/2.0 + float64(iv.area.H)/4.0)) //iv.mouseLoc.y/2
}

func (iv *ImageView) centerImage() {
	iv.canvas.X = int32(float64(iv.area.W)/2.0 - float64(iv.canvas.W)/2.0)
	iv.canvas.Y = int32(float64(iv.area.H)/2.0 - float64(iv.canvas.H)/2.0)
}

func (iv *ImageView) setPixel(x, y int32, color []byte) {
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	gl.TextureSubImage2D(iv.textureID, 0, x, y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&color[0]))
}

func (iv *ImageView) updateMousePos(x, y int32) {
	iv.mouseLoc.x = x
	iv.mouseLoc.y = y
	iv.mousePix.x = int32(float64(iv.mouseLoc.x-iv.canvas.X) / iv.mult)
	iv.mousePix.y = int32(float64(iv.mouseLoc.y-iv.canvas.Y) / iv.mult)
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
	if evt.State == sdl.ButtonLMask() && inBounds(iv.canvas, evt.X, evt.Y) {
		iv.setPixel(iv.mousePix.x, iv.mousePix.y, []byte{0x00, 0x00, 0x00, 0x00})
	}
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (iv *ImageView) OnScroll(evt *sdl.MouseWheelEvent) bool {
	if iv.dragging {
		return true
	}
	if evt.Y > 0 {
		if int32(iv.mult*float64(iv.surf.W)*2.0) > iv.canvas.W && int32(iv.mult*float64(iv.surf.H)*2.0) > iv.canvas.H && iv.mult < 256 {
			iv.zoomIn()
		}
	} else if evt.Y < 0 {
		if int32(iv.mult*float64(iv.surf.W)/2.0) > 0 && int32(iv.mult*float64(iv.surf.H)/2.0) > 0 {
			iv.zoomOut()
		}
	}
	return true
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
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		iv.setPixel(iv.mousePix.x, iv.mousePix.y, []byte{0x00, 0x00, 0x00, 0x00})
	}
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
