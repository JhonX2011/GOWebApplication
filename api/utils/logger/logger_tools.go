package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

/*
FormatNow returns the date and time according to a received parameter. The format to be returned is YYYY/MM/DD - HH:MM:SS.
Parameters:
@ t: Time set in the time.Time data type.
*/
func FormatNow(t time.Time) string {
	return t.Format("2006/01/02 - 15:04:05")
}

/*
FileInfo returns the concatenation of the file and the invoked line number.
Parameters:
@ file: invoked file.
@ line: line number associated with the file.
@ ok: boolean indicating whether it was possible to retrieve the file information or not.
*/
func FileInfo(file string, line int, ok bool) string {
	if !ok {
		file = "<???>"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

/*
FuncInfo splits the correct name of the invoked function.
Parameters:
@ funcname: Name of the function to split.
*/
func FuncInfo(funcname string) string {
	fn := funcname[strings.LastIndex(funcname, "/")+1:]
	return fmt.Sprintf("%s()", fn)
}

// osExitFunc -> Interface that defines the behavior of the Exit function in the os package.
type osExitFunc func(int)

// exitFunc -> Global variable to store the Exit function.
var exitFunc osExitFunc = os.Exit //nolint:gochecknoglobals

// DefaultOSExit -> Function that wraps the call to os.Exit.
func DefaultOSExit(code int) {
	exitFunc(code)
}
