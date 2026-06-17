package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/flowhub"
	"github.com/moximxxx/shifter/registry"
	"github.com/moximxxx/shifter/tui"
)

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "FlowHub — workflow marketplace",
	Long:  `Search, install, and publish workflows on FlowHub (GitHub-backed, zero-cost).`,
}

var flowSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search FlowHub for workflows",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runFlowSearch,
}

var flowInstallCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a workflow from FlowHub",
	Args:  cobra.ExactArgs(1),
	RunE:  runFlowInstall,
}

var flowPublishCmd = &cobra.Command{
	Use:   "publish <name>",
	Short: "Publish current workflow to FlowHub",
	Args:  cobra.ExactArgs(1),
	RunE:  runFlowPublish,
}

var flowListCmd = &cobra.Command{
	Use:   "list",
	Short: "Browse FlowHub (TUI)",
	RunE:  runFlowList,
}

var flowTrendingCmd = &cobra.Command{
	Use:   "trending",
	Short: "Show trending workflows on FlowHub",
	RunE:  runFlowTrending,
}

var (
	flowTags        string
	flowSource      string
	flowDesc        string
	flowInstallTo   string
	flowSubmit      bool
)

func init() {
	flowPublishCmd.Flags().StringVar(&flowTags, "tags", "", "Comma-separated tags")
	flowPublishCmd.Flags().StringVar(&flowSource, "source", "", "Source agent (auto-detect if omitted)")
	flowPublishCmd.Flags().StringVar(&flowDesc, "desc", "", "Description")
	flowPublishCmd.Flags().BoolVar(&flowSubmit, "submit", false, "Auto-submit PR via GitHub API (needs GITHUB_TOKEN)")
	flowInstallCmd.Flags().StringVar(&flowInstallTo, "to", "", "Target agent to apply the workflow to")

	flowCmd.AddCommand(flowSearchCmd, flowInstallCmd, flowPublishCmd, flowListCmd, flowTrendingCmd)
	rootCmd.AddCommand(flowCmd)
}

func runFlowSearch(cmd *cobra.Command, args []string) error {
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	results, err := flowhub.Search(query)
	if err != nil {
		return fmt.Errorf("search failed: %w\n  FlowHub may not be available yet. Visit %s", err, flowhub.RepoURL)
	}

	if len(results) == 0 {
		fmt.Printf("No workflows found for %q\n", query)
		fmt.Printf("\nVisit %s to browse all workflows.\n", flowhub.RepoURL)
		return nil
	}

	fmt.Printf("FlowHub — %d workflow(s)\n\n", len(results))
	for _, w := range results {
		fmt.Printf("  📦 %s", w.Name)
		if w.Version != "" {
			fmt.Printf(" v%s", w.Version)
		}
		fmt.Printf("  ⭐%d\n", w.Downloads)
		if w.Description != "" {
			fmt.Printf("      %s\n", w.Description)
		}
		fmt.Printf("      by %s  •  agent: %s", w.Author, w.Agent)
		if len(w.Tags) > 0 {
			fmt.Printf("  •  %s", strings.Join(w.Tags, ", "))
		}
		fmt.Println()
	}

	fmt.Println("Install: shifter flow install <name>")
	return nil
}

func runFlowInstall(cmd *cobra.Command, args []string) error {
	name := args[0]
	fmt.Printf("↓ Downloading %s from FlowHub...\n", name)

	data, err := flowhub.Download(name)
	if err != nil {
		return err
	}

	var cfg canonical.ShifterConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse workflow: %w", err)
	}

	target := flowInstallTo
	if target == "" {
		// Try to auto-detect which agent is configured here
		results := detect.ScanAll()
		for _, r := range results {
			if r.Found && r.HasProjectConfig {
				target = r.ID
				break
			}
		}
	}
	if target == "" {
		return fmt.Errorf("no target agent found; use --to <agent> to specify")
	}

	tgtAdapter, err := registry.Get(target)
	if err != nil {
		return err
	}

	result, err := tgtAdapter.Write(cmd.Context(), &cfg, adapter.WriteOptions{
		Scope:       "project",
		ProjectRoot: projectRoot,
		Backup:      true,
	})
	if err != nil {
		return fmt.Errorf("apply workflow: %w", err)
	}

	fmt.Printf("✓ Installed %s → %s\n", name, target)
	for _, f := range result.FilesWritten {
		fmt.Printf("  ✓ %s\n", f)
	}
	if len(result.LossWarnings) > 0 {
		fmt.Println("\nLoss warnings:")
		for _, w := range result.LossWarnings {
			fmt.Printf("  ⚠ %s\n", w.Reason)
		}
	}
	return nil
}

func runFlowPublish(cmd *cobra.Command, args []string) error {
	name := args[0]
	source := flowSource
	if source == "" {
		results := detect.ScanAll()
		for _, r := range results {
			if r.Found && r.HasProjectConfig {
				source = r.ID
				break
			}
		}
	}
	if source == "" {
		return fmt.Errorf("no configured agent found; use --source <agent>")
	}

	a, err := registry.Get(source)
	if err != nil {
		return err
	}

	cfg, err := a.Read(cmd.Context(), adapter.ReadOptions{
		Scope:       "project",
		ProjectRoot: projectRoot,
	})
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	// Generate canonical JSON for publishing
	workflowJSON, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	var tags []string
	if flowTags != "" {
		for _, t := range strings.Split(flowTags, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	// Generate metadata
	meta := flowhub.GenerateMetadata(name, source, flowDesc, tags)
	metaJSON, _ := json.MarshalIndent(meta, "", "  ")

	// Auto-generate README
	readme := fmt.Sprintf(`# %s

%s

## Agent
%s

## Tags
%s

## Install
`+"`"+`sh
shifter flow install %s
`+"`"+`

## Contents
- Agents: %d
- Skills: %d
- MCP Servers: %d
- Hooks: %d
`, name, flowDesc, source, strings.Join(tags, ", "), name,
		len(cfg.Agents), len(cfg.Skills), len(cfg.MCPServers), len(cfg.Hooks))

	// Save locally for manual PR submission
	workflowDir := fmt.Sprintf("flowhub-publish/%s", name)
	os.MkdirAll(workflowDir, 0755)
	os.WriteFile(workflowDir+"/workflow.shifter.json", workflowJSON, 0644)
	os.WriteFile(workflowDir+"/metadata.json", metaJSON, 0644)
	os.WriteFile(workflowDir+"/README.md", []byte(readme), 0644)

	// Generate one-click PR URL (GitHub new-file form)
	prURL := fmt.Sprintf("%s/new/main/workflows/%s", flowhub.RepoURL, name)

	fmt.Printf("✓ Workflow %q ready to publish\n\n", name)
	fmt.Printf("📁 Files prepared in %s/:\n", workflowDir)
	fmt.Printf("  ✓ workflow.shifter.json (%d bytes)\n", len(workflowJSON))
	fmt.Printf("  ✓ metadata.json (%d bytes)\n", len(metaJSON))
	fmt.Printf("  ✓ README.md (%d bytes)\n", len(readme))

	// Auto-submit via GitHub API if --submit flag is set
	if flowSubmit {
		fmt.Printf("\n🚀 Auto-submitting via GitHub API...\n")
		files := map[string][]byte{
			"workflow.shifter.json": workflowJSON,
			"metadata.json":         metaJSON,
			"README.md":             []byte(readme),
		}
		prURL, err := flowhub.Publish(flowhub.PublishRequest{
			Name:    name,
			Files:   files,
			Message: fmt.Sprintf("Add workflow: %s\n\n%s", name, flowDesc),
		})
		if err != nil {
			fmt.Printf("⚠ Auto-submit failed: %v\n", err)
			fmt.Println("\nFalling back to manual publish:")
		} else {
			fmt.Printf("✓ PR created: %s\n\n", prURL)
			fmt.Printf("After merge, install with: shifter flow install %s\n", name)
			return nil
		}
	}

	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🚀 Quick Publish:\n\n")
	fmt.Printf("  Set GITHUB_TOKEN and use --submit to auto-publish:\n")
	fmt.Printf("  export GITHUB_TOKEN=ghp_xxxx\n")
	fmt.Printf("  shifter flow publish %s --submit\n\n", name)
	fmt.Printf("  Or manually:\n")
	fmt.Printf("  1. Fork:  %s/fork\n", flowhub.RepoURL)
	fmt.Printf("  2. Upload: %s\n", prURL)
	fmt.Printf("  3. Create PR: %s/compare\n", flowhub.RepoURL)
	fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\n💡 After merge, install with:\n")
	fmt.Printf("   shifter flow install %s\n", name)

	return nil
}

func runFlowList(cmd *cobra.Command, args []string) error {
	return tui.StandaloneFlowHub()
}

func runFlowTrending(cmd *cobra.Command, args []string) error {
	results, err := flowhub.Search("")
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	if len(results) == 0 {
		fmt.Println("No workflows yet.")
		return nil
	}
	// Sort by downloads descending (already done in Search)
	fmt.Printf("🔥 Trending on FlowHub\n\n")
	for i, w := range results {
		if i >= 10 {
			break
		}
		fmt.Printf("  %2d. %s v%s  ⭐%d  %s\n", i+1, w.Name, w.Version, w.Downloads, w.Agent)
	}
	return nil
}
