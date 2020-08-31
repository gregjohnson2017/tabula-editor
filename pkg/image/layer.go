package image

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
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
func (l Layer) Render(view sdl.FRect) {
	fArea := ui.RectToFRect(l.area)
	rect, ok := view.Intersect(&fArea)
	if !ok {
		// not in view
		return
	}

	// update triangles that represent the position and scale of the image (these are SDL/window coordinates, converted to -1,1 gl space coordinates in the vertex shader)
	blx, bly := rect.X-view.X, rect.H+rect.Y-view.Y
	tlx, tly := rect.X-view.X, rect.Y-view.Y
	trx, try := rect.X+rect.W-view.X, rect.Y-view.Y
	brx, bry := rect.X+rect.W-view.X, rect.H+rect.Y-view.Y

	var bls, blt float32 = 0.0, 1.0
	var tls, tlt float32 = 0.0, 0.0
	var trs, trt float32 = 1.0, 0.0
	var brs, brt float32 = 1.0, 1.0

	if rect.X != fArea.X {
		ratio := (rect.X - fArea.X) / fArea.W
		bls = ratio
		tls = ratio
	}

	if rect.X+rect.W != fArea.X+fArea.W {
		ratio := (rect.W + rect.X - fArea.X) / fArea.W
		brs = ratio
		trs = ratio
	}

	if rect.Y != fArea.Y {
		ratio := (rect.Y - fArea.Y) / fArea.H
		tlt = ratio
		trt = ratio
	}

	if rect.Y+rect.H != fArea.Y+fArea.H {
		ratio := (rect.H + rect.Y - fArea.Y) / fArea.H
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
