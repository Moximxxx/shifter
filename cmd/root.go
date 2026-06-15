// Package cmd contains the CLI command definitions for Shifter.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/port"
	"github.com/moximxxx/shifter/pkg/env"
	"github.com/moximxxx/shifter/pkg/log"
	"github.com/moximxxx/shifter/tui"
)

var rootCmd = &cobra.Command{
	Use:   "shifter",
	Short: "One-click config porting between coding agents",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if enableLog {
			log.Init()
			log.Info("cmd", "%s %v", cmd.CommandPath(), args)
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if enableLog {
			log.Info("cmd", "%s completed", cmd.CommandPath())
		}
	},
	Long: `Supported agents: Claude Code, Codex CLI, OpenCode, Gemini CLI, Qoder, Cline, Aider.

Examples:
  shifter detect                           # Scan for configured agents
  shifter port claude-code --to codex      # Port config from Claude Code to Codex
  shifter port claude-code --to codex --dry-run  # Preview changes
  shifter ui                               # Launch interactive TUI`,
	SilenceUsage: true,
}

// Set custom help template
func init() {
	cobra.AddTemplateFunc("groupStyle", func(s string) string { return s })
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🔄 Transfer
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  port        Transfer config between agents
  sync        Bidirectional sync between agents
  export      Export to canonical JSON
  import      Import from canonical JSON

💾 Profiles
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  save        Save config as reusable profile
  load        Apply saved profile to project
  profiles    List/delete saved profiles

🔧 Tools
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  detect      Scan for configured agents
  status      Show project agent status
  backup      Create/restore backups
  config      Manage Shifter settings
  env         Environment configuration

🎨 Interface
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ui          Interactive terminal wizard
  wizard      Same as ui

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Use "shifter [command] --help" for more information.
`)}

var portCmd = &cobra.Command{
	Use:   "port <source> --to <target>",
	Short: "Port configuration from one agent to another",
	Long: `Read configuration from the source agent and write it to the target agent.

The canonical model handles semantic gaps:
  • Slash commands → embedded in instructions (for agents without native commands)
  • Agents without subagent systems get agent definitions as markdown instructions
  • Loss warnings are shown for features that cannot be perfectly ported`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPort,
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Scan for configured coding agents",
	Long:  `Scan the filesystem for configured coding agents and report what was found.`,
	RunE:  runDetect,
}

var (
	portTarget    string
	portScope     string
	portDryRun    bool
	portBackup    bool
	portForce     bool
	portAspects   string
	detectJSON    bool
	projectRoot   string
	enableLog     bool
)

func init() {
	// Read defaults from environment variables
	if portTarget == "" {
		portTarget = env.Target()
	}
	if portScope == "" {
		portScope = env.Scope()
	}
	if !portDryRun {
		portDryRun = env.DryRun()
	}
	if portBackup {
		portBackup = !env.NoBackup()
	}
	if !portForce {
		portForce = env.Force()
	}
	if portAspects == "" {
		if aspects := env.Aspects(); len(aspects) > 0 {
			portAspects = strings.Join(aspects, ",")
		}
	}
	if projectRoot == "." {
		if pr := env.ProjectRoot(); pr != "." {
			projectRoot = pr
		}
	}

	portCmd.Flags().StringVar(&portTarget, "to", portTarget, "Target agent (required)")
	portCmd.Flags().StringVar(&portScope, "scope", portScope, "Scope: global, project, or all")
	portCmd.Flags().BoolVar(&portDryRun, "dry-run", false, "Preview changes without writing")
	portCmd.Flags().BoolVar(&portBackup, "backup", true, "Create backup before writing")
	portCmd.Flags().BoolVar(&portForce, "force", false, "Skip confirmation prompts")
	portCmd.Flags().StringVar(&portAspects, "aspects", "", "Comma-separated aspects to port (agents,skills,mcp,permissions,hooks,commands,instructions,settings)")
	// --to can also be set via SHIFTER_TARGET env var
	// portCmd.MarkFlagRequired("to") — handled manually in runPort

	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output as JSON")

	rootCmd.PersistentFlags().StringVar(&projectRoot, "project-root", ".", "Project root directory")
	rootCmd.PersistentFlags().BoolVar(&enableLog, "log", false, "Enable debug logging to ~/.shifter/logs/")

	rootCmd.AddCommand(portCmd)
	rootCmd.AddCommand(detectCmd)
}

// Version info
var versionInfo string

// SetVersion sets the version string (called from main with ldflags).
func SetVersion(v string) {
	versionInfo = v
	rootCmd.Version = v
	tui.SetVersion(v)
}

// VersionString returns the current version.
func VersionString() string {
	if versionInfo == "" {
		return "dev"
	}
	return versionInfo
}

// LaunchTUI starts the interactive TUI directly (no cobra).
func LaunchTUI() {
	m := tui.NewWizardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Execute runs the root command.
func Execute() {
	defer log.Close()
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		log.Error("cli", "%v", err)
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func runPort(cmd *cobra.Command, args []string) error {
	log.Info("port", "start — project=%s", projectRoot)

	var source string
	if len(args) > 0 {
		source = args[0]
	} else {
		source = env.Source()
	}
	if source == "" {
		return fmt.Errorf("source agent is required (set SHIFTER_SOURCE or pass as argument)")
	}

	// Validate target
	if portTarget == "" {
		portTarget = env.Target()
	}
	if portTarget == "" {
		return fmt.Errorf("--to flag is required (set SHIFTER_TARGET or use --to)")
	}

	// Parse aspects
	var aspects []string
	if portAspects != "" {
		for _, a := range strings.Split(portAspects, ",") {
			aspects = append(aspects, strings.TrimSpace(a))
		}
	}

	opts := port.Options{
		Source:      source,
		Target:      portTarget,
		Scope:       portScope,
		ProjectRoot: projectRoot,
		DryRun:      portDryRun,
		Backup:      portBackup,
		Force:       portForce,
		Aspects:     aspects,
	}

	ctx := context.Background()

	if portDryRun {
		diff, err := port.Preview(ctx, opts)
		if err != nil {
			return err
		}
		fmt.Println("=== DRY RUN ===")
		fmt.Printf("Source: %s → Target: %s\n\n", source, portTarget)
		if len(diff.FilesAdded) > 0 {
			fmt.Println("Files to create:")
			for _, f := range diff.FilesAdded {
				fmt.Printf("  + %s\n", f.Path)
			}
		}
		if len(diff.FilesModified) > 0 {
			fmt.Println("Files to modify:")
			for _, f := range diff.FilesModified {
				fmt.Printf("  ~ %s\n", f.Path)
			}
		}
		if len(diff.FilesDeleted) > 0 {
			fmt.Println("Files to delete:")
			for _, f := range diff.FilesDeleted {
				fmt.Printf("  - %s\n", f.Path)
			}
		}
		if len(diff.LossWarnings) > 0 {
			fmt.Println("\nLoss warnings:")
			for _, w := range diff.LossWarnings {
				fmt.Printf("  ⚠ [%s] %s: %s\n", w.Severity, w.Feature, w.Reason)
			}
		}
		if len(diff.FilesAdded)+len(diff.FilesModified)+len(diff.FilesDeleted) == 0 {
			fmt.Println("No changes needed.")
		}
		return nil
	}

	result, err := port.Port(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Port complete: %s → %s\n\n", result.Source, result.Target)

	if len(result.FilesWritten) > 0 {
		fmt.Println("📄 Files written:")
		for _, f := range result.FilesWritten {
			fmt.Printf("  ✅ %s\n", f)
		}
	}
	if len(result.FilesSkipped) > 0 {
		fmt.Println("⏭  Files skipped:")
		for _, f := range result.FilesSkipped {
			fmt.Printf("  - %s\n", f)
		}
	}
	if len(result.LossWarnings) > 0 {
		fmt.Println("\n⚠️  Loss warnings:")
		for _, w := range result.LossWarnings {
			icon := "ℹ️"
			switch w.Severity {
			case "warning":
				icon = "⚠️"
			case "critical":
				icon = "🔴"
			}
			fmt.Printf("  %s [%s] %s: %s\n", icon, w.Severity, w.Feature, w.Reason)
		}
	}

	if result.Config != nil {
		fmt.Printf("\n📊 Summary: %d agents, %d skills, %d commands, %d MCP servers, %d hooks\n",
			len(result.Config.Agents),
			len(result.Config.Skills),
			len(result.Config.Commands),
			len(result.Config.MCPServers),
			len(result.Config.Hooks),
		)
	}

	return nil
}

func runDetect(cmd *cobra.Command, args []string) error {
	log.Info("detect", "scanning")
	results := detect.ScanAll()

	if detectJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	fmt.Println("🔍 Scanning for configured coding agents...")
	fmt.Println()
	configured := 0
	for _, r := range results {
		if r.Found {
			configured++
			parts := []string{"✅", r.Name}
			details := []string{}
			for k, v := range r.Summary {
				details = append(details, fmt.Sprintf("%d %s", v, k))
			}
			if len(details) > 0 {
				parts = append(parts, strings.Join(details, ", "))
			}
			fmt.Printf("  %s\n", strings.Join(parts, "  "))
		} else {
			fmt.Printf("  ╳  %-16s (not configured)\n", r.Name)
		}
	}

	fmt.Printf("\n  %d of %d agents configured.\n", configured, len(results))
	fmt.Println("\n💡 Use 'shifter port <source> --to <target>' to port configs.")
	return nil
}
