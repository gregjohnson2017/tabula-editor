package tabula

import "github.com/veandco/go-sdl2/sdl"

// ImageTool defines the behavior of tools used for the image view
type ImageTool interface {
	OnClick(evt *sdl.MouseButtonEvent, iv *ImageView)
	OnMotion(evt *sdl.MouseMotionEvent, iv *ImageView)
}

// Make sure the tools satisfy the interface
var _ ImageTool = ImageTool(EmptyTool{})
var _ ImageTool = ImageTool(PixelSelectionTool{})

// EmptyTool does nothing
type EmptyTool struct {
}

// OnClick does nothing
func (t EmptyTool) OnClick(evt *sdl.MouseButtonEvent, iv *ImageView) {
}

// OnMotion does nothing
func (t EmptyTool) OnMotion(evt *sdl.MouseMotionEvent, iv *ImageView) {
}

// PixelSelectionTool selects any clicked pixels
type PixelSelectionTool struct {
}

// OnClick is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelSelectionTool) OnClick(evt *sdl.MouseButtonEvent, iv *ImageView) {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		iv.SelectPixel(iv.mousePix.x, iv.mousePix.y)
	}
}

// OnMotion is called when the user clicks within the Image View's region and the tool is
// currently active for the image view.
func (t PixelSelectionTool) OnMotion(evt *sdl.MouseMotionEvent, iv *ImageView) {
	if evt.State == sdl.ButtonLMask() && inBounds(iv.canvas, evt.X, evt.Y) {
		iv.selection.Add(iv.mousePix.x + iv.mousePix.y*iv.origW)
	}
}
