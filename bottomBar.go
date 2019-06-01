package main

import (
	"strconv"

	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area       *sdl.Rect
	mouseComms <-chan coord
	ctx        *context
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
func NewBottomBar(area *sdl.Rect, mouseComms <-chan coord, ctx *context) (*BottomBar, error) {
	return &BottomBar{
		area:       area,
		mouseComms: mouseComms,
		ctx:        ctx,
	}, nil
}

// GetBoundary returns the clickable region of the UIComponent
func (bb *BottomBar) GetBoundary() *sdl.Rect {
	return bb.area
}

// Render draws the UIComponent
func (bb *BottomBar) Render(rend *sdl.Renderer) error {
	mousePix := <-bb.mouseComms
	var err error
	g := uint8(0x80)
	var bottomBarTex *sdl.Texture
	if bottomBarTex, err = createSolidColorTexture(rend, bb.ctx.Conf.screenWidth, bb.ctx.Conf.bottomBarHeight, g, g, g, 0xFF); err != nil {
		return err
	}
	if err = rend.SetViewport(bb.area); err != nil {
		return err
	}
	// first render grey background
	if err = rend.Copy(bottomBarTex, nil, nil); err != nil {
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
