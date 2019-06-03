package main

import (
	"math"

	set "github.com/kroppt/IntSet"
	"github.com/veandco/go-sdl2/sdl"
)

func (iv *ImageView) zoomIn() {
	iv.mult *= 2.0
	diffW := int32(float64(iv.surf.W)*iv.mult) - iv.canvas.W
	diffH := int32(float64(iv.surf.H)*iv.mult) - iv.canvas.H
	iv.canvas.W += diffW
	iv.canvas.H += diffH
	iv.canvas.X = 2*iv.canvas.X - int32(math.Round(float64(iv.area.W)/2.0)) //iv.mouseLoc.x
	iv.canvas.Y = 2*iv.canvas.Y - int32(math.Round(float64(iv.area.H)/2.0)) //iv.mouseLoc.y
}

func (iv *ImageView) zoomOut() {
	iv.mult /= 2.0
	diffW := int32(float64(iv.surf.W)*iv.mult) - iv.canvas.W
	diffH := int32(float64(iv.surf.H)*iv.mult) - iv.canvas.H
	iv.canvas.W += diffW
	iv.canvas.H += diffH
	iv.canvas.X = int32(math.Round(float64(iv.canvas.X)/2.0 + float64(iv.area.W)/4.0)) //iv.mouseLoc.x/2
	iv.canvas.Y = int32(math.Round(float64(iv.canvas.Y)/2.0 + float64(iv.area.H)/4.0)) //iv.mouseLoc.y/2
}

var _ UIComponent = UIComponent(&ImageView{})

// ImageView defines an interactable image viewing pane
type ImageView struct {
	area      *sdl.Rect
	canvas    *sdl.Rect
	mouseLoc  coord
	mousePix  coord
	dragging  bool
	dragPoint coord
	mult      float64
	sel       set.Set
	surf      *sdl.Surface
	tex       *sdl.Texture
	selSurf   *sdl.Surface
	selTex    *sdl.Texture
	backTex   *sdl.Texture
	comms     chan<- imageComm
	ctx       *context
	fileName  string
}

type imageComm struct {
	fileName string
	mousePix coord
	mult     float64
}

// NewImageView returns a pointer to a new ImageView struct that implements UIComponent
func NewImageView(area *sdl.Rect, fileName string, comms chan<- imageComm, ctx *context) (*ImageView, error) {
	surf, tex, err := loadImage(ctx.Rend, fileName)
	if err != nil {
		return nil, err
	}
	var selSurf *sdl.Surface
	if selSurf, err = sdl.CreateRGBSurfaceWithFormat(0, surf.W, surf.H, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
		return nil, err
	}
	if err = selSurf.FillRect(nil, mapRGBA(selSurf.Format, 0, 0, 0, 0)); err != nil {
		return nil, err
	}
	var selTex *sdl.Texture
	if selTex, err = ctx.Rend.CreateTexture(selSurf.Format.Format, sdl.TEXTUREACCESS_STREAMING, selSurf.W, selSurf.H); err != nil {
		return nil, err
	}
	if err = selTex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nil, err
	}
	var backSurf *sdl.Surface
	if backSurf, err = sdl.CreateRGBSurfaceWithFormat(0, area.W, area.H, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
		return nil, err
	}
	light := mapRGBA(backSurf.Format, 0xEE, 0xEE, 0xEE, 0xFF)
	backSurf.FillRect(nil, light)
	rects := []sdl.Rect{}
	sqsize := int32(8)
	for i := int32(0); i < backSurf.W; i += 2 * sqsize {
		for j := int32(0); j < backSurf.H; j += sqsize {
			off := ((j/sqsize + 1) % 2) * sqsize
			r := sdl.Rect{X: i + off, Y: j, W: sqsize, H: sqsize}
			rects = append(rects, r)
		}
	}
	dark := mapRGBA(backSurf.Format, 0x99, 0x99, 0x99, 0xFF)
	backSurf.FillRects(rects, dark)
	var backTex *sdl.Texture
	if backTex, err = ctx.Rend.CreateTextureFromSurface(backSurf); err != nil {
		return nil, err
	}
	backSurf.Free()
	var canvas = &sdl.Rect{
		X: int32(float64(area.W)/2.0 - float64(surf.W)/2.0),
		Y: int32(float64(area.H)/2.0 - float64(surf.H)/2.0),
		W: surf.W,
		H: surf.H,
	}

	return &ImageView{
		area:     area,
		canvas:   canvas,
		surf:     surf,
		tex:      tex,
		mult:     1.0,
		sel:      set.NewSet(),
		selSurf:  selSurf,
		selTex:   selTex,
		backTex:  backTex,
		comms:    comms,
		ctx:      ctx,
		fileName: fileName,
	}, nil
}

// Destroy frees all surfaces and textures in the ImageView
func (iv *ImageView) Destroy() {
	iv.surf.Free()
	iv.selSurf.Free()
	iv.tex.Destroy()
	iv.selTex.Destroy()
	iv.backTex.Destroy()
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
	go func() {
		iv.comms <- imageComm{fileName: iv.fileName, mousePix: iv.mousePix, mult: iv.mult}
	}()
	iv.sel.Range(func(n int) bool {
		y := int32(n) % iv.selSurf.W
		x := int32(n) - y*iv.selSurf.W
		setPixel(iv.selSurf, coord{x: x, y: y}, sdl.Color{R: 0, G: 0, B: 0, A: 128})
		return true
	})
	var err error
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
	if err = iv.ctx.Rend.SetViewport(r); err != nil {
		return err
	}
	if err = iv.ctx.Rend.Copy(iv.backTex, r, nil); err != nil {
		return err
	}
	if err = iv.ctx.Rend.SetViewport(iv.canvas); err != nil {
		return err
	}
	if err = copyToTexture(iv.tex, iv.surf.Pixels(), nil); err != nil {
		return err
	}
	if err = iv.ctx.Rend.Copy(iv.tex, nil, nil); err != nil {
		return err
	}
	if err = copyToTexture(iv.selTex, iv.selSurf.Pixels(), nil); err != nil {
		return err
	}
	if err = iv.ctx.Rend.Copy(iv.selTex, nil, nil); err != nil {
		return err
	}
	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (iv *ImageView) OnEnter() {}

// OnLeave is called when the cursor leaves the UIComponent's region
func (iv *ImageView) OnLeave() {}

// OnMotion is called when the cursor moves within the UIComponent's region
func (iv *ImageView) OnMotion(evt *sdl.MouseMotionEvent) bool {
	iv.updateMousePos(evt.X, evt.Y)
	if !inBounds(iv.canvas, evt.X, evt.Y) {
		return false
	}
	if evt.State == sdl.ButtonRMask() && iv.dragging {
		iv.canvas.X += evt.X - iv.dragPoint.x
		iv.canvas.Y += evt.Y - iv.dragPoint.y
		iv.dragPoint.x = evt.X
		iv.dragPoint.y = evt.Y
	}
	if evt.State == sdl.ButtonLMask() && inBounds(iv.canvas, evt.X, evt.Y) {
		i := int(iv.surf.W*iv.mousePix.y + iv.mousePix.x)
		if !iv.sel.Contains(i) {
			iv.sel.Add(i)
		}
	}
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (iv *ImageView) OnScroll(evt *sdl.MouseWheelEvent) bool {
	if iv.dragging {
		return true
	}
	if evt.Y > 0 {
		if int32(iv.mult*float64(iv.surf.W)*2.0) < iv.ctx.RendInfo.MaxTextureWidth && int32(iv.mult*float64(iv.surf.H)*2.0) < iv.ctx.RendInfo.MaxTextureHeight {
			iv.zoomIn()
		}
	} else if evt.Y < 0 {
		if int32(iv.mult*float64(iv.surf.W)/2.0) > 0 && int32(iv.mult*float64(iv.surf.H)/2.0) > 0 && iv.mult > 0.25 {
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

		i := int(iv.surf.W*iv.mousePix.y + iv.mousePix.x)
		if !iv.sel.Contains(i) {
			iv.sel.Add(i)
		}
	}
	return true
}
