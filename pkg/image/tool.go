package image

import (
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

// Tool defines the behavior of tools used for the image view
type Tool interface {
	OnClick(evt *sdl.MouseButtonEvent, iv *View)
	OnMotion(evt *sdl.MouseMotionEvent, iv *View)
}

// Make sure the tools satisfy the interface
var _ Tool = Tool(EmptyTool{})
var _ Tool = Tool(PixelSelectionTool{})

// EmptyTool does nothing
type EmptyTool struct {
}

// OnClick does nothing
func (t EmptyTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
}

// OnMotion does nothing
func (t EmptyTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
}

// PixelSelectionTool selects any clicked pixels
type PixelSelectionTool struct {
}

// OnClick is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelSelectionTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		iv.SelectPixel(iv.mousePix.X, iv.mousePix.Y)
	}
}

// OnMotion is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelSelectionTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
	if evt.State == sdl.ButtonLMask() && ui.InBounds(*iv.canvas, sdl.Point{evt.X, evt.Y}) {
		iv.selection.Add(iv.mousePix.X + iv.mousePix.Y*iv.origW)
	}
}

// PixelColorTool colors any clicked pixels purple
type PixelColorTool struct {
}

// OnClick is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelColorTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		iv.setPixel(iv.mousePix.X, iv.mousePix.Y, []byte{0xff, 0x00, 0xff, 0xff})
	}
}

// OnMotion is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelColorTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
	if evt.State == sdl.ButtonLMask() && ui.InBounds(*iv.canvas, sdl.Point{evt.X, evt.Y}) {
		iv.setPixel(iv.mousePix.X, iv.mousePix.Y, []byte{0xff, 0x00, 0xff, 0xff})
	}
}
