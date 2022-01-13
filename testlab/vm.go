//usr/bin/env go run $0 $@ ; exit "$?"

//go:build ignore

// This is a program that launches QEMU (https://qemu.org) VMs for experiments
// and does some other interesting things.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"text/tabwriter"

	"git.astrophena.name/infra/util/run"
)

type command struct {
	f    func(args []string) error
	desc string
}

// commands is a list of available commands. Please keep it sorted.
var commands = map[string]command{
	"debian": command{
		f:    startFunc("debian"),
		desc: "Start Debian VM.",
	},
	"plan9": command{
		f:    startFunc("plan9"),
		desc: "Start Plan 9 VM.",
	},
}

func main() {
	log.SetFlags(0)

	args := os.Args[1:]
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" {
		usage()
		return
	}

	cmd, ok := commands[args[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "%s: unknown command\n\n", args[0])
		os.Exit(127)
	}

	if err := cmd.f(args[1:]); err != nil {
		log.Fatalf("%s: %v", args[0], err)
	}
}

func usage() {
	w := tabwriter.NewWriter(os.Stderr, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(w, "Usage: ./vm.go [command]\n\n")
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

func startFunc(name string) func(args []string) error {
	return func(args []string) error {
		flags := flag.NewFlagSet(name, flag.ContinueOnError)
		var (
			cdrom    = flags.String("cdrom", "", "Path to the ISO `file` that should be attached to VM.")
			userData = flags.String("user-data", "", "Path to the user data `file`.")
			gui      = flags.Bool("gui", name == "plan9", "Run in GUI mode.")
		)
		if err := flags.Parse(args); errors.Is(err, flag.ErrHelp) {
			return nil
		} else if err != nil {
			return err
		}

		qemu := exec.Command("qemu-system-x86_64")

		if !*gui {
			qemu.Args = append(qemu.Args, "-nographic")
		}
		qemu.Stdout = os.Stdout
		qemu.Stderr = os.Stderr
		// See https://wiki.gentoo.org/wiki/QEMU/Options for all available QEMU
		// options.
		qemu.Args = append(qemu.Args,
			"-enable-kvm",
			"-m", "1024",
			"-nic", "user,model=virtio",
			"-drive", "file="+filepath.Join("images", name)+".qcow2,media=disk,if=virtio",
			"-device", "virtio-net-pci,netdev=net0",
			"-netdev", "user,id=net0,hostfwd=tcp::2222-:22",
		)

		if *cdrom != "" {
			qemu.Args = append(qemu.Args, "-cdrom", *cdrom, "-boot", "-d")
		}

		if *userData != "" {
			tmpdir, err := os.MkdirTemp("", "testlab-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpdir)

			seedImg := filepath.Join(tmpdir, "seed.img")
			if err := run.Command("cloud-localds", seedImg, *userData).Run(); err != nil {
				return err
			}
			qemu.Args = append(qemu.Args, "-drive", "if=virtio,format=raw,file="+seedImg)
		}

		if err := qemu.Run(); err != nil {
			return err
		}

		return nil
	}
}
