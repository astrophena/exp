// Command tsserver is a simple HTTP server that listens directly on
// Tailscale.
package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"

	"tailscale.com/tsnet"
)

func main() {
	var (
		auth = flag.Bool("auth", false, "authenticate with Tailscale")
	)
	flag.Parse()

	// https://github.com/tailscale/tailscale/blob/main/tsnet/tsnet.go#L92
	os.Setenv("TAILSCALE_USE_WIP_CODE", "true")
	// https://github.com/tailscale/tailscale/blob/main/tsnet/tsnet.go#L184
	if *auth {
		os.Setenv("TS_LOGIN", "1")
	}

	s := &tsnet.Server{}

	l, err := s.Listen("tcp", ":80")
	if err != nil {
		log.Fatalf("Unable to listen: %v", err)
	}

	log.Fatal(http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who, ok := s.WhoIs(r.RemoteAddr)
		if !ok {
			http.Error(w, "WhoIs failed", 500)
			return
		}
		fmt.Fprintf(w, "<html><body><h1>Hello, world!</h1>\n")
		fmt.Fprintf(w, "<p>You are <b>%s</b> from <b>%s</b> (%s)</p>",
			html.EscapeString(who.UserProfile.LoginName),
			html.EscapeString(firstLabel(who.Node.ComputedName)),
			r.RemoteAddr)
	})))
}

func firstLabel(s string) string {
	if i := strings.Index(s, "."); i != -1 {
		return s[:i]
	}
	return s
}
