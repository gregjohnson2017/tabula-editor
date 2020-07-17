package gfx

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

// ConfigureVAO configures a VAO & VBO pair with a specified vertex layout
// example vertex layout: (x,y,z, s,t) -> layout = (3, 2)
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

// ErrProgramLink indicates that a program failed to link
const ErrProgramLink log.ConstErr = "failed to compile shader"

// CreateShaderProgram compiles a vertex and fragment shader,
// attaches them to a new shader program and returns its ID.
func CreateShaderProgram(vshStr, fshStr string) (uint32, error) {
	prog := gl.CreateProgram()
	vsh, err := compileShader(vshStr, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fsh, err := compileShader(fshStr, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}
	gl.AttachShader(prog, vsh)
	gl.AttachShader(prog, fsh)
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("%w: %v", ErrProgramLink, log)
	}

	return prog, nil
}

// ErrCompileShader indicates that a shader failed to compile
const ErrCompileShader log.ConstErr = "failed to compile shader"

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("%w: %v", ErrCompileShader, log)
	}

	return shader, nil
}
