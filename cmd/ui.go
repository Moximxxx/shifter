package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/tui"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the interactive terminal UI",
	Long: `Launch the interactive Shifter TUI for one-click config porting.

The TUI guides you through:
  1. Select source agent
  2. Select target agent
  3. Choose what to port (agents, skills, MCP, etc.)
  4. Preview and apply changes`,
	RunE: runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	m := tui.NewModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		return err
	}

	return nil
}
