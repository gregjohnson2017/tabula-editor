package menu

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/font"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/veandco/go-sdl2/sdl"
)

// Button defines an interactive button
type Button struct {
	area               *sdl.Rect
	cfg                *config.Config
	defaultBackColor   [4]float32
	highlightBackColor [4]float32
	defaultTextColor   [4]float32
	highlightTextColor [4]float32
	strTriangles       []float32
	backProgram        gfx.Program
	textProgram        gfx.Program
	backBuf            *gfx.BufferArray
	textBuf            *gfx.BufferArray
	fontInfo           font.Info
	text               string
	pressed            bool
	hovering           bool
	action             func()
}

var _ ui.Component = ui.Component(&Button{})

// NewButton returns a pointer to a Button struct
// defaultColor and highlightColor default to light grey (0xD6CFCFFF) and blue (0X0046AFFF) respectively, if nil
func NewButton(area *sdl.Rect, cfg *config.Config, text string, action func()) (*Button, error) {
	var err error
	v1, err := gfx.NewShader(gfx.SolidColorVertex, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f1, err := gfx.NewShader(gfx.SolidColorFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	backProgram, err := gfx.NewProgram([]gfx.Shader{v1, f1})
	if err != nil {
		return nil, err
	}

	v2, err := gfx.NewShader(gfx.GlyphShaderVertex, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f2, err := gfx.NewShader(gfx.GlyphShaderFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	textProgram, err := gfx.NewProgram([]gfx.Shader{v2, f2})
	if err != nil {
		return nil, err
	}

	fnt, err := font.LoadFontTexture("NotoMono-Regular.ttf", 14)
	if err != nil {
		return nil, err
	}

	backColor := [4]float32{0.8392, 0.8118, 0.8118, 1.0}
	textColor := [4]float32{0.0, 0.0, 0.0, 1.0}

	backProgram.UploadUniform("uni_color", backColor[0], backColor[1], backColor[2], backColor[3])
	textProgram.UploadUniform("screen_size", float32(cfg.ScreenWidth), float32(cfg.ScreenHeight))
	textProgram.UploadUniform("tex_size", float32(fnt.GetTexture().GetWidth()),
		float32(fnt.GetTexture().GetHeight()))
	textProgram.UploadUniform("text_color", textColor[0], textColor[1], textColor[2], textColor[3])

	backTriangles := []float32{
		-1.0, -1.0, // bottom-left
		-1.0, +1.0, // top-left
		+1.0, +1.0, // top-right

		-1.0, -1.0, // bottom-left
		+1.0, +1.0, // top-right
		+1.0, -1.0, // bottom-right
	}
	pos := sdl.Point{X: area.X + area.W/2, Y: cfg.ScreenHeight - area.Y - area.H/2}
	align := ui.Align{V: ui.AlignMiddle, H: ui.AlignCenter}
	textTriangles := font.MapString(text, fnt, pos, align)

	backBuf := gfx.NewBufferArray(gl.TRIANGLES, []int32{2})
	err = backBuf.Load(backTriangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load button background triangles: %v", err)
	}

	textBuf := gfx.NewBufferArray(gl.TRIANGLES, []int32{2, 2})
	err = textBuf.Load(textTriangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load button text triangles: %v", err)
	}

	if action == nil {
		action = func() {}
	}

	return &Button{
		area:               area,
		defaultBackColor:   backColor,
		highlightBackColor: [4]float32{0.0, 0.2745, 0.6863, 1.0},
		defaultTextColor:   textColor,
		highlightTextColor: [4]float32{1.0, 1.0, 1.0, 1.0},
		strTriangles:       textTriangles,
		backProgram:        backProgram,
		textProgram:        textProgram,
		backBuf:            backBuf,
		textBuf:            textBuf,
		fontInfo:           fnt,
		text:               text,
		cfg:                cfg,
		pressed:            false,
		hovering:           false,
		action:             action,
	}, nil
}

// SetDefaultBackgroundColor changes the default background color
func (b *Button) SetDefaultBackgroundColor(color [4]float32) {
	b.defaultBackColor = color
	b.backProgram.UploadUniform("text_color", b.defaultBackColor[0], b.defaultBackColor[1], b.defaultBackColor[2], b.defaultBackColor[3])
}

// SetHighlightBackgroundColor changes the highlight background color
func (b *Button) SetHighlightBackgroundColor(color [4]float32) {
	b.highlightBackColor = color
}

// SetDefaultTextColor changes the default text color
func (b *Button) SetDefaultTextColor(color [4]float32) {
	b.defaultTextColor = color
	b.textProgram.UploadUniform("text_color", b.defaultTextColor[0], b.defaultTextColor[1], b.defaultTextColor[2], b.defaultTextColor[3])
}

// SetHighlightTextColor changes the highlight text color
func (b *Button) SetHighlightTextColor(color [4]float32) {
	b.highlightTextColor = color
}

// Destroy frees all assets obtained by the ui.Component
func (b *Button) Destroy() {
	b.backBuf.Destroy()
	b.textBuf.Destroy()
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (b *Button) InBoundary(pt sdl.Point) bool {
	return ui.InBounds(*b.area, pt)
}

// Render draws the ui.Component
func (b *Button) Render() {
	sw := util.Start()

	// render solid color background
	gl.Viewport(b.area.X, b.cfg.ScreenHeight-b.area.Y-b.area.H, b.area.W, b.area.H)
	b.backProgram.Bind()
	b.backBuf.Draw()
	b.backProgram.Unbind()

	// render text on top
	gl.Viewport(0, 0, b.cfg.ScreenWidth, b.cfg.ScreenHeight)
	b.textProgram.Bind()
	b.fontInfo.GetTexture().Bind()
	b.textBuf.Draw()
	b.fontInfo.GetTexture().Unbind()
	b.textProgram.Unbind()

	sw.StopRecordAverage(b.String() + ".Render")
}

// OnEnter is called when the cursor enters the ui.Component's region
func (b *Button) OnEnter() {
	b.hovering = true

	b.backProgram.UploadUniform("uni_color", b.highlightBackColor[0], b.highlightBackColor[1], b.highlightBackColor[2], b.highlightBackColor[3])

	b.textProgram.UploadUniform("text_color", b.highlightTextColor[0], b.highlightTextColor[1], b.highlightTextColor[2], b.highlightTextColor[3])
}

// OnLeave is called when the cursor leaves the ui.Component's region
func (b *Button) OnLeave() {
	b.hovering = false
	b.pressed = false

	b.backProgram.UploadUniform("uni_color", b.defaultBackColor[0], b.defaultBackColor[1], b.defaultBackColor[2], b.defaultBackColor[3])
	b.textProgram.UploadUniform("text_color", b.defaultTextColor[0], b.defaultTextColor[1], b.defaultTextColor[2], b.defaultTextColor[3])
}

// OnMotion is called when the cursor moves within the ui.Component's region
func (b *Button) OnMotion(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (b *Button) OnScroll(evt *sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the ui.Component's region
func (b *Button) OnClick(evt *sdl.MouseButtonEvent) bool {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		b.pressed = true
	} else if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.RELEASED && b.pressed {
		b.pressed = false
		b.action()
	}
	return true
}

// OnResize is called when the user resizes the window
func (b *Button) OnResize(x, y int32) {
	// recompute text triangles
	b.textProgram.UploadUniform("screen_size", float32(b.cfg.ScreenWidth), float32(b.cfg.ScreenHeight))

	pos := sdl.Point{X: b.area.X + b.area.W/2, Y: b.cfg.ScreenHeight - b.area.Y - b.area.H/2}
	align := ui.Align{V: ui.AlignMiddle, H: ui.AlignCenter}
	textTriangles := font.MapString(b.text, b.fontInfo, pos, align)
	err := b.textBuf.Load(textTriangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load button text triangles: %v", err)
	}
}

// String  returns the name of the component type
func (b *Button) String() string {
	return "menu.Button"
}
