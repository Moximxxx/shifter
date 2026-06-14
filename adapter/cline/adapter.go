package cline

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/convert"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string   { return "cline" }
func (a *Adapter) Name() string { return "Cline" }

func (a *Adapter) SearchPaths() []string {
	return []string{".clinerules", ".clinerules/"}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		Notes: map[string]string{
			"agents":      "Cline has no subagent system",
			"skills":      "Cline has no skill system",
			"commands":    "Cline has only /newrule",
			"mcp":         "MCP configured via VS Code extension, not files",
			"permissions": "Permissions via VS Code extension settings",
			"hooks":       "Cline has no hook system",
		},
	}
}

func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}

	// Check for .clinerules file
	if _, err := os.Stat(".clinerules"); err == nil {
		result.Found = true
		result.ProjectPath = ".clinerules"
		result.Summary["rules_file"] = 1
	}

	// Check for .clinerules/ directory
	if info, err := os.Stat(".clinerules"); err == nil && info.IsDir() {
		entries, err := os.ReadDir(".clinerules")
		if err == nil {
			count := 0
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".md") {
					count++
				}
			}
			if count > 0 {
				result.Found = true
				result.ProjectPath = ".clinerules/"
				result.Summary["rules_files"] = count
			}
		}
	}

	return result, nil
}

func (a *Adapter) Read(ctx context.Context, opts adapter.ReadOptions) (*canonical.ShifterConfig, error) {
	cfg := &canonical.ShifterConfig{
		Meta: canonical.ConfigMeta{
			SourceAdapter: a.ID(),
			Version:       "1.0.0",
		},
	}

	// Read .clinerules (single file)
	singlePath := filepath.Join(opts.ProjectRoot, ".clinerules")
	if info, err := os.Stat(singlePath); err == nil && !info.IsDir() {
		data, err := os.ReadFile(singlePath)
		if err == nil {
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    ".clinerules",
				Content: string(data),
				Scope:   "root",
			})
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, singlePath)
		}
	}

	// Read .clinerules/ directory
	rulesDir := filepath.Join(opts.ProjectRoot, ".clinerules")
	if info, err := os.Stat(rulesDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(rulesDir)
		if err == nil {
			for _, entry := range entries {
				if !strings.HasSuffix(entry.Name(), ".md") {
					continue
				}
				rulePath := filepath.Join(rulesDir, entry.Name())
				data, err := os.ReadFile(rulePath)
				if err == nil {
					cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
						Path:    filepath.Join(".clinerules", entry.Name()),
						Content: string(data),
						Scope:   "named",
					})
					cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, rulePath)
				}
			}
		}
	}

	return cfg, nil
}

func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	// Collect non-root instructions into .clinerules/ directory
	// Root instruction → .clinerules (single file)
	var namedInstructions []canonical.Instruction
	var rootInstruction *canonical.Instruction

	for _, inst := range cfg.Instructions {
		if inst.Scope == "root" {
			rootInstruction = &inst
		} else {
			namedInstructions = append(namedInstructions, inst)
		}
	}

	// If we have named instructions, use directory format
	if len(namedInstructions) > 0 {
		rulesDir := filepath.Join(opts.ProjectRoot, ".clinerules")
		if !opts.DryRun {
			os.MkdirAll(rulesDir, 0755)
		}

		// Write root instruction as general.md
		if rootInstruction != nil {
			p := filepath.Join(rulesDir, "general.md")
			if !opts.DryRun {
				os.WriteFile(p, []byte(rootInstruction.Content), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, p)
		}

		for _, inst := range namedInstructions {
			name := strings.TrimSuffix(filepath.Base(inst.Path), ".md") + ".md"
			p := filepath.Join(rulesDir, name)
			if !opts.DryRun {
				os.WriteFile(p, []byte(inst.Content), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, p)
		}
	} else if rootInstruction != nil {
		// Single .clinerules file
		p := filepath.Join(opts.ProjectRoot, ".clinerules")
		if !opts.DryRun {
			os.WriteFile(p, []byte(rootInstruction.Content), 0644)
		}
		result.FilesWritten = append(result.FilesWritten, p)
	}

	// Agents → embedded as rules (lossy)
	if len(cfg.Agents) > 0 {
		var b strings.Builder
		b.WriteString("## Imported Agents\n\n")
		for _, agent := range cfg.Agents {
			b.WriteString("### " + agent.Name + "\n\n")
			b.WriteString(agent.SystemPrompt + "\n\n")
		}
		rulesDir := filepath.Join(opts.ProjectRoot, ".clinerules")
		p := filepath.Join(rulesDir, "imported-agents.md")
		if !opts.DryRun {
			os.MkdirAll(rulesDir, 0755)
			os.WriteFile(p, []byte(b.String()), 0644)
		}
		result.FilesWritten = append(result.FilesWritten, p)
		result.LossWarnings = append(result.LossWarnings, convert.NewLossWarning("agents", cfg.Meta.SourceAdapter, a.ID(), "Cline has no subagent system; agent definitions saved as .clinerules/imported-agents.md", "warning"))

	}

	return result, nil
}

func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}
