// Command cmdtop displays the top of most used Bash commands.
package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	log.SetFlags(0)

	num := int64(10)
	if len(os.Args) > 1 {
		var err error
		num, err = strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(filepath.Join(home, ".bash_history"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	m := make(map[string]int)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		cmd := strings.Fields(scanner.Text())
		if len(cmd) > 0 && cmd[0] != "" {
			m[cmd[0]]++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// https://stackoverflow.com/a/44380276
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for i, kv := range ss {
		if int64(i) == num {
			break
		}
		log.Printf("%d. %s (%d)", i+1, kv.Key, kv.Value)
	}
}
