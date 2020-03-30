package util

import (
	"github.com/jcmuller/gozenity"
	"github.com/veandco/go-sdl2/sdl"
)

// OpenFileDialog uses a system file picker to get a filename from the user
func OpenFileDialog(win *sdl.Window) (string, error) {
	files, err := gozenity.FileSelection("Choose a picture to open", nil)
	if err != nil {
		panic(err)
	}
	return files[0], nil
}
