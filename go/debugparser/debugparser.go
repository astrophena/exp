package debugparser

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "go.astrophena.name/infra/shared"
)

type Debug struct {
	Service    string
	Version    string
	Goroutines int
	Uptime     time.Duration
}

func Fetch(hostname string) (*Debug, error) {
	log.Printf("#%v", reflect.TypeOf(http.DefaultTransport))

	res, err := http.Get("https://" + hostname + "/_debug")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d := &Debug{}
	s := bufio.NewScanner(res.Body)

	for s.Scan() {
		a := strings.Split(s.Text(), "=")
		if len(a) < 2 {
			return nil, fmt.Errorf("invalid format of entry %q", s.Text())
		}

		a[1] = strings.TrimPrefix(a[1], `"`)
		a[1] = strings.TrimSuffix(a[1], `"`)

		var err error

		switch a[0] {
		case "service":
			d.Service = a[1]
		case "version":
			d.Version = a[1]
		case "goroutines":
			c, err := strconv.ParseInt(a[1], 10, 32)
			if err != nil {
				return nil, err
			}
			d.Goroutines = int(c)
		case "uptime":
			d.Uptime, err = time.ParseDuration(a[1])
			if err != nil {
				return nil, err
			}
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return d, nil
}
