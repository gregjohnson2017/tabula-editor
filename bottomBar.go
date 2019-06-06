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
	glSquare      []float32
	backVaoID     uint32
	backVboID     uint32
	uniColorID    int32
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements UIComponent
// the background color defaults to grey (0x808080FF)
func NewBottomBar(area *sdl.Rect, comms <-chan imageComm, fontName string, fontSize int32) (*BottomBar, error) {
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
	uniColorID := gl.GetUniformLocation(backProgramID, &[]byte("uni_color")[0])
	gl.UseProgram(backProgramID)
	gl.Uniform4f(uniColorID, color[0], color[1], color[2], color[3])
	gl.UseProgram(0)

	glSquare := []float32{
		-1.0, -1.0, // top-left
		-1.0, +1.0, // bottom-left
		+1.0, +1.0, // bottom-right
		-1.0, -1.0, // top-left
		+1.0, +1.0, // bottom-right
		+1.0, -1.0, // top-right
	}
	var backVaoID uint32
	gl.GenVertexArrays(1, &backVaoID)
	gl.BindVertexArray(backVaoID)
	gl.EnableVertexAttribArray(0)

	var backVboID uint32
	gl.GenBuffers(1, &backVboID)
	gl.BindBuffer(gl.ARRAY_BUFFER, backVboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(glSquare), gl.Ptr(&glSquare[0]), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	return &BottomBar{
		area:          area,
		comms:         comms,
		color:         color,
		font:          font,
		fontSize:      fontSize,
		backProgramID: backProgramID,
		backVaoID:     backVaoID,
		backVboID:     backVboID,
		glSquare:      glSquare,
		uniColorID:    uniColorID,
	}, nil
}

// SetBackgroundColor sets the color for the bottom bar's background texture
func (bb *BottomBar) SetBackgroundColor(color *sdl.Color) error {
	gl.UseProgram(bb.backProgramID)
	gl.Uniform4f(bb.uniColorID, float32(color.R), float32(color.G), float32(color.B), float32(color.A))
	gl.UseProgram(0)
	return nil
}

// Destroy frees all surfaces and textures in the BottomBar
func (bb *BottomBar) Destroy() {
	gl.DeleteBuffers(1, &bb.backVaoID)
	gl.DeleteVertexArrays(1, &bb.backVboID)
}

// GetBoundary returns the clickable region of the UIComponent
func (bb *BottomBar) GetBoundary() *sdl.Rect {
	return bb.area
}

// Render draws the UIComponent
func (bb *BottomBar) Render() error {
	//msg := <-bb.comms

	gl.Viewport(bb.area.X, bb.area.Y, bb.area.W, bb.area.H)
	gl.UseProgram(bb.backProgramID)

	gl.BindVertexArray(bb.backVaoID)
	gl.EnableVertexAttribArray(0)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(bb.glSquare)/2))
	gl.DisableVertexAttribArray(0)
	gl.BindVertexArray(0)

	//// second render white text on top
	//coords := "(" + strconv.Itoa(int(msg.mousePix.x)) + ", " + strconv.Itoa(int(msg.mousePix.y)) + ")"
	//pos := coord{bb.area.W, int32(float64(bb.area.H) / 2.0)}
	//if err = renderText(bb.ctx.Rend, bb.font, bb.fontSize, coords, pos, Align{AlignMiddle, AlignRight}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
	//	return err
	//}
	//pos.x = 0
	//if err = renderText(bb.ctx.Rend, bb.font, bb.fontSize, msg.fileName, pos, Align{AlignMiddle, AlignLeft}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
	//	return err
	//}
	//pos.x = bb.area.W / 2
	//if err = renderText(bb.ctx.Rend, bb.font, bb.fontSize, strconv.FormatFloat(msg.mult, 'f', 4, 64)+"x", pos, Align{AlignMiddle, AlignCenter}, &sdl.Color{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}); err != nil {
	//	return err
	//}

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
}

// String  returns the name of the component type
func (bb *BottomBar) String() string {
	return "BottomBar"
}
