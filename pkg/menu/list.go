package menu

import (
	"fmt"
	"math"

	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/font"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/gregjohnson2017/tabula-editor/pkg/util"
	"github.com/veandco/go-sdl2/sdl"
)

// list is a vertical menu list
type list struct {
	area    sdl.Rect
	cfg     *config.Config
	entries []*entry
	hover   *entry
}

// newList returns a pointer to a new list struct that implements ui.Component
func newList(cfg *config.Config, pos sdl.Point, menus []Definition) (*list, error) {
	l := list{
		cfg: cfg,
	}

	l.entries = make([]*entry, 0, len(menus))
	fnt, err := font.LoadFontTexture("NotoMono-Regular.ttf", 14)
	if err != nil {
		return nil, err
	}

	l.area.X = pos.X
	l.area.Y = pos.Y

	// normalize width
	var max int32
	for _, c := range menus {
		w, _ := font.CalcStringDims(c.Text, fnt)
		w32 := int32(math.Ceil(w)) + 14
		if w32 > max {
			max = w32
		}
	}

	// populate list of menu entries with appropriate boundaries
	var off int32
	for _, child := range menus {
		_, h := font.CalcStringDims(child.Text, fnt)
		h32 := int32(math.Ceil(h)) + 10
		area := &sdl.Rect{X: l.area.X, Y: l.area.Y + off, W: max, H: h32}
		entry, err := newEntry(l.cfg, area, child.Text, ui.Align{H: ui.AlignRight}, child.Children, child.Action)
		if err != nil {
			return nil, err
		}
		l.entries = append(l.entries, entry)
		off += area.H
	}
	l.area.W = max
	l.area.H = off

	return &l, nil
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (l *list) InBoundary(pt sdl.Point) bool {
	for _, c := range l.entries {
		if c.InBoundary(pt) {
			return true
		}
	}
	return false
}

// Render draws the ui.Component
func (l *list) Render() {
	sw := util.Start()
	for _, e := range l.entries {
		e.Render()
	}
	sw.StopRecordAverage(l.String() + ".Render")
}

// Destroy frees all assets acquired by the ui.Component
func (l *list) Destroy() {
	for _, e := range l.entries {
		e.Destroy()
	}
}

// OnEnter is called when the cursor enters the ui.Component's region
func (l *list) OnEnter() {
}

// OnLeave is called when the cursor leaves the ui.Component's region
func (l *list) OnLeave() {
	if l.hover != nil {
		l.hover.OnLeave()
	}
	l.hover = nil
}

// GetEntryAt returns the menu entry below the given mouse coordinates
func (l *list) GetEntryAt(x int32, y int32) (*entry, error) {
	for _, e := range l.entries {
		if e.InBoundary(sdl.Point{X: x, Y: y}) {
			return e, nil
		} else if e.enabled && e.list != nil {
			if me, err := e.list.GetEntryAt(x, y); err == nil {
				return me, nil
			}
		}
	}
	return nil, fmt.Errorf("GetEntryAt(%v, %v): %w", x, y, ErrNoEntryAtPosition)
}

// OnMotion is called when the cursor moves within the ui.Component's region - bad comment
func (l *list) OnMotion(evt *sdl.MouseMotionEvent) bool {
	e, err := l.GetEntryAt(evt.X, evt.Y)
	if err != nil {
		return false
	}

	if e == l.hover {
		e.OnMotion(evt)
		return true
	}

	if l.hover != nil {
		l.hover.OnLeave()
		l.hover.enabled = false
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
	l.hover = e
	e.OnMotion(evt)
	return true
}

// OnScroll is called when the user scrolls within the ui.Component's region
func (l *list) OnScroll(*sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the ui.Component's region
func (l *list) OnClick(evt *sdl.MouseButtonEvent) bool {
	e, err := l.GetEntryAt(evt.X, evt.Y)
	if err != nil {
		return false
	}
	return e.OnClick(evt)
}

// OnResize is called when the user resizes the window
func (l *list) OnResize(x, y int32) {
	for _, c := range l.entries {
		c.OnResize(x, y)
	}
}

func (l *list) String() string {
	return "menu.list"
}
