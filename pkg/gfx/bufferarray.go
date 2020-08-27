package gfx

import (
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

// BufferArray is used to efficiently render shapes with OpenGL.
type BufferArray struct {
	vaoID uint32
	vboID uint32
	mode  uint32
	size  int32
	data  []float32
}

// NewBufferArray creates the structure necessary for efficiently rendering
// shapes in OpenGL. It configures a VAO & VBO pair with a specified mode and
// vertex layout. Example mode: gl.TRIANGLES. Example vertex layout: (x,y,z,
// s,t) -> layout = (3, 2).
func NewBufferArray(mode uint32, layout []int32) *BufferArray {
	var vaoID, vboID uint32
	gl.GenVertexArrays(1, &vaoID)
	gl.GenBuffers(1, &vboID)
	ConfigureVAO(vaoID, vboID, layout)
	var size int32
	for _, i := range layout {
		size += i
	}
	return &BufferArray{vaoID, vboID, mode, size, nil}
}

// Bind binds its GL primitives. This is required for Load and Draw.
func (ts *BufferArray) Bind() {
	gl.BindBuffer(gl.ARRAY_BUFFER, ts.vboID)
	gl.BindVertexArray(ts.vaoID)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
}

// Unbind unbinds its GL primitives. This is recommended to help reduce bugs.
func (ts *BufferArray) Unbind() {
	gl.DisableVertexAttribArray(1)
	gl.DisableVertexAttribArray(0)
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
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(ts.data), gl.Ptr(&ts.data[0]), gl.STATIC_DRAW)
	return nil
}

// Draw renders the shapes from previously loaded data.
func (ts *BufferArray) Draw() {
	gl.DrawArrays(ts.mode, 0, int32(len(ts.data))/ts.size)
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
