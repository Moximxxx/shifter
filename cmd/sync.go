package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/engine/sync"
)

var (
	syncStrategy string
	syncScope    string
	syncDryRun   bool
	syncBackup   bool
	syncForce    bool
)

var syncCmd = &cobra.Command{
	Use:   "sync <agent-a> <agent-b>",
	Short: "Bidirectional sync between two agents",
	Long: `Synchronize configurations bidirectionally between two coding agents.

Reads config from both agents, merges them, and writes the unified config
back to both agents. Conflicts are resolved using the specified strategy.

Strategies:
  newer   - For each field, use the version with the most recent modification (default)
  merge   - Union of lists, newer-wins for scalars
  source  - Prefer the first agent (agent-a) on all conflicts`,
	Args: cobra.ExactArgs(2),
	RunE: runSync,
}

func init() {
	syncCmd.Flags().StringVar(&syncStrategy, "strategy", "newer", "Merge strategy: newer, merge, or source")
	syncCmd.Flags().StringVar(&syncScope, "scope", "project", "Scope: global or project")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Preview changes without writing")
	syncCmd.Flags().BoolVar(&syncBackup, "backup", true, "Create backup before writing")
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Skip confirmation prompts")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	agentA := args[0]
	agentB := args[1]

	var strategy canonical.MergeStrategy
	switch strings.ToLower(syncStrategy) {
	case "newer":
		strategy = canonical.MergeNewer
	case "merge":
		strategy = canonical.MergeUnion
	case "source":
		strategy = canonical.MergePreferSource
	default:
		return fmt.Errorf("unknown strategy: %s (use: newer, merge, source)", syncStrategy)
	}

	opts := sync.Options{
		AgentA:      agentA,
		AgentB:      agentB,
		Strategy:    strategy,
		Scope:       syncScope,
		ProjectRoot: projectRoot,
		DryRun:      syncDryRun,
		Backup:      syncBackup,
		Force:       syncForce,
	}

	ctx := context.Background()
	result, err := sync.Sync(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Sync complete: %s ↔ %s (strategy: %s)\n\n", result.AgentA, result.AgentB, syncStrategy)

	if len(result.FilesWritten) > 0 {
		fmt.Println("Files synced:")
		for _, f := range result.FilesWritten {
			fmt.Printf("  ✓ %s\n", f)
		}
	}

	if len(result.Conflicts) > 0 {
		fmt.Println("\nConflicts resolved:")
		for _, c := range result.Conflicts {
			fmt.Printf("  ⚡ %s: %s → resolved as %q\n", c.Field, c.Resolved, "side "+c.Resolved)
		}
	}

	if len(result.LossWarnings) > 0 {
		fmt.Println("\nLoss warnings:")
		for _, w := range result.LossWarnings {
			fmt.Printf("  ⚠ [%s] %s: %s\n", w.Severity, w.Feature, w.Reason)
		}
	}

	if result.MergedConfig != nil {
		fmt.Printf("\nMerged: %d agents, %d skills, %d commands, %d MCP servers, %d hooks\n",
			len(result.MergedConfig.Agents),
			len(result.MergedConfig.Skills),
			len(result.MergedConfig.Commands),
			len(result.MergedConfig.MCPServers),
			len(result.MergedConfig.Hooks),
		)
	}

	return nil
}
