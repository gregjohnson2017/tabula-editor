package menu

import (
	"fmt"
	"math"

	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/font"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

// Bar is a horizontal menu bar
type Bar struct {
	area    sdl.Rect // the area encompassing the menubar buttons
	entries []*entry
	cfg     *config.Config
	hover   *entry
}

// NewBar returns the menubar matching the given definition
func NewBar(cfg *config.Config, menus []Definition) (*Bar, error) {
	b := Bar{
		cfg: cfg,
	}

	b.entries = make([]*entry, 0, len(menus))
	fnt, err := font.LoadFontTexture("NotoMono-Regular.ttf", 14)
	if err != nil {
		return nil, err
	}

	b.area.X = 0
	b.area.Y = 0

	// normalize height
	var max int32
	for _, c := range menus {
		_, h := font.CalcStringDims(c.Text, fnt)
		h32 := int32(math.Ceil(h)) + 10
		if h32 > max {
			max = h32
		}
	}

	// populate list of menu entries with appropriate boundaries
	var off int32
	for _, child := range menus {
		w, _ := font.CalcStringDims(child.Text, fnt)
		w32 := int32(math.Ceil(w)) + 14
		var area *sdl.Rect
		area = &sdl.Rect{X: b.area.X + off, Y: b.area.Y, W: w32, H: max}
		entry, err := newEntry(b.cfg, area, child.Text, ui.Align{V: ui.AlignBelow}, child.Children, child.Action)
		if err != nil {
			return nil, err
		}
		b.entries = append(b.entries, entry)
		off += area.W
	}
	b.area.W = off
	b.area.H = max

	return &b, nil
}

// InBoundary returns whether a point is handled by the menubar
func (b *Bar) InBoundary(pt sdl.Point) bool {
	for _, e := range b.entries {
		if e.InBoundary(pt) {
			return true
		}
	}
	return false
}

// Render renders the menubar and all its components
func (b *Bar) Render() {
	for _, e := range b.entries {
		e.Render()
	}
}

// Destroy tears down the menubar's state
func (b *Bar) Destroy() {
	for _, e := range b.entries {
		e.Destroy()
	}
}

// OnEnter updates the menubar for when the mouse enters its boundary
func (b *Bar) OnEnter() {
}

// OnLeave updates the menubar for when the mouse leaves its boundary
func (b *Bar) OnLeave() {
	if b.hover != nil {
		b.hover.OnLeave()
	}
	b.hover = nil
}

// GetEntryAt returns the menu entry below the given mouse coordinates
func (b *Bar) GetEntryAt(x int32, y int32) (*entry, error) {
	for _, e := range b.entries {
		if e.InBoundary(sdl.Point{X: x, Y: y}) {
			return e, nil
		}
	}
	return nil, fmt.Errorf("no entry at given position")
}

// OnMotion updates the menubar for when the mouse moves within its boundary
func (b *Bar) OnMotion(evt *sdl.MouseMotionEvent) bool {
	e, err := b.GetEntryAt(evt.X, evt.Y)
	if err != nil {
		return false
	}

	if e == b.hover {
		e.OnMotion(evt)
		return true
	}

	if b.hover != nil {
		b.hover.OnLeave()
		b.hover.enabled = false
	}
	e.OnEnter()
	if evt.State == sdl.ButtonLMask() {
		btnEvt := sdl.MouseButtonEvent{
			Type:      sdl.MOUSEBUTTONDOWN,
			Timestamp: evt.Timestamp,
			WindowID:  evt.WindowID,
			Which:     evt.Which,
			State:     sdl.PRESSED,
			X:         evt.X,
			Y:         evt.Y,
			Button:    sdl.BUTTON_LEFT,
		}
		e.OnClick(&btnEvt)
	}

	b.hover = e
	e.OnMotion(evt)
	return true
}

// OnScroll updates the menubar for when the mouse scrolls within its boundary
func (b *Bar) OnScroll(*sdl.MouseWheelEvent) bool {
	return true
}

// OnClick updates the menubar for when the mouse clicks within its boundary
func (b *Bar) OnClick(evt *sdl.MouseButtonEvent) bool {
	e, err := b.GetEntryAt(evt.X, evt.Y)
	if err != nil {
		return false
	}
	return e.OnClick(evt)
}

// OnResize updates the menubar for when the window is resized
func (b *Bar) OnResize(x, y int32) {
	for _, c := range b.entries {
		c.OnResize(x, y)
	}
}

// String returns the name of the component type
func (b *Bar) String() string {
	return "Bar"
}
