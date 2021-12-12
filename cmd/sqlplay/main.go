// The sqlplay binary is a little internal tool that's like the Go playground
// but for SQL queries (so people can write queries & share them).
//
// Stolen from
// https://gist.github.com/bradfitz/a7db110a6bd7d9c9bd02352adaea389b.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"git.astrophena.name/infra/web"

	_ "modernc.org/sqlite"
)

func main() {
	log.SetFlags(0)

	addr := flag.String("addr", "localhost:3000", "Listen on `host:port or Unix socket`.")
	dbPath := flag.String("db", "", "Path to the SQLite database.")
	flag.Parse()

	if *dbPath == "" {
		log.Fatal("Set the -db flag to the SQLite database path.")
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", &server{db: db, dbPath: *dbPath})
	schemaQuery := url.Values{}
	// https://stackoverflow.com/a/6617764
	schemaQuery.Set("q", "SELECT name, sql FROM sqlite_master WHERE type='table' ORDER BY name;")
	mux.Handle("/schema", http.RedirectHandler("/?"+schemaQuery.Encode(), http.StatusFound))

	web.ListenAndServe(&web.ListenAndServeConfig{
		Mux:  mux,
		Addr: *addr,
		OnShutdown: func() {
			db.Close()
		},
	})
}

type server struct {
	db     *sql.DB
	dbPath string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		web.NotFound(w, r)
		return
	}

	sql := r.FormValue("q")
	fmt.Fprintf(w, `<html>
<head><style>

tr:nth-child(even){background-color: #f2f2f2;}

td, th {
  border: 1px solid #ddd;
  padding: 8px;
}

table {
    border-collapse: collapse;
    width: 100%%;
}

tr:hover {background-color: #ddd;}

th {
  padding-top: 12px;
  padding-bottom: 12px;
  text-align: left;
  background-color: #04AA6D;
  color: white;
}

</style></head>
<body>
   <h1>SQL Playground</h1>
	 <p>Using database <code>%s</code>.</p>
	 <form method=GET>
   <textarea name="q" rows="5" cols="80" style='width: 100%%'>%s</textarea>
   <p><input type="submit" value="Query"> [<a href="/schema">Schema</a>]</p>
   </form>
</body>`, s.dbPath, html.EscapeString(sql))

	if sql != "" {
		rows, err := s.db.Query(sql)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		io.WriteString(w, `<html><body><table><tr>`)
		cols, _ := rows.Columns()
		for _, c := range cols {
			fmt.Fprintf(w, "<th>%s</th>\n", html.EscapeString(c))
		}
		io.WriteString(w, `</tr>`)

		for rows.Next() {
			val := make([]interface{}, len(cols))
			valPtr := make([]interface{}, len(cols))
			for i := range cols {
				valPtr[i] = &val[i]
			}
			if err := rows.Scan(valPtr...); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			io.WriteString(w, `<tr>`)

			for _, v := range val {
				fmt.Fprintf(w, "<td>%s</td>\n", colHTML(v))
			}
			io.WriteString(w, "</tr>\n")
		}
		io.WriteString(w, "</table>\n")
		return
	}

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
	if strings.HasPrefix(s, "cus_") {
		return fmt.Sprintf("<a href=\"https://dashboard.stripe.com/customers/%s\" rel=\"noopener noreferrer\">%s</a>", h, h)
	}
	return h
}
