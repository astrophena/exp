package deploy

import "bytes"

// Service represents a service to run.
type Service struct {
	// Name is the name of the service. This is the name of the systemd service
	// (without .service).
	Name string // Example: tgbotd.
	// NeedState reports whether the service needs to keep a persistent state.
	NeedState bool // False by default.
}

func (s *Service) genSystemdUnit() []byte {
	b := new(bytes.Buffer)

	var (
		firstSection = true
		section      = func(name string) {
			if !firstSection {
				b.WriteString("\n")
			}
			b.WriteString("[" + name + "]" + "\n")
			firstSection = false
		}
		opt = func(name, val string) { b.WriteString(name + "=" + val + "\n") }
	)

	b.WriteString("# Generated by git.astrophena.name/testlab/devtools/deploy; DO NOT EDIT.\n")
	b.WriteString("# To update this file, change devtools/deploy/config.go and redeploy.\n\n")

	section("Unit")
	opt("Description", s.Name)

	section("Service")
	opt("Type", "notify")                            // https://www.freedesktop.org/software/systemd/man/systemd.service.html#Type=
	opt("ExecStart", "/usr/local/lib/infra/"+s.Name) // https://www.freedesktop.org/software/systemd/man/systemd.service.html#ExecStart=
	opt("Restart", "on-failure")                     // https://www.freedesktop.org/software/systemd/man/systemd.service.html#Restart=
	opt("User", s.Name)                              // https://www.freedesktop.org/software/systemd/man/systemd.exec.html#User=
	opt("DynamicUser", "yes")                        // https://www.freedesktop.org/software/systemd/man/systemd.exec.html#DynamicUser=
	// https://www.freedesktop.org/software/systemd/man/systemd.exec.html#RuntimeDirectory=
	opt("RuntimeDirectory", s.Name)
	if s.NeedState {
		opt("StateDirectory", s.Name)
	}
	// For sandboxing options used by all services, see sandboxing.conf.

	section("Install")
	opt("Alias", "infra-"+s.Name+".service")
	opt("WantedBy", "multi-user.target")

	return b.Bytes()
}
