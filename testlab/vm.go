//usr/bin/env go run $0 $@ ; exit "$?"

//go:build ignore

// This is a program that launches QEMU (https://qemu.org) VMs for experiments.
// Dependencies: qemu-kvm, cloud-image-utils.
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"go.astrophena.name/exp/cmd"
)

func main() {
	log.SetFlags(0)

	userData := flag.String("user-data", "", "Path to the user data `file`.")
	ssh := flag.Bool("ssh", false, "SSH into machine")
	cmd.HandleStartup()

	name := "debian"
	if len(flag.Args()) > 0 {
		name = flag.Args()[0]
	}

	if *ssh {
		ssh := exec.Command("ssh", "-p", "8022", "-o", "StrictHostKeyChecking=no", "localhost")
		ssh.Stdin = os.Stdin
		ssh.Stdout = os.Stdout
		ssh.Stderr = os.Stderr
		if err := ssh.Run(); err != nil {
			log.Fatal(err)
		}
		return
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
		"-device", "virtio-net-pci,netdev=net0",
		"-netdev", "user,id=net0,hostfwd=tcp::8022-:22",
		"-drive", "file="+name+".qcow2,media=disk,if=virtio",
	)

	if userData != "" {
		tmpdir, err := os.MkdirTemp("", "testlab-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpdir)

		seedImg := filepath.Join(tmpdir, "seed.img")
		if err := exec.Command("cloud-localds", seedImg, userData).Run(); err != nil {
			return err
		}
		qemu.Args = append(qemu.Args, "-drive", "if=virtio,format=raw,file="+seedImg)
	}

	if err := qemu.Run(); err != nil {
		return err
	}

	return nil
}
