// Command vars displays some stats about internal services.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"text/tabwriter"
	"time"

	"git.astrophena.name/infra/version"
)

var services = []string{
	"bot",
	"go",
	"webdav",
}

type vars struct {
	Cmdline          []string          `json:"cmdline"`
	Goroutines       int               `json:"goroutines"`
	MemStats         *runtime.MemStats `json:"memstats"`
	ProcessStartTime time.Time         `json:"process_start_time"`
	Uptime           string            `json:"uptime"`
	Version          version.Info      `json:"version"`
}

func formatSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func main() {
	w := tabwriter.NewWriter(os.Stderr, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(w, "Service\tMemory\tGoroutines\tUptime\tVersion\n")

	for _, s := range services {
		r, err := http.Get(fmt.Sprintf("https://%s.astrophena.name/debug/vars", s))
		if err != nil {
			log.Fatal(err)
		}
		defer r.Body.Close()

		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		v := &vars{}
		if err := json.Unmarshal(b, v); err != nil {
			log.Fatal(err)
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\n", s, formatSize(int64(v.MemStats.Alloc)), v.Goroutines, v.Uptime, v.Version.Commit)
	}

	w.Flush()
}
