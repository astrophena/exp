// comparable is the simple thing that tests Go generics.
//
// Run this with 'gotip':
//
//  $ gotip run main.go
package main

func findFunc[T comparable](a []T, v T) int {
	for i, e := range a {
		if e == v {
			return i
		}
	}
	return -1
}

func main() {
	print(findFunc([]int{1, 2, 3, 4, 5, 6}, 5))
}
