// Package main is the entry point for the Shifter CLI.
package main

import (
	"fmt"
	"os"

	"github.com/moximxxx/shifter/cmd"
)

// Build information — set via ldflags at compile time:
//
//	go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -Iseconds)"
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Inject version into root command
	cmd.SetVersion(fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date))

	// Handle --version / -v
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("shifter version %s\n", cmd.VersionString())
		os.Exit(0)
	}

	cmd.Execute()
}
