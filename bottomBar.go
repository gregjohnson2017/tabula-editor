package main

import (
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area     *sdl.Rect
	tex      *sdl.Texture
	comms    <-chan imageComm
	mousePix coord
	ctx      *context
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
func NewBottomBar(area *sdl.Rect, comms <-chan imageComm, ctx *context, color *sdl.Color) (*BottomBar, error) {
	var err error
	var bottomBarTex *sdl.Texture
	if bottomBarTex, err = createSolidColorTexture(ctx.Rend, ctx.Conf.screenWidth, ctx.Conf.bottomBarHeight, color.R, color.G, color.B, color.A); err != nil {
		return nil, err
	}
	return &BottomBar{
		area:  area,
		tex:   bottomBarTex,
		comms: comms,
		ctx:   ctx,
	}, nil
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
	pos := coord{bb.ctx.Conf.screenWidth, int32(float64(bb.ctx.Conf.bottomBarHeight) / 2.0)}
	if err = renderText(bb.ctx.Rend, bb.ctx.Conf.fontName, 24, coords, pos, Align{AlignMiddle, AlignRight}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
		return err
	}
	pos.x = 0
	if err = renderText(bb.ctx.Rend, bb.ctx.Conf.fontName, 24, msg.fileName, pos, Align{AlignMiddle, AlignLeft}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
		return err
	}
	pos.x = bb.ctx.Conf.screenWidth / 2
	if err = renderText(bb.ctx.Rend, bb.ctx.Conf.fontName, 24, strconv.FormatFloat(msg.mult, 'f', 2, 64)+"x", pos, Align{AlignMiddle, AlignCenter}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
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
