package font

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"time"
	"unicode"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/golang/freetype/truetype"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"golang.org/x/image/math/fixed"
)

const minASCII = 32

func PrintRune(mask image.Image, maskp image.Point, rec image.Rectangle) {
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
	advance  float32
}

type pointF32 struct {
	x float32
	y float32
}

func CalcStringDims(str string, font Info) (float64, float64) {
	var strWidth, largestBearingY float32
	for _, r := range str {
		info := font.runeMap[r-minASCII]
		if info.bearingY > largestBearingY {
			largestBearingY = info.bearingY
		}
		strWidth += info.advance
	}
	// adjust strWidth if last rune's width + bearingX > advance
	lastInfo := font.runeMap[str[len(str)-1]-minASCII]
	if float32(lastInfo.width)+lastInfo.bearingX > lastInfo.advance {
		strWidth += (float32(lastInfo.width) + lastInfo.bearingX - lastInfo.advance)
	}

	return float64(strWidth), float64(font.textHeight)
}

// GetMaxVerticalBearing gets the amount of vertical bearing needed
// to render this string with the given font
func GetMaxVerticalBearing(str string, font Info) float32 {
	var largestBearingY float32
	for _, r := range str {
		info := font.runeMap[r-minASCII]
		if info.bearingY > largestBearingY {
			largestBearingY = info.bearingY
		}
	}
	return largestBearingY
}

// MapString turns each character in the string into a pair of
// (x,y,s,t)-vertex triangles using glyph information from a
// pre-loaded font. The vertex info is returned as []float32
func MapString(str string, font Info, pos sdl.Point, align ui.Align) []float32 {
	return MapStringWithBearing(str, font, GetMaxVerticalBearing(str, font), pos, align)
}

// MapStringWithBearing turns each character in the string into
// a pair of (x,y,s,t)-vertex triangles using glyph information
// from a pre-loaded font and the font information, with the
// vertical bearing provided. The vertex info is returned as []float32
func MapStringWithBearing(str string, font Info, maxBearingY float32, pos sdl.Point, align ui.Align) []float32 {
	// 2 triangles per rune, 3 vertices per triangle, 4 float32's per vertex (x,y,s,t)
	buffer := make([]float32, 0, len(str)*24)
	// get glyph information for alignment
	var strWidth float32
	for _, r := range str {
		info := font.runeMap[r-minASCII]
		strWidth += info.advance
	}
	// adjust strWidth if last rune's width + bearingX > advance
	lastInfo := font.runeMap[str[len(str)-1]-minASCII]
	if float32(lastInfo.width)+lastInfo.bearingX > lastInfo.advance {
		strWidth += (float32(lastInfo.width) + lastInfo.bearingX - lastInfo.advance)
	}

	w2 := float64(strWidth) / 2.0
	h2 := float64(font.textHeight) / 2.0
	offx := int32(-w2 - float64(align.H)*w2)
	offy := int32(-h2 - float64(align.V)*h2)
	// offset origin to account for alignment
	origin := pointF32{float32(pos.X + offx), float32(pos.Y+offy) + maxBearingY}
	for _, r := range str {
		info := font.runeMap[r-minASCII]

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
			posBL.x, posBL.y, texBL.x, texBL.y, // bottom-left
			posTL.x, posTL.y, texTL.x, texTL.y, // top-left
			posTR.x, posTR.y, texTR.x, texTR.y, // top-right

			posBL.x, posBL.y, texBL.x, texBL.y, // bottom-left
			posTR.x, posTR.y, texTR.x, texTR.y, // top-right
			posBR.x, posBR.y, texBR.x, texBR.y, // bottom-right
		}
		buffer = append(buffer, triangles...)

		origin.x += info.advance
	}

	return buffer
}

type fontKey struct {
	fontName string
	fontSize int32
}

type Info struct {
	TextureID  uint32     // OpenGL texture ID of cached glyph data
	runeMap    []runeInfo // map of character-specific spacing info
	textHeight float32    // how much space is between two lines of text in this font
}

// TODO save cached fonts to local direct
// fontMap caches previously loaded fonts
var fontMap map[fontKey]Info

// loadFontTexture caches all of the glyph pixel data in an OpenGL texture for
// a given font at a given size. It returns the OpenGL ID for this texture,
// along with a runeInfo array for indexing into the texture per rune at runtime
func LoadFontTexture(fontName string, fontSize int32) (Info, error) {
	if fontMap == nil {
		fontMap = make(map[fontKey]Info)
	}
	if val, ok := fontMap[fontKey{fontName, fontSize}]; ok {
		return val, nil
	}
	sw := util.Start()

	var err error
	var fontBytes []byte
	var font *truetype.Font
	if fontBytes, err = ioutil.ReadFile(fontName); err != nil {
		return Info{0, nil, 0}, err
	}
	if font, err = truetype.Parse(fontBytes); err != nil {
		return Info{0, nil, 0}, err
	}
	face := truetype.NewFace(font, &truetype.Options{Size: float64(fontSize)})

	var runeMap [unicode.MaxASCII - minASCII]runeInfo
	var glyphBytes []byte
	var currentIndex int32
	for i := minASCII; i < unicode.MaxASCII; i++ {
		c := rune(i)

		roundedRect, mask, maskp, advance, okGlyph := face.Glyph(fixed.Point26_6{X: 0, Y: 0}, c)
		if !okGlyph {
			return Info{0, nil, 0}, fmt.Errorf("%v does not contain glyph for '%c'", fontName, c)
		}
		accurateRect, _, okBounds := face.GlyphBounds(c)
		glyph, okCast := mask.(*image.Alpha)
		if !okBounds || !okCast {
			return Info{0, nil, 0}, fmt.Errorf("%v does not contain glyph for '%c'", fontName, c)
		}

		runeMap[i-minASCII] = runeInfo{
			row:      currentIndex,
			width:    int32(roundedRect.Dx()),
			height:   int32(roundedRect.Dy()),
			bearingX: float32(math.Round(float64(accurateRect.Min.X.Ceil()))),
			bearingY: float32(accurateRect.Max.Y.Ceil()),
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
		return Info{0, nil, 0}, fmt.Errorf("Failed to get glyph for 'A'")
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

	fontHeight := int26_6ToFloat32(face.Metrics().Height)

	fmt.Printf("Loaded %v at size %v:\t%v\n", fontName, fontSize, time.Duration(int64(time.Nanosecond)*sw.StopGetNano()))
	InfoLoaded := Info{fontTextureID, runeMap[:], fontHeight}
	fontMap[fontKey{fontName, fontSize}] = InfoLoaded
	return InfoLoaded, nil
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
