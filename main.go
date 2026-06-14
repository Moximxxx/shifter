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

	// Compact version format: v0.2.1(2230e1c)
	ver := fmt.Sprintf("v%s(%s)", version, commit)
	if version == "dev" {
		ver = "dev"
	}
	cmd.SetVersion(ver)

	// Handle --version / -v
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println(ver)
		os.Exit(0)
	}

	// No arguments → launch interactive TUI
	if len(os.Args) <= 1 {
		cmd.LaunchTUI()
		return
	}

	cmd.Execute()
}
