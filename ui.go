package main

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/golang/freetype/truetype"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"golang.org/x/image/math/fixed"
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

func writeRuneToFile(fileName string, mask image.Image, maskp image.Point, rec image.Rectangle) error {
	if alpha, ok := mask.(*image.Alpha); ok {
		tofile := alpha.SubImage(image.Rectangle{maskp, maskp.Add(image.Point{rec.Dx(), rec.Dy()})})
		if f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755); err != nil {
			png.Encode(f, tofile)
			return err
		}
	}
	return nil
}

func printRune(mask image.Image, maskp image.Point, rec image.Rectangle) {
	if alpha, ok := mask.(*image.Alpha); ok {
		for y := maskp.Y; y < maskp.Y+rec.Dy(); y++ {
			for x := maskp.X; x < maskp.X+rec.Dx(); x++ {
				if _, _, _, a := alpha.At(x, y).RGBA(); a > 0 {
					fmt.Printf("%02x ", byte(a))
				} else {
					fmt.Printf(".  ")
				}
			}
			fmt.Printf("\n")
		}
	}
}

func int26_6ToFloat32(x fixed.Int26_6) float32 {
	top := float32(x >> 6)
	bottom := float32(x&0x3F) / 64.0
	return top + bottom
}

type runeInfo struct {
	index   int32
	width   int32
	height  int32
	bearing float32
	advance float32
}

// mapString turns each character in the string into a pair of
// (x,y,s,t)-vertex triangles using glyph information from a
// pre-loaded font. This information is stored in a float32
// array and returned with the total width and height for OpenGL.
func mapString(str string, runeMap []runeInfo) ([]float32, int32, int32) {
	// topLeft := (int32(origin.X + bearingX), origin.Y - rect.MinY)
	// topRight := (topLeft.X + rectDx(), topLeft.Y)
	// bottomLeft := (topLeft.X, origin.Y - rect.MaxY)
	// bottomRight := (topRight.X, bottomLeft.Y)

	// greg's calcs
	// topLeft := (rect.MinX, origin.Y - rect.MinY)
	// topRight := (rect.MaxX, topLeft.Y)
	// bottomLeft := (topLeft.X, origin.Y - rect.MaxY)
	// bottomRight := (topRight.X, bottomLeft.Y)
	return nil, 0, 0
}

// loadFontTexture caches all of the glyph pixel data in an OpenGL texture for
// a given font at a given size. It returns the OpenGL ID for this texture,
// along with a runeInfo array for indexing into the texture per rune at runtime
func loadFontTexture(fontName string, fontSize int32) (uint32, []runeInfo, error) {
	// sw := start()

	var err error
	var fontBytes []byte
	var font *truetype.Font
	if fontBytes, err = ioutil.ReadFile(fontName); err != nil {
		panic(err)
	}
	if font, err = truetype.Parse(fontBytes); err != nil {
		panic(err)
	}
	face := truetype.NewFace(font, &truetype.Options{Size: float64(fontSize)})

	var runeMap [unicode.MaxASCII - minASCII]runeInfo
	var glyphBytes []byte
	var currentIndex int32
	for i := minASCII; i < unicode.MaxASCII; i++ {
		c := rune(i)

		roundedRect, mask, maskp, advance, glyphOK := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, c)
		accurateRect, _, glyphBoundsOK := face.GlyphBounds(c)
		glyph, castOK := mask.(*image.Alpha)
		if !glyphOK || !glyphBoundsOK || !castOK {
			return 0, nil, fmt.Errorf("%v does not contain glyph for '%c'", fontName, c)
		}

		runeMap[i-minASCII] = runeInfo{
			index:   currentIndex,
			width:   int32(roundedRect.Dx()),
			height:  int32(roundedRect.Dy()),
			bearing: int26_6ToFloat32(accurateRect.Min.X),
			advance: int26_6ToFloat32(advance),
		}

		// alternatively, upload entire glyph cache into OpenGL texture
		// ... but this doesnt take that long and cuts texture size by 95%
		for row := 0; row < roundedRect.Dy(); row++ {
			glyphBytes = append(glyphBytes, glyph.Pix[(maskp.Y+row)*glyph.Stride:(maskp.Y+row+1)*glyph.Stride]...)
			currentIndex += int32(glyph.Stride)
		}

		// fmt.Printf("\n")
		// printRune(mask, maskp, roundedRect)
		// fmt.Printf("['%c'] maskp = %v, size=%vx%v\n", c, maskp, roundedRect.Dx(), roundedRect.Dy())
		// fmt.Printf("['%c'] maskp: %v\n", c, maskp)
		// fmt.Printf("['%c'] roundedRect: %v\n", c, roundedRect)
		// fmt.Printf("['%c'] accurateRect: %v\n", c, accurateRect)
	}

	// aRect, mask, maskp, _, _ := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, 'c')
	// fmt.Printf("glyphBytes = %v bytes\n", len(glyphBytes))
	// fmt.Printf("entire cache = %v bytes\n", mask.Bounds().Dx()*mask.Bounds().Dy())

	// runeInfo demo
	// info := runeMap['A'-minASCII]
	// if glyph, ok := mask.(*image.Alpha); ok {
	// 	fmt.Printf("A's maskp = %v, stride = %v\n", maskp, glyph.Stride)
	// 	for row := 0; row < aRect.Dy(); row++ {
	// 		fmt.Printf("%03v\n", glyph.Pix[(maskp.Y+row)*glyph.Stride:(maskp.Y+row+1)*glyph.Stride])
	// 	}
	// }

	// runemap demonstration
	// for i := 0; i < unicode.MaxASCII-minASCII; i++ {
	// 	fmt.Printf("RUNEMAP ['%c'] index = %v, (%vx%v)=%v bytes\n", rune(i+minASCII), runeMap[i].index, runeMap[i].width, runeMap[i].height, runeMap[i].width*runeMap[i].height)
	// }
	// fmt.Printf("glyphBytes = %v bytes\n", len(glyphBytes))

	_, mask, _, _, _ := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, 'A')
	glyph, _ := mask.(*image.Alpha)
	textureWidth := int32(glyph.Stride)
	textureHeight := int32(len(glyphBytes) / glyph.Stride)
	// pass glyphBytes to OpenGL texture
	var fontTextureID uint32
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1) // Disable byte-alignment restriction
	gl.GenTextures(1, &fontTextureID)
	gl.BindTexture(gl.TEXTURE_2D, fontTextureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, textureWidth, textureHeight, 0, uint32(gl.RED), gl.UNSIGNED_BYTE, unsafe.Pointer(&glyphBytes[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4) // TODO Enable byte-alignment restriction ?
	face.Close()
	// fmt.Printf("loaded %v at size %v in %v ns\n", fontName, fontSize, sw.stopGetNano())
	return fontTextureID, runeMap[:], nil
}

func renderText(font *ttf.Font, fontSize int32, text string, pos coord, align Align, col *sdl.Color, maxH int32) ([]byte, *sdl.Rect, error) {
	var surf *sdl.Surface
	var err error
	if surf, err = font.RenderUTF8Blended(text, *col); err != nil {
		return nil, nil, err
	}

	h, err := font.GlyphMetrics('h')
	rowsFromTop := int32(font.Ascent() - h.MaxY)
	sliceStart := surf.Pitch * (rowsFromTop)
	rowsFromBottom := surf.H - fontSize - rowsFromTop
	sliceStop := surf.Pitch * (surf.H - rowsFromBottom)
	// copy the pixels so we can free surf before returning
	slice := make([]byte, len(surf.Pixels()))
	copy(slice, surf.Pixels()[sliceStart:sliceStop])

	w2 := int32(float64(surf.W) / 2.0)
	h2 := int32(float64(fontSize) / 2.0)
	offx := -w2 - int32(align.h)*int32(w2)
	offy := -h2 - int32(align.v)*int32(h2)
	var rect = &sdl.Rect{
		X: pos.x + offx,
		// apply coordinate conversion from SDL to OpenGL
		Y: maxH - fontSize - pos.y - offy,
		W: int32(surf.W),
		H: fontSize,
	}

	surf.Free()
	return slice, rect, nil
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

	// convert pixel format to RGBA32 if necessary
	if surf.Format.Format != uint32(sdl.PIXELFORMAT_RGBA32) {
		convertedSurf, err := surf.ConvertFormat(uint32(sdl.PIXELFORMAT_RGBA32), 0)
		surf.Free()
		if err != nil {
			return nil, err
		}
		return convertedSurf, nil
	}

	return surf, err
}

// bufferData buffers vertex data and returns a VAO and VBO
// the vertex data must be accompanied by a description of its layout
// ex: vertex layout: (x,y,z, s,t) -> layout: (3, 2)
// returns the VAO and VBO id
func bufferData(data []float32, layout []int32) (uint32, uint32) {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(data), gl.Ptr(&data[0]), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	var vertexSize int32
	for i := 0; i < len(layout); i++ {
		gl.EnableVertexAttribArray(uint32(i))
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
	for i := 0; i < len(layout); i++ {
		gl.DisableVertexAttribArray(uint32(i))
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	return vao, vbo
}

const (
	minASCII = 32

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
	uniform vec2 screenSize;
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		vec2 glSpace = vec2(2.0, -2.0) * (position_in / screenSize) + vec2(-1.0, 1.0);
		gl_Position = vec4(glSpace, 0.0, 1.0);
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

	vshTexturePassthrough = `
	#version 460
	layout(location = 0) in vec2 position_in;
	layout(location = 1) in vec2 tex_coords_in;
	out vec2 tex_coords;
	void main() {
		gl_Position = vec4(position_in, 0.0, 1.0);
		tex_coords = tex_coords_in;
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
