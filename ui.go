package main

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"unicode"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/golang/freetype/truetype"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"

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
		diff := image.Point{rec.Dx(), rec.Dy()}
		tofile := alpha.SubImage(image.Rectangle{maskp, maskp.Add(diff)})
		if f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755); err != nil {
			png.Encode(f, tofile)
			return err
		}
	}
	return nil
}

func printRune(mask image.Image, maskp image.Point, rec image.Rectangle) {
	var alpha *image.Alpha
	var ok bool
	if alpha, ok = mask.(*image.Alpha); !ok {
		fmt.Println("printRune image not Alpha")
		return
	}
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

func int26_6ToFloat32(x fixed.Int26_6) float32 {
	top := float32(x >> 6)
	bottom := float32(x&0x3F) / 64.0
	return top + bottom
}

type runeInfo struct {
	row      int32
	width    int32
	height   int32
	bearingX float32
	bearingY float32
	ascent   int32
	descent  int32
	advance  float32
}

type pointF32 struct {
	x float32
	y float32
}

// mapString turns each character in the string into a pair of
// (x,y,s,t)-vertex triangles using glyph information from a
// pre-loaded font. The vertex info is returned as []float32
func mapString(str string, runeMap []runeInfo, pos coord, align Align) []float32 {
	// 2 triangles per rune, 3 vertices per triangle, 4 float32's per vertex (x,y,s,t)
	buffer := make([]float32, 0, len(str)*24)
	// get glyph information for alignment
	var strHeight, maxAscent, maxDescent int32
	var strWidth, largestBearingY float32
	for _, r := range str {
		info := runeMap[r-minASCII]
		if info.ascent > maxAscent {
			maxAscent = info.ascent
		}
		if info.descent > maxDescent {
			maxDescent = info.descent
		}
		if info.bearingY > largestBearingY {
			largestBearingY = info.bearingY
		}
		strWidth += info.advance
	}
	// adjust strWidth if last rune's width + bearingX > advance
	lastInfo := runeMap[str[len(str)-1]-minASCII]
	if float32(lastInfo.width)+lastInfo.bearingX > lastInfo.advance {
		strWidth += (float32(lastInfo.width) + lastInfo.bearingX - lastInfo.advance)
	}

	strHeight = maxAscent + maxDescent
	w2 := float64(strWidth) / 2.0
	h2 := float64(strHeight) / 2.0
	offx := int32(-w2 - float64(align.h)*w2)
	offy := int32(-h2 - float64(align.v)*h2)
	// offset origin to account for alignment
	origin := pointF32{float32(pos.x + offx), float32(pos.y+offy) + largestBearingY}
	for _, r := range str {
		info := runeMap[r-minASCII]

		// calculate x,y position coordinates - use bottom left as (0,0); shader converts for you
		posTL := pointF32{origin.x + info.bearingX, origin.y + (float32(info.height) - info.bearingY)}
		posTR := pointF32{posTL.x + float32(info.width), posTL.y}
		posBL := pointF32{posTL.x, origin.y - info.bearingY}
		posBR := pointF32{posTR.x, posBL.y}
		// calculate s,t texture coordinates - use top left as (0,0); shader converts for you
		texTL := pointF32{0, float32(info.row)}
		texTR := pointF32{float32(info.width), texTL.y}
		texBL := pointF32{texTL.x, texTL.y + float32(info.height)}
		texBR := pointF32{texTR.x, texBL.y}
		// create 2 triangles
		triangles := []float32{
			// TODO do something better... please
			float32(math.Ceil(float64(posBL.x))), float32(math.Ceil(float64(posBL.y))), texBL.x, texBL.y, // bottom-left
			float32(math.Ceil(float64(posTL.x))), float32(math.Ceil(float64(posTL.y))), texTL.x, texTL.y, // top-left
			float32(math.Ceil(float64(posTR.x))), float32(math.Ceil(float64(posTR.y))), texTR.x, texTR.y, // top-right

			float32(math.Ceil(float64(posBL.x))), float32(math.Ceil(float64(posBL.y))), texBL.x, texBL.y, // bottom-left
			float32(math.Ceil(float64(posTR.x))), float32(math.Ceil(float64(posTR.y))), texTR.x, texTR.y, // top-right
			float32(math.Ceil(float64(posBR.x))), float32(math.Ceil(float64(posBR.y))), texBR.x, texBR.y, // bottom-right
		}
		buffer = append(buffer, triangles...)

		origin.x += info.advance
	}

	return buffer
}

// loadFontTexture caches all of the glyph pixel data in an OpenGL texture for
// a given font at a given size. It returns the OpenGL ID for this texture,
// along with a runeInfo array for indexing into the texture per rune at runtime
func loadFontTexture(fontName string, fontSize int32) (uint32, []runeInfo, error) {
	sw := start()

	var err error
	var fontBytes []byte
	var font *truetype.Font
	if fontBytes, err = ioutil.ReadFile(fontName); err != nil {
		return 0, nil, err
	}
	if font, err = truetype.Parse(fontBytes); err != nil {
		return 0, nil, err
	}
	face := truetype.NewFace(font, &truetype.Options{Size: float64(fontSize)})

	var runeMap [unicode.MaxASCII - minASCII]runeInfo
	var glyphBytes []byte
	var currentIndex int32
	for i := minASCII; i < unicode.MaxASCII; i++ {
		c := rune(i)

		roundedRect, mask, maskp, advance, okGlyph := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, c)
		if !okGlyph {
			return 0, nil, fmt.Errorf("%v does not contain glyph for '%c'", fontName, c)
		}
		accurateRect, _, okBounds := face.GlyphBounds(c)
		glyph, okCast := mask.(*image.Alpha)
		if !okBounds || !okCast {
			return 0, nil, fmt.Errorf("%v does not contain glyph for '%c'", fontName, c)
		}

		runeMap[i-minASCII] = runeInfo{
			row:      currentIndex,
			width:    int32(roundedRect.Dx()),
			height:   int32(roundedRect.Dy()),
			bearingX: float32(math.Round(float64(accurateRect.Min.X.Ceil()))),
			bearingY: float32(accurateRect.Max.Y.Ceil()),
			ascent:   int32(math.Abs(float64(roundedRect.Max.Y))),
			descent:  int32(math.Abs(float64(roundedRect.Min.Y))),
			advance:  float32(math.Round(float64(int26_6ToFloat32(advance)))),
		}
		// alternatively, upload entire glyph cache into OpenGL texture
		// ... but this doesnt take that long and cuts texture size by 95%
		for row := 0; row < roundedRect.Dy(); row++ {
			beg := (maskp.Y + row) * glyph.Stride
			end := (maskp.Y + row + 1) * glyph.Stride
			glyphBytes = append(glyphBytes, glyph.Pix[beg:end]...)
			currentIndex++
		}
	}

	_, mask, _, _, aOK := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, 'A')
	if !aOK {
		return 0, nil, fmt.Errorf("Failed to get glyph for 'A'")
	}

	// writeme, _ := mask.(*image.Alpha)
	// writeme.Pix = glyphBytes
	// writeme.Rect = image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{int(writeme.Stride), int(len(glyphBytes) / writeme.Stride)}}
	// file, _ := os.OpenFile("glyphBytes.png", os.O_CREATE|os.O_RDWR, 0755)
	// png.Encode(file, writeme)

	glyph, _ := mask.(*image.Alpha)
	texWidth := int32(glyph.Stride)
	texHeight := int32(len(glyphBytes) / glyph.Stride)

	// pass glyphBytes to OpenGL texture
	var fontTextureID uint32
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1) // Disable byte-alignment restriction
	gl.GenTextures(1, &fontTextureID)
	gl.BindTexture(gl.TEXTURE_2D, fontTextureID)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RED, texWidth, texHeight, 0, uint32(gl.RED), gl.UNSIGNED_BYTE, unsafe.Pointer(&glyphBytes[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4) // TODO Enable byte-alignment restriction ?

	fmt.Printf("Loaded %v at size %v in %v ns\n", fontName, fontSize, sw.stopGetNano())
	return fontTextureID, runeMap[:], nil
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

// configureVAO configures a VAO & VBO pair with a specified vertex layout
// example vertex layout: (x,y,z, s,t) -> layout = (3, 2)
func configureVAO(vaoID uint32, vboID uint32, layout []int32) {
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

	// Uniform `tex_size` is the (width, height) of the texture.
	// Input `position_in` is typical openGL position coordinates.
	// Input `tex_pixels` is the (x, y) of the vertex in the texture starting
	// at (left, top).
	// Output `tex_coords` is typical texture coordinates for fragment shader.
	glyphShaderVertex = `
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

	glyphShaderFragment = `
	#version 460
	uniform sampler2D frag_tex;
	in vec2 tex_coords;
	out vec4 frag_color;
	void main() {
		frag_color = vec4(1.0, 1.0, 1.0, texture(frag_tex, tex_coords).r);
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
