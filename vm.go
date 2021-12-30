//usr/bin/env go run $0 $@ ; exit "$?"

//go:build ignore

// This is a program that launches QEMU (https://qemu.org) VM running Debian for
// experiments.
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"git.astrophena.name/exp/testlab/devtools/deploy"
)

func main() {
	log.SetFlags(0)
	var (
		doDeploy = flag.Bool("deploy", false, "deploy everything to the testlab, VM must be running")
	)
	flag.Parse()

	if *doDeploy {
		if err := deploy.Do(); err != nil {
			log.Fatalf("Deploy failed: %v", err)
		}
		return
	}

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
		log.Fatalf("QEMU exited: %v", err)
	}
}
