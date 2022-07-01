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
	var args []string
	// Enable dark mode.
	args = append(args, "--enable-features=WebUIDarkMode", "--force-dark-mode")
	for _, bookmark := range bar.Children {
		args = append(args, bookmark.URL)
	}

	chrome, err := exec.LookPath("google-chrome-stable")
	if err != nil {
		log.Fatal(err)
	}
	if err := syscall.Exec(chrome, args, os.Environ()); err != nil {
		log.Fatal(err)
	}
}
