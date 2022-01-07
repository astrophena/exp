// The hello binary is a HTTP server on https://testlab.coin-tone.ts.net.
package main

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"runtime/debug"

	"git.astrophena.name/infra/util/systemd"

	"tailscale.com/client/tailscale"
)

//go:embed memes/*
var memes embed.FS

func main() {
	log.SetFlags(0)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "Hello, world!\n")
	})
	mux.HandleFunc("/who", func(w http.ResponseWriter, r *http.Request) {
		who, err := tailscale.WhoIs(r.Context(), r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(who)
	})
	mux.HandleFunc("/lol", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "LOL!\n")
	})
	mux.Handle("/memes/", http.FileServer(neuteredFileSystem{http.FS(memes)}))
	mux.HandleFunc("/buildinfo", func(w http.ResponseWriter, r *http.Request) {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			http.Error(w, "no build info", http.StatusInternalServerError)
			return
		}
		b, err := bi.MarshalText()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})
	mux.Handle("/piper", httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "localhost:3000",
	}))

	s := &http.Server{
		TLSConfig: &tls.Config{
			GetCertificate: tailscale.GetCertificate,
		},
		Handler: mux,
	}
	systemd.Ready()
	log.Fatal(s.ListenAndServeTLS("", ""))
}

// neuteredFileSystem is an implementation of http.FileSystem which prevents
// showing directory listings when using http.FileServer.
type neuteredFileSystem struct {
	fs http.FileSystem
}

// Open implements the http.FileSystem interface.
func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}
