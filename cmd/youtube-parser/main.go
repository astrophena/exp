// The youtube-parser binary converts YouTube CSV playlist exports to
// HTML files with thumbnails and links.
package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/csv"
	"flag"
	"html/template"
	"log"
	"os"
	"os/exec"

	"github.com/neilotoole/errgroup"
)

type video struct {
	Title string
	ID    string
}

var (
	//go:embed template.html
	tplStr string

	tpl = template.Must(template.New("videos").Parse(tplStr))
)

func main() {
	log.SetFlags(0)

	title := flag.String("title", "Videos", "Page title.")
	flag.Parse()

	if len(flag.Args()) < 2 {
		log.Fatal("Path to CSV and the result files is required.")
	}
	from := flag.Args()[0]
	to := flag.Args()[1]

	f, err := os.Open(from)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	recs, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	for n, id := range recs {
		if n == 0 {
			continue
		}
		ids = append(ids, id[0])
	}

	var (
		videos     []video
		videosChan = make(chan video)
	)

	g, _ := errgroup.WithContextN(context.Background(), 20, 700)

	all := len(ids)
	for i, id := range ids {
		i, id := i, id
		g.Go(func() error {
			log.Printf("Processing video %s (%d/%d)...", id, i+1, all)
			title, err := fetchTitle(id)
			if err != nil {
				log.Printf("Failed to fetch video title for ID %s: %v", id, err)
				return nil
			}
			videosChan <- video{ID: id, Title: title}
			return nil
		})
	}

	go func() {
		for v := range videosChan {
			videos = append(videos, v)
		}
	}()

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, map[string]interface{}{
		"title":  *title,
		"videos": videos,
	}); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(to, buf.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}

func fetchTitle(id string) (string, error) {
	b, err := exec.Command("youtube-dl", "-e", "https://www.youtube.com/watch?v="+id).Output()
	if err != nil {
		return "", err
	}
	return string(b), nil
}
