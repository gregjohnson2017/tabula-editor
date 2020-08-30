package image

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/veandco/go-sdl2/sdl"
)

type Layer struct {
	area    sdl.Rect
	origW   int32
	origH   int32
	buffer  *gfx.BufferArray
	texture gfx.Texture
}

func NewLayer(offset sdl.Point, texture gfx.Texture) *Layer {
	return &Layer{
		area: sdl.Rect{
			X: offset.X,
			Y: offset.Y,
			W: texture.GetWidth(),
			H: texture.GetHeight(),
		},
		origW:   texture.GetWidth(),
		origH:   texture.GetHeight(),
		buffer:  gfx.NewBufferArray(gl.TRIANGLES, []int32{2, 2}),
		texture: texture,
	}
}

// Render draws the ui.Component
func (l Layer) Render(view sdl.Rect) {
	rect, ok := view.Intersect(&l.area)
	if !ok {
		// not in view
		return
	}

	// update triangles that represent the position and scale of the image (these are SDL/window coordinates, converted to -1,1 gl space coordinates in the vertex shader)
	blx, bly := float32(rect.X-view.X), float32(rect.H+rect.Y-view.Y)
	tlx, tly := float32(rect.X-view.X), float32(rect.Y-view.Y)
	trx, try := float32(rect.X+rect.W-view.X), float32(rect.Y-view.Y)
	brx, bry := float32(rect.X+rect.W-view.X), float32(rect.H+rect.Y-view.Y)

	var bls, blt float32 = 0.0, 1.0
	var tls, tlt float32 = 0.0, 0.0
	var trs, trt float32 = 1.0, 0.0
	var brs, brt float32 = 1.0, 1.0

	if rect.X != l.area.X {
		ratio := float32(rect.X-l.area.X) / float32(l.area.W)
		bls = ratio
		tls = ratio
	}

	if rect.X+rect.W != l.area.X+l.area.W {
		ratio := float32(rect.W+rect.X-l.area.X) / float32(l.area.W)
		brs = ratio
		trs = ratio
	}

	if rect.Y != l.area.Y {
		ratio := float32(rect.Y-l.area.Y) / float32(l.area.H)
		tlt = ratio
		trt = ratio
	}

	if rect.Y+rect.H != l.area.Y+l.area.H {
		ratio := float32(rect.H+rect.Y-l.area.Y) / float32(l.area.H)
		blt = ratio
		brt = ratio
	}

	triangles := []float32{
		blx, bly, bls, blt, // bottom-left
		tlx, tly, tls, tlt, // top-left
		trx, try, trs, trt, // top-right

		blx, bly, bls, blt, // bottom-left
		trx, try, trs, trt, // top-right
		brx, bry, brs, brt, // bottom-right
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
