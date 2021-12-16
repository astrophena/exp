// The sqlplay binary is a little internal tool that's like the Go playground
// but for SQL queries (so people can write queries & share them). It's based on
// https://gist.github.com/bradfitz/a7db110a6bd7d9c9bd02352adaea389b.
//
// For example, you can run this query on sqlite/liked.db:
//
//  SELECT 'https://youtube.com/embed/' || id AS url, title FROM videos ORDER BY title LIMIT 10;
//
// Or visit http://infra:6969/?query=SELECT+%27https%3A%2F%2Fyoutube.com%2Fembed%2F%27+%7C%7C+id+AS+url%2C+title+FROM+videos+ORDER+BY+title+LIMIT+10%3B.
package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"flag"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.astrophena.name/infra/cmd"
	"git.astrophena.name/infra/version"
	"git.astrophena.name/infra/web"

	_ "modernc.org/sqlite"
)

//go:embed template.html
var tpl string

func main() {
	var (
		addr   = flag.String("addr", "localhost:3000", "Listen on `host:port or Unix socket`.")
		dbPath = flag.String("db", "", "Path to the SQLite database.")
	)
	cmd.SetDescription("SQL playground.")
	cmd.HandleStartup()

	if *dbPath == "" {
		log.Fatal("Set the -db flag to the SQLite database path.")
	}

	s, err := newServer(*dbPath)
	if err != nil {
		log.Fatal("Failed to initialize the server: %v", err)
	}

	log.Printf("Using database %s.", s.dbPath)
	web.ListenAndServe(&web.ListenAndServeConfig{
		Mux:  s.mux,
		Addr: *addr,
		OnShutdown: func() {
			s.db.Close()
		},
	})
}

func newServer(dbPath string) (*server, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	s := &server{db: db, dbPath: dbPath}

	s.tpl, err = template.New("sqlplay").Funcs(template.FuncMap{
		"cmdName": func() string {
			return version.CmdName
		},
		"env": func() string {
			return version.Env
		},
	}).Parse(tpl)
	if err != nil {
		return nil, err
	}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.serve)
	schemaQuery := url.Values{}
	// https://stackoverflow.com/a/6617764
	schemaQuery.Set("query", "SELECT name, sql FROM sqlite_master WHERE type='table' ORDER BY name;")
	s.mux.Handle("/schema", http.RedirectHandler("/?"+schemaQuery.Encode(), http.StatusFound))
	s.mux.Handle("/about", http.RedirectHandler("https://godoc.astrophena.name/pkg/git.astrophena.name/exp/cmd/sqlplay/", http.StatusFound))

	// Remind me that sqlplay code is in exp, not infra.
	web.Debugger(s.mux).KV("Repo", "exp")

	return s, nil
}

type server struct {
	db     *sql.DB
	dbPath string
	mux    *http.ServeMux
	tpl    *template.Template
}

func (s *server) serve(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		web.NotFound(w, r)
		return
	}

	nonce := generateRandomString(16)
	w.Header().Set("Content-Security-Policy", fmt.Sprintf("default-src 'self'; style-src 'self' 'nonce-%s'", nonce))

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
				val := make([]interface{}, len(cols))
				valPtr := make([]interface{}, len(cols))
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
		QueryErr             error
		Duration             time.Duration
		Nonce, DBPath, Query string
		Table                template.HTML
	}{queryErr, dur, nonce, s.dbPath, query, template.HTML(tb.String())}

	var buf bytes.Buffer
	if err := s.tpl.Execute(&buf, d); err != nil {
		web.Error(w, r, err)
		return
	}
	buf.WriteTo(w)
}

func colFmt(v interface{}) string {
	switch v := v.(type) {
	case []byte:
		return string(v)
	default:
		s := fmt.Sprint(v)
		s = strings.TrimSuffix(s, " 00:00:00 +0000 +0000") // so a time.Time of a single day formats nicely
		return s
	}
}

func colHTML(v interface{}) string {
	s := colFmt(v)
	h := html.EscapeString(s)
	// Convert valid URLs into links.
	if isValidURL(s) {
		return fmt.Sprintf(`<a href="%[1]s" rel="noopener noreferrer">%[1]s</a>`, s)
	}
	return h
}

// isValidURL tests a string to determine if it is a well-structured URL or not.
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

// generateRandomBytes returns random bytes.
func generateRandomBytes(size int) []byte {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

// generateRandomString returns a random string.
func generateRandomString(size int) string {
	return base64.URLEncoding.EncodeToString(generateRandomBytes(size))
}
