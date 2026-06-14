package aider

import (
	"context"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/paths"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string   { return "aider" }
func (a *Adapter) Name() string { return "Aider" }

func (a *Adapter) SearchPaths() []string {
	home := paths.MustHomeDir()
	return []string{
		filepath.Join(home, ".aider.conf.yml"),
		".aider.conf.yml",
	}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		SupportsPermissions:  true,
		SupportsSettings:     true,
		Notes: map[string]string{
			"agents":   "Aider has no subagent system",
			"skills":   "Aider has no skill system; use CONVENTIONS.md instead",
			"commands": "Aider has no slash command system",
			"mcp":      "Aider has no native MCP support",
			"hooks":    "Aider has no hook system",
		},
	}
}

// aiderConfig mirrors the YAML structure of .aider.conf.yml.
type aiderConfig struct {
	Model              string   `yaml:"model,omitempty"`
	WeakModel          string   `yaml:"weak-model,omitempty"`
	EditorModel        string   `yaml:"editor-model,omitempty"`
	EditFormat         string   `yaml:"edit-format,omitempty"`
	ReasoningEffort    string   `yaml:"reasoning-effort,omitempty"`
	Git                *bool    `yaml:"git,omitempty"`
	AutoCommits        *bool    `yaml:"auto-commits,omitempty"`
	DarkMode           *bool    `yaml:"dark-mode,omitempty"`
	Pretty             *bool    `yaml:"pretty,omitempty"`
	Stream             *bool    `yaml:"stream,omitempty"`
	CachePrompts       *bool    `yaml:"cache-prompts,omitempty"`
	AutoLint           *bool    `yaml:"auto-lint,omitempty"`
	AutoTest           *bool    `yaml:"auto-test,omitempty"`
	TestCmd            string   `yaml:"test-cmd,omitempty"`
	LintCmd            []string `yaml:"lint-cmd,omitempty"`
	Read               []string `yaml:"read,omitempty"`
	File               []string `yaml:"file,omitempty"`
	MapTokens          int      `yaml:"map-tokens,omitempty"`
}

func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}
	home := paths.MustHomeDir()
	candidates := []string{
		filepath.Join(home, ".aider.conf.yml"),
		".aider.conf.yml",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			result.Found = true
			if filepath.IsAbs(p) {
				result.GlobalPaths = append(result.GlobalPaths, p)
			} else {
				result.ProjectPath = p
			}
			result.Summary["config_file"] = 1
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

	var configPath string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		configPath = filepath.Join(home, ".aider.conf.yml")
	default:
		configPath = filepath.Join(opts.ProjectRoot, ".aider.conf.yml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, nil // not configured is not an error
	}

	var ac aiderConfig
	if err := yaml.Unmarshal(data, &ac); err != nil {
		return cfg, nil
	}

	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, configPath)

	if ac.Model != "" {
		cfg.Settings.Model = ac.Model
	}
	if ac.WeakModel != "" {
		cfg.Settings.SmallModel = ac.WeakModel
	}
	if ac.ReasoningEffort != "" {
		cfg.Settings.ReasoningEffort = ac.ReasoningEffort
	}
	if ac.DarkMode != nil && *ac.DarkMode {
		cfg.Settings.DarkMode = true
	}
	if ac.AutoCommits != nil && *ac.AutoCommits {
		cfg.Settings.AutoCommit = true
	}

	// Read files → instructions (approximation)
	for _, f := range ac.Read {
		data, err := os.ReadFile(filepath.Join(opts.ProjectRoot, f))
		if err == nil {
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    f,
				Content: string(data),
				Scope:   "named",
			})
		}
	}

	// Check for CONVENTIONS.md as root instruction
	conventionsPath := filepath.Join(opts.ProjectRoot, "CONVENTIONS.md")
	if data, err := os.ReadFile(conventionsPath); err == nil {
		cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
			Path:    "CONVENTIONS.md",
			Content: string(data),
			Scope:   "root",
		})
	}

	// Approximate permissions from .aiderignore
	aiderignorePath := filepath.Join(opts.ProjectRoot, ".aiderignore")
	if data, err := os.ReadFile(aiderignorePath); err == nil {
		cfg.Permissions = &canonical.PermissionSet{
			DenyRules: []canonical.PermissionRule{
				{Pattern: string(data), Action: "deny"},
			},
		}
	}

	return cfg, nil
}

func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	ac := aiderConfig{}

	if cfg.Settings.Model != "" {
		ac.Model = cfg.Settings.Model
	}
	if cfg.Settings.SmallModel != "" {
		ac.WeakModel = cfg.Settings.SmallModel
	}
	if cfg.Settings.ReasoningEffort != "" {
		ac.ReasoningEffort = cfg.Settings.ReasoningEffort
	}
	if cfg.Settings.DarkMode {
		b := true
		ac.DarkMode = &b
	}
	if cfg.Settings.AutoCommit {
		b := true
		ac.AutoCommits = &b
	}

	// Instructions → read list
	for _, inst := range cfg.Instructions {
		ac.Read = append(ac.Read, inst.Path)
	}

	// If there are agents/skills/commands that Aider can't natively support, warn
	if len(cfg.Agents) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "agents",
			Field:       "agents",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Aider has no subagent system; agent definitions dropped",
			Severity:    "warning",
		})
	}
	if len(cfg.Skills) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "skills",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Aider has no skill system; use CONVENTIONS.md instead",
			Severity:    "info",
		})
	}
	if len(cfg.Commands) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "commands",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Aider has no slash command system",
			Severity:    "info",
		})
	}
	if len(cfg.MCPServers) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "mcp",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Aider has no native MCP support",
			Severity:    "warning",
		})
	}

	var configPath string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		configPath = filepath.Join(home, ".aider.conf.yml")
	default:
		configPath = filepath.Join(opts.ProjectRoot, ".aider.conf.yml")
	}

	if !opts.DryRun {
		data, err := yaml.Marshal(ac)
		if err != nil {
			return result, err
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return result, err
		}
	}
	result.FilesWritten = append(result.FilesWritten, configPath)

	// Write instructions as files
	for _, inst := range cfg.Instructions {
		p := filepath.Join(opts.ProjectRoot, inst.Path)
		if !opts.DryRun {
			if err := os.WriteFile(p, []byte(inst.Content), 0644); err != nil {
				return result, err
			}
		}
		result.FilesWritten = append(result.FilesWritten, p)
	}

	return result, nil
}

func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}
