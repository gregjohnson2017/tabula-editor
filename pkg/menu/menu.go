// Package menu provides menu-related structs for use in tabula. Specifically,
// the package provides Bar for a menu bar implementation and List for a menu
// entry container implementation.
// Bar and List satisfy the tabula.ui.Component interface.
package menu

import (
	"github.com/gregjohnson2017/tabula-editor/pkg/log"
)

// Definition defines menu bar menus
type Definition struct {
	Text     string
	Children []Definition
	Action   func()
}

// ErrNoEntryAtPosition indicates that no entry exists at a given position
const ErrNoEntryAtPosition log.ConstErr = "no entry at given position"
