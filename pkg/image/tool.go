package image

import (
	"fmt"

	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

// Tool defines the behavior of tools used for the image view.
type Tool interface {
	OnClick(evt *sdl.MouseButtonEvent, iv *View)
	OnMotion(evt *sdl.MouseMotionEvent, iv *View)
	fmt.Stringer
}

// Make sure the tools satisfy the interface
var _ Tool = Tool(EmptyTool{})
var _ Tool = Tool(&PixelSelectionTool{})
var _ Tool = Tool(&PixelColorTool{})

// EmptyTool does nothing.
type EmptyTool struct {
}

// OnClick does nothing.
func (t EmptyTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
}

// OnMotion does nothing.
func (t EmptyTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
}

func (t EmptyTool) String() string {
	return "image.EmptyTool"
}

// PixelSelectionTool selects any clicked pixels
type PixelSelectionTool struct {
	lastDrag sdl.Point
}

// OnClick is called when the user clicks within the Image View's region and the
// tool is currently active for the image view.
func (t *PixelSelectionTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		_ = iv.SelectPixel(iv.mousePix)
		t.lastDrag = iv.mousePix
	}
}

// OnMotion is called when the user clicks within the Image View's region and
// the tool is currently active for the image view.
func (t *PixelSelectionTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
	if evt.State == sdl.ButtonLMask() {
		points := ui.Interpolate(iv.mousePix, t.lastDrag)
		for _, p := range points {
			_ = iv.SelectPixel(p)
		}
		t.lastDrag = iv.mousePix
	}
}

func (t *PixelSelectionTool) String() string {
	return "image.PixelSelectionTool"
}

// PixelColorTool colors any clicked pixels purple
type PixelColorTool struct {
	lastDrag sdl.Point
}

// OnClick is called when the user clicks within the Image View's region and the
// tool is currently active for the image view.
func (t *PixelColorTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		_ = iv.setPixel(iv.mousePix, []byte{0xFF, 0x00, 0xFF, 0xFF})
		t.lastDrag = iv.mousePix
	}
}

// OnMotion is called when the user clicks within the Image View's region and
// the tool is currently active for the image view.
func (t *PixelColorTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
	if evt.State == sdl.ButtonLMask() {
		for _, p := range ui.Interpolate(iv.mousePix, t.lastDrag) {
			_ = iv.setPixel(p, []byte{0xFF, 0x00, 0xFF, 0xFF})
		}
		t.lastDrag = iv.mousePix
	}
}

func (t *PixelColorTool) String() string {
	return "image.PixelColorTool"
}
