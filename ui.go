package main

import (
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

// UIComponent says what functions a UIComponent must implement
type UIComponent interface {
	GetBoundary() *sdl.Rect
	Render(*sdl.Renderer) error
	Destroy()
	OnEnter(*sdl.MouseMotionEvent) bool
	OnLeave(*sdl.MouseMotionEvent) bool
	OnMotion(*sdl.MouseMotionEvent) bool
	OnScroll(*sdl.MouseWheelEvent) bool
	OnClick(*sdl.MouseButtonEvent) bool
}

type context struct {
	Win      *sdl.Window
	Rend     *sdl.Renderer
	RendInfo *sdl.RendererInfo
	Conf     *config
}

// AlignV is used for the positioning of elements vertically
type AlignV int

const (
	// AlignBelow puts the top side at the y coordinate
	AlignBelow AlignV = iota - 1
	// AlignMiddle puts the top and bottom sides equidistant from the middle
	AlignMiddle
	// AlignAbove puts the bottom side on the y coordinate
	AlignAbove
)

// AlignH is used for the positioning of elements horizontally
type AlignH int

const (
	// AlignLeft puts the left side on the x coordinate
	AlignLeft AlignH = iota - 1
	//AlignCenter puts the left and right sides equidistant from the center
	AlignCenter
	// AlignRight puts the right side at the x coordinate
	AlignRight
)

// Align holds vertical and horizontal alignments
type Align struct {
	v AlignV
	h AlignH
}

type coord struct {
	x int32
	y int32
}

func createSolidColorTexture(rend *sdl.Renderer, w int32, h int32, r uint8, g uint8, b uint8, a uint8) (*sdl.Texture, error) {
	var surf *sdl.Surface
	var err error
	if surf, err = sdl.CreateRGBSurfaceWithFormat(0, w, h, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
		return nil, err
	}
	if err = surf.FillRect(nil, mapRGBA(surf.Format, r, g, b, a)); err != nil {
		return nil, err
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTextureFromSurface(surf); err != nil {
		return nil, err
	}
	surf.Free()
	return tex, nil
}

func renderText(conf *config, rend *sdl.Renderer, text string, pos coord, align Align) error {
	col := sdl.Color{
		R: 255,
		G: 255,
		B: 255,
		A: 0,
	}
	var surf *sdl.Surface
	var err error
	if surf, err = conf.font.RenderUTF8Blended(text, col); err != nil {
		return err
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTexture(surf.Format.Format, sdl.TEXTUREACCESS_STREAMING, surf.W, int32(conf.fontSize)); err != nil {
		surf.Free()
		return err
	}
	if err = tex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		surf.Free()
		tex.Destroy()
		return err
	}
	sliceOffset := surf.Pitch * (surf.H - conf.fontSize)
	copyToTexture(tex, surf.Pixels()[sliceOffset:], nil)

	w2 := int32(float64(surf.W) / 2.0)
	h2 := int32(float64(conf.fontSize) / 2.0)
	offx := -w2 - int32(align.h)*int32(w2)
	offy := -h2 - int32(align.v)*int32(h2)
	var rect = &sdl.Rect{
		X: pos.x + offx,
		Y: pos.y + offy,
		W: int32(surf.W),
		H: int32(conf.fontSize),
	}
	err = rend.Copy(tex, nil, rect)
	surf.Free()
	tex.Destroy()
	return err
}

func mapRGBA(form *sdl.PixelFormat, r, g, b, a uint8) uint32 {
	ur := uint32(r)
	ur |= ur<<8 | ur<<16 | ur<<24
	ug := uint32(g)
	ug |= ug<<8 | ug<<16 | ug<<24
	ub := uint32(b)
	ub |= ub<<8 | ub<<16 | ub<<24
	ua := uint32(a)
	ua |= ua<<8 | ua<<16 | ua<<24
	return form.Rmask&ur |
		form.Gmask&ug |
		form.Bmask&ub |
		form.Amask&ua
}

func setPixel(surf *sdl.Surface, p coord, c sdl.Color) {
	d := mapRGBA(surf.Format, c.R, c.G, c.B, c.A)
	bs := []byte{byte(d), byte(d >> 8), byte(d >> 16), byte(d >> 24)}
	i := int32(surf.BytesPerPixel())*p.x + surf.Pitch*p.y
	copy(surf.Pixels()[i:], bs)
}

func copyToTexture(tex *sdl.Texture, pixels []byte, region *sdl.Rect) error {
	var bytes []byte
	var err error
	bytes, _, err = tex.Lock(region)
	copy(bytes, pixels)
	tex.Unlock()
	return err
}

func loadImage(rend *sdl.Renderer, fileName string) (*sdl.Surface, *sdl.Texture, error) {
	var tex *sdl.Texture
	var surf *sdl.Surface
	var err error
	if surf, err = img.Load(fileName); err != nil {
		return nil, nil, err
	}
	if tex, err = rend.CreateTexture(surf.Format.Format, sdl.TEXTUREACCESS_STREAMING, surf.W, surf.H); err != nil {
		return nil, nil, err
	}
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return nil, nil, err
	}
	if err = copyToTexture(tex, surf.Pixels(), nil); err != nil {
		return nil, nil, err
	}
	return surf, tex, err
}
