package util

import (
	"fmt"
	"time"
)

// MenuAction is an identifier for the event handling to know which menu button
// has been pressed.
type MenuAction uint32

// Constants for all possible menu actions.
const (
	MenuNone MenuAction = iota
	MenuExit
)

// StopWatch is a time.Time with a stopping methods
type StopWatch struct {
	t time.Time
}

// Start returns a newly started stopwatch
func Start() StopWatch {
	return StopWatch{time.Now()}
}

// Stop prints the time duration since the stopwatch start
func (sw StopWatch) Stop(str string) {
	fmt.Printf("%v=%v\n", str, time.Since(sw.t))
}

// StopGetNano returns the nanoseconds from the stopwatch start
func (sw StopWatch) StopGetNano() int64 {
	return time.Since(sw.t).Nanoseconds()
}
