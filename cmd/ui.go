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
	Long: `Launch the interactive Shifter wizard for managing coding agent configs.

The wizard guides you through:
  💾 Save — Capture a project's agent config as a reusable global profile
  📥 Load — Apply a saved profile to another project's agent
  🔀 Port — Transfer config between agents in the current project

Profiles are stored in ~/.shifter/profiles/ and can be applied in any directory.`,
	RunE: runUI,
}

var wizardCmd = &cobra.Command{
	Use:   "wizard",
	Short: "Interactive config management wizard",
	Long:  `Same as 'shifter ui' — launches the interactive wizard.`,
	RunE:  runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
	rootCmd.AddCommand(wizardCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	m := tui.NewWizardModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running wizard: %v\n", err)
		return err
	}

	return nil
}
