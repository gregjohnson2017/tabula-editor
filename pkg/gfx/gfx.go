package gfx

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

// ConfigureVAO configures a VAO & VBO pair with a specified vertex layout
// example vertex layout: (x,y,z, s,t) -> layout = (3, 2)
// TODO move to NewBufferArray
func ConfigureVAO(vaoID uint32, vboID uint32, layout []int32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vboID)
	gl.BindVertexArray(vaoID)
	var vertexSize int32
	for i := 0; i < len(layout); i++ {
		vertexSize += layout[i]
	}
	// calculate vertex size in bytes
	// ex: (x,y,z,s,t) -> 5*4 = 20 bytes
	vertexStride := vertexSize * 4
	var offset int32
	for i := 0; i < len(layout); i++ {
		gl.VertexAttribPointer(uint32(i), layout[i], gl.FLOAT, false, vertexStride, unsafe.Pointer(uintptr(offset*4)))
		offset += layout[i]
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}
