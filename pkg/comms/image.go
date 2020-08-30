package comms

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Image struct {
	FileName string
	MousePix sdl.Point
	Mult     int32
}
