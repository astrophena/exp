// Command rofi-wiki-menu implements a Rofi (https://github.com/davatorium/rofi)
// mode for quickly opening Vimwiki pages.
//
// See rofi-script(5) to learn more about Rofi script mode API.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"go.astrophena.name/exp/cmd"
)

func dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("os.UserHomeDir(): %v", err)
	}
	return filepath.Join(home, "src", "wiki")
}

func parseTitle(path string) (title string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "#") {
			title = strings.TrimPrefix(s.Text(), "# ")
			break
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}

	return title, nil
}

func openPage(name string) error {
	// https://vi.stackexchange.com/a/26735
	vim := exec.Command(term[0], term[1], "vim", "-c", "VimwikiIndex", "-c", "VimwikiGoto "+strings.TrimSuffix(name, filepath.Ext(name)))
	vim.Env = os.Environ()
	vim.Stderr = os.Stderr
	return vim.Start()
}

var term = []string{"kitty", "--single-instance"}

// https://regex101.com/r/00YyXk/2
var spanRe = regexp.MustCompile(`\p{L}+\s+<span font_size="small">(.*?)</span>`)

func main() {
	dir := flag.String("dir", dir(), "Directory where wiki files are stored.")
	cmd.HandleStartup()

	// User has selected an entry, open it.
	hasSelection := len(flag.Args()) == 1
	if hasSelection {
		rawName := flag.Args()[0]
		if spanRe.MatchString(rawName) {
			match := spanRe.FindStringSubmatch(rawName)
			if len(match) == 2 {
				if err := openPage(match[1]); err != nil {
					log.Fatalf("openPage(%s): %v", match[1], err)
				}
				os.Exit(0)
			}
		}
	}

	pages := make(map[string]string)
	if err := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		title, err := parseTitle(path)
		if err != nil {
			return fmt.Errorf("parseTitle(%s): %v", path, err)
		}
		pages[d.Name()] = title

		return nil
	}); err != nil {
		log.Fatalf("filepath.WalkDir: %v", err)
	}

	// Don't allow custom entries.
	io.WriteString(os.Stdout, "\x00no-custom\x1ftrue\n")
	// Use markup.
	io.WriteString(os.Stdout, "\x00markup-rows\x1ftrue\n")
	// Write the prompt.
	io.WriteString(os.Stdout, "\x00prompt\x1fOpen a wiki page\n")

	hasIndex := false
	const indexName = "index.md"

	names := make([]string, 0, len(pages))
	for name := range pages {
		if name == indexName {
			hasIndex = true
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if hasIndex {
		printPage(pages[indexName], indexName)
	}
	for _, name := range names {
		printPage(pages[name], name)
	}
}

func printPage(title, name string) {
	fmt.Fprintf(os.Stdout, `%s <span font_size="small">%s</span>`+"\n", title, name)
}
