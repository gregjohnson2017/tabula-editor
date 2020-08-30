package gfx

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

// BufferArray is used to efficiently render shapes with OpenGL.
type BufferArray struct {
	vaoID      uint32
	vboID      uint32
	mode       uint32
	vertSize   int32
	numAttribs uint32
	data       []float32
}

// NewBufferArray creates the structure necessary for efficiently rendering
// shapes in OpenGL. It configures a VAO & VBO pair with a specified mode and
// vertex layout. Example mode: gl.TRIANGLES. Example vertex layout: (x,y,z,
// s,t) -> layout = (3, 2).
func NewBufferArray(mode uint32, layout []int32) *BufferArray {
	var vaoID, vboID uint32
	gl.GenVertexArrays(1, &vaoID)
	gl.GenBuffers(1, &vboID)
	var vertSize int32
	for _, s := range layout {
		vertSize += s
	}
	configureVAO(vaoID, vboID, layout, vertSize)
	return &BufferArray{vaoID, vboID, mode, vertSize, uint32(len(layout)), nil}
}

// configureVAO configures a VAO & VBO pair with a specified vertex layout
// example vertex layout: (x,y,z, s,t) -> layout = (3, 2)
func configureVAO(vaoID uint32, vboID uint32, layout []int32, vertSize int32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vboID)
	gl.BindVertexArray(vaoID)

	// calculate vertex size in bytes
	// ex: (x,y,z,s,t) -> 5*4 = 20 bytes
	vertexStride := vertSize * 4
	var offset int32
	for i := 0; i < len(layout); i++ {
		gl.VertexAttribPointer(uint32(i), layout[i], gl.FLOAT, false, vertexStride, unsafe.Pointer(uintptr(offset*4)))
		offset += layout[i]
	}

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}

// ErrEmptyData indiciates that the given data is empty.
const ErrEmptyData log.ConstErr = "data is empty so cannot be used"

// Load uploads the given data to the graphics card using the given usage. An
// example usage is gl.STATIC_DRAW.
func (ts *BufferArray) Load(data []float32, usage uint32) error {
	if len(data) == 0 {
		return ErrEmptyData
	}
	ts.data = data
	gl.BindBuffer(gl.ARRAY_BUFFER, ts.vboID)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(ts.data), gl.Ptr(&ts.data[0]), usage)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	return nil
}

// Draw renders the shapes from previously loaded data.
func (ts *BufferArray) Draw() {
	var i uint32
	gl.BindVertexArray(ts.vaoID)
	for i = 0; i < ts.numAttribs; i++ {
		gl.EnableVertexAttribArray(i)
	}
	gl.DrawArrays(ts.mode, 0, int32(len(ts.data))/ts.vertSize)
	for i = 0; i < ts.numAttribs; i++ {
		gl.DisableVertexAttribArray(i)
	}
	gl.BindVertexArray(0)
}

// Destroy frees the resources.
func (ts *BufferArray) Destroy() {
	gl.DeleteBuffers(1, &ts.vboID)
	gl.DeleteVertexArrays(1, &ts.vaoID)
	ts.mode = 0
	ts.vboID = 0
	ts.vaoID = 0
	ts.data = nil
}
