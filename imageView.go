package main

import (
	"math"
	"os"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

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

// func (iv *ImageView) centerImage() {
// 	iv.canvas.X = int32(float64(iv.area.W)/2.0 - float64(iv.canvas.W)/2.0)
// 	iv.canvas.Y = int32(float64(iv.area.H)/2.0 - float64(iv.canvas.H)/2.0)
// }

// NewImageView returns a pointer to a new ImageView struct that implements UIComponent
func NewImageView(area *sdl.Rect, fileName string, comms chan<- imageComm) (*ImageView, error) {
	var err error
	var iv = &ImageView{}
	iv.area = area
	if err = iv.loadFromFile(fileName); err != nil {
		return nil, err
	}

	if iv.programID, err = CreateShaderProgram(vertexShaderSource, fragmentShaderSource); err != nil {
		return nil, err
	}
	screenSizeLoc := gl.GetUniformLocation(iv.programID, &[]byte("screenSize")[0])
	gl.UseProgram(iv.programID)
	gl.Uniform2f(screenSizeLoc, float32(iv.area.W), float32(iv.area.H))
	gl.UseProgram(0)

	iv.comms = comms
	iv.mult = 1.0
	// iv.sel = set.NewSet()

	return iv, nil
}

// Destroy frees all surfaces and textures in the ImageView
func (iv *ImageView) Destroy() {
	iv.surf.Free()
	// iv.selSurf.Free()
	// iv.tex.Destroy()
	// iv.selTex.Destroy()
	// iv.backTex.Destroy()
}

func (iv *ImageView) updateMousePos(x, y int32) {
	iv.mouseLoc.x = x
	iv.mouseLoc.y = y
	iv.mousePix.x = int32(float64(iv.mouseLoc.x-iv.canvas.X) / iv.mult)
	iv.mousePix.y = int32(float64(iv.mouseLoc.y-iv.canvas.Y) / iv.mult)
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

	gl.UseProgram(iv.programID)

	// copy pixels to texture
	// TODO find correct format
	format := int32(gl.RGBA)
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, iv.surf.W, iv.surf.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&iv.surf.Pixels()[0]))
	// this might fudge the pixels
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// render canvas rectangle
	sfw, sfh := float32(iv.surf.W), float32(iv.surf.H)
	square := []float32{
		0.0, 0.0, 0, 1, // top-left
		0.0, sfh, 0, 0, // bottom-left
		sfw, sfh, 1, 0, // bottom-right
		0.0, 0.0, 0, 1, // top-left
		sfw, sfh, 1, 0, // bottom-right
		sfw, 0.0, 1, 1, // top-right
	}
	// for i := 0; i < len(square); i += 4 {
	// 	sx, sy := square[i], square[i+1]
	// 	x, y := 2.0*sx/scw-1.0, 2.0*sy/sch-1.0
	// 	fmt.Printf("%v: %v,%v\n", i, x, y)
	// }
	vao := makeVao(square)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/4))
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	// unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (iv *ImageView) OnEnter() {}

// OnLeave is called when the cursor leaves the UIComponent's region
func (iv *ImageView) OnLeave() {
	iv.dragging = false
}

// OnMotion is called when the cursor moves within the UIComponent's region
func (iv *ImageView) OnMotion(evt *sdl.MouseMotionEvent) bool {
	// iv.updateMousePos(evt.X, evt.Y)
	// if !inBounds(iv.canvas, evt.X, evt.Y) {
	// 	return false
	// }
	// if evt.State == sdl.ButtonRMask() && iv.dragging {
	// 	iv.canvas.X += evt.X - iv.dragPoint.x
	// 	iv.canvas.Y += evt.Y - iv.dragPoint.y
	// 	iv.dragPoint.x = evt.X
	// 	iv.dragPoint.y = evt.Y
	// }
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
	// if iv.dragging {
	// 	return true
	// }
	// if evt.Y > 0 {
	// 	if int32(iv.mult*float64(iv.surf.W)*2.0) < iv.ctx.RendInfo.MaxTextureWidth && int32(iv.mult*float64(iv.surf.H)*2.0) < iv.ctx.RendInfo.MaxTextureHeight {
	// 		iv.zoomIn()
	// 	}
	// } else if evt.Y < 0 {
	// 	if int32(iv.mult*float64(iv.surf.W)/2.0) > 0 && int32(iv.mult*float64(iv.surf.H)/2.0) > 0 {
	// 		iv.zoomOut()
	// 	}
	// }
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
	iv.area.W = x
	iv.area.H += (y - iv.area.H)
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
