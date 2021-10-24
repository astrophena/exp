package watchtime

import (
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	var cases = []struct {
		id   string
		want time.Duration
	}{
		{"HLrqNhgdiC0", (6 * time.Minute) + (20 * time.Second)},
		{"LZa5KKfqHtA", (5 * time.Minute) + (41 * time.Second)},
		{"yIxEEgEuhT4", (51 * time.Minute) + (52 * time.Second)},
		{"bpHf1XcoiFs", (1 * time.Hour) + (20 * time.Minute) + (42 * time.Second)},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			got, err := Fetch(tc.id)
			if err != nil {
				t.Fatalf("Got an error when fetching %q: %v", tc.id, err)
			}

			if tc.want != got {
				t.Fatalf(`got %v, want %q`, got, tc.want)
			}
		})
	}
}
