package util

import (
	"fmt"

	"github.com/kroppt/winfileask"
	"github.com/veandco/go-sdl2/sdl"
)

// OpenFileDialog uses a system file picker to get a filename from the user
func OpenFileDialog(win *sdl.Window) (string, error) {
	var wm *sdl.SysWMInfo
	var err error
	if wm, err = win.GetWMInfo(); err != nil {
		return "", err
	}
	info := wm.GetWindowsInfo()
	filter := winfileask.FileFilter{winfileask.Filter{}}
	str, ok, err := winfileask.GetOpenFileName(info.Window, "Open an Image", filter, "")
	if !ok {
		err = ErrNoImageChosen
	}
	if err != nil {
		return "", fmt.Errorf("OpenFileDialog: %w", err)
	}
	return str, nil
}

// SaveFileDialog uses a system file picker to get a file path from the user for the purpose of saving an image
func SaveFileDialog(win *sdl.Window) (string, error) {
	var wm *sdl.SysWMInfo
	var err error
	if wm, err = win.GetWMInfo(); err != nil {
		return "", err
	}
	info := wm.GetWindowsInfo()
	filter := winfileask.FileFilter{
		winfileask.Filter{
			Name:    "JPEG (*.jpeg;*.jpg;*.jpe;*.jfif)",
			Pattern: "*.jpeg;*.jpg;*.jpe;*.jfif",
		},
		winfileask.Filter{
			Name:    "PNG (*.png)",
			Pattern: "*.png",
		},
	}
	str, ok, err := winfileask.GetSaveFileName(info.Window, "Save an Image", filter, "")
	if !ok {
		err = ErrNoImageChosen
	}
	if err != nil {
		return "", fmt.Errorf("SaveFileDialog: %w", err)
	}
	return str, nil
}
