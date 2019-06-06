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
	programID uint32
	mouseLoc  coord
	mousePix  coord
	dragging  bool
	dragPoint coord
	mult      float64
	glSquare  []float32
	vaoID     uint32
	vboID     uint32
	textureID uint32
	// screenSizeID int32

	// sel       set.Set
	surf *sdl.Surface
	// tex       *sdl.Texture
	// selSurf   *sdl.Surface
	// selTex    *sdl.Texture
	// backTex   *sdl.Texture
	comms    chan<- imageComm
	fileName string
	fullPath string
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
	// var selSurf *sdl.Surface
	// if selSurf, err = sdl.CreateRGBSurfaceWithFormat(0, surf.W, surf.H, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
	// 	return err
	// }
	// if err = selSurf.FillRect(nil, mapRGBA(selSurf.Format, 0, 0, 0, 0)); err != nil {
	// 	return err
	// }
	// var selTex *sdl.Texture
	// if selTex, err = iv.ctx.Rend.CreateTexture(selSurf.Format.Format, sdl.TEXTUREACCESS_STREAMING, selSurf.W, selSurf.H); err != nil {
	// 	return err
	// }
	// if err = selTex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
	// 	return err
	// }
	// if err = iv.createBackTex(); err != nil {
	// 	return err
	// }
	var canvas = &sdl.Rect{
		X: int32(float64(iv.area.W)/2.0 - float64(surf.W)/2.0),
		Y: int32(float64(iv.area.H)/2.0 - float64(surf.H)/2.0),
		W: surf.W,
		H: surf.H,
	}
	iv.surf = surf
	// iv.selSurf = selSurf
	// iv.tex = tex
	// iv.selTex = selTex
	iv.canvas = canvas
	parts := strings.Split(fileName, string(os.PathSeparator))
	iv.fileName = parts[len(parts)-1]
	iv.fullPath = fileName
	return nil
}

// func (iv *ImageView) createBackTex() error {
// 	var backSurf *sdl.Surface
// 	var err error
// 	if backSurf, err = sdl.CreateRGBSurfaceWithFormat(0, iv.area.W, iv.area.H, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
// 		return err
// 	}
// 	light := mapRGBA(backSurf.Format, 0xEE, 0xEE, 0xEE, 0xFF)
// 	backSurf.FillRect(nil, light)
// 	rects := []sdl.Rect{}
// 	sqsize := int32(8)
// 	for i := int32(0); i < backSurf.W; i += 2 * sqsize {
// 		for j := int32(0); j < backSurf.H; j += sqsize {
// 			off := ((j/sqsize + 1) % 2) * sqsize
// 			r := sdl.Rect{X: i + off, Y: j, W: sqsize, H: sqsize}
// 			rects = append(rects, r)
// 		}
// 	}
// 	dark := mapRGBA(backSurf.Format, 0x99, 0x99, 0x99, 0xFF)
// 	backSurf.FillRects(rects, dark)
// 	var backTex *sdl.Texture
// 	if backTex, err = iv.ctx.Rend.CreateTextureFromSurface(backSurf); err != nil {
// 		return err
// 	}
// 	backSurf.Free()
// 	iv.backTex = backTex
// 	return nil
// }

// NewImageView returns a pointer to a new ImageView struct that implements UIComponent
func NewImageView(area *sdl.Rect, fileName string, comms chan<- imageComm) (*ImageView, error) {
	var err error
	var iv = &ImageView{}
	iv.area = area
	iv.comms = comms
	iv.mult = 1.0
	// iv.sel = set.NewSet()
	if err = iv.loadFromFile(fileName); err != nil {
		return nil, err
	}

	if iv.programID, err = CreateShaderProgram(vertexShaderSource, fragmentShaderSource); err != nil {
		return nil, err
	}

	// iv.screenSizeID = gl.GetUniformLocation(iv.programID, &[]byte("screenSize")[0])
	// gl.UseProgram(iv.programID)
	// gl.Uniform2f(iv.screenSizeID, float32(iv.area.W), float32(iv.area.H))
	// gl.UseProgram(0)

	// TODO find correct format
	format := int32(gl.RGBA)
	gl.GenTextures(1, &iv.textureID)
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	// copy pixels to texture
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, iv.surf.W, iv.surf.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&iv.surf.Pixels()[0]))
	// https://www.khronos.org/registry/OpenGL-Refpages/es2.0/xhtml/glTexParameter.xml
	// TODO pick right minify filter param
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	iv.glSquare = []float32{
		-1.0, -1.0, 0.0, 1.0, // top-left
		-1.0, +1.0, 0.0, 0.0, // bottom-left
		+1.0, +1.0, 1.0, 0.0, // bottom-right
		-1.0, -1.0, 0.0, 1.0, // top-left
		+1.0, +1.0, 1.0, 0.0, // bottom-right
		+1.0, -1.0, 1.0, 1.0, // top-right
	}
	iv.vaoID, iv.vboID = makeVAO(iv.glSquare)

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
	// go func() {
	// 	iv.comms <- imageComm{fileName: iv.fileName, mousePix: iv.mousePix, mult: iv.mult}
	// }()
	// iv.sel.Range(func(n int) bool {
	// 	y := int32(n) % iv.selSurf.W
	// 	x := int32(n) - y*iv.selSurf.W
	// 	setPixel(iv.selSurf, coord{x: x, y: y}, sdl.Color{R: 0, G: 0, B: 0, A: 128})
	// 	return true
	// })
	// var err error
	r := &sdl.Rect{X: iv.canvas.X, Y: iv.canvas.Y, W: iv.canvas.W, H: iv.canvas.H}
	if r.X < 0 {
		r.W += r.X
		r.X = 0
	}
	if r.Y < 0 {
		r.H += r.Y
		r.Y = 0
	}
	if r.X+r.W > iv.area.W {
		r.W = iv.area.W - r.X
	}
	if r.Y+r.H > iv.area.H {
		r.H = iv.area.H - r.Y
	}

	gl.Viewport(iv.canvas.X, iv.canvas.Y, iv.canvas.W, iv.canvas.H)
	gl.UseProgram(iv.programID)

	gl.BindVertexArray(iv.vaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.BindTexture(gl.TEXTURE_2D, iv.textureID)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(iv.glSquare)/4))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindVertexArray(0)

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
	iv.updateMousePos(evt.X, evt.Y)
	if !iv.dragging && !inBounds(iv.canvas, evt.X, evt.Y) {
		return false
	}
	if evt.State == sdl.ButtonRMask() && iv.dragging {
		iv.canvas.X += evt.X - iv.dragPoint.x
		iv.canvas.Y += evt.Y - iv.dragPoint.y
		iv.dragPoint.x = evt.X
		iv.dragPoint.y = evt.Y
	}
	// if evt.State == sdl.ButtonLMask() && inBounds(iv.canvas, evt.X, evt.Y) {
	// 	i := int(iv.surf.W*iv.mousePix.y + iv.mousePix.x)
	// 	if !iv.sel.Contains(i) {
	// 		iv.sel.Add(i)
	// 	}
	// }
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (iv *ImageView) OnScroll(evt *sdl.MouseWheelEvent) bool {
	if iv.dragging {
		return true
	}
	if evt.Y > 0 {
		if int32(iv.mult*float64(iv.surf.W)*2.0) > iv.canvas.W && int32(iv.mult*float64(iv.surf.H)*2.0) > iv.canvas.H {
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
	// if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
	// 	i := int(iv.surf.W*iv.mousePix.y + iv.mousePix.x)
	// 	if !iv.sel.Contains(i) {
	// 		iv.sel.Add(i)
	// 	}
	// }
	return true
}

// OnResize is called when the user resizes the window
func (iv *ImageView) OnResize(x, y int32) {
	// var err error
	iv.area.W += x
	iv.area.H += y
	// might need to upload screen size for alpha background calculation
	// gl.UseProgram(iv.programID)
	// gl.Uniform2f(iv.screenSizeID, float32(iv.area.W), float32(iv.area.H))
	// gl.UseProgram(0)
	iv.centerImage()
	// if err = iv.backTex.Destroy(); err != nil {
	// 	panic(err)
	// }
	// if err = iv.createBackTex(); err != nil {
	// 	panic(err)
	// }
}

// String returns the name of the component type
func (iv *ImageView) String() string {
	return "ImageView"
}
