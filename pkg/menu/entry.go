package menu

import (
	"fmt"

	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

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
func NewMenuEntry(area *sdl.Rect, label string, list *MenuList, cfg *config.Config, act func()) (MenuEntry, error) {
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
	if ui.InBounds(*me.GetBoundary(), sdl.Point{evt.X, evt.Y}) {
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
