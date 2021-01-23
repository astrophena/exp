// Package watchingtime fetches the watching time of YouTube videos.
package watchingtime // import "go.astrophena.name/exp/watchingtime"

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/senseyeio/duration"
)

// Fetch returns a watching time of YouTube video with the ID videoID.
func Fetch(videoID string) (time.Duration, error) {
	url := "https://www.youtube.com/watch?v=" + videoID

	r, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("unable to fetch %s: %w", url, err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("server returned status code %d", r.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(r.Body)
	if err != nil {
		return 0, fmt.Errorf("unable to initialize document: %w", err)
	}

	var durs string

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		itemprop, ok := s.Attr("itemprop")
		if !ok || itemprop != "duration" {
			return
		}

		durs, ok = s.Attr("content")
		if !ok || durs == "" {
			return
		}

		if itemprop == "duration" && durs != "" {
			return
		}
	})

	if durs == "" {
		return 0, fmt.Errorf("no duration found")
	}

	dur, err := duration.ParseISO8601(durs)
	if err != nil {
		return 0, fmt.Errorf("unable to parse duration: %w", err)
	}

	return timeDuration(dur), nil
}

func timeDuration(d duration.Duration) time.Duration {
	var dur time.Duration
	dur = dur + (time.Duration(d.TH) * time.Hour)
	dur = dur + (time.Duration(d.TM) * time.Minute)
	dur = dur + (time.Duration(d.TS) * time.Second)
	return dur
}
