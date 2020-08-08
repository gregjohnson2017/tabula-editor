package ui

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// Component describes which functions a UI component must implement
type Component interface {
	InBoundary(sdl.Point) bool
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

// Interpolate returns all the points hit by the line connecting points a and b.
// The interpolated points are returned in random order.
// The returned points excludes a (beginning) and includes b (ending).
func Interpolate(a, b sdl.Point) []sdl.Point {
	// The calculations are done using int32s so are not the most accurate,
	// but a solution with floating point numbers would be excessively costly.
	lp, rp := b, a
	if a.X < b.X {
		lp, rp = a, b
	}
	bp, tp := b, a
	if a.Y < b.Y {
		bp, tp = a, b
	}
	pointMap := make(map[sdl.Point]struct{}, (rp.X-lp.X)+(tp.Y-bp.Y))

	// interpolate points left to right
	denomX := rp.X - lp.X
	for x := lp.X + 1; x < rp.X; x++ {
		nomin := lp.Y*(rp.X-x) + rp.Y*(x-lp.X)
		y := nomin / denomX
		pointMap[sdl.Point{X: x, Y: y}] = struct{}{}
	}

	// interpolate points bottom to top
	denomY := tp.Y - bp.Y
	for y := bp.Y + 1; y < tp.Y; y++ {
		nomin := bp.X*(tp.Y-y) + tp.X*(y-bp.Y)
		x := nomin / denomY
		pointMap[sdl.Point{X: x, Y: y}] = struct{}{}
	}

	// aggregate and output
	points := make([]sdl.Point, 0, len(pointMap)+1)
	for p := range pointMap {
		points = append(points, p)
	}
	return append(points, b)
}
