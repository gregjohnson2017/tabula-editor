package main

import (
	"strconv"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

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

func renderText(conf *config, rend *sdl.Renderer, text string, relx int32, rely int32, right bool) {
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
	if right {
		relx -= int32(w)
	}
	var rect = &sdl.Rect{
		X: relx,
		Y: rely,
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
	mult  float64
	origW float64
	origH float64
	maxW  int32
	maxH  int32
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

func main() {
	conf := initConfig()

	var err error
	if err = sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()
	if img.Init(img.INIT_PNG) != img.INIT_PNG {
		panic("could not initialize PNG")
	}
	defer img.Quit()
	if err = ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()
	if conf.font, err = ttf.OpenFont(conf.fontName, conf.fontSize); err != nil {
		panic(err)
	}
	defer conf.font.Close()

	var win *sdl.Window
	if win, err = sdl.CreateWindow("test", 0, 0, conf.screenWidth, conf.screenHeight, 0); err != nil {
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
	var format uint32
	format, _ = win.GetPixelFormat()
	if tex, err = rend.CreateTexture(format, sdl.TEXTUREACCESS_STREAMING, surf.W, surf.H); err != nil {
		panic(err)
	}

	var canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: surf.W,
		H: surf.H,
	}
	var bytes []byte
	bytes, _, err = tex.Lock(canvas)
	copy(bytes, surf.Pixels())
	tex.Unlock()

	var framerate = &gfx.FPSmanager{}
	gfx.InitFramerate(framerate)
	if gfx.SetFramerate(framerate, conf.framerate) != true {
		panic("could not set Framerate")
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
	var info sdl.RendererInfo
	if info, err = rend.GetInfo(); err != nil {
		panic(err)
	}
	var zoom = &zoomer{
		1.0,
		float64(surf.W),
		float64(surf.H),
		info.MaxTextureWidth,
		info.MaxTextureHeight,
	}
	for running {
		var e sdl.Event
		for e = sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			switch evt := e.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseButtonEvent:
				if evt.Button == sdl.BUTTON_RIGHT {
					if evt.State == sdl.PRESSED && evt.Y < bottomBar.Y {
						rmouseDown = true
					} else if evt.State == sdl.RELEASED {
						rmouseDown = false
					}
					rmousePoint.x = evt.X
					rmousePoint.y = evt.Y
				}
			case *sdl.MouseMotionEvent:
				mouseLoc.x = evt.X
				mouseLoc.y = evt.Y
				if evt.State == sdl.ButtonRMask() && rmouseDown {
					canvas.X += evt.X - rmousePoint.x
					canvas.Y += evt.Y - rmousePoint.y
					rmousePoint.x = evt.X
					rmousePoint.y = evt.Y
				}
			case *sdl.MouseWheelEvent:
				if evt.Y > 0 {
					zoom.In()
				} else if evt.Y < 0 {
					zoom.Out()
				}
			}
		}
		diffW := zoom.MultW() - canvas.W
		diffH := zoom.MultH() - canvas.H
		canvas.W += diffW
		canvas.H += diffH
		canvas.X -= int32(diffW / 2.0)
		canvas.Y -= int32(diffH / 2.0)
		var mousePix coord
		mousePix.x = int32(float64(mouseLoc.x-canvas.X) / zoom.mult)
		mousePix.y = int32(float64(mouseLoc.y-canvas.Y) / zoom.mult)
		if err = rend.Clear(); err != nil {
			panic(err)
		}
		if err = rend.SetViewport(canvas); err != nil {
			panic(err)
		}
		if err = rend.Copy(tex, nil, nil); err != nil {
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
		renderText(conf, rend, coords, conf.screenWidth, 0, true)
		renderText(conf, rend, strconv.Itoa(fps)+" FPS", 0, 0, false)
		lastTime = time
		rend.Present()
	}
}
