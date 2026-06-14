// Package main is the entry point for the Shifter CLI.
package main

import (
	"fmt"
	"os"

	"github.com/moximxxx/shifter/cmd"
	"github.com/moximxxx/shifter/pkg/i18n"
	"github.com/moximxxx/shifter/pkg/settings"
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
	// Initialize i18n from user settings
	s, _ := settings.Load()
	if s != nil && s.Lang != "" {
		i18n.Init(s.Lang)
	} else {
		i18n.Init("en")
	}

	// Inject version into root command
	cmd.SetVersion(fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date))

	// Handle --version / -v
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("shifter version %s\n", cmd.VersionString())
		os.Exit(0)
	}

	cmd.Execute()
}
