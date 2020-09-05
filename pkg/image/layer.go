package image

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	set "github.com/kroppt/Int32Set"
	"github.com/veandco/go-sdl2/sdl"
)

type Layer struct {
	area     sdl.Rect
	buffer   *gfx.BufferArray
	selBuf   *gfx.BufferArray
	texture  gfx.Texture
	selTex   gfx.Texture
	selSet   set.Set
	selDirty bool
	selData  []float32
}

func NewLayer(offset sdl.Point, texture gfx.Texture) (*Layer, error) {
	selData := make([]byte, texture.GetWidth()*texture.GetHeight())
	selTex, err := gfx.NewTexture(texture.GetWidth(), texture.GetHeight(), selData, gl.RED, 1)
	if err != nil {
		return nil, err
	}
	return &Layer{
		area: sdl.Rect{
			X: offset.X,
			Y: offset.Y,
			W: texture.GetWidth(),
			H: texture.GetHeight(),
		},
		buffer:  gfx.NewBufferArray(gl.TRIANGLES, []int32{2, 2}),
		selBuf:  gfx.NewBufferArray(gl.POINTS, []int32{2}),
		texture: texture,
		selTex:  selTex,
		selSet:  set.NewSet(),
	}, nil
}

func (l Layer) Render(view sdl.FRect, program gfx.Program) {
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
	program.Bind()
	l.texture.Bind()
	l.buffer.Draw()
	l.texture.Unbind()
	program.Unbind()

}

func (l *Layer) RenderSelection(view sdl.FRect, program gfx.Program) {
	fArea := ui.RectToFRect(l.area)
	_, ok := view.Intersect(&fArea)
	if !ok {
		// not in view
		return
	}
	// selections
	// *2 for x,y
	if l.selDirty {
		sw := util.Start()
		l.selData = make([]float32, 0, l.selSet.Size()*2)
		l.selSet.Range(func(i int32) bool {
			// i is every y*width+x index
			texelX := float32(i % l.area.W)
			texelY := float32((float32(i) - texelX) / float32(l.area.W))
			l.selData = append(l.selData, texelX, texelY)
			return true
		})
		l.selDirty = false
		sw.StopRecordAverage("selection set")
	}

	if len(l.selData) == 0 {
		return
	}
	program.UploadUniform("layerArea", fArea.X, fArea.Y)
	err := l.selBuf.Load(l.selData, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load selection points: %v", err)
	}

	program.Bind()
	l.selTex.Bind()
	glq := util.StartGLQuery()
	l.selBuf.Draw()
	glq.Stop("selection shaders")
	l.selTex.Unbind()
	program.Unbind()
}

func (l *Layer) SelectTexel(p sdl.Point) error {
	if p.X < 0 || p.Y < 0 || p.X >= l.area.W || p.Y >= l.area.H {
		return fmt.Errorf("SelectTexel(%v, %v): %w", p.X, p.Y, ErrCoordOutOfRange)
	}
	l.selSet.Add(p.X + p.Y*l.area.W)
	l.selDirty = true
	return l.selTex.SetPixel(p, []byte{1}, false)
}

func (l *Layer) SelectRegion(r sdl.Rect) error {
	if r.X < 0 || r.Y < 0 || r.X >= l.area.W || r.Y >= l.area.H {
		return fmt.Errorf("SelectRegion(%v, %v, %v, %v): %w", r.X, r.Y, r.W, r.H, ErrCoordOutOfRange)
	}
	if r.W > l.area.W || r.H > l.area.H {
		return fmt.Errorf("SelectRegion(%v, %v, %v, %v): %w", r.X, r.Y, r.W, r.H, ErrCoordOutOfRange)
	}
	for i := r.X; i < r.X+r.W; i++ {
		for j := r.Y; j < r.Y+r.H; j++ {
			l.selSet.Add(i + j*l.area.W)
		}
	}
	data := make([]byte, r.W*r.H)
	for i := range data {
		data[i] = 1
	}
	l.selDirty = true
	return l.selTex.SetPixelArea(r, data, false)
}

func (l *Layer) SelectWorstCase() error {
	r := sdl.Rect{
		X: 0,
		Y: 0,
		W: l.area.W,
		H: l.area.H,
	}
	if r.X < 0 || r.Y < 0 || r.X >= l.area.W || r.Y >= l.area.H {
		return fmt.Errorf("SelectRegion(%v, %v, %v, %v): %w", r.X, r.Y, r.W, r.H, ErrCoordOutOfRange)
	}
	if r.W > l.area.W || r.H > l.area.H {
		return fmt.Errorf("SelectRegion(%v, %v, %v, %v): %w", r.X, r.Y, r.W, r.H, ErrCoordOutOfRange)
	}
	data := make([]byte, r.W*r.H)
	for i := r.X; i < r.X+r.W; i++ {
		for j := r.Y; j < r.Y+r.H; j++ {
			if i%2 == j%2 {
				l.selSet.Add(i + j*l.area.W)
				data[i+j*l.area.W] = 1
			}
		}
	}
	l.selDirty = true
	return l.selTex.SetPixelArea(r, data, false)
}

func (l Layer) GetSelTex() gfx.Texture {
	return l.selTex
}

// Destroy destroys OpenGL assets associated with the Layer
func (l Layer) Destroy() {
	l.buffer.Destroy()
	l.texture.Destroy()
}

// MarshalBinary fulfills a requirement for gob to encode Layer
func (l Layer) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var err error
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(l.area); err != nil {
		return nil, err
	}
	if err = enc.Encode(l.texture.GetData()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary fulfills a requirement for gob to decode Layer
func (l *Layer) UnmarshalBinary(data []byte) error {
	var err error
	dec := gob.NewDecoder(bytes.NewReader(data))

	var area sdl.Rect
	if err = dec.Decode(&area); err != nil {
		return err
	}

	var texData = make([]byte, area.W*area.H*4)
	if err = dec.Decode(&texData); err != nil {
		return err
	}
	tex, err := gfx.NewTexture(area.W, area.H, texData, gl.RGBA, 4)
	if err != nil {
		return err
	}
	tex.SetParameter(gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	tex.SetParameter(gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	layer, err := NewLayer(sdl.Point{X: area.X, Y: area.Y}, tex)
	if err != nil {
		return err
	}

	*l = *layer
	return nil
}
