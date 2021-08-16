package watchtime

import (
	"testing"
	"time"
)

func TestFetch(t *testing.T) {
	var cases = []struct {
		videoID  string
		expected time.Duration
	}{
		{"HLrqNhgdiC0", (6 * time.Minute) + (20 * time.Second)},
		{"LZa5KKfqHtA", (5 * time.Minute) + (41 * time.Second)},
		{"yIxEEgEuhT4", (51 * time.Minute) + (52 * time.Second)},
		{"bpHf1XcoiFs", (1 * time.Hour) + (20 * time.Minute) + (42 * time.Second)},
	}

	for _, tc := range cases {
		result, err := Fetch(tc.videoID)
		if err != nil {
			t.Errorf("Got an error when fetching %q: %v", tc.videoID, err)
		}

		if tc.expected != result {
			t.Errorf(`Unexpected result, got %v for duration %q`, result, tc.expected)
		}
	}
}
