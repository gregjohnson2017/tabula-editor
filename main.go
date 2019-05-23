package main

import (
	"strconv"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

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
	if tex, err = rend.CreateTextureFromSurface(surf); err != nil {
		panic(err)
	}
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

	var realW, realH int32
	if _, _, realW, realH, err = tex.Query(); err != nil {
		panic(err)
	}

	var canvas = &sdl.Rect{
		X: 0,
		Y: 0,
		W: realW,
		H: realH,
	}

	running := true
	var time, lastTime uint32
	lastTime = sdl.GetTicks()
	rmouseDown := false
	var rmousePoint = struct{ x, y int32 }{
		x: 0,
		y: 0,
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
				if evt.State == sdl.ButtonRMask() && rmouseDown {
					canvas.X += evt.X - rmousePoint.x
					canvas.Y += evt.Y - rmousePoint.y
					rmousePoint.x = evt.X
					rmousePoint.y = evt.Y
				}
			}
		}
		rend.Clear()
		rend.SetViewport(canvas)
		rend.Copy(tex, nil, nil)
		rend.SetViewport(bottomBar)
		rend.Copy(bottomBarTex, nil, nil)
		gfx.FramerateDelay(framerate)
		time = sdl.GetTicks()
		fps := int(1.0 / (float32(time-lastTime) / 1000.0))
		coords := "(" + strconv.Itoa(int(rmousePoint.x)) + ", " + strconv.Itoa(int(rmousePoint.y)) + ")"
		renderText(conf, rend, coords, conf.screenWidth, 0, true)
		renderText(conf, rend, strconv.Itoa(fps)+" FPS", 0, 0, false)
		lastTime = time
		rend.Present()
	}
}
