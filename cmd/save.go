package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/engine/profile"
	"github.com/moximxxx/shifter/registry"
)

var (
	saveName        string
	saveDescription string
	saveSource      string
)

var saveCmd = &cobra.Command{
	Use:   "save <profile-name>",
	Short: "Save current project's agent config as a global profile",
	Long: `Capture the configuration of a coding agent in the current project
and save it as a named global profile. The profile can later be
loaded in another project with 'shifter load'.

Examples:
  shifter save my-workflow --source claude-code
  shifter save team-setup --source codex --desc "Team standard config"`,
	Args: cobra.ExactArgs(1),
	RunE: runSave,
}

var loadCmd = &cobra.Command{
	Use:   "load <profile-name>",
	Short: "Load a saved profile into the current project",
	Long: `Apply a previously saved profile to an agent in the current project.

Examples:
  shifter load my-workflow --to codex
  shifter load team-setup --to opencode`,
	Args: cobra.ExactArgs(1),
	RunE: runLoad,
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List saved profiles",
	Long:  `List all saved global configuration profiles.`,
	RunE:  runProfiles,
}

var profilesDeleteCmd = &cobra.Command{
	Use:   "delete <profile-name>",
	Short: "Delete a saved profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfilesDelete,
}

var (
	loadTarget string
)

func init() {
	saveCmd.Flags().StringVar(&saveSource, "source", "", "Source agent (e.g., claude-code, codex)")
	saveCmd.Flags().StringVar(&saveDescription, "desc", "", "Profile description")
	saveCmd.MarkFlagRequired("source")

	loadCmd.Flags().StringVar(&loadTarget, "to", "", "Target agent to apply the profile to")
	loadCmd.MarkFlagRequired("to")

	profilesCmd.AddCommand(profilesDeleteCmd)

	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(loadCmd)
	rootCmd.AddCommand(profilesCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
	name := args[0]

	a, err := registry.Get(saveSource)
	if err != nil {
		return fmt.Errorf("unknown agent %q (available: %v)", saveSource, registry.ListIDs())
	}

	ctx := context.Background()
	cfg, err := a.Read(ctx, adapter.ReadOptions{
		Scope:       "project",
		ProjectRoot: projectRoot,
	})
	if err != nil {
		return fmt.Errorf("read %s config: %w", saveSource, err)
	}

	p := profile.Profile{
		Name:        name,
		Description: saveDescription,
		SourceAgent: saveSource,
		Config:      cfg,
		CreatedAt:   time.Now(),
	}

	if err := profile.Save(p); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}

	fmt.Printf("✓ Profile %q saved (%s)\n", name, p.Summary())
	fmt.Printf("  Source: %s\n", a.Name())
	if len(cfg.Agents) > 0 {
		fmt.Printf("  Agents: %d\n", len(cfg.Agents))
	}
	if len(cfg.Skills) > 0 {
		fmt.Printf("  Skills: %d\n", len(cfg.Skills))
	}
	if len(cfg.MCPServers) > 0 {
		fmt.Printf("  MCP Servers: %d\n", len(cfg.MCPServers))
	}
	fmt.Println("\nUse 'shifter load " + name + " --to <agent>' to apply in another project.")

	return nil
}

func runLoad(cmd *cobra.Command, args []string) error {
	name := args[0]

	p, err := profile.Load(name)
	if err != nil {
		return err
	}

	if p.Config == nil {
		return fmt.Errorf("profile %q has no configuration data", name)
	}

	tgtAdapter, err := registry.Get(loadTarget)
	if err != nil {
		return fmt.Errorf("unknown target agent %q (available: %v)", loadTarget, registry.ListIDs())
	}

	ctx := context.Background()
	result, err := tgtAdapter.Write(ctx, p.Config, adapter.WriteOptions{
		Scope:       "project",
		ProjectRoot: projectRoot,
		Backup:      true,
	})
	if err != nil {
		return fmt.Errorf("apply profile: %w", err)
	}

	fmt.Printf("✓ Profile %q applied to %s\n\n", name, tgtAdapter.Name())
	fmt.Printf("Source: %s (originally from %s)\n", p.Name, p.SourceAgent)
	fmt.Println("Files written:")
	for _, f := range result.FilesWritten {
		fmt.Printf("  ✓ %s\n", f)
	}
	if len(result.LossWarnings) > 0 {
		fmt.Println("\nLoss warnings:")
		for _, w := range result.LossWarnings {
			fmt.Printf("  ⚠ [%s] %s: %s\n", w.Severity, w.Feature, w.Reason)
		}
	}

	return nil
}

func runProfiles(cmd *cobra.Command, args []string) error {
	profiles, err := profile.List()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		fmt.Println("No saved profiles found.")
		fmt.Println("\nUse 'shifter save <name> --source <agent>' to create one.")
		return nil
	}

	fmt.Printf("Saved profiles (%d):\n\n", len(profiles))
	for _, p := range profiles {
		fmt.Printf("  📁 %s\n", p.Name)
		fmt.Printf("     Source: %s\n", p.SourceAgent)
		fmt.Printf("     Contents: %s\n", p.Summary())
		if p.Description != "" {
			fmt.Printf("     Description: %s\n", p.Description)
		}
		fmt.Printf("     Updated: %s\n", p.UpdatedAt.Format("2006-01-02 15:04"))
		fmt.Println()
	}

	fmt.Println("Use 'shifter load <name> --to <agent>' to apply a profile.")
	fmt.Println("Use 'shifter profiles delete <name>' to remove a profile.")

	return nil
}

func runProfilesDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := profile.Delete(name); err != nil {
		return err
	}

	fmt.Printf("✓ Profile %q deleted.\n", name)
	return nil
}
