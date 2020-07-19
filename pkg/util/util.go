package util

import (
	"time"

	"github.com/gregjohnson2017/tabula-editor/pkg/log"
	"github.com/gregjohnson2017/tabula-editor/pkg/perf"
)

// MenuAction is an identifier for the event handling to know which menu button
// has been pressed.
type MenuAction uint32

// Constants for all possible menu actions.
const (
	MenuNone MenuAction = iota
	MenuExit
)

// ErrNoImageChosen indicates that an image selection was cancelled
const ErrNoImageChosen log.ConstErr = "no image chosen"

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
	log.Perff("%v=%v\n", str, time.Since(sw.t))
}

// StopGetNano returns the nanoseconds from the stopwatch start
func (sw StopWatch) StopGetNano() int64 {
	return time.Since(sw.t).Nanoseconds()
}

func (sw StopWatch) StopRecordAverage(key string) {
	perf.RecordAverageTime(key, time.Since(sw.t).Nanoseconds())
}
