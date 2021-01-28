// Command git-shell-commands implements the Git server SSH commands.
//
// See https://git-scm.com/docs/git-shell#_commands.
//
// TODO(astrophena): merge this into infra after the freeze ends.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

var commands = map[string]func(){
	"new":                  newRepo,
	"no-interactive-login": noInteractiveLogin,
	"ls":                   ls,
}

func main() {
	log.SetFlags(0)
	color.NoColor = false

	callname := filepath.Base(os.Args[0])

	cmd, exists := commands[callname]
	if !exists {
		fmt.Fprintf(os.Stderr, "%s: no such command.\n", callname)
		os.Exit(127)
	}
	cmd()
}

func newRepo() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Repo name is required.\n")
		os.Exit(1)
	}
	repo := os.Args[1]

	if !strings.HasSuffix(repo, ".git") {
		repo = repo + ".git"
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	repopath := filepath.Join(homedir, repo)
	log.Printf("Creating %s", repo)
	if err := exec.Command("git", "init", "--bare", repopath).Run(); err != nil {
		log.Fatal(err)
	}
}

func noInteractiveLogin() {
	fmt.Fprint(os.Stderr, `Hi! You've successfully authenticated, but we do not
provide interactive shell access.
`)
	os.Exit(128)
}

func ls() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	host, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	repos, err := filepath.Glob(filepath.Join(homedir, "*.git"))
	if err != nil {
		log.Fatal(err)
	}

	color.Yellow("Repos on %s:", host)
	for _, repo := range repos {
		desc, err := ioutil.ReadFile(filepath.Join(repo, "description"))
		if err != nil {
			if !os.IsNotExist(err) {
				log.Fatal(err)
			}
			desc = []byte("No description.")
		}

		log.Printf("%s: %s", color.MagentaString(filepath.Base(repo)), desc)
	}
}
