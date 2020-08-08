package image

import (
	"fmt"

	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

// Tool defines the behavior of tools used for the image view
type Tool interface {
	OnClick(evt *sdl.MouseButtonEvent, iv *View)
	OnMotion(evt *sdl.MouseMotionEvent, iv *View)
	fmt.Stringer
}

// Make sure the tools satisfy the interface
var _ Tool = Tool(EmptyTool{})
var _ Tool = Tool(PixelSelectionTool{})
var _ Tool = Tool(PixelColorTool{})

// EmptyTool does nothing
type EmptyTool struct {
}

// OnClick does nothing
func (t EmptyTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
}

// OnMotion does nothing
func (t EmptyTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
}

func (t EmptyTool) String() string {
	return "image.EmptyTool"
}

// PixelSelectionTool selects any clicked pixels
type PixelSelectionTool struct {
}

// OnClick is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelSelectionTool) OnClick(evt *sdl.MouseButtonEvent, iv *View) {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		err := iv.SelectPixel(iv.mousePix.X, iv.mousePix.Y)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// OnMotion is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelSelectionTool) OnMotion(evt *sdl.MouseMotionEvent, iv *View) {
	if evt.State == sdl.ButtonLMask() {
		err := iv.SelectPixel(iv.mousePix.X, iv.mousePix.Y)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (t PixelSelectionTool) String() string {
	return "image.PixelSelectionTool"
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
	if evt.State == sdl.ButtonLMask() && ui.InBounds(*iv.canvas, sdl.Point{X: evt.X, Y: evt.Y}) {
		iv.setPixel(iv.mousePix.X, iv.mousePix.Y, []byte{0xff, 0x00, 0xff, 0xff})
	}
}

func (t PixelColorTool) String() string {
	return "image.PixelColorTool"
}
