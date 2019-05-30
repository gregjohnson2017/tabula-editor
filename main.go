package main

import (
	"fmt"
	"math"
	"strconv"

	set "github.com/kroppt/IntSet"
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
	fontSize        int
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

func createSolidColorTexture(rend *sdl.Renderer, w int32, h int32, r uint8, g uint8, b uint8, a uint8) *sdl.Texture {
	var surf *sdl.Surface
	var err error
	if surf, err = sdl.CreateRGBSurfaceWithFormat(0, w, h, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
		panic(err)
	}
	if err = surf.FillRect(nil, sdl.MapRGBA(surf.Format, r, g, b, a)); err != nil {
		panic(err)
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTextureFromSurface(surf); err != nil {
		panic(err)
	}
	surf.Free()
	return tex
}

func renderText(conf *config, rend *sdl.Renderer, text string, pos coord, align Align) {
	col := sdl.Color{
		R: 255,
		G: 255,
		B: 255,
		A: 0,
	}
	var surf *sdl.Surface
	var err error
	if surf, err = conf.font.RenderUTF8Blended(text, col); err != nil {
		panic(err)
	}
	var tex *sdl.Texture
	if tex, err = rend.CreateTextureFromSurface(surf); err != nil {
		panic(err)
	}
	var w, h int
	if w, h, err = conf.font.SizeUTF8(text); err != nil {
		surf.Free()
		tex.Destroy()
		panic(err)
	}
	w2 := int32(float64(w) / 2.0)
	h2 := int32(float64(h) / 2.0)
	offx := -w2 - int32(align.h)*int32(w2)
	offy := -h2 - int32(align.v)*int32(h2)
	var rect = &sdl.Rect{
		X: pos.x + offx,
		Y: pos.y + offy,
		W: int32(w),
		H: int32(h),
	}
	if err = rend.Copy(tex, nil, rect); err != nil {
		panic(err)
	}
	surf.Free()
	tex.Destroy()
}

type zoomer struct {
	lastMult float64
	mult     float64
	origW    float64
	origH    float64
	maxW     int32
	maxH     int32
}

func (z *zoomer) In() {
	if int32(z.mult*z.origW*2.0) < z.maxW && int32(z.mult*z.origH*2.0) < z.maxH {
		z.mult *= 2
	}
}

func (z *zoomer) Out() {
	if int32(z.mult*z.origW/2.0) > 0 && int32(z.mult*z.origH/2.0) > 0 {
		z.mult /= 2
	}
}

func (z *zoomer) MultW() int32 {
	return int32(z.origW * z.mult)
}

func (z *zoomer) MultH() int32 {
	return int32(z.origH * z.mult)
}

func (z *zoomer) IsIn() bool {
	return z.lastMult < z.mult
}

func (z *zoomer) IsOut() bool {
	return z.lastMult > z.mult
}

func (z *zoomer) Update() {
	z.lastMult = z.mult
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
	if conf.font, err = ttf.OpenFont(conf.fontName, conf.fontSize); err != nil {
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

func copyToTexture(tex *sdl.Texture, surf *sdl.Surface, region *sdl.Rect) error {
	var bytes []byte
	var err error
	bytes, _, err = tex.Lock(region)
	copy(bytes, surf.Pixels())
	tex.Unlock()
	return err
}

func main() {
	conf := initConfig()
	var err error
	if err = initialize(conf); err != nil {
		panic(err)
	}

	var win *sdl.Window
	if win, err = sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, conf.screenWidth, conf.screenHeight, 0); err != nil {
		panic(err)
	}
	var rend *sdl.Renderer
	if rend, err = sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED); err != nil {
		panic(err)
	}
	if err = rend.SetDrawColor(0xFF, 0xFF, 0xFF, 0xFF); err != nil {
		panic(err)
	}
	var tex *sdl.Texture
	var surf *sdl.Surface
	if surf, err = img.Load("monkaW.png"); err != nil {
		panic(err)
	}
	if tex, err = rend.CreateTexture(surf.Format.Format, sdl.TEXTUREACCESS_STREAMING, surf.W, surf.H); err != nil {
		panic(err)
	}
	err = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	var canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: surf.W,
		H: surf.H,
	}

	func() {
		var bytes []byte
		bytes, _, err = tex.Lock(canvas)
		copy(bytes, surf.Pixels())
		tex.Unlock()
	}()

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
	bottomBarTex := createSolidColorTexture(rend, conf.screenWidth, conf.bottomBarHeight, g, g, g, 0xFF)

	running := true
	var time, lastTime uint32
	lastTime = sdl.GetTicks()
	rmouseDown := false
	var rmousePoint = coord{
		x: 0,
		y: 0,
	}
	var mouseLoc = coord{
		x: 0,
		y: 0,
	}
	var mousePix coord
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
	var zoom = &zoomer{
		1.0,
		1.0,
		float64(surf.W),
		float64(surf.H),
		info.MaxTextureWidth,
		info.MaxTextureHeight,
	}
	updateMousePos := func(x, y int32) {
		mouseLoc.x = x
		mouseLoc.y = y
		mousePix.x = int32(float64(mouseLoc.x-canvas.X) / zoom.mult)
		mousePix.y = int32(float64(mouseLoc.y-canvas.Y) / zoom.mult)
	}
	sel := set.NewSet()
	var selSurf *sdl.Surface
	if selSurf, err = sdl.CreateRGBSurfaceWithFormat(0, surf.W, surf.H, 32, uint32(sdl.PIXELFORMAT_RGBA32)); err != nil {
		panic(err)
	}
	if err = selSurf.FillRect(nil, sdl.MapRGBA(selSurf.Format, 0, 0, 0, 0)); err != nil {
		panic(err)
	}
	var selTex *sdl.Texture
	if selTex, err = rend.CreateTexture(selSurf.Format.Format, sdl.TEXTUREACCESS_STREAMING, selSurf.W, selSurf.H); err != nil {
		panic(err)
	}
	selTex.SetBlendMode(sdl.BLENDMODE_BLEND)
	selFunc := func(n int) bool {
		y := int32(n) % selSurf.W
		x := int32(n) - y*selSurf.W
		setPixel(selSurf, coord{x: x, y: y}, sdl.Color{R: 0, G: 0, B: 0, A: 128})
		return true
	}
	onCanvas := func(x, y int32) bool {
		if x < canvas.X || x >= canvas.X+canvas.W {
			return false
		}
		if y < canvas.Y || y >= canvas.Y+canvas.H {
			return false
		}
		if y >= bottomBar.Y {
			return false
		}
		return true
	}
	for running {
		diffW := zoom.MultW() - canvas.W
		diffH := zoom.MultH() - canvas.H
		canvas.W += diffW
		canvas.H += diffH
		if zoom.IsIn() {
			canvas.X = 2*canvas.X - mouseLoc.x
			canvas.Y = 2*canvas.Y - mouseLoc.y
		}
		if zoom.IsOut() {
			canvas.X = canvas.X/2 + mouseLoc.x/2
			canvas.Y = canvas.Y/2 + mouseLoc.y/2
		}
		zoom.Update()
		var e sdl.Event
		for e = sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch evt := e.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				updateMousePos(evt.X, evt.Y)
				if evt.Button == sdl.BUTTON_RIGHT {
					if evt.State == sdl.PRESSED && evt.Y < bottomBar.Y {
						rmouseDown = true
					} else if evt.State == sdl.RELEASED {
						rmouseDown = false
					}
					rmousePoint.x = evt.X
					rmousePoint.y = evt.Y
				}
				if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED && onCanvas(evt.X, evt.Y) {
					i := int(surf.W*mousePix.y + mousePix.x)
					if !sel.Contains(i) {
						sel.Add(i)
					}
				}
			case *sdl.MouseMotionEvent:
				updateMousePos(evt.X, evt.Y)
				if evt.State == sdl.ButtonRMask() && rmouseDown {
					canvas.X += evt.X - rmousePoint.x
					canvas.Y += evt.Y - rmousePoint.y
					rmousePoint.x = evt.X
					rmousePoint.y = evt.Y
				}
				if evt.State == sdl.ButtonLMask() && onCanvas(evt.X, evt.Y) {
					i := int(surf.W*mousePix.y + mousePix.x)
					if !sel.Contains(i) {
						sel.Add(i)
					}
				}
			case *sdl.MouseWheelEvent:
				if evt.Y > 0 {
					zoom.In()
				} else if evt.Y < 0 {
					zoom.Out()
				}
			}
		}

		if err = rend.Clear(); err != nil {
			panic(err)
		}
		if err = rend.SetViewport(canvas); err != nil {
			panic(err)
		}
		if err = rend.Copy(tex, nil, nil); err != nil {
			panic(err)
		}

		var bytes []byte
		if bytes, _, err = selTex.Lock(nil); err != nil {
			panic(err)
		}
		sel.Range(selFunc)
		copy(bytes, selSurf.Pixels())
		selTex.Unlock()
		if err = rend.Copy(selTex, nil, nil); err != nil {
			panic(err)
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
		coords := "(" + strconv.Itoa(int(mousePix.x)) + ", " + strconv.Itoa(int(mousePix.y)) + ")"
		pos := coord{conf.screenWidth, int32(float64(bottomBar.H) / 2.0)}
		renderText(conf, rend, coords, pos, Align{AlignMiddle, AlignRight})
		pos.x = 0
		renderText(conf, rend, strconv.Itoa(fps)+" FPS", pos, Align{AlignMiddle, AlignLeft})
		lastTime = time
		rend.Present()
	}

	quit(conf)
}
