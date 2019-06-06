package main

import (
	"strconv"
	"unsafe"

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
	glSquare      []float32
	glTexSquare   []float32
	backVaoID     uint32
	backVboID     uint32
	textVaoID     uint32
	textVboID     uint32
	uniColorID    int32
	textureID     uint32
	cfg           *config
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
// the background color defaults to grey (0x808080FF)
func NewBottomBar(area *sdl.Rect, comms <-chan imageComm, fontName string, fontSize int32, cfg *config) (*BottomBar, error) {
	var m float32 = 255.0
	color := [4]float32{128.0 / m, 128.0 / m, 128.0 / m, 1.0}
	var err error
	var font *ttf.Font
	if font, err = ttf.OpenFont(fontName, int(fontSize)); err != nil {
		return nil, err
	}
	var backProgramID uint32
	if backProgramID, err = CreateShaderProgram(solidColorVertex, solidColorFragment); err != nil {
		return nil, err
	}
	var textProgramID uint32
	if textProgramID, err = CreateShaderProgram(vshTexturePassthrough, fragmentShaderSource); err != nil {
		return nil, err
	}
	uniColorID := gl.GetUniformLocation(backProgramID, &[]byte("uni_color")[0])
	gl.UseProgram(backProgramID)
	gl.Uniform4f(uniColorID, color[0], color[1], color[2], color[3])
	gl.UseProgram(0)

	glSquare := []float32{
		-1.0, -1.0, // bottom-left
		-1.0, +1.0, // top-left
		+1.0, +1.0, // top-right
		-1.0, -1.0, // bottom-left
		+1.0, +1.0, // top-right
		+1.0, -1.0, // bottom-right
	}
	glTexSquare := []float32{
		-1.0, -1.0, 0.0, 1.0, // bottom-left
		-1.0, +1.0, 0.0, 0.0, // top-left
		+1.0, +1.0, 1.0, 0.0, // top-right
		-1.0, -1.0, 0.0, 1.0, // bottom-left
		+1.0, +1.0, 1.0, 0.0, // top-right
		+1.0, -1.0, 1.0, 1.0, // bottom-right
	}

	backVaoID, backVboID := bufferData(glSquare, []int32{2})
	textVaoID, textVboID := bufferData(glTexSquare, []int32{2, 2})

	return &BottomBar{
		area:          area,
		comms:         comms,
		color:         color,
		font:          font,
		fontSize:      fontSize,
		backProgramID: backProgramID,
		textProgramID: textProgramID,
		backVaoID:     backVaoID,
		backVboID:     backVboID,
		textVaoID:     textVaoID,
		textVboID:     textVboID,
		glSquare:      glSquare,
		glTexSquare:   glTexSquare,
		uniColorID:    uniColorID,
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
	gl.DeleteBuffers(1, &bb.backVaoID)
	gl.DeleteVertexArrays(1, &bb.backVboID)
	gl.DeleteBuffers(1, &bb.textVaoID)
	gl.DeleteVertexArrays(1, &bb.textVboID)
	bb.font.Close()
}

// GetBoundary returns the clickable region of the UIComponent
func (bb *BottomBar) GetBoundary() *sdl.Rect {
	return bb.area
}

// Render draws the UIComponent
func (bb *BottomBar) Render() error {
	msg := <-bb.comms

	// first render solid color background
	gl.Viewport(bb.area.X, 0, bb.area.W, bb.area.H)
	gl.UseProgram(bb.backProgramID)

	gl.BindVertexArray(bb.backVaoID)
	gl.EnableVertexAttribArray(0)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(bb.glSquare)/2))
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)

	// second render white text on top
	format := int32(gl.RGBA)
	gl.UseProgram(bb.textProgramID)
	gl.GenTextures(1, &bb.textureID)
	gl.BindTexture(gl.TEXTURE_2D, bb.textureID)
	gl.BindVertexArray(bb.textVaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)

	// draw x,y image pixel coordinates on right
	coords := "(" + strconv.Itoa(int(msg.mousePix.x)) + ", " + strconv.Itoa(int(msg.mousePix.y)) + ")"
	pos := coord{bb.area.W, int32(float64(bb.area.H) / 2.0)}
	slice, rect, err := renderText(bb.font, bb.fontSize, coords, pos, Align{AlignMiddle, AlignRight}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, bb.area.H)
	if err != nil {
		return err
	}
	gl.Viewport(rect.X, rect.Y, rect.W, rect.H)
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, rect.W, rect.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&slice[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(bb.glTexSquare)/4))
	gl.DeleteTextures(1, &bb.textureID)

	// draw image filename on left
	pos.x = 0
	slice, rect, err = renderText(bb.font, bb.fontSize, msg.fileName, pos, Align{AlignMiddle, AlignLeft}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, bb.area.H)
	if err != nil {
		return err
	}
	gl.Viewport(rect.X, rect.Y, rect.W, rect.H)
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, rect.W, rect.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&slice[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(bb.glTexSquare)/4))
	gl.DeleteTextures(1, &bb.textureID)

	// draw zoom multiplier in middle
	pos.x = bb.area.W / 2
	slice, rect, err = renderText(bb.font, bb.fontSize, strconv.FormatFloat(msg.mult, 'f', 4, 64)+"x", pos, Align{AlignMiddle, AlignCenter}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, bb.area.H)
	if err != nil {
		return err
	}
	gl.Viewport(rect.X, rect.Y, rect.W, rect.H)
	gl.TexImage2D(gl.TEXTURE_2D, 0, format, rect.W, rect.H, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&slice[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(bb.glTexSquare)/4))
	gl.DeleteTextures(1, &bb.textureID)

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
}

// String  returns the name of the component type
func (bb *BottomBar) String() string {
	return "BottomBar"
}
