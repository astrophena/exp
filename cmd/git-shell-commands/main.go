// Command git-shell-commands implements the Git server SSH commands.
//
// See https://git-scm.com/docs/git-shell#_commands.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var commands = map[string]func(){
	"no-interactive-login": noInteractiveLogin,
}

func main() {
	callname := filepath.Base(os.Args[0])

	cmd, ok := commands[callname]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: no such command\n", callname)
		os.Exit(127)
	}
	cmd()
}

func noInteractiveLogin() {
	fmt.Fprint(os.Stderr, `Hi! You've successfully authenticated, but we do not
provide interactive shell access.
`)
	os.Exit(128)
}
