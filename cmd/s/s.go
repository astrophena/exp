// Command s is a simple HTTP server that serves files.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"go.astrophena.name/exp/cmd"
)

func main() {
	addr := flag.String("addr", "localhost:3000", "Listen on `host:port`.")
	cmd.SetDescription("Simple HTTP server that serves files.")
	cmd.SetArgsUsage("[dir]")
	cmd.HandleStartup()

	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	fullDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:    *addr,
		Handler: http.FileServer(http.Dir(fullDir)),
	}
	errCh := make(chan error, 1)

	go func() {
		log.Printf("Serving %s on %s.", fullDir, *addr)
		log.Printf("Use Ctrl+C to shut down the server.")
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("Received %s, gracefully shutting down.", sig)
	case err := <-errCh:
		log.Fatal(err)
	}
}
