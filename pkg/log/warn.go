package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func init() {
	warn = log.New(ioutil.Discard, warnLabel+" ", log.LstdFlags)
}

var warn *log.Logger

// Print calls l.Output to print to the warning logger.
// Arguments are handled in the manner of fmt.Print.
func Warn(v ...interface{}) {
	_ = warn.Output(2, fmt.Sprint(v...))
}

// Printf calls l.Output to print to the warning logger.
// Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, v ...interface{}) {
	_ = warn.Output(2, fmt.Sprintf(format, v...))
}

// Println calls l.Output to print to the warning logger.
// Arguments are handled in the manner of fmt.Println.
func Warnln(v ...interface{}) {
	_ = warn.Output(2, fmt.Sprintln(v...))
}

// SetOutput sets the output destination for the warning logger.
func SetWarnOutput(out io.Writer) {
	warn.SetOutput(out)
}
