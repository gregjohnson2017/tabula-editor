package main

// import (
// 	"github.com/veandco/go-sdl2/sdl"
// 	"github.com/veandco/go-sdl2/ttf"
// )

// var _ UIComponent = UIComponent(&Button{})

// // Button defines an interactive button
// type Button struct {
// 	area               *sdl.Rect
// 	defaultTex         *sdl.Texture
// 	highlightTex       *sdl.Texture
// 	defaultTextColor   *sdl.Color
// 	highlightTextColor *sdl.Color
// 	font               *ttf.Font
// 	fontSize           int32
// 	text               string
// 	ctx                *context
// 	pressed            bool
// 	hovering           bool
// 	menu               *Menu
// 	action             func()
// }

// // NewMenuButton returns a pointer to a Button struct with special menu functionality
// // defaultColor and highlightColor default to light grey (0xD6CFCFFF) and blue (0X0046AFFF) respectively, if nil
// func NewMenuButton(area *sdl.Rect, ctx *context, defaultColor *sdl.Color, highlightColor *sdl.Color, text string, fontName string, fontSize int32, action func(), menu *Menu) (*Button, error) {
// 	if defaultColor == nil {
// 		defaultColor = &sdl.Color{R: 0xD6, G: 0xCF, B: 0xCF, A: 0xFF}
// 	}
// 	if highlightColor == nil {
// 		highlightColor = &sdl.Color{R: 0x00, G: 0x46, B: 0xAF, A: 0xFF}
// 	}
// 	var err error
// 	var defaultTex *sdl.Texture
// 	if defaultTex, err = createSolidColorTexture(ctx.Rend, area.W, area.H, defaultColor.R, defaultColor.G, defaultColor.B, defaultColor.A); err != nil {
// 		return nil, err
// 	}
// 	var highlightTex *sdl.Texture
// 	if highlightTex, err = createSolidColorTexture(ctx.Rend, area.W, area.H, highlightColor.R, highlightColor.G, highlightColor.B, highlightColor.A); err != nil {
// 		return nil, err
// 	}
// 	var font *ttf.Font
// 	if font, err = ttf.OpenFont(fontName, int(fontSize)); err != nil {
// 		return nil, err
// 	}
// 	return &Button{
// 		area:               area,
// 		defaultTex:         defaultTex,
// 		defaultTextColor:   &sdl.Color{R: 0x00, G: 0x00, B: 0x00, A: 0xFF},
// 		highlightTex:       highlightTex,
// 		highlightTextColor: &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF},
// 		text:               text,
// 		ctx:                ctx,
// 		pressed:            false,
// 		hovering:           false,
// 		action:             action,
// 		menu:               menu,
// 		font:               font,
// 		fontSize:           fontSize,
// 	}, nil
// }

// // NewButton returns a pointer to a new Button struct that implements UIComponent
// // defaultColor and highlightColor default to light grey (0xD6CFCFFF) and blue (0X0046AFFF) respectively, if nil
// func NewButton(area *sdl.Rect, ctx *context, text string, fontName string, fontSize int32, action func()) (*Button, error) {
// 	return NewMenuButton(area, ctx, nil, nil, text, fontName, fontSize, action, nil)
// }

// // SetDefaultBackgroundColor recreates the button's defaultTex with the given background color
// func (b *Button) SetDefaultBackgroundColor(color *sdl.Color) error {
// 	var err error
// 	var tex *sdl.Texture
// 	if tex, err = createSolidColorTexture(b.ctx.Rend, b.area.W, b.area.H, color.R, color.G, color.B, color.A); err != nil {
// 		return err
// 	}
// 	if b.defaultTex != nil {
// 		b.defaultTex.Destroy()
// 	}
// 	b.defaultTex = tex
// 	return nil
// }

// // SetHighlightBackgroundColor recreates the button's highlightTex with the given background color
// func (b *Button) SetHighlightBackgroundColor(color *sdl.Color) error {
// 	var err error
// 	var tex *sdl.Texture
// 	if tex, err = createSolidColorTexture(b.ctx.Rend, b.area.W, b.area.H, color.R, color.G, color.B, color.A); err != nil {
// 		return err
// 	}
// 	if b.highlightTex != nil {
// 		b.highlightTex.Destroy()
// 	}
// 	b.highlightTex = tex
// 	return nil
// }

// // SetDefaultTextColor changes the default text color
// func (b *Button) SetDefaultTextColor(color *sdl.Color) {
// 	b.defaultTextColor = color
// }

// // SetHighlightTextColor changes the highlight text color
// func (b *Button) SetHighlightTextColor(color *sdl.Color) {
// 	b.highlightTextColor = color
// }

// // Destroy frees all surfaces and textures in the BottomBar
// func (b *Button) Destroy() {
// 	b.defaultTex.Destroy()
// 	b.highlightTex.Destroy()
// }

// // GetBoundary returns the clickable region of the UIComponent
// func (b *Button) GetBoundary() *sdl.Rect {
// 	return b.area
// }

// // Render draws the UIComponent
// func (b *Button) Render() error {
// 	// choose correct pair of text/background color
// 	var backgroundTex *sdl.Texture
// 	var textColor *sdl.Color
// 	if b.hovering {
// 		backgroundTex = b.highlightTex
// 		textColor = b.highlightTextColor
// 	} else {
// 		backgroundTex = b.defaultTex
// 		textColor = b.defaultTextColor
// 	}

// 	// background first
// 	var err error
// 	if err = b.ctx.Rend.SetViewport(b.area); err != nil {
// 		return err
// 	}
// 	if err = b.ctx.Rend.Copy(backgroundTex, nil, nil); err != nil {
// 		return err
// 	}
// 	// text on top
// 	if err = renderText(b.ctx.Rend, b.font, b.fontSize, b.text, coord{b.area.W / 2, b.area.H / 2}, Align{AlignMiddle, AlignCenter}, textColor); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // OnEnter is called when the cursor enters the UIComponent's region
// func (b *Button) OnEnter() {
// 	b.hovering = true
// }

// // OnLeave is called when the cursor leaves the UIComponent's region
// func (b *Button) OnLeave() {
// 	b.hovering = false
// 	b.pressed = false
// }

// // OnMotion is called when the cursor moves within the UIComponent's region
// func (b *Button) OnMotion(evt *sdl.MouseMotionEvent) bool {
// 	return true
// }

// // OnScroll is called when the user scrolls within the UIComponent's region
// func (b *Button) OnScroll(evt *sdl.MouseWheelEvent) bool {
// 	return true
// }

// // OnClick is called when the user clicks within the UIComponent's region
// func (b *Button) OnClick(evt *sdl.MouseButtonEvent) bool {
// 	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
// 		b.pressed = true
// 	} else if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.RELEASED && b.pressed {
// 		b.pressed = false
// 		b.action()
// 	}
// 	return true
// }

// // OnResize is called when the user resizes the window
// func (b *Button) OnResize(x, y int32) {}

// // String  returns the name of the component type
// func (b *Button) String() string {
// 	return "Button"
// }
