// Command sqlplay is a playground for SQLite databases. It's based on
// https://gist.github.com/bradfitz/a7db110a6bd7d9c9bd02352adaea389b.
package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/tailscale/sqlite"
)

var (
	//go:embed template.html
	tpl string

	//go:embed style.css
	css string
)

func main() {
	log.SetFlags(0)

	addr := flag.String("addr", "localhost:3000", "Listen on `host:port`.")
	flag.Parse()

	dbPath := flag.Arg(0)
	if dbPath == "" {
		log.Fatal("You need to specify a path to the SQLite database.")
	}

	s, err := newServer(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize the server: %v", err)
	}

	log.Printf("Using database %s.", s.dbPath)

	httpSrv := &http.Server{
		Addr:    *addr,
		Handler: s,
	}
	go func() {
		log.Printf("Listening on %s...", *addr)
		if err := httpSrv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("HTTP server crashed: %v", err)
			}
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-exit
	log.Printf("Received %s, gracefully shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	httpSrv.Shutdown(ctx)
}

func newServer(dbPath string) (*server, error) {
	fullPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", "file://"+fullPath)
	if err != nil {
		return nil, err
	}

	s := &server{db: db, dbPath: dbPath}

	s.tpl, err = template.New("sqlplay").Parse(tpl)
	if err != nil {
		return nil, err
	}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.serve)
	schemaQuery := url.Values{}
	// https://stackoverflow.com/a/6617764
	schemaQuery.Set("query", "SELECT name, sql FROM sqlite_master WHERE type='table' ORDER BY name;")
	s.mux.Handle("/schema", http.RedirectHandler("/?"+schemaQuery.Encode(), http.StatusFound))
	s.mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "style.css", time.Now(), strings.NewReader(css))
	})

	return s, nil
}

type server struct {
	db     *sql.DB
	dbPath string
	mux    *http.ServeMux
	tpl    *template.Template
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }

func (s *server) serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	query := r.FormValue("query")

	var dur time.Duration
	var tb strings.Builder
	var queryErr error
	if query != "" {
		start := time.Now()
		rows, err := s.db.Query(query)
		if err != nil {
			queryErr = err
		}
		done := time.Now()
		dur = done.Sub(start)

		if rows != nil {
			io.WriteString(&tb, `<table><tr>`)
			cols, _ := rows.Columns()
			for _, c := range cols {
				fmt.Fprintf(&tb, "<th>%s</th>\n", html.EscapeString(c))
			}
			io.WriteString(&tb, `</tr>`)

			for rows.Next() {
				val := make([]any, len(cols))
				valPtr := make([]any, len(cols))
				for i := range cols {
					valPtr[i] = &val[i]
				}
				if err := rows.Scan(valPtr...); err != nil {
					queryErr = err
				}
				io.WriteString(&tb, `<tr>`)

				for _, v := range val {
					fmt.Fprintf(&tb, "<td>%s</td>\n", colHTML(v))
				}
				io.WriteString(&tb, "</tr>\n")
			}
			io.WriteString(&tb, "</table>\n")
		}
	}

	d := struct {
		QueryErr      error
		Duration      time.Duration
		DBPath, Query string
		Table         template.HTML
	}{queryErr, dur, s.dbPath, query, template.HTML(tb.String())}

	var buf bytes.Buffer
	if err := s.tpl.Execute(&buf, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func colFmt(v any) string {
	switch v := v.(type) {
	case []byte:
		return string(v)
	default:
		s := fmt.Sprint(v)
		s = strings.TrimSuffix(s, " 00:00:00 +0000 +0000") // so a time.Time of a single day formats nicely
		return s
	}
}

func colHTML(v any) string {
	s := colFmt(v)
	h := html.EscapeString(s)
	// Convert valid URLs into links.
	if isValidURL(s) {
		return fmt.Sprintf(`<a href="%[1]s" rel="noopener noreferrer">%[1]s</a>`, s)
	}
	return h
}

func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}
