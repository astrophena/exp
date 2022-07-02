//go:build linux

// Command chrome-open is a Chrome wrapper that does some useful things.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"go.astrophena.name/exp/cmd"
)

func main() {
	cmd.SetArgsUsage("[option] [path|url]")
	cmd.SetDescription("Chrome wrapper that does some useful things. To see Chrome help, run 'google-chrome --help'.")

	openBookmarksBar := flag.Bool("bookmarks-bar", false, "Open everything in the bookmarks bar.")
	cmd.HandleStartup()

	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	if *openBookmarksBar {
		args = getBookmarksBar()
	}
	execChrome(args)
}

func getBookmarksBar() []string {
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

	var urls []string
	for _, bookmark := range bar.Children {
		urls = append(urls, bookmark.URL)
	}
	return urls
}

func execChrome(args []string) {
	chrome, err := exec.LookPath("google-chrome-stable")
	if err != nil {
		log.Fatalf("failed to find Chrome binary: %v", err)
	}

	// By convention, the first of these strings (i.e., argv[0]) should contain
	// the filename associated with the file being executed
	// (https://man7.org/linux/man-pages/man2/execve.2.html).
	argv := []string{chrome}
	// Enable dark mode.
	argv = append(argv, "--enable-features=WebUIDarkMode", "--force-dark-mode")
	argv = append(argv, args...)

	if err := syscall.Exec(chrome, argv, os.Environ()); err != nil {
		log.Fatalf("failed to exec Chrome binary: %v", err)
	}
}
