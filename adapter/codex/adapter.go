// Package codex implements the AgentAdapter for OpenAI Codex CLI.
//
// Codex stores configuration in:
//   - ~/.codex/config.toml (global user config)
//   - .codex/config.toml (project config)
//   - .codex/codex.md (project instructions)
package codex

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
)

// Adapter implements adapter.AgentAdapter for Codex CLI.
type Adapter struct{}

// New creates a new Codex adapter.
func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ID() string   { return "codex" }
func (a *Adapter) Name() string { return "Codex CLI" }

func (a *Adapter) SearchPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".codex"),
		".codex",
	}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		SupportsAgents:       true,
		SupportsSkills:       true,
		SupportsCommands:     false, // no native slash commands — embed in instructions
		SupportsMCP:          true,
		SupportsPermissions:  true,
		SupportsHooks:        true,
		SupportsSettings:     true,
		Notes: map[string]string{
			"commands": "Codex has no native slash commands; they will be embedded as instructions",
			"skills":   "Codex skills are experimental (features.skills flag)",
			"agents":   "Codex agents use [agents.*] TOML tables",
		},
	}
}

// codexConfig mirrors the structure of Codex's config.toml.
type codexConfig struct {
	Model                  string                       `toml:"model,omitempty"`
	ApprovalPolicy         string                       `toml:"approval_policy,omitempty"`
	SandboxMode            string                       `toml:"sandbox_mode,omitempty"`
	ModelReasoningEffort   string                       `toml:"model_reasoning_effort,omitempty"`
	Personality            string                       `toml:"personality,omitempty"`
	WebSearch              string                       `toml:"web_search,omitempty"`
	LogDir                 string                       `toml:"log_dir,omitempty"`
	ShellEnvironmentPolicy map[string]interface{}        `toml:"shell_environment_policy,omitempty"`
	Features               map[string]interface{}        `toml:"features,omitempty"`
	MCPServers             map[string]codexMCPServer     `toml:"mcp_servers,omitempty"`
	Agents                 map[string]codexAgent         `toml:"agents,omitempty"`
	Hooks                  []codexHook                   `toml:"hooks,omitempty"`
}

type codexMCPServer struct {
	Command string            `toml:"command,omitempty"`
	Args    []string          `toml:"args,omitempty"`
	Env     map[string]string `toml:"env,omitempty"`
}

type codexAgent struct {
	Description  string   `toml:"description,omitempty"`
	SystemPrompt string   `toml:"system_prompt"` // multi-line string
	Tools        []string `toml:"tools,omitempty"`
	Model        string   `toml:"model,omitempty"`
}

type codexHook struct {
	Event   string   `toml:"event"`
	Command string   `toml:"command"`
	Args    []string `toml:"args,omitempty"`
}

// Detect checks whether Codex is configured.
func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}

	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".codex"),
		".codex",
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			result.Found = true
			if filepath.IsAbs(dir) {
				result.GlobalPaths = append(result.GlobalPaths, dir)
			} else {
				result.ProjectPath = dir
			}

			configPath := filepath.Join(dir, "config.toml")
			if _, err := os.Stat(configPath); err == nil {
				result.Summary["config_file"] = 1
			}

			codexMDPath := filepath.Join(dir, "codex.md")
			if _, err := os.Stat(codexMDPath); err == nil {
				result.Summary["instructions"] = 1
			}
		}
	}

	return result, nil
}

// Read converts Codex configuration to the canonical model.
func (a *Adapter) Read(ctx context.Context, opts adapter.ReadOptions) (*canonical.ShifterConfig, error) {
	cfg := &canonical.ShifterConfig{
		Meta: canonical.ConfigMeta{
			SourceAdapter: a.ID(),
			Version:       "1.0.0",
		},
	}

	var codexDir string
	switch opts.Scope {
	case "global":
		home, _ := os.UserHomeDir()
		codexDir = filepath.Join(home, ".codex")
	case "project":
		codexDir = filepath.Join(opts.ProjectRoot, ".codex")
	default:
		codexDir = filepath.Join(opts.ProjectRoot, ".codex")
		if _, err := os.Stat(codexDir); os.IsNotExist(err) {
			home, _ := os.UserHomeDir()
			codexDir = filepath.Join(home, ".codex")
		}
	}

	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, codexDir)

	// Read config.toml
	configPath := filepath.Join(codexDir, "config.toml")
	if data, err := os.ReadFile(configPath); err == nil {
		cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, configPath)
		var cc codexConfig
		if _, err := toml.Decode(string(data), &cc); err == nil {
			a.readConfig(&cc, cfg)
		}
	}

	// Read codex.md
	for _, p := range []string{
		filepath.Join(opts.ProjectRoot, ".codex", "codex.md"),
		filepath.Join(opts.ProjectRoot, "codex.md"),
	} {
		if content, err := os.ReadFile(p); err == nil {
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    ".codex/codex.md",
				Content: string(content),
				Scope:   "root",
			})
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, p)
			break
		}
	}

	return cfg, nil
}

// Write converts the canonical model back to Codex native config.
func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	var codexDir string
	switch opts.Scope {
	case "global":
		home, _ := os.UserHomeDir()
		codexDir = filepath.Join(home, ".codex")
	default:
		codexDir = filepath.Join(opts.ProjectRoot, ".codex")
	}

	cc := a.buildConfig(cfg, &result.LossWarnings)
	configPath := filepath.Join(codexDir, "config.toml")

	if !opts.DryRun {
		if err := os.MkdirAll(codexDir, 0755); err != nil {
			return result, fmt.Errorf("create codex dir: %w", err)
		}

		var buf strings.Builder
		if err := toml.NewEncoder(&buf).Encode(cc); err != nil {
			return result, fmt.Errorf("marshal config.toml: %w", err)
		}
		if err := os.WriteFile(configPath, []byte(buf.String()), 0644); err != nil {
			return result, fmt.Errorf("write config.toml: %w", err)
		}
	}
	result.FilesWritten = append(result.FilesWritten, configPath)

	// Write codex.md
	for _, inst := range cfg.Instructions {
		if inst.Scope == "root" {
			codexMDPath := filepath.Join(codexDir, "codex.md")
			content := inst.Content

			// If the source agent has commands and Codex doesn't support them natively,
			// embed them in the instructions
			if len(cfg.Commands) > 0 {
				content += a.embedCommands(cfg.Commands)
			}

			if !opts.DryRun {
				if err := os.WriteFile(codexMDPath, []byte(content), 0644); err != nil {
					return result, fmt.Errorf("write codex.md: %w", err)
				}
			}
			result.FilesWritten = append(result.FilesWritten, codexMDPath)
			break
		}
	}

	// If no instructions but we have commands, create instructions
	if len(cfg.Instructions) == 0 && len(cfg.Commands) > 0 {
		codexMDPath := filepath.Join(codexDir, "codex.md")
		content := a.embedCommands(cfg.Commands)
		if !opts.DryRun {
			if err := os.WriteFile(codexMDPath, []byte(content), 0644); err != nil {
				return result, fmt.Errorf("write codex.md: %w", err)
			}
		}
		result.FilesWritten = append(result.FilesWritten, codexMDPath)
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "commands",
			Field:       "slash commands",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Codex has no native slash commands; embedded as instructions in codex.md",
			Severity:    "info",
		})
	}

	return result, nil
}

// Preview shows proposed changes without writing.
func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}

func (a *Adapter) readConfig(cc *codexConfig, cfg *canonical.ShifterConfig) {
	// Model
	if cc.Model != "" {
		cfg.Settings.Model = cc.Model
	}

	// Approval
	if cc.ApprovalPolicy != "" {
		cfg.Settings.ApprovalMode = cc.ApprovalPolicy
	}

	// Sandbox
	if cc.SandboxMode != "" {
		cfg.Settings.SandboxMode = cc.SandboxMode
	}

	// Reasoning
	if cc.ModelReasoningEffort != "" {
		cfg.Settings.ReasoningEffort = cc.ModelReasoningEffort
	}

	// Personality
	if cc.Personality != "" {
		cfg.Settings.Personality = cc.Personality
	}

	// Web search
	if cc.WebSearch != "" {
		cfg.Settings.WebSearch = cc.WebSearch
	}

	// MCP servers
	for name, srv := range cc.MCPServers {
		cfg.MCPServers = append(cfg.MCPServers, canonical.MCPServerDef{
			Name:    name,
			Type:    "stdio",
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
			Enabled: true,
		})
	}

	// Agents
	for name, agent := range cc.Agents {
		cfg.Agents = append(cfg.Agents, canonical.AgentDef{
			Name:         name,
			Description:  agent.Description,
			Model:        agent.Model,
			Tools:        agent.Tools,
			SystemPrompt: agent.SystemPrompt,
		})
	}

	// Hooks
	for _, h := range cc.Hooks {
		cfg.Hooks = append(cfg.Hooks, canonical.HookDef{
			Event:   h.Event,
			Command: h.Command,
			Args:    h.Args,
		})
	}
}

func (a *Adapter) buildConfig(cfg *canonical.ShifterConfig, warnings *[]canonical.LossWarning) *codexConfig {
	cc := &codexConfig{
		Features: make(map[string]interface{}),
	}

	if cfg.Settings.Model != "" {
		cc.Model = cfg.Settings.Model
	}
	if cfg.Settings.ApprovalMode != "" {
		cc.ApprovalPolicy = cfg.Settings.ApprovalMode
	}
	if cfg.Settings.SandboxMode != "" {
		cc.SandboxMode = cfg.Settings.SandboxMode
	}
	if cfg.Settings.ReasoningEffort != "" {
		cc.ModelReasoningEffort = cfg.Settings.ReasoningEffort
	}
	if cfg.Settings.Personality != "" {
		cc.Personality = cfg.Settings.Personality
	}
	if cfg.Settings.WebSearch != "" {
		cc.WebSearch = cfg.Settings.WebSearch
	}

	// MCP servers
	if len(cfg.MCPServers) > 0 {
		cc.MCPServers = make(map[string]codexMCPServer)
		for _, mcp := range cfg.MCPServers {
			cc.MCPServers[mcp.Name] = codexMCPServer{
				Command: mcp.Command,
				Args:    mcp.Args,
				Env:     mcp.Env,
			}
		}
	}

	// Agents
	if len(cfg.Agents) > 0 {
		cc.Agents = make(map[string]codexAgent)
		for _, agent := range cfg.Agents {
			cc.Agents[agent.Name] = codexAgent{
				Description:  agent.Description,
				SystemPrompt: agent.SystemPrompt,
				Tools:        agent.Tools,
				Model:        agent.Model,
			}
		}
	}

	// Hooks
	for _, hook := range cfg.Hooks {
		cc.Hooks = append(cc.Hooks, codexHook{
			Event:   hook.Event,
			Command: hook.Command,
			Args:    hook.Args,
		})
	}

	// If there are skills, note experimental support
	if len(cfg.Skills) > 0 {
		cc.Features["skills"] = true
		if warnings != nil {
			*warnings = append(*warnings, canonical.LossWarning{
				Feature:     "skills",
				Field:       "skills",
				SourceAgent: cfg.Meta.SourceAdapter,
				TargetAgent: a.ID(),
				Reason:      "Codex skill support is experimental; enable with features.skills=true",
				Severity:    "info",
			})
		}
	}

	// Permissions → approval_policy mapping
	if cfg.Permissions != nil {
		if cfg.Permissions.DefaultMode == "edit" {
			cc.ApprovalPolicy = "on-request"
		}
	}

	return cc
}

// embedCommands formats slash commands as instructions embedded in Codex's codex.md.
func (a *Adapter) embedCommands(commands []canonical.CommandDef) string {
	if len(commands) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n\n---\n\n## Custom Commands (ported from another agent)\n\n")
	b.WriteString("> **Note:** Codex CLI does not have native slash commands. ")
	b.WriteString("These are provided as instruction patterns you can reference.\n\n")

	for _, cmd := range commands {
		b.WriteString(fmt.Sprintf("### /%s\n\n", cmd.Name))
		if cmd.Description != "" {
			b.WriteString(fmt.Sprintf("*%s*\n\n", cmd.Description))
		}
		b.WriteString(cmd.Prompt)
		b.WriteString("\n\n---\n\n")
	}

	return b.String()
}
