package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func init() {
	fatal = log.New(ioutil.Discard, fatalLabel+" ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
}

var fatal *log.Logger

// Print calls l.Output to print to the fatal logger.
// Arguments are handled in the manner of fmt.Print.
func Fatal(v ...interface{}) {
	_ = fatal.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Printf calls l.Output to print to the fatal logger.
// Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, v ...interface{}) {
	_ = fatal.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Println calls l.Output to print to the fatal logger.
// Arguments are handled in the manner of fmt.Println.
func Fatalln(v ...interface{}) {
	_ = fatal.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// SetOutput sets the output destination for the fatal logger.
func SetFatalOutput(out io.Writer) {
	fatal.SetOutput(out)
}
