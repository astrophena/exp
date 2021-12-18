// The debugstats binary periodically scrapes debug variables from services and
// persists them in a JSON file for later analysis.
package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"git.astrophena.name/infra/cmd"
	"git.astrophena.name/infra/util/atomicfile"
	"git.astrophena.name/infra/web"

	"github.com/peterbourgon/unixtransport"
)

var services = []string{
	"go-import-redirector",
	"tgbotd",
	"webdavd",
}

type vars struct {
	Cmdline          []string          `json:"cmdline"`
	Goroutines       int               `json:"goroutines"`
	MemStats         *runtime.MemStats `json:"memstats"`
	ProcessStartTime time.Time         `json:"process_start_time"`
	Uptime           string            `json:"uptime"`
	Version          string            `json:"version"`
}

func init() {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		panic("http.DefaultTransport is not a pointer to http.Transport")
	}
	unixtransport.Register(t)
}

func main() {
	var (
		addr           = flag.String("addr", "localhost:3000", "Listen on `host:port or Unix socket`.")
		file           = flag.String("file", "data.json.gz", "Path to the database file.")
		scrapeInterval = flag.Duration("scrape-interval", 1*time.Minute, "Scrape interval")
	)
	cmd.HandleStartup()

	s, err := newServer(*file)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	go s.scrapeLoop(&wg, *scrapeInterval)

	web.ListenAndServe(&web.ListenAndServeConfig{
		Addr: *addr,
		Mux:  s.mux,
		OnShutdown: func() {
			s.stop <- true
		},
	})

	wg.Wait()
}

func newServer(file string) (*server, error) {
	s := &server{
		file: file,
		mux:  http.NewServeMux(),
		stop: make(chan bool, 1),
		data: make(map[int64]map[string]vars),
	}

	if _, err := os.Stat(s.file); os.IsNotExist(err) {
		return s, nil
	} else if err != nil {
		return nil, err
	}

	f, err := os.Open(s.file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if err := r.Close(); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(b, &s.data); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return s, nil
}

type server struct {
	file string

	mu   sync.Mutex
	data map[int64]map[string]vars // Unix time -> vars map

	mux  *http.ServeMux
	stop chan bool
}

func (s *server) scrapeLoop(wg *sync.WaitGroup, scrapeInterval time.Duration) {
	wg.Add(1)
	log.Printf("Starting scraping...")

	ticker := time.NewTicker(scrapeInterval)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			v, err := getVars()
			if err != nil {
				log.Printf("Got an error when scraping: %v", err)
			}

			s.mu.Lock()
			s.data[time.Now().Unix()] = v
			if err := s.save(); err != nil {
				log.Printf("Got an error when saving file: %v", err)
			}
			s.mu.Unlock()
		case <-s.stop:
			break loop
		}
	}

	log.Printf("Stopping scraping...")
	wg.Done()
}

// dump returns the JSON encoding of data contents. s.mu must be held.
func (s *server) dump() ([]byte, error) { return json.Marshal(s.data) }

// save writes data contents to the file. s.mu must be held.
func (s *server) save() error {
	b, err := s.dump()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return atomicfile.WriteFile(s.file, buf.Bytes(), 0o644)
}

func getVars() (map[string]vars, error) {
	kv := make(map[string]vars)

	for _, k := range services {
		r, err := http.Get(fmt.Sprintf("http+unix:///run/%s/socket:/debug/vars", k))
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		v := new(vars)
		if err := json.Unmarshal(b, v); err != nil {
			return nil, err
		}

		kv[k] = *v
	}

	return kv, nil
}
