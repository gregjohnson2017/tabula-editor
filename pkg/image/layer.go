package image

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/gfx"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

type Layer struct {
	area      sdl.Rect
	imgVAO    *gfx.VAO
	selVAO    *gfx.VAO
	texture   gfx.Texture
	selTex    gfx.Texture
	selDirty  bool
	setupBuf  *gfx.BufferObject
	offsetBuf *gfx.BufferObject
	vertsBuf  *gfx.BufferObject
	chunkSize int32
	workers   int32
}

func NewLayer(offset sdl.Point, texture gfx.Texture) (*Layer, error) {
	selData := make([]byte, texture.GetWidth()*texture.GetHeight())
	selTex, err := gfx.NewTexture(texture.GetWidth(), texture.GetHeight(), selData, gl.RED, 1, 1)
	if err != nil {
		return nil, err
	}
	var maxWorkers int32
	gl.GetIntegeri_v(gl.MAX_COMPUTE_WORK_GROUP_COUNT, 0, &maxWorkers)
	// TODO calculate intelligent chunk size on a per-image basis
	chunkSize := int32(256)
	workers := texture.GetWidth() * texture.GetHeight() / chunkSize
	if workers > maxWorkers {
		workers = maxWorkers
	} else if workers == 0 {
		workers = 1
	}
	log.Debugf("compute shader workers = %v", workers)

	setupBuf := gfx.NewBufferObject()
	setupBuf.BufferData(gl.SHADER_STORAGE_BUFFER, uint32(4*workers), gl.Ptr(nil), gl.DYNAMIC_READ)
	offsetBuf := gfx.NewBufferObject()
	offsetBuf.BufferData(gl.SHADER_STORAGE_BUFFER, uint32(4*(workers+1)), gl.Ptr(nil), gl.DYNAMIC_READ)
	vertsBuf := gfx.NewBufferObject()
	selVAO := gfx.NewVAO(gl.POINTS, []int32{2})
	selVAO.SetVBO(vertsBuf)

	return &Layer{
		area: sdl.Rect{
			X: offset.X,
			Y: offset.Y,
			W: texture.GetWidth(),
			H: texture.GetHeight(),
		},
		imgVAO:    gfx.NewVAO(gl.TRIANGLES, []int32{2, 2}),
		selVAO:    selVAO,
		texture:   texture,
		selTex:    selTex,
		setupBuf:  setupBuf,
		offsetBuf: offsetBuf,
		vertsBuf:  vertsBuf,
		chunkSize: chunkSize,
		workers:   workers,
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

	err := l.imgVAO.Load(triangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load image triangles: %v", err)
	}

	// draw image
	program.Bind()
	l.texture.Bind()
	l.imgVAO.Draw()
	l.texture.Unbind()
	program.Unbind()

}

func (l *Layer) RenderSelection(view sdl.FRect, program, cs1, cs2, cs3 gfx.Program) {
	fArea := ui.RectToFRect(l.area)
	_, ok := view.Intersect(&fArea)
	if !ok {
		// not in view
		return
	}
	if l.selDirty {
		l.genPointsComputeShader(cs1, cs2, cs3)
		l.selDirty = false
	}
	program.UploadUniform("layerArea", fArea.X, fArea.Y)

	program.Bind()
	l.selTex.Bind()
	l.selVAO.Draw()
	l.selTex.Unbind()
	program.Unbind()
}

func (l *Layer) genPointsComputeShader(cs1, cs2, cs3 gfx.Program) {
	l.setupBuf.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 0)
	l.vertsBuf.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 1)
	l.offsetBuf.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 2)
	cs1.UploadUniformui("chunkSize", uint32(l.chunkSize))
	cs3.UploadUniformui("chunkSize", uint32(l.chunkSize))
	// workers count selections in their respective chunks
	cs1.Bind()
	l.selTex.Bind()
	gl.DispatchCompute(uint32(l.workers), 1, 1)
	l.selTex.Unbind()
	cs1.Unbind()

	gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)
	// parallel prefix algorithm on compute shader
	// to determine worker indices to fill in final answer
	passes := int(math.Ceil(math.Log2(float64(l.workers))))
	for i := 0; i <= passes; i++ {
		cs2.UploadUniformi("pass", int32(i))
		cs2.Bind()
		gl.DispatchCompute(uint32(l.workers), 1, 1)
		cs2.Unbind()

		gl.MemoryBarrier(gl.SHADER_STORAGE_BARRIER_BIT)
	}
	// allocate space for final answer in SSBO
	var sum uint32
	l.offsetBuf.GetSubData(gl.SHADER_STORAGE_BUFFER, 0, 4, gl.Ptr(&sum))
	l.vertsBuf.BufferData(gl.SHADER_STORAGE_BUFFER, 4*sum, gl.Ptr(nil), gl.DYNAMIC_READ)
	// workers fill in final answer in an SSBO, where the final answer
	// is an array of (X,Y) points, where each point is a selected texel
	cs3.Bind()
	l.selTex.Bind()
	gl.DispatchCompute(uint32(l.workers), 1, 1)
	l.selTex.Unbind()
	cs3.Unbind()
	gl.MemoryBarrier(gl.ALL_BARRIER_BITS)
}

func (l *Layer) SelectTexel(p sdl.Point) error {
	if p.X < 0 || p.Y < 0 || p.X >= l.area.W || p.Y >= l.area.H {
		return fmt.Errorf("SelectTexel(%v, %v): %w", p.X, p.Y, ErrCoordOutOfRange)
	}
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
	if r.W > l.area.W || r.H > l.area.H {
		return fmt.Errorf("SelectWorstCase(%v, %v): %w", r.W, r.H, ErrCoordOutOfRange)
	}
	data := make([]byte, r.W*r.H)
	for i := int32(0); i < r.W; i++ {
		for j := int32(0); j < r.H; j++ {
			if i%2 == j%2 {
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
	l.imgVAO.Destroy()
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
	tex, err := gfx.NewTexture(area.W, area.H, texData, gl.RGBA, 4, 4)
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
