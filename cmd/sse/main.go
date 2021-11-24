// The sse binary is an example of using Server Sent Events (SSE) from Go
// together with the htmx JavaScript library.
package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexandrevicenzi/go-sse"
)

//go:embed htmx.js
var htmx string

//go:embed tmpl.html
var tmpl string

func main() {
	s := sse.NewServer(nil)
	defer s.Shutdown()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, tmpl)
	})
	mux.HandleFunc("/htmx.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "htmx.js", time.Time{}, strings.NewReader(htmx))
	})
	mux.Handle("/events", s)

	go func() {
		for range time.Tick(1 * time.Second) {
			s.SendMessage("", sse.SimpleMessage(time.Now().Format("2006/02/01/ 15:04:05")))
		}
	}()

	log.Fatal(http.ListenAndServe(":3000", mux))
}
