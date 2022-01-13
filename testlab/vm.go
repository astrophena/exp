//usr/bin/env go run $0 $@ ; exit "$?"

//go:build ignore

// This is a program that launches QEMU (https://qemu.org) VMs for experiments.
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"git.astrophena.name/infra/util/run"
)

type command struct {
	f    func(args []string) error
	desc string
}

func main() {
	log.SetFlags(0)

	userData := flag.String("user-data", "", "Path to the user data `file`.")
	flag.Parse()

	name := "debian"
	if len(flag.Args()) > 0 {
		name = flag.Args()[0]
	}

	if err := vm(name, *userData); err != nil {
		log.Fatal(err)
	}
}

func vm(name, userData string) error {
	qemu := exec.Command("qemu-system-x86_64")

	qemu.Stdout = os.Stdout
	qemu.Stderr = os.Stderr
	// See https://wiki.gentoo.org/wiki/QEMU/Options for all available QEMU
	// options.
	qemu.Args = append(qemu.Args,
		"-enable-kvm",
		"-nographic",
		"-m", "1024",
		"-nic", "user,model=virtio",
		"-drive", "file="+filepath.Join("images", name)+".qcow2,media=disk,if=virtio",
	)

	if userData != "" {
		tmpdir, err := os.MkdirTemp("", "testlab-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpdir)

		seedImg := filepath.Join(tmpdir, "seed.img")
		if err := run.Command("cloud-localds", seedImg, userData).Run(); err != nil {
			return err
		}
		qemu.Args = append(qemu.Args, "-drive", "if=virtio,format=raw,file="+seedImg)
	}

	if err := qemu.Run(); err != nil {
		return err
	}

	return nil
}
