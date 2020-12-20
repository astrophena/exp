// Command renamer renames files sequentially.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(0)

	var (
		dir   = flag.String("dir", ".", "modify files in `path`")
		start = flag.Int("start", 1, "start from")
	)
	flag.Parse()

	if err := filepath.Walk(*dir, rename(*start)); err != nil {
		log.Fatal(err)
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

		log.Printf("renaming %s to %s", basename, newname)
		if err := os.Rename(basename, newname); err != nil {
			return err
		}

		start++

		return nil
	}
}
