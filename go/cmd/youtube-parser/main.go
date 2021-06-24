package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	f, err := os.Open("liked.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	for n, id := range records {
		if n == 0 {
			continue
		}
		ids = append(ids, id[0])
	}

	for _, id := range ids {
		b, err := fetchTitle(id)
		if err != nil {
			log.Printf("%s: %v", id, err)
		}
		fmt.Printf("%s: %s\n", id, string(b))
	}
}

func fetchTitle(id string) ([]byte, error) {
	cmd := exec.Command("youtube-dl", "-e", "https://www.youtube.com/watch?v="+id)
	return cmd.Output()
}
