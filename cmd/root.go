// Package cmd contains the CLI command definitions for Shifter.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/engine/port"
	"github.com/moximxxx/shifter/registry"
)

var rootCmd = &cobra.Command{
	Use:   "shifter",
	Short: "Shifter — one-click config porting between coding agents",
	Long: `Shifter lets you port configurations (agents, skills, MCP servers,
permissions, hooks, and more) between different coding agents.

Supported agents: Claude Code, Codex CLI, OpenCode, Gemini CLI, Qoder, Cline, Aider.

Examples:
  shifter detect                           # Scan for configured agents
  shifter port claude-code --to codex      # Port config from Claude Code to Codex
  shifter port claude-code --to codex --dry-run  # Preview changes
  shifter ui                               # Launch interactive TUI`,
	SilenceUsage: true,
}

var portCmd = &cobra.Command{
	Use:   "port <source> --to <target>",
	Short: "Port configuration from one agent to another",
	Long: `Read configuration from the source agent and write it to the target agent.

The canonical model handles semantic gaps:
  • Slash commands → embedded in instructions (for agents without native commands)
  • Agents without subagent systems get agent definitions as markdown instructions
  • Loss warnings are shown for features that cannot be perfectly ported`,
	Args: cobra.ExactArgs(1),
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
)

func init() {
	portCmd.Flags().StringVar(&portTarget, "to", "", "Target agent (required)")
	portCmd.Flags().StringVar(&portScope, "scope", "project", "Scope: global, project, or all")
	portCmd.Flags().BoolVar(&portDryRun, "dry-run", false, "Preview changes without writing")
	portCmd.Flags().BoolVar(&portBackup, "backup", true, "Create backup before writing")
	portCmd.Flags().BoolVar(&portForce, "force", false, "Skip confirmation prompts")
	portCmd.Flags().StringVar(&portAspects, "aspects", "", "Comma-separated aspects to port (agents,skills,mcp,permissions,hooks,commands,instructions,settings)")
	portCmd.MarkFlagRequired("to")

	detectCmd.Flags().BoolVar(&detectJSON, "json", false, "Output as JSON")

	rootCmd.PersistentFlags().StringVar(&projectRoot, "project-root", ".", "Project root directory")

	rootCmd.AddCommand(portCmd)
	rootCmd.AddCommand(detectCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runPort(cmd *cobra.Command, args []string) error {
	source := args[0]

	// Validate target
	if portTarget == "" {
		return fmt.Errorf("--to flag is required")
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

	fmt.Printf("✓ Port complete: %s → %s\n\n", result.Source, result.Target)

	if len(result.FilesWritten) > 0 {
		fmt.Println("Files written:")
		for _, f := range result.FilesWritten {
			fmt.Printf("  ✓ %s\n", f)
		}
	}
	if len(result.FilesSkipped) > 0 {
		fmt.Println("Files skipped:")
		for _, f := range result.FilesSkipped {
			fmt.Printf("  - %s\n", f)
		}
	}
	if len(result.LossWarnings) > 0 {
		fmt.Println("\nLoss warnings:")
		for _, w := range result.LossWarnings {
			fmt.Printf("  ⚠ [%s] %s: %s\n", w.Severity, w.Feature, w.Reason)
		}
	}

	if result.Config != nil {
		fmt.Printf("\nSummary: %d agents, %d skills, %d commands, %d MCP servers, %d hooks\n",
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
	ids := registry.ListIDs()

	if detectJSON {
		type detectOutput struct {
			ID      string                  `json:"id"`
			Name    string                  `json:"name"`
			Found   bool                    `json:"found"`
			Summary map[string]int          `json:"summary,omitempty"`
			Paths   []string                `json:"paths,omitempty"`
		}
		var results []detectOutput
		for _, id := range ids {
			a, err := registry.Get(id)
			if err != nil {
				continue
			}
			dr, err := a.Detect()
			if err != nil {
				continue
			}
			out := detectOutput{
				ID:    id,
				Name:  a.Name(),
				Found: dr.Found,
				Summary: dr.Summary,
			}
			out.Paths = append(out.Paths, dr.GlobalPaths...)
			if dr.ProjectPath != "" {
				out.Paths = append(out.Paths, dr.ProjectPath)
			}
			results = append(results, out)
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	fmt.Println("Scanning for configured coding agents...")
	fmt.Println()
	for _, id := range ids {
		a, err := registry.Get(id)
		if err != nil {
			continue
		}
		dr, err := a.Detect()
		if err != nil {
			continue
		}

		status := "✗ not configured"
		if dr.Found {
			parts := []string{"✓"}
			parts = append(parts, a.Name())
			for k, v := range dr.Summary {
				parts = append(parts, fmt.Sprintf("%d %s", v, k))
			}
			status = strings.Join(parts, "  ")
		} else {
			status = fmt.Sprintf("✗ %s (not configured)", a.Name())
		}
		fmt.Printf("  %s\n", status)
	}

	fmt.Println("\nUse 'shifter port <source> --to <target>' to port configs.")
	return nil
}
