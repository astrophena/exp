// club is the beginning of the Astrophena Club server.
package main

import (
	"context"
	"embed"
	"expvar"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/zserge/metric"
)

//go:embed static/*
var static embed.FS

type server struct {
	router     *mux.Router
	httpServer *http.Server
	templates  *template.Template
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.router.ServeHTTP(w, r) }

func newServer(addr string) (*server, error) {
	s := &server{router: mux.NewRouter()}
	s.httpServer = &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      s,
	}

	s.router.HandleFunc("/", s.handleHome())
	s.router.Handle("/debug/metrics", metric.Handler(metric.Exposed))

	fss, err := fs.Sub(static, "static")
	if err != nil {
		return nil, err
	}
	s.router.PathPrefix("/").Handler(http.FileServer(http.FS(fss)))

	return s, nil
}

func (s *server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, world!\n")
	}
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		addr = flags.String("addr", "localhost:3000", "Listen on `host:port`.")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	s, err := newServer(*addr)
	if err != nil {
		return fmt.Errorf("Server initialization failed: %v", err)
	}

	runMetricsCollector()
	go func() {
		log.Printf("Listening on %s...", *addr)
		if err := s.httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("HTTP server crashed: %v", err)
			}
		}
	}()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("Shutting down gracefully...")
	s.httpServer.Shutdown(ctx)

	return nil
}

func runMetricsCollector() {
	expvar.Publish("go:numgoroutine", metric.NewGauge("2m1s", "15m30s", "1h1m"))
	expvar.Publish("go:alloc", metric.NewGauge("2m1s", "15m30s", "1h1m"))
	expvar.Publish("go:alloctotal", metric.NewGauge("2m1s", "15m30s", "1h1m"))

	go func() {
		for range time.Tick(100 * time.Millisecond) {
			m := &runtime.MemStats{}
			runtime.ReadMemStats(m)

			expvar.Get("go:numgoroutine").(metric.Metric).Add(float64(runtime.NumGoroutine()))
			expvar.Get("go:alloc").(metric.Metric).Add(float64(m.Alloc) / 1000000)
			expvar.Get("go:alloctotal").(metric.Metric).Add(float64(m.TotalAlloc) / 1000000)
		}
	}()
}
