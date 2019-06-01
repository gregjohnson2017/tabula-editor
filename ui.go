package main

import "github.com/veandco/go-sdl2/sdl"

// UIComponent says what functions a UIComponent must implement
type UIComponent interface {
	getBoundary() *sdl.Rect
	render(*sdl.Renderer) error
	onEnter(*sdl.MouseMotionEvent) bool
	onLeave(*sdl.MouseMotionEvent) bool
	onMotion(*sdl.MouseMotionEvent) bool
	onScroll(*sdl.MouseWheelEvent) bool
	onClick(*sdl.MouseButtonEvent) bool
}

type context struct {
	Win      *sdl.Window
	Rend     *sdl.Renderer
	RendInfo *sdl.RendererInfo
}
