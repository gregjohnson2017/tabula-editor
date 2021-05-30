package app

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/comms"
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/shaders"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/kroppt/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

var _ ui.Component = ui.Component(&BottomBar{})

// BottomBar defines a solid color bar with text displays
type BottomBar struct {
	area        *sdl.Rect
	comms       <-chan comms.Image
	backProgram gfx.Program
	textProgram gfx.Program
	backBuf     *gfx.VAO
	textBuf     *gfx.VAO
	font        *gfx.FontInfo
	cfg         *config.Config
}

// NewBottomBar returns a pointer to a new BottomBar struct that implements ui.Component
// the background color defaults to grey (0x808080FF) and the text white
func NewBottomBar(area *sdl.Rect, comms <-chan comms.Image, cfg *config.Config) (*BottomBar, error) {
	var err error

	v1, err := gfx.NewShader(shaders.SolidColorVertex, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f1, err := gfx.NewShader(shaders.SolidColorFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}
	backProgram, err := gfx.NewProgram(v1, f1)
	if err != nil {
		return nil, err
	}
	v2, err := gfx.NewShader(shaders.GlyphShaderVertex, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	f2, err := gfx.NewShader(shaders.GlyphShaderFragment, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}
	textProgram, err := gfx.NewProgram(v2, f2)
	if err != nil {
		return nil, err
	}

	fnt, err := gfx.LoadFontTexture("NotoMono-Regular.ttf", 24)
	if err != nil {
		return nil, err
	}

	barColor := [4]float32{0.5, 0.5, 0.5, 1.0}
	textColor := [4]float32{1.0, 1.0, 1.0, 1.0}

	err = backProgram.UploadUniform("uni_color", barColor[0], barColor[1], barColor[2], barColor[3])
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "uni_color", err)
	}

	err = textProgram.UploadUniform("screen_size", float32(cfg.ScreenWidth), float32(cfg.ScreenHeight))
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "screen_size", err)
	}
	err = textProgram.UploadUniform("tex_size", float32(fnt.GetTexture().GetWidth()), float32(fnt.GetTexture().GetHeight()))
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "tex_size", err)
	}
	err = textProgram.UploadUniform("text_color", textColor[0], textColor[1], textColor[2], textColor[3])
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "text_color", err)
	}

	backTriangles := []float32{
		-1.0, -1.0, // bottom-left
		-1.0, +1.0, // top-left
		+1.0, +1.0, // top-right

		-1.0, -1.0, // bottom-left
		+1.0, +1.0, // top-right
		+1.0, -1.0, // bottom-right
	}

	backBuf := gfx.NewVAO(gl.TRIANGLES, []int32{2})

	err = backBuf.Load(backTriangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load bottom bar background triangles: %v", err)
	}

	textBuf := gfx.NewVAO(gl.TRIANGLES, []int32{2, 2})

	return &BottomBar{
		area:        area,
		comms:       comms,
		backProgram: backProgram,
		textProgram: textProgram,
		backBuf:     backBuf,
		textBuf:     textBuf,
		font:        fnt,
		cfg:         cfg,
	}, nil
}

// SetBackgroundColor sets the color for the bottom bar's background texture
func (bb *BottomBar) SetBackgroundColor(color []float32) {
	err := bb.backProgram.UploadUniform("uni_color", float32(color[0]), float32(color[1]), float32(color[2]), float32(color[3]))
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "uni_color", err)
	}
}

// SetTextColor sets the color for the bottom bar's text elements
func (bb *BottomBar) SetTextColor(color []float32) {
	err := bb.textProgram.UploadUniform("text_color", float32(color[0]), float32(color[1]), float32(color[2]), float32(color[3]))
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "text_color", err)
	}
}

// Destroy frees all assets obtained by the ui.Component
func (bb *BottomBar) Destroy() {
	bb.backBuf.Destroy()
	bb.textBuf.Destroy()
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (bb *BottomBar) InBoundary(pt sdl.Point) bool {
	return ui.InBounds(*bb.area, pt)
}

// Render draws the ui.Component
func (bb *BottomBar) Render() {
	sw := util.Start()
	msg := <-bb.comms

	// first render solid color background
	gl.Viewport(bb.area.X, 0, bb.area.W, bb.area.H)
	bb.backProgram.Bind()
	bb.backBuf.Draw()
	bb.backProgram.Unbind()

	// second render text on top
	// TODO optimize rendering by no-oping if string hasn't changed (or window size)
	fileNameMessage := msg.FileName
	zoomMessage := fmt.Sprintf("2^(%v)", msg.Mult)
	mousePixMessage := fmt.Sprintf("(%v, %v)", msg.MousePix.X, msg.MousePix.Y)

	pos := gfx.Point{X: 0, Y: bb.cfg.BottomBarHeight / 2}
	align := gfx.Align{V: gfx.AlignMiddle, H: gfx.AlignLeft}
	fileNameTriangles := bb.font.MapString(fileNameMessage, pos, align)
	pos = gfx.Point{X: bb.cfg.ScreenWidth / 2, Y: bb.cfg.BottomBarHeight / 2}
	align = gfx.Align{V: gfx.AlignMiddle, H: gfx.AlignCenter}
	zoomTriangles := bb.font.MapString(zoomMessage, pos, align)
	pos = gfx.Point{X: bb.cfg.ScreenWidth, Y: bb.cfg.BottomBarHeight / 2}
	align = gfx.Align{V: gfx.AlignMiddle, H: gfx.AlignRight}
	mousePixTriangles := bb.font.MapString(mousePixMessage, pos, align)
	triangles := make([]float32, 0, len(fileNameTriangles)+len(zoomTriangles)+len(mousePixTriangles))
	triangles = append(triangles, fileNameTriangles...)
	triangles = append(triangles, zoomTriangles...)
	triangles = append(triangles, mousePixTriangles...)

	gl.Viewport(0, 0, bb.cfg.ScreenWidth, bb.cfg.ScreenHeight)
	bb.textProgram.Bind()
	bb.font.GetTexture().Bind()

	err := bb.textBuf.Load(triangles, gl.STATIC_DRAW)
	if err != nil {
		log.Warnf("failed to load bottom bar text triangles: %v", err)
	} else {
		bb.textBuf.Draw()
	}

	bb.font.GetTexture().Unbind()
	bb.textProgram.Unbind()

	sw.StopRecordAverage(bb.String() + ".Render")
}

// OnEnter is called when the cursor enters the ui.Component's region
func (bb *BottomBar) OnEnter() {}

// OnLeave is called when the cursor leaves the ui.Component's region
func (bb *BottomBar) OnLeave() {}

// OnMotion is called when the cursor moves within the ui.Component's region
func (bb *BottomBar) OnMotion(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (bb *BottomBar) OnScroll(evt *sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the ui.Component's region
func (bb *BottomBar) OnClick(evt *sdl.MouseButtonEvent) bool {
	return true
}

// OnResize is called when the user resizes the window
func (bb *BottomBar) OnResize(x, y int32) {
	bb.area.W += x
	bb.area.Y += y

	err := bb.textProgram.UploadUniform("screen_size", float32(bb.cfg.ScreenWidth), float32(bb.cfg.ScreenHeight))
	if err != nil {
		log.Warnf("failed to upload uniform \"%v\": %v", "screen_size", err)
	}
}

// String returns the name of the component type
func (bb *BottomBar) String() string {
	return "app.BottomBar"
}
