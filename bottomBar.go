package main

import (
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area       *sdl.Rect
	tex        *sdl.Texture
	mouseComms <-chan coord
	ctx        *context
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
func NewBottomBar(area *sdl.Rect, mouseComms <-chan coord, ctx *context) (*BottomBar, error) {
	var err error
	var bottomBarTex *sdl.Texture
	if bottomBarTex, err = createSolidColorTexture(ctx.Rend, ctx.Conf.screenWidth, ctx.Conf.bottomBarHeight, 0x80, 0x80, 0x80, 0xFF); err != nil {
		return nil, err
	}
	return &BottomBar{
		area:       area,
		tex:        bottomBarTex,
		mouseComms: mouseComms,
		ctx:        ctx,
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
func (bb *BottomBar) Render(rend *sdl.Renderer) error {
	mousePix := <-bb.mouseComms
	var err error
	if err = rend.SetViewport(bb.area); err != nil {
		return err
	}
	// first render grey background
	if err = rend.Copy(bb.tex, nil, nil); err != nil {
		return err
	}
	// second render white text on top
	coords := "(" + strconv.Itoa(int(mousePix.x)) + ", " + strconv.Itoa(int(mousePix.y)) + ")"
	pos := coord{bb.ctx.Conf.screenWidth, int32(float64(bb.ctx.Conf.bottomBarHeight) / 2.0)}
	if err = renderText(bb.ctx.Conf, rend, coords, pos, Align{AlignMiddle, AlignRight}); err != nil {
		return err
	}
	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (bb *BottomBar) OnEnter(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnLeave is called when the cursor leaves the UIComponent's region
func (bb *BottomBar) OnLeave(evt *sdl.MouseMotionEvent) bool {
	return true
}

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
