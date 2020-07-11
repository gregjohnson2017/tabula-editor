package menu

import (
	"github.com/gregjohnson2017/tabula-editor/pkg/config"
	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

// entry is the clickable entry which opens a list
type entry struct {
	enabled bool
	button  *Button
	list    *list
}

// newEntry returns the struct with the given label and list
func newEntry(cfg *config.Config, area *sdl.Rect, label string, align ui.Align, menus []Definition, act func()) (*entry, error) {
	btn, err := NewButton(area, cfg, label, act)
	if err != nil {
		return nil, err
	}

	pos := sdl.Point{X: area.X, Y: area.Y}
	if align.H == ui.AlignRight {
		pos.X += area.W
	}
	if align.V == ui.AlignBelow {
		pos.Y += area.H
	}
	list, err := newList(cfg, pos, menus)
	if err != nil {
		return nil, err
	}

	return &entry{
		enabled: false,
		button:  btn,
		list:    list,
	}, nil
}

// Destroy calls destroy on underlying ui.Components
func (e *entry) Destroy() {
	e.button.Destroy()
	e.list.Destroy()
}

// InBoundary returns whether a point is in this ui.Component's bounds
func (e *entry) InBoundary(pt sdl.Point) bool {
	if e.button.InBoundary(pt) {
		return true
	}
	if e.enabled && e.list != nil && e.list.InBoundary(pt) {
		return true
	}
	return false
}

// OnEnter calls the underlying button's OnEnter method
func (e *entry) OnEnter() {
	e.button.OnEnter()
}

// OnLeave calls the underlying button's OnLeave method
func (e *entry) OnLeave() {
	e.button.OnLeave()
	e.list.OnLeave()
	e.enabled = false
}

// OnClick calls the underlying button's OnClick method
func (e *entry) OnClick(evt *sdl.MouseButtonEvent) bool {
	if evt.Button != sdl.BUTTON_LEFT {
		return true
	}
	if e.button.InBoundary(sdl.Point{X: evt.X, Y: evt.Y}) {
		e.button.OnClick(evt)
		if evt.State == sdl.RELEASED {
			e.enabled = !e.enabled
		}
		return true
	}
	if e, err := e.list.GetEntryAt(evt.X, evt.Y); err == nil {
		e.OnClick(evt)
	}
	return true
}

// Render calls the underlying button's render function
func (e *entry) Render() {
	e.button.Render()
	if e.list != nil && e.enabled {
		e.list.Render()
	}
}

// OnResize calls the underlying ui.Components' OnResize function
func (e *entry) OnResize(x, y int32) {
	e.button.OnResize(x, y)
	e.list.OnResize(x, y)
}

// OnMotion is called when the cursor moves within the ui.Component's region - bad comment
func (e *entry) OnMotion(evt *sdl.MouseMotionEvent) bool {
	if e.list.InBoundary(sdl.Point{X: evt.X, Y: evt.Y}) {
		e.list.OnMotion(evt)
	}
	return true
}
