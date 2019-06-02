package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&Button{})

// Button defines an interactive button
type Button struct {
	area         *sdl.Rect
	defaultTex   *sdl.Texture
	highlightTex *sdl.Texture
	text         string
	ctx          *context
	pressed      bool
	hovering     bool
	action       func()
}

// NewButton returns a pointer to a new BottomBar struct that implements UIComponent
// defaultColor and highlightColor default to white and blue respectively, if nil
func NewButton(area *sdl.Rect, ctx *context, defaultColor *sdl.Color, highlightColor *sdl.Color, text string, action func()) (*Button, error) {
	if defaultColor == nil {
		defaultColor = &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	}
	if highlightColor == nil {
		highlightColor = &sdl.Color{R: 0x00, G: 0x46, B: 0xAF, A: 0xFF}
	}
	var err error
	var defaultTex *sdl.Texture
	if defaultTex, err = createSolidColorTexture(ctx.Rend, area.W, area.H, defaultColor.R, defaultColor.G, defaultColor.B, defaultColor.A); err != nil {
		return nil, err
	}
	var highlightTex *sdl.Texture
	if highlightTex, err = createSolidColorTexture(ctx.Rend, area.W, area.H, highlightColor.R, highlightColor.G, highlightColor.B, highlightColor.A); err != nil {
		return nil, err
	}
	return &Button{
		area:         area,
		defaultTex:   defaultTex,
		highlightTex: highlightTex,
		text:         text,
		ctx:          ctx,
		pressed:      false,
		action:       action,
	}, nil
}

// Destroy frees all surfaces and textures in the BottomBar
func (b *Button) Destroy() {
	b.defaultTex.Destroy()
	b.highlightTex.Destroy()
}

// GetBoundary returns the clickable region of the UIComponent
func (b *Button) GetBoundary() *sdl.Rect {
	return b.area
}

// Render draws the UIComponent
func (b *Button) Render() error {
	// choose correct pair of text/background color
	var backgroundTex *sdl.Texture
	var textColor *sdl.Color
	if b.hovering {
		backgroundTex = b.highlightTex
		textColor = &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	} else {
		backgroundTex = b.defaultTex
		textColor = &sdl.Color{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	}

	// background first
	var err error
	if err = b.ctx.Rend.SetViewport(b.area); err != nil {
		return err
	}
	if err = b.ctx.Rend.Copy(backgroundTex, nil, nil); err != nil {
		return err
	}
	// text on top
	if err = renderText(b.ctx.Rend, b.ctx.Conf.fontName, 14, b.text, coord{b.area.W / 2, b.area.H / 2}, Align{AlignMiddle, AlignCenter}, textColor); err != nil {
		return err
	}
	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (b *Button) OnEnter(evt *sdl.MouseMotionEvent) bool {
	b.hovering = true
	return true
}

// OnLeave is called when the cursor leaves the UIComponent's region
func (b *Button) OnLeave(evt *sdl.MouseMotionEvent) bool {
	b.hovering = false
	b.pressed = false
	return true
}

// OnMotion is called when the cursor moves within the UIComponent's region
func (b *Button) OnMotion(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (b *Button) OnScroll(evt *sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the UIComponent's region
func (b *Button) OnClick(evt *sdl.MouseButtonEvent) bool {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		b.pressed = true
	} else if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.RELEASED && b.pressed {
		b.pressed = false
		b.action()
	}
	return true
}
