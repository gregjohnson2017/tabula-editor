package image

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/veandco/go-sdl2/sdl"
)

type Layer struct {
	area         sdl.Rect
	origW, origH int32
	buffer       *gfx.BufferArray
	texture      gfx.Texture
}

func NewLayer(offset sdl.Point, texture gfx.Texture, mult float64) *Layer {
	return &Layer{
		area: sdl.Rect{
			X: int32(float64(offset.X) * mult),
			Y: int32(float64(offset.Y) * mult),
			W: int32(float64(texture.GetWidth()) * mult),
			H: int32(float64(texture.GetHeight()) * mult),
		},
		origW:   texture.GetWidth(),
		origH:   texture.GetHeight(),
		buffer:  gfx.NewBufferArray(gl.TRIANGLES, []int32{2, 2}),
		texture: texture,
	}
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

	// draw image
	l.texture.Bind()
	l.buffer.Draw()
	l.texture.Unbind()
}

func (l Layer) Destroy() {
	l.buffer.Destroy()
	l.texture.Destroy()
}
