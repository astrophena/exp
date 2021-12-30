//usr/bin/env go run $0 $@ ; exit "$?"

//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"text/tabwriter"
)

type command struct {
	f    func() error
	desc string
}

// commands is a list of available commands. Please keep it sorted.
var commands = map[string]command{
	"start": {
		f:    start,
		desc: "Start VM.",
	},
}

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) < 1 || args[0] == "help" || args[0] == "-h" {
		usage()
		return
	}

	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: unknown command\n\n", args[0])
		os.Exit(127)
	}

	if err := cmd.f(); err != nil {
		log.Fatalf("%s: %v", args[0], err)
	}
}

func usage() {
	w := tabwriter.NewWriter(os.Stderr, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(w, "Usage: ./run.go [command]\n\n")
	fmt.Fprintf(w, "Available commands:\n\n")

	keys := make([]string, 0, len(commands))
	for key := range commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Fprintf(w, "  %s\t%s\t\n", key, commands[key].desc)
	}

	fmt.Fprintln(w, "")

	w.Flush()
}

func logf(format string, args ...any) { log.Printf("==> "+format, args...) }

func start() error {
	args := []string{
		"-enable-kvm",
		"-m", "1024",
		"-nic", "user,model=virtio",
		"-drive", "file=disk.qcow2,media=disk,if=virtio",
		"-nographic",
	}

	qemu := exec.Command("qemu-system-x86_64", args...)
	qemu.Stdin = os.Stdin
	qemu.Stdout = os.Stdout
	qemu.Stderr = os.Stderr
	if err := qemu.Run(); err != nil {
		return err
	}

	return nil
}
