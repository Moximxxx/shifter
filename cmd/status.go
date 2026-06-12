package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/engine/detect"
)

var (
	statusJSON     bool
	statusExitCode bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configuration status for the current project",
	Long: `Scan and display the configuration status of all coding agents
in the current project. Shows which agents are configured and their
configuration state (agents, skills, MCP servers, etc.).`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	statusCmd.Flags().BoolVar(&statusExitCode, "exit-code", false, "Exit with code 1 if any agent has uncommitted config changes")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	results := detect.ScanAll()

	if statusJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	// Get project info
	projectName := filepath.Base(projectRoot)
	if projectRoot == "." {
		if abs, err := filepath.Abs(projectRoot); err == nil {
			projectName = filepath.Base(abs)
		}
	}

	fmt.Printf("Project: %s\n\n", projectName)
	fmt.Println("Coding Agent Status:")
	fmt.Println(strings.Repeat("─", 60))

	configuredCount := 0
	for _, r := range results {
		if r.Found {
			configuredCount++
			fmt.Printf("  ✓ %-16s", r.Name)
			parts := []string{}
			for k, v := range r.Summary {
				parts = append(parts, fmt.Sprintf("%d %s", v, k))
			}
			fmt.Println(strings.Join(parts, ", "))
		} else {
			fmt.Printf("  ✗ %-16s (not configured)\n", r.Name)
		}
	}

	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("\n%d of %d agents configured.\n", configuredCount, len(results))

	if configuredCount > 0 {
		fmt.Println("\nUse 'shifter port <source> --to <target>' to port configurations.")
	}

	if statusExitCode && configuredCount == 0 {
		os.Exit(1)
	}

	return nil
}
