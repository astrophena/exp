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
	goId := C.GoString(id)
	t, err := watchtime.Fetch(goId)
	if err != nil {
		return C.CString(err.Error())
	}
	return C.CString(t.String())
}

func main() {}