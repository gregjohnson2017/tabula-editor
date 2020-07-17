package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func init() {
	debug = log.New(ioutil.Discard, debugLabel+" ", log.LstdFlags|log.Lshortfile)
}

var debug *log.Logger

// Print calls l.Output to print to the debug logger.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	_ = debug.Output(2, fmt.Sprint(v...))
}

// Printf calls l.Output to print to the debug logger.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	_ = debug.Output(2, fmt.Sprintf(format, v...))
}

// Println calls l.Output to print to the debug logger.
// Arguments are handled in the manner of fmt.Println.
func Debugln(v ...interface{}) {
	_ = debug.Output(2, fmt.Sprintln(v...))
}

// SetOutput sets the output destination for the debug logger.
func SetDebugOutput(out io.Writer) {
	debug.SetOutput(out)
}
