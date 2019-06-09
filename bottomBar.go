package main

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var _ UIComponent = UIComponent(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area          *sdl.Rect
	comms         <-chan imageComm
	color         [4]float32
	mousePix      coord
	font          *ttf.Font
	fontSize      int32
	backProgramID uint32
	textProgramID uint32
	backVaoID     uint32
	backVboID     uint32
	textVaoID     uint32
	textVboID     uint32
	uniColorID    int32
	uniScrSizeID  int32
	fontTexID     uint32
	runeMap       []runeInfo
	cfg           *config
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
// the background color defaults to grey (0x808080FF)
func NewBottomBar(area *sdl.Rect, comms <-chan imageComm, cfg *config) (*BottomBar, error) {
	var m float32 = 255.0
	color := [4]float32{128.0 / m, 128.0 / m, 128.0 / m, 1.0}
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
	fontTexID, runeMap, err := loadFontTexture("NotoMono-Regular.ttf", 24)
	if err != nil {
		return nil, err
	}

	uniColorID := gl.GetUniformLocation(backProgramID, &[]byte("uni_color\x00")[0])
	gl.UseProgram(backProgramID)
	gl.Uniform4f(uniColorID, color[0], color[1], color[2], color[3])

	uniScrSizeID := gl.GetUniformLocation(textProgramID, &[]byte("screen_size\x00")[0])
	texSizeID := gl.GetUniformLocation(textProgramID, &[]byte("tex_size\x00")[0])
	var texSheetWidth, texSheetHeight int32
	gl.BindTexture(gl.TEXTURE_2D, fontTexID)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_WIDTH, &texSheetWidth)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_HEIGHT, &texSheetHeight)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.UseProgram(textProgramID)
	gl.Uniform2f(texSizeID, float32(texSheetWidth), float32(texSheetHeight))
	gl.Uniform2f(uniScrSizeID, float32(cfg.screenWidth), float32(cfg.screenHeight))

	glSquare := []float32{
		-1.0, -1.0, // bottom-left
		-1.0, +1.0, // top-left
		+1.0, +1.0, // top-right
		-1.0, -1.0, // bottom-left
		+1.0, +1.0, // top-right
		+1.0, -1.0, // bottom-right
	}

	var backVaoID, backVboID uint32
	gl.GenVertexArrays(1, &backVaoID)
	gl.GenBuffers(1, &backVboID)
	configureVAO(backVaoID, backVboID, []int32{2})
	gl.BindBuffer(gl.ARRAY_BUFFER, backVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(glSquare), gl.Ptr(&glSquare[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	var textVaoID, textVboID uint32
	gl.GenVertexArrays(1, &textVaoID)
	gl.GenBuffers(1, &textVboID)
	configureVAO(textVaoID, textVboID, []int32{2, 2})

	gl.UseProgram(0)

	return &BottomBar{
		area:          area,
		comms:         comms,
		color:         color,
		backProgramID: backProgramID,
		textProgramID: textProgramID,
		backVaoID:     backVaoID,
		backVboID:     backVboID,
		textVaoID:     textVaoID,
		textVboID:     textVboID,
		fontTexID:     fontTexID,
		runeMap:       runeMap,
		uniColorID:    uniColorID,
		uniScrSizeID:  uniScrSizeID,
		cfg:           cfg,
	}, nil
}

// SetBackgroundColor sets the color for the bottom bar's background texture
func (bb *BottomBar) SetBackgroundColor(color *sdl.Color) error {
	gl.UseProgram(bb.backProgramID)
	gl.Uniform4f(bb.uniColorID, float32(color.R), float32(color.G), float32(color.B), float32(color.A))
	gl.UseProgram(0)
	return nil
}

// Destroy frees all assets obtained by the UIComponent
func (bb *BottomBar) Destroy() {
	gl.DeleteBuffers(1, &bb.backVboID)
	gl.DeleteBuffers(1, &bb.textVboID)
	gl.DeleteVertexArrays(1, &bb.backVaoID)
	gl.DeleteVertexArrays(1, &bb.textVaoID)
}

// GetBoundary returns the clickable region of the UIComponent
func (bb *BottomBar) GetBoundary() *sdl.Rect {
	return bb.area
}

// Render draws the UIComponent
func (bb *BottomBar) Render() error {
	// msg := <-bb.comms

	// first render solid color background
	gl.Viewport(bb.area.X, 0, bb.area.W, bb.area.H)
	gl.UseProgram(bb.backProgramID)

	gl.BindVertexArray(bb.backVaoID)
	gl.EnableVertexAttribArray(0)
	gl.DrawArrays(gl.TRIANGLES, 0, 6) // always 6 vertices for background rectangle
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)

	// second render text on top
	testStr := "The quick brown fox jumped over the lazy dog"
	// TODO implement alignment, requiring the unused width and height returns
	// TODO optimize rendering by no-oping if string hasn't changed (or window size)
	glyphData, _, _ := mapString(testStr, bb.runeMap)
	gl.BindBuffer(gl.ARRAY_BUFFER, bb.textVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(glyphData), gl.Ptr(&glyphData[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.Viewport(0, 0, bb.cfg.screenWidth, bb.cfg.screenHeight)
	gl.UseProgram(bb.textProgramID)

	gl.BindVertexArray(bb.textVaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.BindTexture(gl.TEXTURE_2D, bb.fontTexID)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(glyphData)/4))
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindVertexArray(0)

	gl.UseProgram(0)
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
	bb.area.W += x
	bb.area.Y += y

	gl.UseProgram(bb.textProgramID)
	gl.Uniform2f(bb.uniScrSizeID, float32(bb.cfg.screenWidth), float32(bb.cfg.screenHeight))
	gl.UseProgram(0)
}

// String  returns the name of the component type
func (bb *BottomBar) String() string {
	return "BottomBar"
}
