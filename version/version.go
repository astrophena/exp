// Package version provides the version and build information.
package version

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
)

// Info is the version and build information of the current binary.
type Info struct {
	Commit  string `json:"commit"`   // BuildInfo's vcs.revision
	BuiltAt string `json:"built_at"` // BuildInfo's vcs.date
	Go      string `json:"go"`       // runtime.Version()
	OS      string `json:"os"`       // runtime.GOOS
	Arch    string `json:"arch"`     // runtime.GOARCH
}

// String implements the fmt.Stringer interface.
func (i Info) String() string {
	return fmt.Sprintf(`Commit: %s
Built at: %s
Go: %s
OS: %s
Architecture: %s
`, i.Commit, i.BuiltAt, i.Go, i.OS, i.Arch)
}

var (
	once    sync.Once
	cmdName string
	info    Info
)

// CmdName returns the base name of the current binary.
func CmdName() string {
	once.Do(initOnce)
	return cmdName
}

// Version returns the version and build information of the current binary.
func Version() Info {
	once.Do(initOnce)
	return info
}

func initOnce() {
	i := &Info{
		Commit:  "HEAD",
		BuiltAt: "undefined",
		Go:      runtime.Version(),
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
	cmdName = "cmd"

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Printf("version: failed to read build information")
		return
	}

	exe, err := os.Executable()
	if err == nil {
		cmdName = filepath.Base(exe)
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			i.Commit = s.Value
		case "vcs.time":
			i.BuiltAt = s.Value
		}
	}
	info = *i
}
