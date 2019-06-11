package main

import (
	"fmt"
	"time"
)

// Constants for menu bar buttons
const (
	MenuExit = iota
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
