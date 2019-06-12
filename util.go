package main

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

type stopWatch struct {
	t time.Time
}

func start() stopWatch {
	return stopWatch{time.Now()}
}

func (sw stopWatch) stop(str string) {
	fmt.Printf("%v=%v\n", str, time.Since(sw.t))
}

func (sw stopWatch) stopGetNano() int64 {
	return time.Since(sw.t).Nanoseconds()
}
