package main

import (
	"fmt"

	"github.com/kroppt/winfileask"
	"github.com/veandco/go-sdl2/sdl"
)

func openFileDialog(win *sdl.Window) (string, error) {
	var wm *sdl.SysWMInfo
	var err error
	if wm, err = win.GetWMInfo(); err != nil {
		return "", err
	}
	info := wm.GetWindowsInfo()
	filter := winfileask.FileFilter{winfileask.Filter{}}
	str, ok, err := winfileask.GetOpenFileName(info.Window, "Open an Image", filter, "")
	if !ok {
		err = fmt.Errorf("no image chosen")
	}
	if err != nil {
		return "", err
	}
	return str, nil
}
