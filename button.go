package main

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&Button{})

// Button defines an interactive button
type Button struct {
	area               *sdl.Rect
	cfg                *config
	defaultBackColor   [4]float32
	highlightBackColor [4]float32
	defaultTextColor   [4]float32
	highlightTextColor [4]float32
	strTriangles       []float32
	backProgramID      uint32
	textProgramID      uint32
	backVaoID          uint32
	backVboID          uint32
	textVaoID          uint32
	textVboID          uint32
	fontTexID          uint32
	runeMap            []runeInfo
	text               string
	pressed            bool
	hovering           bool
	action             func()
}

// NewButton returns a pointer to a Button struct
// defaultColor and highlightColor default to light grey (0xD6CFCFFF) and blue (0X0046AFFF) respectively, if nil
func NewButton(area *sdl.Rect, cfg *config, text string, action func()) (*Button, error) {
	var err error
	var backProgramID uint32
	if backProgramID, err = CreateShaderProgram(solidColorVertex, solidColorFragment); err != nil {
		return nil, err
	}
	var textProgramID uint32
	if textProgramID, err = CreateShaderProgram(glyphShaderVertex, glyphShaderFragment); err != nil {
		return nil, err
	}

	// TODO preload fonts & sizes in main.go and store in a collection for all UIComponents
	fontTexID, runeMap, err := loadFontTexture("NotoMono-Regular.ttf", 14)
	if err != nil {
		return nil, err
	}

	backColor := [4]float32{0.8392, 0.8118, 0.8118, 1.0}
	textColor := [4]float32{0.0, 0.0, 0.0, 1.0}

	backColorID := gl.GetUniformLocation(backProgramID, &[]byte("uni_color\x00")[0])
	gl.UseProgram(backProgramID)
	gl.Uniform4f(backColorID, backColor[0], backColor[1], backColor[2], backColor[3])

	uniScrSizeID := gl.GetUniformLocation(textProgramID, &[]byte("screen_size\x00")[0])
	texSizeID := gl.GetUniformLocation(textProgramID, &[]byte("tex_size\x00")[0])
	textColorID := gl.GetUniformLocation(textProgramID, &[]byte("text_color\x00")[0])

	var texSheetWidth, texSheetHeight int32
	gl.BindTexture(gl.TEXTURE_2D, fontTexID)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_WIDTH, &texSheetWidth)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_HEIGHT, &texSheetHeight)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.UseProgram(textProgramID)
	gl.Uniform2f(uniScrSizeID, float32(cfg.screenWidth), float32(cfg.screenHeight))
	gl.Uniform2f(texSizeID, float32(texSheetWidth), float32(texSheetHeight))
	gl.Uniform4f(textColorID, textColor[0], textColor[1], textColor[2], textColor[3])

	backTriangles := []float32{
		-1.0, -1.0, // bottom-left
		-1.0, +1.0, // top-left
		+1.0, +1.0, // top-right

		-1.0, -1.0, // bottom-left
		+1.0, +1.0, // top-right
		+1.0, -1.0, // bottom-right
	}
	textTriangles := mapString(text, runeMap, coord{area.X + area.W/2, cfg.screenHeight - area.Y - area.H/2}, Align{AlignMiddle, AlignCenter})

	var backVaoID, backVboID uint32
	gl.GenVertexArrays(1, &backVaoID)
	gl.GenBuffers(1, &backVboID)
	configureVAO(backVaoID, backVboID, []int32{2})
	gl.BindBuffer(gl.ARRAY_BUFFER, backVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(backTriangles), gl.Ptr(&backTriangles[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	var textVaoID, textVboID uint32
	gl.GenVertexArrays(1, &textVaoID)
	gl.GenBuffers(1, &textVboID)
	configureVAO(textVaoID, textVboID, []int32{2, 2})
	gl.BindBuffer(gl.ARRAY_BUFFER, textVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(textTriangles), gl.Ptr(&textTriangles[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.UseProgram(0)

	return &Button{
		area:               area,
		defaultBackColor:   backColor,
		highlightBackColor: [4]float32{0.0, 0.2745, 0.6863, 1.0},
		defaultTextColor:   textColor,
		highlightTextColor: [4]float32{1.0, 1.0, 1.0, 1.0},
		strTriangles:       textTriangles,
		backProgramID:      backProgramID,
		textProgramID:      textProgramID,
		backVaoID:          backVaoID,
		backVboID:          backVboID,
		textVaoID:          textVaoID,
		textVboID:          textVboID,
		fontTexID:          fontTexID,
		runeMap:            runeMap,
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
	backColorID := gl.GetUniformLocation(b.textProgramID, &[]byte("text_color\x00")[0])
	gl.UseProgram(b.backProgramID)
	gl.Uniform4f(backColorID, b.defaultBackColor[0], b.defaultBackColor[1], b.defaultBackColor[2], b.defaultBackColor[3])
	gl.UseProgram(0)
}

// SetHighlightBackgroundColor changes the highlight background color
func (b *Button) SetHighlightBackgroundColor(color [4]float32) {
	b.highlightBackColor = color
}

// SetDefaultTextColor changes the default text color
func (b *Button) SetDefaultTextColor(color [4]float32) {
	b.defaultTextColor = color
	textColorID := gl.GetUniformLocation(b.textProgramID, &[]byte("text_color\x00")[0])
	gl.UseProgram(b.textProgramID)
	gl.Uniform4f(textColorID, b.defaultTextColor[0], b.defaultTextColor[1], b.defaultTextColor[2], b.defaultTextColor[3])
	gl.UseProgram(0)
}

// SetHighlightTextColor changes the highlight text color
func (b *Button) SetHighlightTextColor(color [4]float32) {
	b.highlightTextColor = color
}

// Destroy frees all assets obtained by the UIComponent
func (b *Button) Destroy() {
	gl.DeleteBuffers(1, &b.backVboID)
	gl.DeleteBuffers(1, &b.textVboID)
	gl.DeleteVertexArrays(1, &b.backVaoID)
	gl.DeleteVertexArrays(1, &b.textVaoID)
}

// GetBoundary returns the clickable region of the UIComponent
func (b *Button) GetBoundary() *sdl.Rect {
	return b.area
}

// Render draws the UIComponent
func (b *Button) Render() error {
	// render solid color background
	gl.Viewport(b.area.X, b.cfg.screenHeight-b.area.Y-b.area.H, b.area.W, b.area.H)
	gl.UseProgram(b.backProgramID)

	gl.BindVertexArray(b.backVaoID)
	gl.EnableVertexAttribArray(0)
	gl.DrawArrays(gl.TRIANGLES, 0, 6) // always 6 vertices for background rectangle
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)

	// render text on top
	gl.Viewport(0, 0, b.cfg.screenWidth, b.cfg.screenHeight)
	gl.UseProgram(b.textProgramID)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.textVboID)
	gl.BindVertexArray(b.textVaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.BindTexture(gl.TEXTURE_2D, b.fontTexID)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(b.strTriangles)/4))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.UseProgram(0)

	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (b *Button) OnEnter() {
	b.hovering = true

	backColorID := gl.GetUniformLocation(b.backProgramID, &[]byte("uni_color\x00")[0])
	gl.UseProgram(b.backProgramID)
	gl.Uniform4f(backColorID, b.highlightBackColor[0], b.highlightBackColor[1], b.highlightBackColor[2], b.highlightBackColor[3])
	textColorID := gl.GetUniformLocation(b.textProgramID, &[]byte("text_color\x00")[0])
	gl.UseProgram(b.textProgramID)
	gl.Uniform4f(textColorID, b.highlightTextColor[0], b.highlightTextColor[1], b.highlightTextColor[2], b.highlightTextColor[3])
	gl.UseProgram(0)
}

// OnLeave is called when the cursor leaves the UIComponent's region
func (b *Button) OnLeave() {
	b.hovering = false
	b.pressed = false

	backColorID := gl.GetUniformLocation(b.backProgramID, &[]byte("uni_color\x00")[0])
	gl.UseProgram(b.backProgramID)
	gl.Uniform4f(backColorID, b.defaultBackColor[0], b.defaultBackColor[1], b.defaultBackColor[2], b.defaultBackColor[3])
	textColorID := gl.GetUniformLocation(b.textProgramID, &[]byte("text_color\x00")[0])
	gl.UseProgram(b.textProgramID)
	gl.Uniform4f(textColorID, b.defaultTextColor[0], b.defaultTextColor[1], b.defaultTextColor[2], b.defaultTextColor[3])
	gl.UseProgram(0)
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

// OnResize is called when the user resizes the window
func (b *Button) OnResize(x, y int32) {
	// recompute text triangles
	uniformID := gl.GetUniformLocation(b.textProgramID, &[]byte("screen_size\x00")[0])
	gl.UseProgram(b.textProgramID)
	gl.Uniform2f(uniformID, float32(b.cfg.screenWidth), float32(b.cfg.screenHeight))
	gl.UseProgram(0)

	textTriangles := mapString(b.text, b.runeMap, coord{b.area.X + b.area.W/2, b.cfg.screenHeight - b.area.Y - b.area.H/2}, Align{AlignMiddle, AlignCenter})
	gl.BindBuffer(gl.ARRAY_BUFFER, b.textVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(textTriangles), gl.Ptr(&textTriangles[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

}

// String  returns the name of the component type
func (b *Button) String() string {
	return "Button"
}
