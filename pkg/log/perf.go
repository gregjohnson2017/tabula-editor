package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func init() {
	perf = log.New(ioutil.Discard, perfLabel+" ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
}

var perf *log.Logger

// Print calls l.Output to print to the performance logger.
// Arguments are handled in the manner of fmt.Print.
func Perf(v ...interface{}) {
	_ = perf.Output(2, fmt.Sprint(v...))
}

// Printf calls l.Output to print to the performance logger.
// Arguments are handled in the manner of fmt.Printf.
func Perff(format string, v ...interface{}) {
	_ = perf.Output(2, fmt.Sprintf(format, v...))
}

// Println calls l.Output to print to the performance logger.
// Arguments are handled in the manner of fmt.Println.
func Perfln(v ...interface{}) {
	_ = perf.Output(2, fmt.Sprintln(v...))
}

// SetOutput sets the output destination for the performance logger.
func SetPerfOutput(out io.Writer) {
	perf.SetOutput(out)
}
