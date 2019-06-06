package main

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// UIComponent says what functions a UIComponent must implement
type UIComponent interface {
	GetBoundary() *sdl.Rect
	Render() error
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
	// AlignBelow puts the top side at the y coordinate
	AlignBelow AlignV = iota - 1
	// AlignMiddle puts the top and bottom sides equidistant from the middle
	AlignMiddle
	// AlignAbove puts the bottom side on the y coordinate
	AlignAbove
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
	v AlignV
	h AlignH
}

type coord struct {
	x int32
	y int32
}

func createSolidColorTexture(rend *sdl.Renderer, w int32, h int32, r uint8, g uint8, b uint8, a uint8) (*sdl.Texture, error) {
	var surf *sdl.Surface
	var err error
	if surf, err = sdl.CreateRGBSurfaceWithFormat(0, w, h, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
		return nil, err
	}
	if err = surf.FillRect(nil, mapRGBA(surf.Format, r, g, b, a)); err != nil {
		return nil, err
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTextureFromSurface(surf); err != nil {
		return nil, err
	}
	surf.Free()
	return tex, nil
}

func renderText(rend *sdl.Renderer, font *ttf.Font, fontSize int32, text string, pos coord, align Align, col *sdl.Color) error {
	var surf *sdl.Surface
	var err error
	if surf, err = font.RenderUTF8Blended(text, *col); err != nil {
		return err
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTexture(surf.Format.Format, sdl.TEXTUREACCESS_STREAMING, surf.W, int32(fontSize)); err != nil {
		surf.Free()
		return err
	}
	if err = tex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		surf.Free()
		tex.Destroy()
		return err
	}

	h, err := font.GlyphMetrics('h')
	rowsFromTop := int32(font.Ascent() - h.MaxY)
	sliceStart := surf.Pitch * (rowsFromTop)
	rowsFromBottom := surf.H - fontSize - rowsFromTop
	sliceStop := surf.Pitch * (surf.H - rowsFromBottom)
	copyToTexture(tex, surf.Pixels()[sliceStart:sliceStop], nil)

	w2 := int32(float64(surf.W) / 2.0)
	h2 := int32(float64(fontSize) / 2.0)
	offx := -w2 - int32(align.h)*int32(w2)
	offy := -h2 - int32(align.v)*int32(h2)
	var rect = &sdl.Rect{
		X: pos.x + offx,
		Y: pos.y + offy,
		W: int32(surf.W),
		H: int32(fontSize),
	}

	err = rend.Copy(tex, nil, rect)
	surf.Free()
	tex.Destroy()
	return err
}

func mapRGBA(form *sdl.PixelFormat, r, g, b, a uint8) uint32 {
	ur := uint32(r)
	ur |= ur<<8 | ur<<16 | ur<<24
	ug := uint32(g)
	ug |= ug<<8 | ug<<16 | ug<<24
	ub := uint32(b)
	ub |= ub<<8 | ub<<16 | ub<<24
	ua := uint32(a)
	ua |= ua<<8 | ua<<16 | ua<<24
	return form.Rmask&ur |
		form.Gmask&ug |
		form.Bmask&ub |
		form.Amask&ua
}

func setPixel(surf *sdl.Surface, p coord, c sdl.Color) {
	d := mapRGBA(surf.Format, c.R, c.G, c.B, c.A)
	bs := []byte{byte(d), byte(d >> 8), byte(d >> 16), byte(d >> 24)}
	i := int32(surf.BytesPerPixel())*p.x + surf.Pitch*p.y
	copy(surf.Pixels()[i:], bs)
}

func copyToTexture(tex *sdl.Texture, pixels []byte, region *sdl.Rect) error {
	var bytes []byte
	var err error
	bytes, _, err = tex.Lock(region)
	copy(bytes, pixels)
	tex.Unlock()
	return err
}

func loadImage(fileName string) (*sdl.Surface, error) {
	var surf *sdl.Surface
	var err error
	if surf, err = img.Load(fileName); err != nil {
		return nil, err
	}
	// TODO check bytes per pixel == 4 and convert if necessary
	return surf, err
}

func makeVAO(points []float32) (uint32, uint32) {
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)

	var vbo uint32
	var vertexStride int32 = 4 * 4
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, vertexStride, nil)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, vertexStride, unsafe.Pointer(uintptr(2*4)))

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindVertexArray(0)
	return vao, vbo
}

const (
	solidColorVertex = `
	#version 460
	uniform vec4 uni_color;
	in vec2 position_in;
	out vec4 color;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		color = uni_color;
	}

` + "\x00"
	solidColorFragment = `
	#version 460
	in vec4 color;
	out vec4 frag_color;
	void main() {
		frag_color = color;
	}
` + "\x00"

	vertexShaderSource = `
	#version 460
	// uniform vec2 screenSize;
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	// TODO add colors
	out vec2 tex_coords;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		tex_coords = tex_coords_in;
	}
` + "\x00"

	fragmentShaderSource = `
	#version 460
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = texture(frag_tex, tex_coords);
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
