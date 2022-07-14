//go:build linux

/*
Command chrome-open is a Chrome launcher that can open all URLs in the bookmarks
bar as a separate tabs.

When running under i3 it also focuses the already or newly opened Chrome window.

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

	"go.i3wm.org/i3/v4"
)

var (
	openBookmarksBar = flag.Bool("bookmarks-bar", false, "Open everything (or some bookmarks) from the bookmarks bar.")
	bookmarksLimit   = flag.Int("bookmarks-limit", 0, "Open n first bookmarks. If 0, open everything.")
	i3Focus          = flag.Bool("i3-focus", true, "When running under i3, focus the current Chrome window if it's already running.")
	binary           = flag.String("chrome-binary", "google-chrome-stable", "Chrome binary name.")
	chromeFlags      = flag.String("chrome-flags", "", "Additional flags to pass to the Chrome binary.")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("chrome-open: ")

	flag.Usage = usage
	flag.Parse()

	var args []string
	if flag.NArg() > 0 {
		args = flag.Args()
	}
	if *i3Focus && len(args) == 0 && !*openBookmarksBar {
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
	if *openBookmarksBar {
		args = getBookmarksBar(configDir, *bookmarksLimit)
	}

	if err := run(configDir, args); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	const doc = `Command chrome-open is a Chrome launcher that can open all URLs in the bookmarks
bar as a separate tabs.

When running under i3 it also focuses the already or newly opened Chrome window.

It launches Chrome with flags defined in $XDG_CONFIG_HOME/chrome-flags.conf
(~/.config/chrome-flags.conf if $XDG_CONFIG_HOME is not set).
`
	fmt.Fprint(os.Stderr, doc)
	fmt.Fprintf(os.Stderr, "\nUsage: chrome-open [flags] [URL]\n\n")
	fmt.Fprintf(os.Stderr, "chrome-open flags:\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nTo see Chrome flags, run 'man google-chrome'.\n")
}

func getBookmarksBar(configDir string, limit int) []string {
	b, err := os.ReadFile(filepath.Join(configDir, "google-chrome", "Default", "Bookmarks"))
	if err != nil {
		log.Fatalf("failed to read the bookmarks file: %v", err)
	}

	var bookmarks struct {
		Roots map[string]struct {
			Children []struct {
				URL  string `json:"url"`
				Type string `json:"type"`
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
		if bookmark.Type == "folder" {
			continue
		}
		urls = append(urls, bookmark.URL)
	}

	if len(urls) <= limit {
		return urls
	}
	return urls[:limit]
}

func focus() (launched bool, err error) {
	// Check if i3 is running.
	if !i3.IsRunningHook() {
		return false, nil
	}

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

func run(configDir string, args []string) error {
	// Read flags from $XDG_CONFIG_HOME/chrome-flags.conf and add them.
	if bs, err := os.ReadFile(filepath.Join(configDir, "chrome-flags.conf")); err == nil {
		flags := strings.Split(string(bs), "\n")
		if len(flags) > 0 {
			for _, flag := range flags {
				// Ignore empty lines and comments.
				if flag == "" || strings.HasPrefix(flag, "#") {
					continue
				}
				args = append(args, flag)
			}
		}
	}

	if *chromeFlags != "" {
		args = append(args, strings.Fields(*chromeFlags)...)
	}

	// Start Chrome in a detached process.
	chrome := exec.Command(*binary, args...)
	chrome.Stdout = os.Stdout
	chrome.Stderr = os.Stderr
	if err := chrome.Start(); err != nil {
		return err
	}

	// Focus the window.
	if _, err := focus(); err != nil {
		return err
	}

	return nil
}
