package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/engine/backup"
)

var (
	backupRestore string
	backupList    bool
)

var backupCmd = &cobra.Command{
	Use:   "backup [agent-id]",
	Short: "Manage backups of agent configurations",
	Long: `Create and restore backups of coding agent configurations.

By default, backup creates a timestamped archive of all config files for
the specified agent. Use --restore to restore from a backup.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBackup,
}

func init() {
	backupCmd.Flags().StringVar(&backupRestore, "restore", "", "Restore from a backup archive path")
	backupCmd.Flags().BoolVarP(&backupList, "list", "l", false, "List all backups")
	rootCmd.AddCommand(backupCmd)
}

func runBackup(cmd *cobra.Command, args []string) error {
	if backupList {
		backups, err := backup.List()
		if err != nil {
			return err
		}
		if len(backups) == 0 {
			fmt.Println("No backups found.")
			return nil
		}
		fmt.Println("Backups:")
		for _, b := range backups {
			info, _ := os.Stat(b)
			size := ""
			if info != nil {
				size = fmt.Sprintf(" (%d bytes)", info.Size())
			}
			fmt.Printf("  %s%s\n", filepath.Base(b), size)
		}
		return nil
	}

	if backupRestore != "" {
		fmt.Printf("Restoring from: %s\n", backupRestore)
		if err := backup.Restore(backupRestore); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}
		fmt.Println("✓ Restore complete.")
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("specify an agent-id to backup, or use --list / --restore")
	}

	agentID := args[0]
	fmt.Printf("Creating backup for %s...\n", agentID)

	// Collect files to backup based on agent
	var files []string
	switch agentID {
	case "claude-code":
		files = []string{".claude"}
	case "codex":
		files = []string{".codex"}
	case "opencode":
		files = []string{".opencode", "opencode.json"}
	case "gemini-cli":
		files = []string{".gemini"}
	case "qoder":
		files = []string{".qoder"}
	case "cline":
		files = []string{".clinerules"}
	case "aider":
		files = []string{".aider.conf.yml"}
	default:
		return fmt.Errorf("unknown agent: %s", agentID)
	}

	archivePath, err := backup.Create(files)
	if err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("✓ Backup created: %s\n", archivePath)
	return nil
}
