package main

import (
	set "github.com/kroppt/IntSet"
	"github.com/veandco/go-sdl2/sdl"
)

type zoomer struct {
	lastMult float64
	mult     float64
	origW    float64
	origH    float64
	maxW     int32
	maxH     int32
}

func (z *zoomer) In() {
	if int32(z.mult*z.origW*2.0) < z.maxW && int32(z.mult*z.origH*2.0) < z.maxH {
		z.mult *= 2
	}
}

func (z *zoomer) Out() {
	if int32(z.mult*z.origW/2.0) > 0 && int32(z.mult*z.origH/2.0) > 0 {
		z.mult /= 2
	}
}

func (z *zoomer) MultW() int32 {
	return int32(z.origW * z.mult)
}

func (z *zoomer) MultH() int32 {
	return int32(z.origH * z.mult)
}

func (z *zoomer) IsIn() bool {
	return z.lastMult < z.mult
}

func (z *zoomer) IsOut() bool {
	return z.lastMult > z.mult
}

func (z *zoomer) Update() {
	z.lastMult = z.mult
}

var _ UIComponent = UIComponent(&imageView{})

type imageView struct {
	area      *sdl.Rect
	canvas    *sdl.Rect
	mouseLoc  coord
	mousePix  coord
	dragging  bool
	dragPoint coord
	zoom      zoomer
	sel       set.Set
	surf      *sdl.Surface
	tex       *sdl.Texture
	selSurf   *sdl.Surface
	selTex    *sdl.Texture
}

func newImageView(area *sdl.Rect, fileName string, ctx *context) (*imageView, error) {
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
	var zoom = zoomer{
		1.0,
		1.0,
		float64(surf.W),
		float64(surf.H),
		ctx.RendInfo.MaxTextureWidth,
		ctx.RendInfo.MaxTextureHeight,
	}
	if err != nil {
		return nil, err
	}
	var canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: surf.W,
		H: surf.H,
	}
	return &imageView{
		area:    area,
		canvas:  canvas,
		surf:    surf,
		tex:     tex,
		zoom:    zoom,
		sel:     set.NewSet(),
		selSurf: selSurf,
		selTex:  selTex,
	}, nil
}

func (iv *imageView) updateMousePos(x, y int32) {
	iv.mouseLoc.x = x
	iv.mouseLoc.y = y
	iv.mousePix.x = int32(float64(iv.mouseLoc.x-iv.canvas.X) / iv.zoom.mult)
	iv.mousePix.y = int32(float64(iv.mouseLoc.y-iv.canvas.Y) / iv.zoom.mult)
}

func (iv *imageView) getBoundary() *sdl.Rect {
	return iv.area
}

func (iv *imageView) render(rend *sdl.Renderer) error {
	diffW := iv.zoom.MultW() - iv.canvas.W
	diffH := iv.zoom.MultH() - iv.canvas.H
	iv.canvas.W += diffW
	iv.canvas.H += diffH
	if iv.zoom.IsIn() {
		iv.canvas.X = 2*iv.canvas.X - iv.mouseLoc.x
		iv.canvas.Y = 2*iv.canvas.Y - iv.mouseLoc.y
	}
	if iv.zoom.IsOut() {
		iv.canvas.X = iv.canvas.X/2 + iv.mouseLoc.x/2
		iv.canvas.Y = iv.canvas.Y/2 + iv.mouseLoc.y/2
	}
	iv.zoom.Update()
	iv.sel.Range(func(n int) bool {
		y := int32(n) % iv.selSurf.W
		x := int32(n) - y*iv.selSurf.W
		setPixel(iv.selSurf, coord{x: x, y: y}, sdl.Color{R: 0, G: 0, B: 0, A: 128})
		return true
	})
	var err error
	if err = rend.SetViewport(iv.canvas); err != nil {
		return err
	}
	if err = copyToTexture(iv.tex, iv.surf.Pixels(), nil); err != nil {
		return err
	}
	if err = rend.Copy(iv.tex, nil, nil); err != nil {
		panic(err)
	}
	if err = copyToTexture(iv.selTex, iv.selSurf.Pixels(), nil); err != nil {
		return err
	}
	if err = rend.Copy(iv.selTex, nil, nil); err != nil {
		panic(err)
	}
	return nil
}

func (iv *imageView) onEnter(evt *sdl.MouseMotionEvent) bool {
	if !inBounds(iv.canvas, evt.X, evt.Y) {
		return false
	}
	return true
}

func (iv *imageView) onLeave(evt *sdl.MouseMotionEvent) bool {
	if !inBounds(iv.canvas, evt.X, evt.Y) {
		return false
	}
	return true
}

func (iv *imageView) onMotion(evt *sdl.MouseMotionEvent) bool {
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

func (iv *imageView) onScroll(evt *sdl.MouseWheelEvent) bool {
	if evt.Y > 0 {
		iv.zoom.In()
	} else if evt.Y < 0 {
		iv.zoom.Out()
	}
	return true
}

func (iv *imageView) onClick(evt *sdl.MouseButtonEvent) bool {
	iv.updateMousePos(evt.X, evt.Y)
	if evt.Button == sdl.BUTTON_RIGHT {
		if evt.State == sdl.PRESSED {
			iv.dragging = true
		} else if evt.State == sdl.RELEASED {
			iv.dragging = false
		}
		iv.dragPoint.x = evt.X
		iv.dragPoint.y = evt.Y
	}
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED && inBounds(iv.canvas, evt.X, evt.Y) {
		i := int(iv.surf.W*iv.mousePix.y + iv.mousePix.x)
		if !iv.sel.Contains(i) {
			iv.sel.Add(i)
		}
	}
	return true
}
