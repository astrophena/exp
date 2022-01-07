// This is an example of a Python program interacting with the Go
// library.
//
// Hacking
//
// Run this:
//
//  $ make
//  $ python prog.py
package main

import "C"
import (
	"runtime"

	"git.astrophena.name/exp/watchtime"
)

//export Add
func Add(a, b int) int {
	return a + b
}

//export Version
func Version() *C.char {
	return C.CString(runtime.Version())
}

//export WatchTime
func WatchTime(id *C.char) *C.char {
	goID := C.GoString(id)
	t, err := watchtime.Fetch(goID)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(t.String())
}

func main() {}
