package main

import (
	"fmt"
	"time"
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
