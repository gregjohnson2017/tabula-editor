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

func (p Program) UploadUniformi(uniformName string, data ...int32) {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		log.Fatalf("\"%s\" is an invalid uniform name", uniformName)
	}
	gl.UseProgram(p.id)
	switch len(data) {
	case 1:
		gl.Uniform1i(uniformID, data[0])
	case 2:
		gl.Uniform2i(uniformID, data[0], data[1])
	case 3:
		gl.Uniform3i(uniformID, data[0], data[1], data[2])
	case 4:
		gl.Uniform4i(uniformID, data[0], data[1], data[2], data[3])
	default:
		log.Fatal("Invalid number of arguments to uploadUniform")
	}
	gl.UseProgram(0)
}

func (p Program) UploadUniformui(uniformName string, data ...uint32) {
	uniformID := gl.GetUniformLocation(p.id, &[]byte(uniformName + "\x00")[0])
	if uniformID == -1 {
		log.Fatalf("\"%s\" is an invalid uniform name", uniformName)
	}
	gl.UseProgram(p.id)
	switch len(data) {
	case 1:
		gl.Uniform1uiEXT(uniformID, data[0])
	case 2:
		gl.Uniform2uiEXT(uniformID, data[0], data[1])
	case 3:
		gl.Uniform3uiEXT(uniformID, data[0], data[1], data[2])
	case 4:
		gl.Uniform4uiEXT(uniformID, data[0], data[1], data[2], data[3])
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

// Destroy destroys OpenGL assets associated with the Shader
func (p Program) Destroy() {
	gl.DeleteProgram(p.id)
}

type ShaderInfo struct {
	Source     string
	ShaderType uint32
}

var sharedPrograms map[Program][]ShaderInfo

func doShadersMatch(arr1, arr2 []ShaderInfo) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for _, s1 := range arr1 {
		match := false
		for _, s2 := range arr2 {
			if s1 == s2 {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}

func GetSharedProgram(requested ...ShaderInfo) (Program, error) {
	if sharedPrograms == nil {
		sharedPrograms = make(map[Program][]ShaderInfo)
	}
	for prog, shaders := range sharedPrograms {
		if doShadersMatch(shaders, requested) {
			return prog, nil
		}
	}
	shaderList := make([]Shader, len(requested))
	for i, request := range requested {
		shader, err := NewShader(request.Source, request.ShaderType)
		if err != nil {
			return Program{}, err
		}
		shaderList[i] = shader
	}
	newProg, err := NewProgram(shaderList...)
	if err != nil {
		return Program{}, err
	}
	sharedPrograms[newProg] = requested
	return newProg, nil
}
