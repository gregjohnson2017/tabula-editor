package ui

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/sdl"
)

// Component describes which functions a UI component must implement
type Component interface {
	InBoundary(sdl.Point) bool
	GetBoundary() *sdl.Rect
	Render()
	Destroy()
	OnEnter()
	OnLeave()
	OnMotion(*sdl.MouseMotionEvent) bool
	OnScroll(*sdl.MouseWheelEvent) bool
	OnClick(*sdl.MouseButtonEvent) bool
	OnResize(x, y int32)
	fmt.Stringer
}

// AlignV is used for the positioning of elements vertically
type AlignV int

const (
	// AlignAbove puts the top side at the y coordinate
	AlignAbove AlignV = iota - 1
	// AlignMiddle puts the top and bottom sides equidistant from the middle
	AlignMiddle
	// AlignBelow puts the bottom side on the y coordinate
	AlignBelow
)

// AlignH is used for the positioning of elements horizontally
type AlignH int

const (
	// AlignLeft puts the left side on the x coordinate
	AlignLeft AlignH = iota - 1
	//AlignCenter puts the left and right sides equidistant from the center
	AlignCenter
	// AlignRight puts the right side at the x coordinate
	AlignRight
)

// Align holds vertical and horizontal alignments
type Align struct {
	V AlignV
	H AlignH
}

func WriteRuneToFile(fileName string, mask image.Image, maskp image.Point, rec image.Rectangle) error {
	if alpha, ok := mask.(*image.Alpha); ok {
		diff := image.Point{rec.Dx(), rec.Dy()}
		tofile := alpha.SubImage(image.Rectangle{maskp, maskp.Add(diff)})
		if f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755); err != nil {
			png.Encode(f, tofile)
			return err
		}
	}
	return nil
}

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

const (
	OutlineVsh = `
	#version 460
	uniform vec4 uni_color;
	uniform vec2 origDims;
	uniform float mult;
	in vec2 position_in;
	out vec4 color;
	void main() {
		vec2 canvasArea = mult * origDims;
		vec2 pos = vec2(mult * position_in.x, canvasArea.y - mult * position_in.y);
		vec2 glSpace = vec2(2.0, 2.0) * (pos / canvasArea) + vec2(-1.0, -1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
		color = uni_color;
	}
` + "\x00"

	OutlineFsh = `
	#version 460
	in vec4 color;
	out vec4 frag_color;
	void main() {
		float scale = 4.0;
		float mx = floor(mod(gl_FragCoord.x / scale, 2.0));
		float my = floor(mod(gl_FragCoord.y / scale, 2.0));
		vec4 col1 = vec4(1.0, 1.0, 1.0, 1.0);
		vec4 col2 = vec4(0.3, 0.3, 0.3, 1.0);
		vec4 checker = mx == my ? col1 : col2;
		frag_color = checker;
	}
` + "\x00"

	SolidColorVertex = `
	#version 460
	uniform vec4 uni_color;
	in vec2 position_in;
	out vec4 color;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		color = uni_color;
	}
` + "\x00"

	SolidColorFragment = `
	#version 460
	in vec4 color;
	out vec4 frag_color;
	void main() {
		frag_color = color;
	}
` + "\x00"

	VertexShaderSource = `
	#version 460
	uniform vec2 area;
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		vec2 glSpace = vec2(2.0, -2.0) * (position_in / area) + vec2(-1.0, 1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
		tex_coords = tex_coords_in;
	}
` + "\x00"

	FragmentShaderSource = `
	#version 460
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = texture(frag_tex, tex_coords);
	}
` + "\x00"

	VshTexturePassthrough = `
	#version 460
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		tex_coords = tex_coords_in;
	}
` + "\x00"

	// Uniform `tex_size` is the (width, height) of the texture.
	// Input `position_in` is typical openGL position coordinates.
	// Input `tex_pixels` is the (x, y) of the vertex in the texture starting
	// at (left, top).
	// Output `tex_coords` is typical texture coordinates for fragment shader.
	GlyphShaderVertex = `
	#version 460
	uniform vec2 tex_size;
	uniform vec2 screen_size;
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_pixels;
	out vec2 tex_coords;
	void main() {
		vec2 glSpace = vec2(2.0, 2.0) * (position_in / screen_size) + vec2(-1.0, -1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
		tex_coords = vec2(tex_pixels.x / tex_size.x, tex_pixels.y / tex_size.y);
	}
` + "\x00"

	GlyphShaderFragment = `
	#version 460
	uniform sampler2D frag_tex;
	uniform vec4 text_color;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = vec4(text_color.xyz, texture(frag_tex, tex_coords).r * text_color.w);
	}
` + "\x00"

	CheckerShaderFragment = `
	#version 460
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		float scale = 10.0;
		float mx = floor(mod(gl_FragCoord.x / scale, 2.0));
		float my = floor(mod(gl_FragCoord.y / scale, 2.0));
		vec4 col1 = vec4(1.0, 1.0, 1.0, 1.0);
		vec4 col2 = vec4(0.7, 0.7, 0.7, 1.0);
		vec4 checker = mx == my ? col1 : col2;
		vec4 tex = texture(frag_tex, tex_coords);
		frag_color = mix(checker, tex, tex.a);
	}
` + "\x00"
)

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

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

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

		return 0, fmt.Errorf("failed to compile program: %v", log)
	}

	return prog, nil
}

func InBounds(area sdl.Rect, point sdl.Point) bool {
	if point.X < area.X || point.X >= area.X+area.W || point.Y < area.Y || point.Y >= area.Y+area.H {
		return false
	}
	return true
}
