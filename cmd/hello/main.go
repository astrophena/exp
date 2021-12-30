package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"tailscale.com/client/tailscale"
)

func main() {
	log.SetFlags(0)
	s := &http.Server{
		TLSConfig: &tls.Config{
			GetCertificate: tailscale.GetCertificate,
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, world!\n")
		}),
	}
	log.Fatal(s.ListenAndServeTLS("", ""))
}
