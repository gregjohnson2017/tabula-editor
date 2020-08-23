package gfx

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/veandco/go-sdl2/sdl"
)

type Texture struct {
	id     uint32
	width  int32
	height int32
}

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
	t, err := NewTexture(int32(width), int32(height), data, gl.RGBA, 4)
	t.SetParameter(gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	t.SetParameter(gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	return t, err
}

func NewTexture(width, height int32, data []byte, format int, alignment int32) (Texture, error) {
	t := Texture{width: width, height: height}

	gl.GenTextures(1, &t.id)
	t.Bind()
	// copy pixels to texture
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, alignment)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), width, height, 0, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	t.Unbind()

	return t, nil
}

func (t Texture) SetParameter(paramName uint32, param int32) {
	t.Bind()
	gl.TexParameteri(gl.TEXTURE_2D, paramName, param)
	t.Unbind()
}

// ErrCoordOutOfRange indicates that given coordinates are out of range
const ErrCoordOutOfRange log.ConstErr = "coordinates out of range"

func (t Texture) SetPixel(p sdl.Point, col color.RGBA) error {
	if p.X < 0 || p.Y < 0 || p.X >= t.width || p.Y >= t.height {
		return fmt.Errorf("setPixel(%v, %v): %w", p.X, p.Y, ErrCoordOutOfRange)
	}
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	gl.TextureSubImage2D(t.id, 0, p.X, p.Y, 1, 1, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&col))
	// TODO update mipmap textures only when needed ?
	t.Bind()
	gl.GenerateMipmap(gl.TEXTURE_2D)
	t.Unbind()
	return nil
}

func (t Texture) GetTextureData() []byte {
	// TODO do this in batches/stream to avoid memory limitations
	var data = make([]byte, t.width*t.height*4)
	t.Bind()
	gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	t.Unbind()
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

func (t Texture) Delete() {
	gl.DeleteTextures(1, &t.id)
}
