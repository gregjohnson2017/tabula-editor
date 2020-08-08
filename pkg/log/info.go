package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

func init() {
	info = log.New(ioutil.Discard, infoLabel+" ", log.LstdFlags)
}

var info *log.Logger

// Print calls l.Output to print to the info logger.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	_ = info.Output(2, fmt.Sprint(v...))
}

// Printf calls l.Output to print to the info logger.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	_ = info.Output(2, fmt.Sprintf(format, v...))
}

// Println calls l.Output to print to the info logger.
// Arguments are handled in the manner of fmt.Println.
func Infoln(v ...interface{}) {
	_ = info.Output(2, fmt.Sprintln(v...))
}

// SetOutput sets the output destination for the info logger.
func SetInfoOutput(out io.Writer) {
	info.SetOutput(out)
}
