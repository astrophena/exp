// Package deploy deploys everything to the server using Ansible and rsync.
package deploy

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"git.astrophena.name/infra/util/run"
)

var host = struct {
	name     string
	user     string
	binaries []string
	services []*Service
}{
	name: "testlab",
	user: "astrophena",
	services: []*Service{
		{Name: "example"},
	},
}

// Do performs a deployment to the server.
func Do() error {
	tmpDir, err := os.MkdirTemp("", "infra-deploy-"+host.name)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var (
		servicesDir = filepath.Join(tmpDir, "services") // service binaries
		unitsDir    = filepath.Join(tmpDir, "units")    // systemd units
	)
	for _, dir := range []string{servicesDir, unitsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	for _, s := range host.services {
		install := run.Command("go", "install", "./cmd"+s.Name)
		install.Env = append(os.Environ(), "GOBIN="+servicesDir)
		if err := install.Run(); err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join(unitsDir, s.Name+".service"), s.genSystemdUnit(), 0o644); err != nil {
			return err
		}
	}

	return nil
}

// logf implements the logger.Logf.
func logf(format string, args ...interface{}) { log.Printf("==> "+format, args...) }

// sync is a wrapper around rsync. It is run and originates on the local host
// where deploy is being run. Argument names corresponds to the rsync command
// options, see rsync(1) to learn more about them.
func sync(src, dest, chown string, delete bool) error {
	args := []string{
		"--quiet",                 // suppresses non-error messages
		"--archive",               // preserves most file properties (permissions, ownership and so on)
		"--rsync-path=sudo rsync", // runs rsync as root
		"--chown", chown,          // changes ownership to the given one
	}

	if delete {
		args = append(args, "--delete") // means files deleted on the source are to be deleted on the destination as well
	}

	args = append(args, src)
	args = append(args, fmt.Sprintf("%s@[%s]:%s", host.user, host.name, dest))

	return run.Command("rsync", args...).Run()
}

// remote runs command on a host.
func remote(args ...string) *exec.Cmd {
	return run.Command("ssh", append([]string{host.user + "@" + host.name}, args...)...)
}

// systemctl runs a systemctl action with arguments.
func systemctl(action string, args ...string) error {
	args = append([]string{"sudo", "systemctl", action}, args...)
	return remote(args...).Run()
}
