package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/registry"
)

var (
	exportOutput string
	exportPretty bool
)

var exportCmd = &cobra.Command{
	Use:   "export <agent-id>",
	Short: "Export agent config to Shifter canonical format (JSON)",
	Long: `Read a coding agent's configuration and export it to Shifter's
universal canonical JSON format. This format can be:
  - Saved and version-controlled
  - Manually edited
  - Imported to any other agent with 'shifter import'
  - Shared across teams

Examples:
  shifter export claude-code
  shifter export claude-code -o my-workflow.shifter.json
  shifter export opencode --pretty > team-standard.shifter.json`,
	Args: cobra.ExactArgs(1),
	RunE: runExport,
}

var importCmd = &cobra.Command{
	Use:   "import <file.shifter.json> --to <agent-id>",
	Short: "Import Shifter canonical JSON and apply to a target agent",
	Long: `Read a Shifter canonical JSON file and write it to the target
agent's native configuration format.

This is the reverse of 'shifter export' — it takes the universal
format and reformats it for a specific coding agent.

Examples:
  shifter import my-workflow.shifter.json --to codex
  shifter import team-standard.shifter.json --to opencode
  cat workflow.shifter.json | shifter import - --to claude-code`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

var importTarget string

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().BoolVar(&exportPretty, "pretty", true, "Pretty-print JSON output")

	importCmd.Flags().StringVar(&importTarget, "to", "", "Target agent to apply the config to")
	importCmd.MarkFlagRequired("to")

	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	agentID := args[0]

	a, err := registry.Get(agentID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cfg, err := a.Read(ctx, adapter.ReadOptions{
		Scope:       "project",
		ProjectRoot: projectRoot,
	})
	if err != nil {
		return fmt.Errorf("read %s: %w", agentID, err)
	}

	// Strip internal fields for clean export
	cfg.LossWarnings = nil

	var output []byte
	if exportPretty {
		output, err = json.MarshalIndent(cfg, "", "  ")
	} else {
		output, err = json.Marshal(cfg)
	}
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, output, 0644); err != nil {
			return fmt.Errorf("write: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Exported %s config to %s\n", agentID, exportOutput)
		fmt.Fprintf(cmd.OutOrStdout(), "  Agents: %d, Skills: %d, Commands: %d, MCP: %d, Hooks: %d\n",
			len(cfg.Agents), len(cfg.Skills), len(cfg.Commands),
			len(cfg.MCPServers), len(cfg.Hooks))
	} else {
		cmd.OutOrStdout().Write(output)
		cmd.OutOrStdout().Write([]byte("\n"))
	}

	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	var data []byte
	var err error

	if filePath == "-" {
		// Read from stdin
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(filePath)
	}
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	var cfg canonical.ShifterConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse canonical JSON: %w", err)
	}

	tgtAdapter, err := registry.Get(importTarget)
	if err != nil {
		return err
	}

	ctx := context.Background()
	result, err := tgtAdapter.Write(ctx, &cfg, adapter.WriteOptions{
		Scope:       "project",
		ProjectRoot: projectRoot,
		Backup:      true,
	})
	if err != nil {
		return fmt.Errorf("write %s: %w", importTarget, err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Imported to %s\n", importTarget)
	for _, f := range result.FilesWritten {
		fmt.Fprintf(cmd.OutOrStdout(), "  ✓ %s\n", f)
	}
	if len(result.LossWarnings) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\nLoss warnings:\n")
		for _, w := range result.LossWarnings {
			fmt.Fprintf(cmd.OutOrStdout(), "  ⚠ [%s] %s: %s\n", w.Severity, w.Feature, w.Reason)
		}
	}

	return nil
}
