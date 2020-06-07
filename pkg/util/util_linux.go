package util

import (
	"fmt"

	"github.com/jcmuller/gozenity"
	"github.com/veandco/go-sdl2/sdl"
)

// OpenFileDialog uses a system file picker to get a filename from the user
func OpenFileDialog(win *sdl.Window) (string, error) {
	files, err := gozenity.FileSelection("Choose a picture to open", nil)
	if err != nil {
		return "", fmt.Errorf("OpenFileDialog: %w", err)
	}
	return files[0], nil
}

// SaveFileDialog uses a system file picker to get a file path from the user for the purpose of saving an image
func SaveFileDialog(*sdl.Window) (string, error) {
	folders, err := gozenity.DirectorySelection("Choose a folder to save in")
	if err != nil {
		return "", fmt.Errorf("SaveFileDialog: %w", err)
	}
	file, err := gozenity.Entry("Enter a name for the picture", "")
	if err != nil {
		return "", fmt.Errorf("SaveFileDialog: %w", err)
	}
	fmt.Printf("SaveFileDialog got back with %v folders and \"%v\" file name", folders, file)
	return folders[0] + "/" + file, nil
}
