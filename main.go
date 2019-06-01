package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

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

type config struct {
	screenWidth     int32
	screenHeight    int32
	bottomBarHeight int32
	fontName        string
	fontSize        int32
	font            *ttf.Font
	framerate       uint32
}

func initConfig() *config {
	c := config{
		screenWidth:     640,
		screenHeight:    480,
		bottomBarHeight: 30,
		fontName:        "NotoMono-Regular.ttf",
		fontSize:        24,
		framerate:       144,
	}
	return &c
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

func renderText(conf *config, rend *sdl.Renderer, text string, pos coord, align Align) error {
	col := sdl.Color{
		R: 255,
		G: 255,
		B: 255,
		A: 0,
	}
	var surf *sdl.Surface
	var err error
	if surf, err = conf.font.RenderUTF8Blended(text, col); err != nil {
		return err
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTexture(surf.Format.Format, sdl.TEXTUREACCESS_STREAMING, surf.W, int32(conf.fontSize)); err != nil {
		surf.Free()
		return err
	}
	if err = tex.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		surf.Free()
		tex.Destroy()
		return err
	}
	sliceOffset := surf.Pitch * (surf.H - conf.fontSize)
	copyToTexture(tex, surf.Pixels()[sliceOffset:], nil)

	w2 := int32(float64(surf.W) / 2.0)
	h2 := int32(float64(conf.fontSize) / 2.0)
	offx := -w2 - int32(align.h)*int32(w2)
	offy := -h2 - int32(align.v)*int32(h2)
	var rect = &sdl.Rect{
		X: pos.x + offx,
		Y: pos.y + offy,
		W: int32(surf.W),
		H: int32(conf.fontSize),
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

func initialize(conf *config) error {
	if sdl.SetHint(sdl.HINT_RENDER_DRIVER, "opengl") != true {
		return fmt.Errorf("failed to set opengl render driver hint")
	}
	var err error
	if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
		return err
	}
	if img.Init(img.INIT_PNG) != img.INIT_PNG {
		return fmt.Errorf("could not initialize PNG")
	}
	if err = ttf.Init(); err != nil {
		return err
	}
	if conf.font, err = ttf.OpenFont(conf.fontName, int(conf.fontSize)); err != nil {
		return err
	}
	return err
}

func quit(conf *config) {
	sdl.Quit()
	img.Quit()
	conf.font.Close()
	ttf.Quit()
}

func copyToTexture(tex *sdl.Texture, pixels []byte, region *sdl.Rect) error {
	var bytes []byte
	var err error
	bytes, _, err = tex.Lock(region)
	copy(bytes, pixels)
	tex.Unlock()
	return err
}

func loadImage(rend *sdl.Renderer, fileName string) (*sdl.Surface, *sdl.Texture, error) {
	var tex *sdl.Texture
	var surf *sdl.Surface
	var err error
	if surf, err = img.Load(fileName); err != nil {
		return nil, nil, err
	}
	if tex, err = rend.CreateTexture(surf.Format.Format, sdl.TEXTUREACCESS_STREAMING, surf.W, surf.H); err != nil {
		return nil, nil, err
	}
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return nil, nil, err
	}
	if err = copyToTexture(tex, surf.Pixels(), nil); err != nil {
		return nil, nil, err
	}
	return surf, tex, err
}

func inBounds(area *sdl.Rect, x int32, y int32) bool {
	if x < area.X || x >= area.X+area.W || y < area.Y || y >= area.Y+area.H {
		return false
	}
	return true
}

func main() {
	conf := initConfig()
	var err error
	if err = initialize(conf); err != nil {
		panic(err)
	}

	var win *sdl.Window
	if win, err = sdl.CreateWindow("Tabula Editor", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, conf.screenWidth, conf.screenHeight, 0); err != nil {
		panic(err)
	}
	var rend *sdl.Renderer
	if rend, err = sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED); err != nil {
		panic(err)
	}
	if err = rend.SetDrawColor(0xFF, 0xFF, 0xFF, 0xFF); err != nil {
		panic(err)
	}

	var framerate = &gfx.FPSmanager{}
	gfx.InitFramerate(framerate)
	if gfx.SetFramerate(framerate, conf.framerate) != true {
		panic(fmt.Errorf("could not set framerate: %v", sdl.GetError()))
	}
	var bottomBar = &sdl.Rect{
		X: 0,
		Y: conf.screenHeight - conf.bottomBarHeight,
		W: conf.screenWidth,
		H: conf.bottomBarHeight,
	}
	g := uint8(0x80)
	var bottomBarTex *sdl.Texture
	if bottomBarTex, err = createSolidColorTexture(rend, conf.screenWidth, conf.bottomBarHeight, g, g, g, 0xFF); err != nil {
		panic(err)
	}

	running := true
	var time, lastTime uint32
	lastTime = sdl.GetTicks()
	var info sdl.RendererInfo
	if info, err = rend.GetInfo(); err != nil {
		panic(err)
	}
	if info.MaxTextureWidth == 0 {
		info.MaxTextureWidth = math.MaxInt32
	}
	if info.MaxTextureHeight == 0 {
		info.MaxTextureHeight = math.MaxInt32
	}
	ctx := &context{win, rend, &info}
	area := &sdl.Rect{
		X: 0,
		Y: 0,
		W: conf.screenWidth,
		H: conf.screenHeight,
	}
	iv, err := newImageView(area, "monkaW.png", ctx)
	for running {
		var e sdl.Event
		for e = sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch evt := e.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				// TODO
				if inBounds(iv.area, evt.X, evt.Y) {
					iv.onClick(evt)
				}
			case *sdl.MouseMotionEvent:
				// TODO
				if inBounds(iv.area, evt.X, evt.Y) {
					iv.onMotion(evt)
				}
			case *sdl.MouseWheelEvent:
				// TODO
				iv.onScroll(evt)
			}
		}

		if err = rend.Clear(); err != nil {
			panic(err)
		}
		// TODO
		if err = rend.SetViewport(iv.getBoundary()); err != nil {
			panic(err)
		}
		textures, err := iv.render()
		if err != nil {
			panic(err)
		}
		for _, tex := range textures {
			if err = rend.Copy(tex, nil, nil); err != nil {
				panic(err)
			}
		}
		if err = rend.SetViewport(bottomBar); err != nil {
			panic(err)
		}
		if err = rend.Copy(bottomBarTex, nil, nil); err != nil {
			panic(err)
		}

		gfx.FramerateDelay(framerate)
		time = sdl.GetTicks()
		fps := int(1.0 / (float32(time-lastTime) / 1000.0))
		coords := "(" + strconv.Itoa(int(iv.mousePix.x)) + ", " + strconv.Itoa(int(iv.mousePix.y)) + ")"
		pos := coord{conf.screenWidth, int32(float64(bottomBar.H) / 2.0)}
		// TODO
		if err = renderText(conf, rend, coords, pos, Align{AlignMiddle, AlignRight}); err != nil {
			panic(err)
		}
		pos.x = 0
		if err = renderText(conf, rend, strconv.Itoa(fps)+" FPS", pos, Align{AlignMiddle, AlignLeft}); err != nil {
			panic(err)
		}
		lastTime = time
		rend.Present()
	}

	quit(conf)
}
