package ui

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// Component describes which functions a UI component must implement
type Component interface {
	InBoundary(sdl.Point) bool
	GetBoundary() *sdl.Rect
	Render()
	Destroy()
	OnEnter()
	OnLeave()
	OnMotion(*sdl.MouseMotionEvent) bool
	OnScroll(*sdl.MouseWheelEvent) bool
	OnClick(*sdl.MouseButtonEvent) bool
	OnResize(x, y int32)
	fmt.Stringer
}

// AlignV is used for the positioning of elements vertically
type AlignV int

const (
	// AlignAbove puts the top side at the y coordinate
	AlignAbove AlignV = iota - 1
	// AlignMiddle puts the top and bottom sides equidistant from the middle
	AlignMiddle
	// AlignBelow puts the bottom side on the y coordinate
	AlignBelow
)

// AlignH is used for the positioning of elements horizontally
type AlignH int

const (
	// AlignLeft puts the left side on the x coordinate
	AlignLeft AlignH = iota - 1
	//AlignCenter puts the left and right sides equidistant from the center
	AlignCenter
	// AlignRight puts the right side at the x coordinate
	AlignRight
)

// Align holds vertical and horizontal alignments
type Align struct {
	V AlignV
	H AlignH
}

func InBounds(area sdl.Rect, point sdl.Point) bool {
	if point.X < area.X || point.X >= area.X+area.W || point.Y < area.Y || point.Y >= area.Y+area.H {
		return false
	}
	return true
}
