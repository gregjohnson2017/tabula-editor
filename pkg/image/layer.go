package image

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/veandco/go-sdl2/sdl"
)

type Layer struct {
	area    sdl.Rect
	buffer  *gfx.BufferArray
	program gfx.Program
	texture gfx.Texture
}

func NewLayer(offset sdl.Point, program gfx.Program, texture gfx.Texture) (Layer, error) {
	return Layer{
		area: sdl.Rect{
			X: offset.X,
			Y: offset.Y,
			W: texture.GetWidth(),
			H: texture.GetHeight(),
		},
		buffer:  gfx.NewBufferArray(gl.TRIANGLES, []int32{2, 2}),
		program: program,
		texture: texture,
	}, nil
}

// Render draws the ui.Component
func (l Layer) Render() {
	// update triangles that represent the position and scale of the image (these are SDL/window coordinates, converted to -1,1 gl space coordinates in the vertex shader)
	tlx, tly := float32(l.area.X), float32(l.area.Y)
	trx, try := float32(l.area.X+l.area.W), float32(l.area.Y)
	blx, bly := float32(l.area.X), float32(l.area.H+l.area.Y)
	brx, bry := float32(l.area.X+l.area.W), float32(l.area.H+l.area.Y)
	triangles := []float32{
		blx, bly, 0.0, 1.0, // bottom-left
		tlx, tly, 0.0, 0.0, // top-left
		trx, try, 1.0, 0.0, // top-right

		blx, bly, 0.0, 1.0, // bottom-left
		trx, try, 1.0, 0.0, // top-right
		brx, bry, 1.0, 1.0, // bottom-right
	}

	err := l.buffer.Load(triangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load image triangles: %v", err)
	}

	// gl viewport x,y is bottom left
	gl.Viewport(l.area.X, l.area.Y, l.area.W, l.area.H)
	// draw image
	l.program.Bind()
	l.texture.Bind()
	l.buffer.Draw()
	l.texture.Unbind()
	l.program.Unbind()
}
