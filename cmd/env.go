package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/pkg/env"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show Shifter environment configuration",
	Long: `Display all SHIFTER_* environment variables and their current values.

Use these variables to set defaults so you don't need to repeat
flags on every command.

Examples:
  shifter env
  export SHIFTER_SOURCE=claude-code
  export SHIFTER_TARGET=codex
  shifter port   # uses env vars for source/target`,
	RunE: runEnv,
}

var envInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Output shell configuration for Shifter environment variables",
	Long: `Generate shell commands to configure Shifter environment variables.

Examples:
  eval "$(shifter env init)"           # bash/zsh
  shifter env init --shell fish        # fish shell`,
	RunE: runEnvInit,
}

var envShell string

func init() {
	envInitCmd.Flags().StringVar(&envShell, "shell", "", "Shell type: bash, zsh, fish (auto-detect if empty)")
	envCmd.AddCommand(envInitCmd)
	rootCmd.AddCommand(envCmd)
}

func runEnv(cmd *cobra.Command, args []string) error {
	config := env.Config()

	if len(config) == 0 {
		fmt.Println("No SHIFTER_* environment variables configured.")
		fmt.Println()
		fmt.Println("Available variables:")
		printAllVars()
		fmt.Println()
		fmt.Println("Use 'shifter env init' to generate shell configuration.")
		return nil
	}

	fmt.Println("Current SHIFTER environment:")
	fmt.Println()
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("  %-24s = %s\n", k, config[k])
	}

	fmt.Println()
	fmt.Println("Active defaults:")
	fmt.Printf("  Source:      %s\n", nvl(env.Source(), "(not set, required)"))
	fmt.Printf("  Target:      %s\n", nvl(env.Target(), "(not set, use --to flag)"))
	fmt.Printf("  Scope:       %s\n", env.Scope())
	fmt.Printf("  Dry-run:     %v\n", env.DryRun())
	fmt.Printf("  No backup:   %v\n", env.NoBackup())
	fmt.Printf("  Force:       %v\n", env.Force())
	fmt.Printf("  Strategy:    %s\n", env.Strategy())
	fmt.Printf("  Aspects:     %s\n", nvl(strings.Join(env.Aspects(), ","), "(all)"))
	fmt.Printf("  Profile:     %s\n", nvl(env.Profile(), "(not set)"))
	fmt.Printf("  Project:     %s\n", env.ProjectRoot())

	return nil
}

func runEnvInit(cmd *cobra.Command, args []string) error {
	shell := envShell
	if shell == "" {
		shell = detectShell()
	}

	var output string
	switch shell {
	case "fish":
		output = fishConfig()
	default:
		output = bashConfig()
	}

	fmt.Print(output)
	return nil
}

func bashConfig() string {
	return `# Shifter — coding agent config migration tool
# Add to ~/.bashrc or ~/.zshrc

export SHIFTER_SOURCE="claude-code"      # default source agent
export SHIFTER_TARGET=""                  # default target agent (set to codex/opencode/qoder)
export SHIFTER_SCOPE="project"            # project | global | all
export SHIFTER_STRATEGY="newer"           # newer | merge | source
# export SHIFTER_DRY_RUN="1"              # uncomment to always preview first
# export SHIFTER_NO_BACKUP="1"            # uncomment to skip backups
# export SHIFTER_FORCE="1"                # uncomment to skip confirmations
# export SHIFTER_ASPECTS="agents,mcp"     # uncomment for default aspect filter
# export SHIFTER_PROFILE="my-workflow"    # uncomment for default profile
`
}

func fishConfig() string {
	return `# Shifter — coding agent config migration tool
# Add to ~/.config/fish/config.fish

set -gx SHIFTER_SOURCE "claude-code"
set -gx SHIFTER_TARGET ""
set -gx SHIFTER_SCOPE "project"
set -gx SHIFTER_STRATEGY "newer"
# set -gx SHIFTER_DRY_RUN "1"
# set -gx SHIFTER_NO_BACKUP "1"
# set -gx SHIFTER_FORCE "1"
`
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	return "bash"
}

func printAllVars() {
	vars := []struct{ name, desc string }{
		{"SHIFTER_SOURCE", "Default source agent (claude-code, codex, opencode, qoder, gemini-cli, cline, aider)"},
		{"SHIFTER_TARGET", "Default target agent"},
		{"SHIFTER_SCOPE", "Default scope: project | global | all"},
		{"SHIFTER_PROJECT_ROOT", "Default project root directory"},
		{"SHIFTER_DRY_RUN", "Set to 1/true to enable dry-run by default"},
		{"SHIFTER_NO_BACKUP", "Set to 1/true to skip automatic backups"},
		{"SHIFTER_FORCE", "Set to 1/true to skip confirmation prompts"},
		{"SHIFTER_ASPECTS", "Default aspects: agents,skills,mcp,permissions,hooks,commands,settings"},
		{"SHIFTER_PROFILE", "Default profile name for save/load"},
		{"SHIFTER_STRATEGY", "Default merge strategy: newer | merge | source"},
		{"SHIFTER_PRETTY", "Set to 0/false to disable pretty output"},
	}
	for _, v := range vars {
		fmt.Printf("  %-24s %s\n", v.name, v.desc)
	}
}

func nvl(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
