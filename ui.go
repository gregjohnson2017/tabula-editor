package main

import "github.com/veandco/go-sdl2/sdl"

// UIComponent says what functions a UIComponent must implement
type UIComponent interface {
	GetBoundary() *sdl.Rect
	Render(*sdl.Renderer) error
	OnEnter(*sdl.MouseMotionEvent) bool
	OnLeave(*sdl.MouseMotionEvent) bool
	OnMotion(*sdl.MouseMotionEvent) bool
	OnScroll(*sdl.MouseWheelEvent) bool
	OnClick(*sdl.MouseButtonEvent) bool
}

type context struct {
	Win      *sdl.Window
	Rend     *sdl.Renderer
	RendInfo *sdl.RendererInfo
	Conf     *config
}
