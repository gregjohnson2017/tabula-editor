package main

import (
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var _ UIComponent = UIComponent(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area     *sdl.Rect
	tex      *sdl.Texture
	comms    <-chan imageComm
	color    *sdl.Color
	mousePix coord
	ctx      *context
	font     *ttf.Font
	fontSize int32
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
// the background color defaults to grey (0x808080FF)
func NewBottomBar(area *sdl.Rect, comms <-chan imageComm, ctx *context, fontName string, fontSize int32) (*BottomBar, error) {
	color := &sdl.Color{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}
	var err error
	var bottomBarTex *sdl.Texture
	if bottomBarTex, err = createSolidColorTexture(ctx.Rend, area.W, area.H, color.R, color.G, color.B, color.A); err != nil {
		return nil, err
	}
	var font *ttf.Font
	if font, err = ttf.OpenFont(fontName, int(fontSize)); err != nil {
		return nil, err
	}
	return &BottomBar{
		area:     area,
		tex:      bottomBarTex,
		comms:    comms,
		ctx:      ctx,
		color:    color,
		font:     font,
		fontSize: fontSize,
	}, nil
}

// SetBackgroundColor sets the color for the bottom bar's background texture
func (bb *BottomBar) SetBackgroundColor(color *sdl.Color) error {
	var bottomBarTex *sdl.Texture
	var err error
	if bottomBarTex, err = createSolidColorTexture(bb.ctx.Rend, bb.area.W, bb.area.H, color.R, color.G, color.B, color.A); err != nil {
		return err
	}
	if bb.tex != nil {
		bb.tex.Destroy()
	}
	bb.tex = bottomBarTex
	return nil
}

// Destroy frees all surfaces and textures in the BottomBar
func (bb *BottomBar) Destroy() {
	bb.tex.Destroy()
}

// GetBoundary returns the clickable region of the UIComponent
func (bb *BottomBar) GetBoundary() *sdl.Rect {
	return bb.area
}

// Render draws the UIComponent
func (bb *BottomBar) Render() error {
	msg := <-bb.comms

	var err error
	if err = bb.ctx.Rend.SetViewport(bb.area); err != nil {
		return err
	}
	// first render grey background
	if err = bb.ctx.Rend.Copy(bb.tex, nil, nil); err != nil {
		return err
	}
	// second render white text on top
	coords := "(" + strconv.Itoa(int(msg.mousePix.x)) + ", " + strconv.Itoa(int(msg.mousePix.y)) + ")"
	pos := coord{bb.area.W, int32(float64(bb.area.H) / 2.0)}
	if err = renderText(bb.ctx.Rend, bb.font, bb.fontSize, coords, pos, Align{AlignMiddle, AlignRight}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
		return err
	}
	pos.x = 0
	if err = renderText(bb.ctx.Rend, bb.font, bb.fontSize, msg.fileName, pos, Align{AlignMiddle, AlignLeft}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
		return err
	}
	pos.x = bb.area.W / 2
	if err = renderText(bb.ctx.Rend, bb.font, bb.fontSize, strconv.FormatFloat(msg.mult, 'f', 4, 64)+"x", pos, Align{AlignMiddle, AlignCenter}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
		return err
	}
	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (bb *BottomBar) OnEnter() {}

// OnLeave is called when the cursor leaves the UIComponent's region
func (bb *BottomBar) OnLeave() {}

// OnMotion is called when the cursor moves within the UIComponent's region
func (bb *BottomBar) OnMotion(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (bb *BottomBar) OnScroll(evt *sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the UIComponent's region
func (bb *BottomBar) OnClick(evt *sdl.MouseButtonEvent) bool {
	return true
}

// OnResize is called when the user resizes the window
func (bb *BottomBar) OnResize(x, y int32) {
	bb.area.W = x
	bb.area.Y = y - bb.area.H
	bb.tex.Destroy()
	var err error
	var bottomBarTex *sdl.Texture
	if bottomBarTex, err = createSolidColorTexture(bb.ctx.Rend, bb.area.W, bb.area.H, bb.color.R, bb.color.G, bb.color.B, bb.color.A); err != nil {
		panic(err)
	}
	bb.tex = bottomBarTex
}
