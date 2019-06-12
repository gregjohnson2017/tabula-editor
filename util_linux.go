package main

import (
	"fmt"
	"os"

	"github.com/jcmuller/gozenity"
	"github.com/veandco/go-sdl2/sdl"
)

func openFileDialog(win *sdl.Window) (string, error) {
	files, err := gozenity.FileSelection("Choose a picture to open", nil)
	if err != nil {
		panic(err)
	}
	return files[0], nil
}

func setupMenuBar(win *sdl.Window) error {
	fmt.Fprintln(os.Stderr, "setupMenuBar not implemented")
	return nil
}

func getMenuAction(evt *sdl.SysWMEvent) MenuAction {
	fmt.Fprintln(os.Stderr, "handleWMEvent not implemented")
	return MenuNone
}
