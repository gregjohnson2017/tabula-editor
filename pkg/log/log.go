// Package log implements Tabula's format for logging different types of logs.
// This extends the already-powerful standard library and provides only the
// necessary features without using external dependencies.
// Provides info, debug, and performance loggers.
package log

import (
	"fmt"
	"log"
)

// The prefix labels for each of the loggers
const (
	infoLabel  = "INFO"
	warnLabel  = "WARN"
	debugLabel = "DBUG"
	fatalLabel = "FATL"
	perfLabel  = "PERF"
)

// ANSI foreground text color codes
const (
	brightRed     = "91"
	brightGreen   = "92"
	brightYellow  = "93"
	brightMagenta = "95"
	brightWhite   = "97"
)

func SetColorized(toggle bool) {
	setColorized(toggle, info, brightWhite, infoLabel)
	setColorized(toggle, warn, brightYellow, warnLabel)
	setColorized(toggle, debug, brightMagenta, debugLabel)
	setColorized(toggle, fatal, brightRed, fatalLabel)
	setColorized(toggle, perf, brightGreen, perfLabel)
}

func setColorized(toggle bool, l *log.Logger, color, label string) {
	if !toggle {
		l.SetPrefix(label + " ")
		return
	}
	prefix := fmt.Sprintf("\033[%vm%v\033[0m ", color, label)
	l.SetPrefix(prefix)
}

type ConstErr string

func (e ConstErr) Error() string {
	return string(e)
}
