//go:build linux

// Command chrome-open-bookmarks-bar opens  everything in the bookmarks bar at
// once in a separate tabs.
package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	log.SetPrefix("chrome-open-bookmarks-bar: ")

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("os.UserConfigDir(): %v", err)
	}

	b, err := os.ReadFile(filepath.Join(configDir, "google-chrome", "Default", "Bookmarks"))
	if err != nil {
		log.Fatalf("failed to read the bookmarks file: %v", err)
	}

	var bookmarks struct {
		Roots map[string]struct {
			Children []struct {
				URL string `json:"url"`
			} `json:"children"`
		} `json:"roots"`
	}
	if err := json.Unmarshal(b, &bookmarks); err != nil {
		log.Fatalf("failed to parse the bookmarks file: %v", err)
	}
	bar, ok := bookmarks.Roots["bookmark_bar"]
	if !ok {
		log.Fatalf("there are no bookmarks in the bookmarks bar")
	}

	chrome, err := exec.LookPath("google-chrome-stable")
	if err != nil {
		log.Fatal(err)
	}

	// By convention, the first of these strings (i.e., argv[0]) should contain
	// the filename associated with the file being executed
	// (https://man7.org/linux/man-pages/man2/execve.2.html).
	argv := []string{chrome}
	// Enable dark mode.
	argv = append(argv, "--enable-features=WebUIDarkMode", "--force-dark-mode")
	for _, bookmark := range bar.Children {
		argv = append(argv, bookmark.URL)
	}

	if err := syscall.Exec(chrome, argv, os.Environ()); err != nil {
		log.Fatal(err)
	}
}
