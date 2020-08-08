package ui_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/gregjohnson2017/tabula-editor/pkg/ui"
	"github.com/veandco/go-sdl2/sdl"
)

func benchPoints(p1, p2 sdl.Point) func(b *testing.B) {
	return func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ui.Interpolate(p1, p2)
		}
	}
}

func point(x, y int32) sdl.Point {
	return sdl.Point{X: x, Y: y}
}

func BenchmarkInterpolate(b *testing.B) {
	// each are distance 10
	b.Run("diagonal line", benchPoints(point(1, 1), point(9, 7)))
	b.Run("vertical line", benchPoints(point(1, 1), point(1, 11)))
	b.Run("horizontal line", benchPoints(point(1, 1), point(11, 1)))
}

type pointSlice []sdl.Point

func (p pointSlice) Len() int           { return len(p) }
func (p pointSlice) Less(i, j int) bool { return p[i].X < p[j].X || p[i].Y < p[j].Y }
func (p pointSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func testPoints(a, b sdl.Point, expected []sdl.Point) func(t *testing.T) {
	return func(t *testing.T) {
		actual := ui.Interpolate(a, b)
		sort.Sort(pointSlice(expected))
		sort.Sort(pointSlice(actual))
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("expected != actual\nexpected: %v\nactual: %v", expected, actual)
		}
	}
}

func TestInterpolate(t *testing.T) {
	t.Run("diagonal line up-right", testPoints(point(1, 1), point(9, 7), []sdl.Point{
		point(2, 1), point(2, 2), point(3, 2), point(3, 3), point(4, 3),
		point(5, 4), point(6, 4), point(6, 5), point(7, 5), point(7, 6),
		point(8, 6), point(9, 7),
	}))
	t.Run("diagonal line down-left", testPoints(point(9, 7), point(1, 1), []sdl.Point{
		point(1, 1), point(2, 1), point(2, 2), point(3, 2), point(3, 3),
		point(4, 3), point(5, 4), point(6, 4), point(6, 5), point(7, 5),
		point(7, 6), point(8, 6),
	}))
	t.Run("vertical line up", testPoints(point(1, 1), point(1, 11), []sdl.Point{
		point(1, 2), point(1, 3), point(1, 4), point(1, 5), point(1, 6),
		point(1, 7), point(1, 8), point(1, 9), point(1, 10), point(1, 11),
	}))
	t.Run("vertical line down", testPoints(point(1, 11), point(1, 1), []sdl.Point{
		point(1, 1), point(1, 2), point(1, 3), point(1, 4), point(1, 5),
		point(1, 6), point(1, 7), point(1, 8), point(1, 9), point(1, 10),
	}))
	t.Run("horizontal line right", testPoints(point(1, 1), point(11, 1), []sdl.Point{
		point(2, 1), point(3, 1), point(4, 1), point(5, 1), point(6, 1),
		point(7, 1), point(8, 1), point(9, 1), point(10, 1), point(11, 1),
	}))
	t.Run("horizontal line left", testPoints(point(11, 1), point(1, 1), []sdl.Point{
		point(1, 1), point(2, 1), point(3, 1), point(4, 1), point(5, 1),
		point(6, 1), point(7, 1), point(8, 1), point(9, 1), point(10, 1),
	}))
}
