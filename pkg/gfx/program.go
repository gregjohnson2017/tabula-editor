package gfx

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

type Program struct {
	id uint32
}

// ErrProgramLink indicates that a program failed to link
const ErrProgramLink log.ConstErr = "failed to compile shader"

// CreateShaderProgram compiles a vertex and fragment shader,
// attaches them to a new shader program and returns its ID.
func NewProgram(shaders ...Shader) (Program, error) {
	prog := gl.CreateProgram()
	for _, shader := range shaders {
		gl.AttachShader(prog, shader.id)
	}
	gl.LinkProgram(prog)

	var status int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)

		log := string(make([]byte, logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(log))

		return Program{}, fmt.Errorf("%w: %v", ErrProgramLink, log)
	}

	return Program{prog}, nil
}

// UploadUniform uploads float32 data in the given uniform variable
// belonging to the given program ID.
//
// If data does not contain between 1 and 4 arguments (inclusive),
// UploadUniform will panic.
func (p Program) UploadUniform(uniformName string, data ...float32) {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		log.Fatalf("\"%s\" is an invalid uniform name", uniformName)
	}
	gl.UseProgram(p.id)
	switch len(data) {
	case 1:
		gl.Uniform1f(uniformID, data[0])
	case 2:
		gl.Uniform2f(uniformID, data[0], data[1])
	case 3:
		gl.Uniform3f(uniformID, data[0], data[1], data[2])
	case 4:
		gl.Uniform4f(uniformID, data[0], data[1], data[2], data[3])
	default:
		log.Fatal("Invalid number of arguments to uploadUniform")
	}
	gl.UseProgram(0)
}

// Bind makes OpenGL use this program
func (p Program) Bind() {
	gl.UseProgram(p.id)
}

// Unbind sets the current program ID to 0
func (p Program) Unbind() {
	gl.UseProgram(0)
}

// Delete tells OpenGL to delete the program ID
func (p Program) Delete() {
	gl.DeleteProgram(p.id)
}
