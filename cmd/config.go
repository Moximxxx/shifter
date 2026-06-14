package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/pkg/i18n"
	"github.com/moximxxx/shifter/pkg/settings"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Shifter settings",
	Long: `View and modify Shifter user settings.

Settings are stored in ~/.shifter/settings.json

Examples:
  shifter config              # Show current settings
  shifter config set lang zh  # Switch to Chinese
  shifter config set lang en  # Switch to English`,
	Args: cobra.MaximumNArgs(4),
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	s, err := settings.Load()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// Show current settings
		fmt.Println(i18n.T("config.show") + ":")
		fmt.Printf("  lang       = %s", s.Lang)
		switch s.Lang {
		case "zh":
			fmt.Println(" (简体中文)")
		case "en":
			fmt.Println(" (English)")
		default:
			fmt.Println()
		}
		fmt.Printf("  first_run  = %v\n", s.FirstRun)
		fmt.Println()
		fmt.Println(i18n.T("config.keys"))
		fmt.Println()
		fmt.Println("Use 'shifter config set <key> <value>' to change a setting.")
		return nil
	}

	if args[0] != "set" || len(args) < 3 {
		return fmt.Errorf("usage: shifter config set <key> <value>")
	}

	key := args[1]
	value := args[2]

	switch strings.ToLower(key) {
	case "lang", "language":
		value = strings.ToLower(value)
		if value != "en" && value != "zh" {
			return fmt.Errorf("unsupported language: %q (supported: en, zh)", value)
		}
		s.Lang = value
		i18n.SetLang(value)
		fmt.Println(i18n.Tf("settings.language_changed", map[string]string{"lang": value}))
		fmt.Println(i18n.T("settings.restart_hint"))

	default:
		return fmt.Errorf("unknown setting: %q (available: lang)", key)
	}

	s.FirstRun = false
	return settings.Save(s)
}
