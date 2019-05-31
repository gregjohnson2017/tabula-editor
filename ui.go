package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

// UIComponent says what functions a UIComponent must implement
type UIComponent interface {
	getBoundary() *sdl.Rect
	render() (*sdl.Texture, error)
	onEnter(*sdl.MouseMotionEvent)
	onLeave(*sdl.MouseMotionEvent)
	onClick(*sdl.MouseButtonEvent)
}
