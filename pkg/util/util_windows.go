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
		err = fmt.Errorf("no image chosen")
	}
	if err != nil {
		return "", err
	}
	return str, nil
}

// SetupMenuBar sets up system window menu bars
func SetupMenuBar(win *sdl.Window) error {
	return nil
	/*
		var wm *sdl.SysWMInfo
		var err error
		if wm, err = win.GetWMInfo(); err != nil {
			return err
		}
		info := wm.GetWindowsInfo()
		var hmenu winmenu.HMenu
		var ok bool
		if hmenu, ok = winmenu.GetMenu(info.Window); !ok {
			if hmenu, ok = winmenu.CreateMenu(); !ok {
				return fmt.Errorf("could not create menu")
			}
		}
		var hfile winmenu.HMenu
		if hfile, ok = winmenu.CreatePopupMenu(); !ok {
			return fmt.Errorf("could not create file menu")
		}
		{
			var mii = winmenu.NewMenuItemInfo()
			if ok = mii.SetAsString("Exit"); !ok {
				return fmt.Errorf("could not set exit menu item as string")
			}
			mii.SetID(uint32(MenuExit))
			if ok = hfile.InsertMenuItem(0, false, mii); !ok {
				return fmt.Errorf("could not insert exit menu item")
			}
		}
		{
			var mii = winmenu.NewMenuItemInfo()
			mii.SetSubMenu(hfile)
			if ok = mii.SetAsString("File"); !ok {
				return fmt.Errorf("could not set file menu item as string")
			}
			if ok = hmenu.InsertMenuItem(0, false, mii); !ok {
				return fmt.Errorf("could not insert file menu item")
			}
		}
		if ok = winmenu.SetMenu(info.Window, hmenu); !ok {
			return fmt.Errorf("could not set current window's menu")
		}
		return nil
	*/
}

/*
// GetMenuAction returns the type of action that has been detected
func GetMenuAction(evt *sdl.SysWMEvent) MenuAction {
	const WmCommand = 0x0111
	winmsg := evt.Msg.Windows()
	if winmsg.Msg != WmCommand {
		return MenuNone
	}
	switch MenuAction(winmsg.WParam) {
	case MenuExit:
		return MenuExit
	}
	return MenuNone
}
*/
