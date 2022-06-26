// Package cmd contains common command-line flags and configuration
// options.
package cmd

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"go.astrophena.name/exp/version"
)

var opts struct {
	description, argsUsage string
}

// SetDescription sets the command description.
func SetDescription(description string) { opts.description = description }

// SetArgsUsage sets the command arguments help string.
func SetArgsUsage(argsUsage string) { opts.argsUsage = argsUsage }

// HandleStartup handles the command startup.
func HandleStartup() {
	log.SetFlags(0)

	if opts.argsUsage == "" {
		opts.argsUsage = "[flags]"
	}
	flag.Usage = usage
	showVersion := flag.Bool("version", false, "Show version.")
	flag.Parse()

	if *showVersion {
		io.WriteString(os.Stderr, version.Version().String())
		os.Exit(0)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s %s\n\n", version.CmdName(), opts.argsUsage)
	if opts.description != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", opts.description)
	}
	fmt.Fprint(os.Stderr, "Available flags:\n\n")
	flag.PrintDefaults()
}
