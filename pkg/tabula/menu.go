package tabula

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&MenuList{})

// MenuEntry is the clickable entry which opens a MenuList
type MenuEntry struct {
	enabled bool
	button  *MenuBarButton
	list    *MenuList
	action  func()
}

// MenuBarButton defines an interactive button, but redefines OnClick to perform action on press, not release
type MenuBarButton struct {
	*Button
}

// OnClick is called when the user clicks within the UIComponent's region
func (mbb *MenuBarButton) OnClick(evt *sdl.MouseButtonEvent) bool {
	if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.PRESSED {
		mbb.pressed = true
		mbb.action()
	} else if evt.Button == sdl.BUTTON_LEFT && evt.State == sdl.RELEASED && mbb.pressed {
		mbb.pressed = false
		// TODO add release action
	}
	return true
}

// NewMenuEntry returns the struct with the given label and list
func NewMenuEntry(area *sdl.Rect, label string, list *MenuList, cfg *Config, act func()) (MenuEntry, error) {
	if cfg == nil {
		return MenuEntry{}, fmt.Errorf("NewMenuEntry found nil Config")
	}
	if list == nil && act == nil {
		return MenuEntry{}, fmt.Errorf("NewMenuEntry needs a list and/or an action")
	}
	var btn *Button
	var err error
	if btn, err = NewButton(area, cfg, label, act); err != nil {
		return MenuEntry{}, err
	}
	return MenuEntry{
		enabled: false,
		button:  &MenuBarButton{btn},
		list:    list,
	}, nil
}

// Destroy calls destroy on underlying UIComponents
func (me MenuEntry) Destroy() {
	me.button.Destroy()
	me.list.Destroy()
}

// InBoundary returns whether a point is in this UIComponent's bounds
func (me MenuEntry) InBoundary(pt sdl.Point) bool {
	if me.button.InBoundary(pt) {
		return true
	}
	if me.enabled && me.list != nil && me.list.InBoundary(pt) {
		return true
	}
	return false
}

// GetBoundary returns the underlying button's boundary
func (me MenuEntry) GetBoundary() *sdl.Rect {
	return me.button.GetBoundary()
}

// OnEnter calls the underlying button's OnEnter method
func (me *MenuEntry) OnEnter() {
	me.button.OnEnter()
}

// OnLeave calls the underlying button's OnLeave method
func (me *MenuEntry) OnLeave() {
	me.button.OnLeave()
	me.list.OnLeave()
}

// OnClick calls the underlying button's OnClick method
func (me *MenuEntry) OnClick(evt *sdl.MouseButtonEvent) bool {
	if evt.Button != sdl.BUTTON_LEFT || evt.State != sdl.PRESSED {
		return true
	}
	if inBounds(me.GetBoundary(), evt.X, evt.Y) {
		me.button.OnClick(evt)
		me.enabled = !me.enabled
		return true
	}
	if e, err := me.list.GetEntryAt(evt.X, evt.Y); err == nil {
		e.OnClick(evt)
	}
	return true
}

// Render calls the underlying button's render function
func (me MenuEntry) Render() {
	me.button.Render()
	if me.list != nil && me.enabled {
		me.list.Render()
	}
}

// OnResize calls the underlying UIComponents' OnResize function
func (me MenuEntry) OnResize(x, y int32) {
	me.button.OnResize(x, y)
	me.list.OnResize(x, y)
}

// OnMotion is called when the cursor moves within the UIComponent's region - bad comment
func (me *MenuEntry) OnMotion(evt *sdl.MouseMotionEvent) bool {
	if me.list.InBoundary(sdl.Point{X: evt.X, Y: evt.Y}) {
		me.list.OnMotion(evt)
	}
	return true
}

// MenuList is the horizontal menu bar
type MenuList struct {
	area    *sdl.Rect
	cfg     *Config
	entries []MenuEntry
	hover   *MenuEntry
	horiz   bool
}

// NewMenuList returns a pointer to a new MenuList struct that implements UIComponent
func NewMenuList(cfg *Config, horiz bool) *MenuList {
	return &MenuList{
		area:  &sdl.Rect{},
		cfg:   cfg,
		horiz: horiz,
	}
}

// SetChildren registers a set of menu entries with the menu bar
func (ml *MenuList) SetChildren(offx int32, offy int32, childs []struct {
	str string
	ml  *MenuList
	act func()
}) error {
	for _, e := range ml.entries {
		e.Destroy()
	}
	ml.entries = make([]MenuEntry, 0, len(childs))
	_, runeMap, err := loadFontTexture("NotoMono-Regular.ttf", 14)
	if err != nil {
		return err
	}
	ml.area.X = offx
	ml.area.Y = offy
	// normalize height or width
	var max int32
	for _, c := range childs {
		w, h := calcStringDims(c.str, runeMap)
		w32 := int32(math.Ceil(w)) + 14
		h32 := int32(math.Ceil(h)) + 10
		if ml.horiz {
			if h32 > max {
				max = h32
			}
		} else {
			if w32 > max {
				max = w32
			}
		}
	}
	// populate list of menu entries with appropriately boundaries
	var off int32
	for _, child := range childs {
		w, h := calcStringDims(child.str, runeMap)
		w32 := int32(math.Ceil(w)) + 14
		h32 := int32(math.Ceil(h)) + 10
		var area *sdl.Rect
		if ml.horiz {
			area = &sdl.Rect{X: ml.area.X + off, Y: ml.area.Y, W: w32, H: max}
		} else {
			area = &sdl.Rect{X: ml.area.X, Y: ml.area.Y + off, W: max, H: h32}
		}
		entry, err := NewMenuEntry(area, child.str, child.ml, ml.cfg, child.act)
		if err != nil {
			return err
		}
		ml.entries = append(ml.entries, entry)
		r := entry.GetBoundary()
		if ml.horiz {
			off += r.W
		} else {
			off += r.H
		}
	}
	if ml.horiz {
		ml.area.W = off
		ml.area.H = max
	} else {
		ml.area.W = max
		ml.area.H = off
	}
	return nil
}

// InBoundary returns whether a point is in this UIComponent's bounds
func (ml MenuList) InBoundary(pt sdl.Point) bool {
	for _, c := range ml.entries {
		if c.InBoundary(pt) {
			return true
		}
	}
	return false
}

// GetBoundary returns the clickable region of the UIComponent
func (ml *MenuList) GetBoundary() *sdl.Rect {
	return ml.area
}

// Render draws the UIComponent
func (ml *MenuList) Render() {
	for _, e := range ml.entries {
		e.Render()
	}
}

// Destroy frees all assets acquired by the UIComponent
func (ml *MenuList) Destroy() {
}

// OnEnter is called when the cursor enters the UIComponent's region
func (ml *MenuList) OnEnter() {
}

// OnLeave is called when the cursor leaves the UIComponent's region
func (ml *MenuList) OnLeave() {
	if ml.hover != nil {
		ml.hover.OnLeave()
	}
	ml.hover = nil
}

// GetEntryAt returns the menu entry below the given mouse coordinates
func (ml *MenuList) GetEntryAt(x int32, y int32) (*MenuEntry, error) {
	for i := range ml.entries {
		c := &ml.entries[i]
		if c.InBoundary(sdl.Point{X: x, Y: y}) {
			return c, nil
		} else if c.enabled && c.list != nil {
			if me, err := c.list.GetEntryAt(x, y); err == nil {
				return me, nil
			}
		}
	}
	return nil, fmt.Errorf("no entry at given position")
}

// OnMotion is called when the cursor moves within the UIComponent's region - bad comment
func (ml *MenuList) OnMotion(evt *sdl.MouseMotionEvent) bool {
	e, err := ml.GetEntryAt(evt.X, evt.Y)
	if err != nil {
		return false
	}
	if e != ml.hover {
		if ml.hover != nil {
			ml.hover.OnLeave()
			ml.hover.enabled = false
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
		ml.hover = e
	}
	e.OnMotion(evt)
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (ml *MenuList) OnScroll(*sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the UIComponent's region
func (ml *MenuList) OnClick(evt *sdl.MouseButtonEvent) bool {
	e, err := ml.GetEntryAt(evt.X, evt.Y)
	if err != nil {
		return false
	}
	return e.OnClick(evt)
}

// OnResize is called when the user resizes the window
func (ml *MenuList) OnResize(x, y int32) {
	for _, c := range ml.entries {
		c.OnResize(x, y)
	}
}

func (ml *MenuList) String() string {
	return "MenuList"
}
