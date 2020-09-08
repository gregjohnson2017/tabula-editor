package gfx

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/veandco/go-sdl2/sdl"
)

type Texture struct {
	id        uint32
	width     int32
	height    int32
	format    uint32
	alignment int32
	texelSize int32
}

// NewTextureFromFile creates a new Texture, loading data from fileName
// with the assumption that it is an image that can be converted to RGBA
// (alpha is black for jpegs)
func NewTextureFromFile(fileName string) (Texture, error) {
	in, err := os.Open(fileName)
	if err != nil {
		return Texture{}, err
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return Texture{}, err
	}
	// TODO load from underlying arrays directly and correctly format in OpenGL
	// switch img.(type) {
	// case *image.Alpha:
	// case *image.Alpha16:
	// case *image.CMYK:
	// case *image.Gray:
	// case *image.Gray16:
	// case *image.NRGBA:
	// case *image.NRGBA64:
	// case *image.Paletted:
	// case *image.RGBA:
	// case *image.RGBA64:
	// case *image.YCbCr, *image.NYCbCrA, *image.Uniform:
	// 	// no Pix array
	// }
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	data := make([]byte, 0, width*height*4)
	for j := 0; j < height; j++ {
		for i := 0; i < width; i++ {
			col := color.NRGBAModel.Convert(img.At(i, j))
			nrgba := col.(color.NRGBA)
			r, g, b, a := nrgba.R, nrgba.G, nrgba.B, nrgba.A
			data = append(data, r, g, b, a)
		}
	}
	t, err := NewTexture(int32(width), int32(height), data, gl.RGBA, 4, 4)
	t.SetParameter(gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	t.SetParameter(gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	return t, err
}

// NewTexture creates a Texture object that wraps the OpenGL texture functions
// alignment is in bytes and is passed to gl.PixelStorei() for unpacking
// format example: gl.RGBA
func NewTexture(width, height int32, data []byte, format int, alignment int32, texelSize int32) (Texture, error) {
	t := Texture{
		width:     width,
		height:    height,
		format:    uint32(format),
		alignment: alignment,
		texelSize: texelSize,
	}
	var ptr unsafe.Pointer
	if data != nil {
		ptr = unsafe.Pointer(&data[0])
	}
	gl.GenTextures(1, &t.id)
	t.Bind()
	// copy pixels to texture
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, t.alignment)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), width, height, 0, uint32(format), gl.UNSIGNED_BYTE, ptr)
	gl.GenerateMipmap(gl.TEXTURE_2D)
	t.Unbind()

	return t, nil
}

// SetParameter wraps gl.TexParameteri()
func (t Texture) SetParameter(paramName uint32, param int32) {
	t.Bind()
	gl.TexParameteri(gl.TEXTURE_2D, paramName, param)
	t.Unbind()
}

// ErrCoordOutOfRange indicates that given coordinates are out of range
const ErrCoordOutOfRange log.ConstErr = "coordinates out of range"

// SetPixelByte sets a texel of a texture at a given point p to the given byte b.
func (t Texture) SetPixelArea(r sdl.Rect, b []byte, genMipmap bool) error {
	if r.X < 0 || r.Y < 0 || r.X >= t.width || r.Y >= t.height {
		return fmt.Errorf("SetPixelArea(%v): %w", r, ErrCoordOutOfRange)
	}
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, t.alignment)
	gl.TextureSubImage2D(t.id, 0, r.X, r.Y, r.W, r.H, t.format, gl.UNSIGNED_BYTE, unsafe.Pointer(&b[0]))
	if genMipmap {
		t.Bind()
		gl.GenerateMipmap(gl.TEXTURE_2D)
		t.Unbind()
	}
	return nil
}

func (t Texture) SetPixel(p sdl.Point, b []byte, genMipmap bool) error {
	return t.SetPixelArea(sdl.Rect{X: p.X, Y: p.Y, W: 1, H: 1}, b, genMipmap)
}

// GetData returns a byte slice of all the texture data
func (t Texture) GetData() []byte {
	// TODO do this in batches/stream to avoid memory limitations
	var data = make([]byte, t.width*t.height*t.texelSize)
	t.Bind()
	gl.PixelStorei(gl.PACK_ALIGNMENT, t.alignment)
	gl.GetTexImage(gl.TEXTURE_2D, 0, t.format, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	t.Unbind()
	return data
}

// GetSubData returns a portion of the texture data starting at x, y and going
// w in the x diretion and h in the y direction
func (t Texture) GetSubData(x, y, w, h int32) []byte {
	// TODO do this in batches/stream to avoid memory limitations
	var data = make([]byte, w*h*t.texelSize)
	gl.PixelStorei(gl.PACK_ALIGNMENT, t.alignment)
	gl.GetTextureSubImage(t.id, 0, x, y, 0, w, h, 1, t.format, gl.UNSIGNED_BYTE, w*h*t.texelSize, unsafe.Pointer(&data[0]))
	return data
}

// Bind tells OpenGL to set this texture as the current texture
func (t Texture) Bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.id)
}

// Unbind sets the bound texture id to 0
func (t Texture) Unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

func (t Texture) GetWidth() int32 {
	return t.width
}

func (t Texture) GetHeight() int32 {
	return t.height
}

// Destroy destroys OpenGL assets associated with the Texture
func (t Texture) Destroy() {
	gl.DeleteTextures(1, &t.id)
}
