package main

import "C"
import "runtime"

//export Add
func Add(a, b int) int {
	return a + b
}

//export Version
func Version() *C.char {
	return C.CString(runtime.Version())
}

func main() {}
