// The tsproxy binary is the simple reverse proxy that listens directly on the
// Tailscale network.
package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
	"tailscale.com/types/logger"
)

func main() {
	log.SetFlags(0)
	var (
		hostname = flag.String("hostname", "tsproxy", "Hostname to use.")
		upstream = flag.String("upstream", "http://localhost:3000", "URL to proxy.")
	)
	flag.Parse()
	logf := prefixLogf("tsproxy: ")

	os.Setenv("TAILSCALE_USE_WIP_CODE", "true")

	s := &tsnet.Server{
		Dir:      os.Getenv("STATE_DIRECTORY"), // https://www.freedesktop.org/software/systemd/man/systemd.exec.html#RuntimeDirectory=
		Hostname: *hostname,
		Logf:     prefixLogf("tailscale: "),
	}

	logf("Starting Tailscale HTTPS listener on %s...", *hostname)
	l, err := s.Listen("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}
	l = tls.NewListener(l, &tls.Config{
		GetCertificate: tailscale.GetCertificate,
	})

	u, err := url.Parse(*upstream)
	if err != nil {
		log.Fatal(err)
	}

	httpsrv := &http.Server{
		Handler: httputil.NewSingleHostReverseProxy(u),
	}

	log.Fatal(httpsrv.Serve(l))
}

func prefixLogf(prefix string) logger.Logf {
	return func(format string, args ...interface{}) {
		log.Printf(prefix+format, args...)
	}
}
