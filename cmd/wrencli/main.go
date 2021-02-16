// Command wrencli tests the Wren (https://wren.io) Go bindings.
package main

import (
	"log"

	"github.com/dradtke/go-wren"
)

func main() {
	vm := wren.NewVM()
	if err := vm.Interpret(`System.print("Hello, Wren!")`); err != nil {
		log.Println(err)
	}
}
