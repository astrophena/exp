// vim: foldmethod=marker

// Command httpsrv is the HTTP server boilerplate.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(0)

	addr := flag.String("addr", "localhost:3000", "Listen on `host:port`.")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	s := &server{
		addr: *addr,
	}
	if err := s.run(ctx); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	// configuration
	addr string // where to listen, in form of host:port

	http *http.Server

	// guards initialization
	init    sync.Once
	initErr error
}

func (s *server) doInit() { // {{{
	if err := func() error {
		s.http = &http.Server{
			Addr:    s.addr,
			Handler: s,
		}

		return nil
	}(); err != nil {
		s.initErr = err
	}
} // }}}

func (s *server) run(ctx context.Context) error { // {{{
	s.init.Do(s.doInit)
	if s.initErr != nil {
		return s.initErr
	}

	errCh := make(chan error, 1)

	go func() {
		log.Printf("Listening on %s for HTTP requests...", s.addr)
		if err := s.http.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("HTTP server crashed: %v", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.shutdown()
	}
} // }}}

func (s *server) shutdown() error { // {{{
	log.Printf("Gracefully shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
} // }}}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Put something useful here.")
}
