// Command cmdtop displays the top of most used commands in bash history.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"go.astrophena.name/exp/cmd"
)

func main() {
	cmd.SetDescription("cmdtop displays the top of most used commands in bash history.")
	cmd.SetArgsUsage("[num] [flags]")
	cmd.HandleStartup()

	num := int64(10)
	args := flag.Args()
	if len(args) > 0 {
		var err error
		num, err = strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			log.Fatalf("Invalid number of commands: %v", err)
		}
	}

	histfile, ok := os.LookupEnv("HISTFILE")
	if !ok {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		histfile = filepath.Join(home, ".bash_history")
	}

	f, err := os.Open(histfile)
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

	type kv struct {
		key   string
		value int
	}
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].value > ss[j].value
	})
	for i, kv := range ss {
		if int64(i) == num {
			break
		}
		log.Printf("%d. %s (%d)", i+1, kv.key, kv.value)
	}
}
