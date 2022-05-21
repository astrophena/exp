// Command renamer renames files sequentially.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	log.SetFlags(0)

	var (
		dir   = flag.String("dir", ".", "modify files in `path`")
		start = flag.Int("start", 1, "start from")
	)
	flag.Parse()

	fullDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Are you sure? This will sequentially rename all files in %s. ", fullDir)
	if !askForConfirmation() {
		log.Printf("Canceled.")
		return
	}

	if err := filepath.Walk(fullDir, rename(*start)); err != nil {
		log.Fatal(err)
	}
}

func askForConfirmation() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Println("I'm sorry but I didn't get what you meant, please type (y)es or (n)o and then press Enter:")
		return askForConfirmation()
	}
}

func rename(start int) filepath.WalkFunc {
	return func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		var (
			ext      = filepath.Ext(path)
			basename = filepath.Base(path)
			newname  = fmt.Sprintf("%d%s", start, ext)
		)

		if basename == "desktop.ini" {
			log.Printf("Skipping desktop.ini.")
			return nil
		}

		log.Printf("Renaming %s to %s.", basename, newname)
		if err := os.Rename(basename, newname); err != nil {
			return err
		}

		start++

		return nil
	}
}
