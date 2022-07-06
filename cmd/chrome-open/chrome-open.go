//go:build linux

/*
Command chrome-open is a Chrome launcher that can open all URLs in the bookmarks
bar as a separate tabs.

When running under i3 if Chrome is already running it focuses the Chrome window
instead (this is disabled when -bookmarks-bar flag is passed).

It launches Chrome with flags defined in $XDG_CONFIG_HOME/chrome-flags.conf
(~/.config/chrome-flags.conf if $XDG_CONFIG_HOME is not set).
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"go.i3wm.org/i3/v4"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("chrome-open: ")

	flag.Usage = usage
	var (
		openBookmarksBar = flag.Bool("bookmarks-bar", false, "Open everything in the bookmarks bar.")
		i3Focus          = flag.Bool("i3-focus", true, "When running under i3, focus the current Chrome window if it's already running.")
		binary           = flag.String("chrome-binary", "google-chrome-stable", "Chrome binary name.")
	)
	flag.Parse()

	if *i3Focus && !*openBookmarksBar {
		launched, err := focus()
		if err != nil {
			log.Printf("focus(): %v", err)
		}
		if launched && err == nil {
			return
		}
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("os.UserConfigDir(): %v", err)
	}

	var args []string
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	if *openBookmarksBar {
		args = getBookmarksBar(configDir)
	}
	if err := run(*binary, configDir, args); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	const doc = `Command chrome-open is a Chrome launcher that can open all URLs in the bookmarks
bar as a separate tabs.

When running under i3 if Chrome is already running it focuses the Chrome window
instead (this is disabled when -bookmarks-bar flag is passed).

It launches Chrome with flags defined in $XDG_CONFIG_HOME/chrome-flags.conf
(~/.config/chrome-flags.conf if $XDG_CONFIG_HOME is not set).
`
	fmt.Fprint(os.Stderr, doc)
	fmt.Fprintf(os.Stderr, "\nUsage: chrome-open [chrome-open and/or Chrome flags] [URL]\n\n")
	fmt.Fprintf(os.Stderr, "chrome-open flags:\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nTo see Chrome flags, run 'man google-chrome'.")
}

func getBookmarksBar(configDir string) []string {
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

func focus() (launched bool, err error) {
	tree, err := i3.GetTree()
	if err != nil {
		return false, err
	}
	if win := tree.Root.FindChild(func(n *i3.Node) bool { return strings.HasSuffix(n.Name, "- Google Chrome") }); win != nil {
		if _, err := i3.RunCommand(fmt.Sprintf(`[con_id="%d"] focus`, win.ID)); err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func run(binary, configDir string, args []string) error {
	// Find the Chrome binary.
	chrome, err := exec.LookPath(binary)
	if err != nil {
		return fmt.Errorf("failed to find Chrome binary: %v", err)
	}

	// By convention, the first of these strings (i.e., argv[0]) should contain
	// the filename associated with the file being executed
	// (https://man7.org/linux/man-pages/man2/execve.2.html).
	argv := []string{chrome}

	// Read flags from $XDG_CONFIG_HOME/chrome-flags.conf and add them.
	if bs, err := os.ReadFile(filepath.Join(configDir, "chrome-flags.conf")); err == nil {
		flags := strings.Split(string(bs), "\n")
		if len(flags) > 0 {
			for _, flag := range flags {
				// Ignore empty lines and comments.
				if flag == "" || strings.HasPrefix(flag, "#") {
					continue
				}
				argv = append(argv, flag)
			}
		}
	}
	if len(args) > 0 {
		argv = append(argv, args...)
	}

	return syscall.Exec(chrome, argv, os.Environ())
}
