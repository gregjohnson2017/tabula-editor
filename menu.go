package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

var _ UIComponent = UIComponent(&Menu{})

// Menu defines a list of buttons
type Menu struct {
	area    *sdl.Rect
	buttons []*Button
	ctx     *context
}

// NewMenu returns a pointer to a Menu struct that implements UIComponent
func NewMenu(ctx *context, buttons []*Button, x, y int32) (*Menu, error) {
	if len(buttons) == 0 {
		return nil, fmt.Errorf("cannot create empty menu")
	}
	var menuH int32
	var menuW = buttons[0].area.W
	for _, b := range buttons {
		menuH += b.area.H
		if b.area.W > menuW {
			menuW = b.area.W
		}
	}

	return &Menu{
		area:    &sdl.Rect{X: x, Y: y, W: menuW, H: menuH},
		ctx:     ctx,
		buttons: buttons,
	}, nil
}

// Destroy frees all surfaces and textures in the Menu
func (m *Menu) Destroy() {}

// GetBoundary returns the clickable region of the UIComponent
func (m *Menu) GetBoundary() *sdl.Rect {
	return m.area
}

// Render draws the UIComponent
func (m *Menu) Render() error {

	return nil
}

// OnEnter is called when the cursor enters the UIComponent's region
func (m *Menu) OnEnter() {}

// OnLeave is called when the cursor leaves the UIComponent's region
func (m *Menu) OnLeave() {}

// OnMotion is called when the cursor moves within the UIComponent's region
func (m *Menu) OnMotion(evt *sdl.MouseMotionEvent) bool {
	return true
}

// OnScroll is called when the user scrolls within the UIComponent's region
func (m *Menu) OnScroll(evt *sdl.MouseWheelEvent) bool {
	return true
}

// OnClick is called when the user clicks within the UIComponent's region
func (m *Menu) OnClick(evt *sdl.MouseButtonEvent) bool {
	return true
}

// OnResize is called when the user resizes the window
func (m *Menu) OnResize(x, y int32) {}

// String returns the name of the component type
func (m *Menu) String() string {
	return "Menu"
}
