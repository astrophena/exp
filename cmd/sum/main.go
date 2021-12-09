// The sum binary sums the numbers in a slice, distributing the work between two
// goroutines. Once both goroutines have completed their computation, it
// calculates the final result.
//
// It's an example from A Tour of Go (https://go.dev/tour/concurrency/2).
package main

import (
	"fmt"
	"log"
	"runtime"
)

func sum(s []int, c chan int) {
	sum := 0
	for _, v := range s {
		sum += v
	}
	c <- sum // send sum to c
}

func main() {
	log.SetFlags(0)
	log.Printf("GOMAXPROCS = %d", runtime.GOMAXPROCS(0))

	s := []int{7, 2, 8, -9, 4, 0}

	c := make(chan int)
	go sum(s[:len(s)/2], c)
	go sum(s[len(s)/2:], c)
	x, y := <-c, <-c // receive from c

	fmt.Println(x, y, x+y)
}
