package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	s := &http.Server{
		Addr: "localhost:3000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "This is piper.\n")
		}),
	}
	log.Fatal(s.ListenAndServe())
}
