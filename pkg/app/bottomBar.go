package app

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/comms"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/font"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

var _ ui.Component = ui.Component(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area          *sdl.Rect
	comms         <-chan comms.Image
	backProgramID uint32
	textProgramID uint32
	backVaoID     uint32
	backVboID     uint32
	textVaoID     uint32
	textVboID     uint32
	fontInfo      font.Info
	cfg           *config.Config
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements ui.Component
// the background color defaults to grey (0x808080FF) and the text white
func NewBottomBar(area *sdl.Rect, comms <-chan comms.Image, cfg *config.Config) (*BottomBar, error) {
	var err error
	var backProgramID uint32
	if backProgramID, err = gfx.CreateShaderProgram(gfx.SolidColorVertex, gfx.SolidColorFragment); err != nil {
		return nil, err
	}
	var textProgramID uint32
	if textProgramID, err = gfx.CreateShaderProgram(gfx.GlyphShaderVertex, gfx.GlyphShaderFragment); err != nil {
		return nil, err
	}

	fnt, err := font.LoadFontTexture("NotoMono-Regular.ttf", 24)
	if err != nil {
		return nil, err
	}

	barColor := [4]float32{0.5, 0.5, 0.5, 1.0}
	textColor := [4]float32{1.0, 1.0, 1.0, 1.0}

	barColorID := gl.GetUniformLocation(backProgramID, &[]byte("uni_color\x00")[0])
	gl.UseProgram(backProgramID)
	gl.Uniform4f(barColorID, barColor[0], barColor[1], barColor[2], barColor[3])

	uniScrSizeID := gl.GetUniformLocation(textProgramID, &[]byte("screen_size\x00")[0])
	texSizeID := gl.GetUniformLocation(textProgramID, &[]byte("tex_size\x00")[0])
	textColorID := gl.GetUniformLocation(textProgramID, &[]byte("text_color\x00")[0])

	var texSheetWidth, texSheetHeight int32
	gl.BindTexture(gl.TEXTURE_2D, fnt.TextureID())
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_WIDTH, &texSheetWidth)
	gl.GetTexLevelParameteriv(gl.TEXTURE_2D, 0, gl.TEXTURE_HEIGHT, &texSheetHeight)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.UseProgram(textProgramID)
	gl.Uniform2f(texSizeID, float32(texSheetWidth), float32(texSheetHeight))
	gl.Uniform2f(uniScrSizeID, float32(cfg.ScreenWidth), float32(cfg.ScreenHeight))
	gl.Uniform4f(textColorID, textColor[0], textColor[1], textColor[2], textColor[3])

	backTriangles := []float32{
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
	gfx.ConfigureVAO(backVaoID, backVboID, []int32{2})
	gl.BindBuffer(gl.ARRAY_BUFFER, backVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(backTriangles), gl.Ptr(&backTriangles[0]), gl.STATIC_DRAW)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	var textVaoID, textVboID uint32
	gl.GenVertexArrays(1, &textVaoID)
	gl.GenBuffers(1, &textVboID)
	gfx.ConfigureVAO(textVaoID, textVboID, []int32{2, 2})

	gl.UseProgram(0)

	return &BottomBar{
		area:          area,
		comms:         comms,
		backProgramID: backProgramID,
		textProgramID: textProgramID,
		backVaoID:     backVaoID,
		backVboID:     backVboID,
		textVaoID:     textVaoID,
		textVboID:     textVboID,
		fontInfo:      fnt,
		cfg:           cfg,
	}, nil
}

// SetBackgroundColor sets the color for the bottom bar's background texture
func (bb *BottomBar) SetBackgroundColor(color []float32) {
	uniformID := gl.GetUniformLocation(bb.backProgramID, &[]byte("uni_color\x00")[0])
	gl.UseProgram(bb.backProgramID)
	gl.Uniform4f(uniformID, float32(color[0]), float32(color[1]), float32(color[2]), float32(color[3]))
	gl.UseProgram(0)
}

// SetTextColor sets the color for the bottom bar's text elements
func (bb *BottomBar) SetTextColor(color []float32) {
	uniformID := gl.GetUniformLocation(bb.textProgramID, &[]byte("text_color\x00")[0])
	gl.UseProgram(bb.textProgramID)
	gl.Uniform4f(uniformID, float32(color[0]), float32(color[1]), float32(color[2]), float32(color[3]))
	gl.UseProgram(0)
}

// Destroy frees all assets obtained by the ui.Component
func (bb *BottomBar) Destroy() {
	gl.DeleteBuffers(1, &bb.backVboID)
	gl.DeleteBuffers(1, &bb.textVboID)
	gl.DeleteVertexArrays(1, &bb.backVaoID)
	gl.DeleteVertexArrays(1, &bb.textVaoID)
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (bb *BottomBar) InBoundary(pt sdl.Point) bool {
	return ui.InBounds(*bb.area, pt)
}

// Render draws the ui.Component
func (bb *BottomBar) Render() {
	msg := <-bb.comms

	// first render solid color background
	gl.Viewport(bb.area.X, 0, bb.area.W, bb.area.H)
	gl.UseProgram(bb.backProgramID)

	gl.BindVertexArray(bb.backVaoID)
	gl.EnableVertexAttribArray(0)
	gl.DrawArrays(gl.TRIANGLES, 0, 6) // always 6 vertices for background rectangle
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)

	// second render text on top
	// TODO optimize rendering by no-oping if string hasn't changed (or window size)
	fileNameMessage := msg.FileName
	zoomMessage := fmt.Sprintf("%vx", msg.Mult)
	mousePixMessage := fmt.Sprintf("(%v, %v)", msg.MousePix.X, msg.MousePix.Y)

	pos := sdl.Point{X: 0, Y: bb.cfg.BottomBarHeight / 2}
	align := ui.Align{V: ui.AlignMiddle, H: ui.AlignLeft}
	fileNameTriangles := font.MapString(fileNameMessage, bb.fontInfo, pos, align)
	pos = sdl.Point{X: bb.cfg.ScreenWidth / 2, Y: bb.cfg.BottomBarHeight / 2}
	align = ui.Align{V: ui.AlignMiddle, H: ui.AlignCenter}
	zoomTriangles := font.MapString(zoomMessage, bb.fontInfo, pos, align)
	pos = sdl.Point{X: bb.cfg.ScreenWidth, Y: bb.cfg.BottomBarHeight / 2}
	align = ui.Align{V: ui.AlignMiddle, H: ui.AlignRight}
	mousePixTriangles := font.MapString(mousePixMessage, bb.fontInfo, pos, align)
	triangles := make([]float32, 0, len(fileNameTriangles)+len(zoomTriangles)+len(mousePixTriangles))
	triangles = append(triangles, fileNameTriangles...)
	triangles = append(triangles, zoomTriangles...)
	triangles = append(triangles, mousePixTriangles...)

	gl.Viewport(0, 0, bb.cfg.ScreenWidth, bb.cfg.ScreenHeight)
	gl.UseProgram(bb.textProgramID)

	gl.BindBuffer(gl.ARRAY_BUFFER, bb.textVboID)
	gl.BindVertexArray(bb.textVaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.BindTexture(gl.TEXTURE_2D, bb.fontInfo.TextureID())

	gl.BufferData(gl.ARRAY_BUFFER, 4*len(triangles), gl.Ptr(&triangles[0]), gl.STATIC_DRAW)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangles)/4))

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	gl.UseProgram(0)
}

// OnEnter is called when the cursor enters the ui.Component's region
func (bb *BottomBar) OnEnter() {}

// OnLeave is called when the cursor leaves the ui.Component's region
func (bb *BottomBar) OnLeave() {}

// OnMotion is called when the cursor moves within the ui.Component's region
func (bb *BottomBar) OnMotion(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (bb *BottomBar) OnScroll(evt *sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the ui.Component's region
func (bb *BottomBar) OnClick(evt *sdl.MouseButtonEvent) bool {
	return true
}

// OnResize is called when the user resizes the window
func (bb *BottomBar) OnResize(x, y int32) {
	bb.area.W += x
	bb.area.Y += y

	uniformID := gl.GetUniformLocation(bb.textProgramID, &[]byte("screen_size\x00")[0])
	gl.UseProgram(bb.textProgramID)
	gl.Uniform2f(uniformID, float32(bb.cfg.ScreenWidth), float32(bb.cfg.ScreenHeight))
	gl.UseProgram(0)
}

// String returns the name of the component type
func (bb *BottomBar) String() string {
	return "app.BottomBar"
}
