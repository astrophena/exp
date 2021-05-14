package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	f, err := os.Open("passwords.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Println(f)
}
