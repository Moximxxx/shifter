// Package claudecode implements the AgentAdapter for Claude Code.
//
// Claude Code stores configuration in:
//   - .claude/settings.json (JSON) — permissions, MCP, hooks, env, model
//   - .claude/agents/*.md (YAML frontmatter + Markdown) — subagent definitions
//   - .claude/skills/<name>/SKILL.md (YAML frontmatter + Markdown) — skills
//   - .claude/commands/*.md (YAML frontmatter + Markdown) — slash commands
//   - CLAUDE.md or .claude/CLAUDE.md — project instructions
package claudecode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/format"
	"github.com/moximxxx/shifter/pkg/paths"
)

// Adapter implements adapter.AgentAdapter for Claude Code.
type Adapter struct{}

// New creates a new Claude Code adapter.
func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ID() string   { return "claude-code" }
func (a *Adapter) Name() string { return "Claude Code" }

func (a *Adapter) SearchPaths() []string {
	home := paths.MustHomeDir()
	return []string{
		filepath.Join(home, ".claude"),
		".claude",
	}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		SupportsAgents:       true,
		SupportsSkills:       true,
		SupportsCommands:     true,
		SupportsMCP:          true,
		SupportsPermissions:  true,
		SupportsHooks:        true,
		SupportsSettings:     true,
	}
}

// claudeSettings mirrors the structure of .claude/settings.json.
type claudeSettings struct {
	Model       string                 `json:"model,omitempty"`
	Permissions *claudePermissions     `json:"permissions,omitempty"`
	Env         map[string]string      `json:"env,omitempty"`
	Hooks       map[string][]claudeHook `json:"hooks,omitempty"`
	MCPServers  map[string]claudeMCPServer `json:"mcpServers,omitempty"`
	Sandbox     *claudeSandbox         `json:"sandbox,omitempty"`
}

type claudePermissions struct {
	Allow                []string `json:"allow,omitempty"`
	Deny                 []string `json:"deny,omitempty"`
	DefaultMode          string   `json:"defaultMode,omitempty"`
	AdditionalDirectories []string `json:"additionalDirectories,omitempty"`
}

type claudeHook struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []claudeHookAction `json:"hooks,omitempty"`
}

type claudeHookAction struct {
	Type    string `json:"type"`
	Command string `json:"command,omitempty"`
}

type claudeMCPServer struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Type    string            `json:"type,omitempty"`
}

type claudeSandbox struct {
	Enabled    bool                `json:"enabled,omitempty"`
	Filesystem *claudeSandboxFS    `json:"filesystem,omitempty"`
}

type claudeSandboxFS struct {
	DenyRead  []string `json:"denyRead,omitempty"`
	DenyWrite []string `json:"denyWrite,omitempty"`
}

// Detect checks whether Claude Code is configured.
func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}

	home := paths.MustHomeDir()
	candidates := []string{
		filepath.Join(home, ".claude"),
		".claude",
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			result.Found = true
			if filepath.IsAbs(dir) {
				result.GlobalPaths = append(result.GlobalPaths, dir)
			} else {
				result.ProjectPath = dir
			}

			// Count contents
			agentsDir := filepath.Join(dir, "agents")
			if entries, err := os.ReadDir(agentsDir); err == nil {
				count := 0
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".md") {
						count++
					}
				}
				if count > 0 {
					result.Summary["agents"] = count
				}
			}

			skillsDir := filepath.Join(dir, "skills")
			if entries, err := os.ReadDir(skillsDir); err == nil {
				count := 0
				for _, e := range entries {
					if e.IsDir() {
						count++
					}
				}
				if count > 0 {
					result.Summary["skills"] = count
				}
			}

			commandsDir := filepath.Join(dir, "commands")
			if entries, err := os.ReadDir(commandsDir); err == nil {
				count := 0
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".md") {
						count++
					}
				}
				if count > 0 {
					result.Summary["commands"] = count
				}
			}

			settingsPath := filepath.Join(dir, "settings.json")
			if _, err := os.Stat(settingsPath); err == nil {
				result.Summary["settings_file"] = 1
			}
		}
	}

	return result, nil
}

// Read converts Claude Code configuration to the canonical model.
func (a *Adapter) Read(ctx context.Context, opts adapter.ReadOptions) (*canonical.ShifterConfig, error) {
	cfg := &canonical.ShifterConfig{
		Meta: canonical.ConfigMeta{
			SourceAdapter: a.ID(),
			Version:       "1.0.0",
		},
	}

	var claudeDir string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		claudeDir = filepath.Join(home, ".claude")
	case "project":
		claudeDir = filepath.Join(opts.ProjectRoot, ".claude")
	default:
		// "all" — prefer project, fall back to global
		claudeDir = filepath.Join(opts.ProjectRoot, ".claude")
		if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
			home := paths.MustHomeDir()
			claudeDir = filepath.Join(home, ".claude")
		}
	}

	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, claudeDir)

	// Read settings.json
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, settingsPath)
		var s claudeSettings
		if err := json.Unmarshal(data, &s); err == nil {
			a.readSettings(&s, cfg)
		}
	}

	// Read CLAUDE.md (project root or .claude/)
	for _, p := range []string{
		filepath.Join(opts.ProjectRoot, "CLAUDE.md"),
		filepath.Join(claudeDir, "CLAUDE.md"),
	} {
		if content, err := os.ReadFile(p); err == nil {
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    filepath.Base(p),
				Content: string(content),
				Scope:   "root",
			})
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, p)
			break // only one
		}
	}

	// Read agents
	agentsDir := filepath.Join(claudeDir, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			agentPath := filepath.Join(agentsDir, entry.Name())
			agent, err := readAgentFile(agentPath)
			if err == nil {
				cfg.Agents = append(cfg.Agents, agent)
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, agentPath)
			}
		}
	}

	// Read skills
	skillsDir := filepath.Join(claudeDir, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(skillsDir, entry.Name())
			skill, err := readSkillDir(skillPath)
			if err == nil {
				cfg.Skills = append(cfg.Skills, skill)
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, skillPath)
			}
		}
	}

	// Read commands
	commandsDir := filepath.Join(claudeDir, "commands")
	if entries, err := os.ReadDir(commandsDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			cmdPath := filepath.Join(commandsDir, entry.Name())
			cmd, err := readCommandFile(cmdPath)
			if err == nil {
				cfg.Commands = append(cfg.Commands, cmd)
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, cmdPath)
			}
		}
	}

	return cfg, nil
}

// Write converts the canonical model back to Claude Code native config.
func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	// Determine target directory
	var claudeDir string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		claudeDir = filepath.Join(home, ".claude")
	default:
		claudeDir = filepath.Join(opts.ProjectRoot, ".claude")
	}

	// Write settings.json
	settings := a.buildSettings(cfg, &result.LossWarnings)
	if settings != nil {
		settingsPath := filepath.Join(claudeDir, "settings.json")
		if !opts.DryRun {
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				return result, fmt.Errorf("create claude dir: %w", err)
			}
			data, err := json.MarshalIndent(settings, "", "  ")
			if err != nil {
				return result, fmt.Errorf("marshal settings: %w", err)
			}
			if err := os.WriteFile(settingsPath, data, 0644); err != nil {
				return result, fmt.Errorf("write settings: %w", err)
			}
		}
		result.FilesWritten = append(result.FilesWritten, settingsPath)
	}

	// Write CLAUDE.md
	for _, inst := range cfg.Instructions {
		if inst.Scope == "root" {
			claudeMDPath := filepath.Join(opts.ProjectRoot, "CLAUDE.md")
			if !opts.DryRun {
				if err := os.WriteFile(claudeMDPath, []byte(inst.Content), 0644); err != nil {
					return result, fmt.Errorf("write CLAUDE.md: %w", err)
				}
			}
			result.FilesWritten = append(result.FilesWritten, claudeMDPath)
			break
		}
	}

	// Write agents
	if len(cfg.Agents) > 0 {
		agentsDir := filepath.Join(claudeDir, "agents")
		if !opts.DryRun {
			if err := os.MkdirAll(agentsDir, 0755); err != nil {
				return result, fmt.Errorf("create agents dir: %w", err)
			}
		}
		for _, agent := range cfg.Agents {
			agentPath := filepath.Join(agentsDir, agent.Name+".md")
			content := formatAgentFile(agent)
			if !opts.DryRun {
				if err := os.WriteFile(agentPath, []byte(content), 0644); err != nil {
					return result, fmt.Errorf("write agent %s: %w", agent.Name, err)
				}
			}
			result.FilesWritten = append(result.FilesWritten, agentPath)
		}
	}

	// Write skills
	for _, skill := range cfg.Skills {
		skillDir := filepath.Join(claudeDir, "skills", skill.Name)
		if !opts.DryRun {
			if err := os.MkdirAll(skillDir, 0755); err != nil {
				return result, fmt.Errorf("create skill dir %s: %w", skill.Name, err)
			}
		}
		skillPath := filepath.Join(skillDir, "SKILL.md")
		content := formatSkillFile(skill)
		if !opts.DryRun {
			if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
				return result, fmt.Errorf("write skill %s: %w", skill.Name, err)
			}
		}
		result.FilesWritten = append(result.FilesWritten, skillPath)

		// Write supporting files
		for name, content := range skill.Scripts {
			p := filepath.Join(skillDir, "scripts", name)
			if !opts.DryRun {
				os.MkdirAll(filepath.Dir(p), 0755)
				os.WriteFile(p, []byte(content), 0755)
			}
			result.FilesWritten = append(result.FilesWritten, p)
		}
		for name, content := range skill.References {
			p := filepath.Join(skillDir, name)
			if !opts.DryRun {
				os.WriteFile(p, []byte(content), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, p)
		}
	}

	// Write commands
	if len(cfg.Commands) > 0 {
		commandsDir := filepath.Join(claudeDir, "commands")
		if !opts.DryRun {
			if err := os.MkdirAll(commandsDir, 0755); err != nil {
				return result, fmt.Errorf("create commands dir: %w", err)
			}
		}
		for _, cmd := range cfg.Commands {
			cmdPath := filepath.Join(commandsDir, cmd.Name+".md")
			content := formatCommandFile(cmd)
			if !opts.DryRun {
				if err := os.WriteFile(cmdPath, []byte(content), 0644); err != nil {
					return result, fmt.Errorf("write command %s: %w", cmd.Name, err)
				}
			}
			result.FilesWritten = append(result.FilesWritten, cmdPath)
		}
	}

	return result, nil
}

// Preview shows proposed changes without writing.
func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	// TODO: implement diff preview
	return &adapter.DiffResult{}, nil
}

// readSettings populates the canonical config from Claude Code settings.
func (a *Adapter) readSettings(s *claudeSettings, cfg *canonical.ShifterConfig) {
	// Model
	if s.Model != "" {
		cfg.Settings.Model = s.Model
	}

	// Permissions
	if s.Permissions != nil {
		ps := &canonical.PermissionSet{
			DefaultMode: s.Permissions.DefaultMode,
		}
		for _, pattern := range s.Permissions.Allow {
			ps.AllowRules = append(ps.AllowRules, canonical.PermissionRule{
				Pattern: pattern,
				Action:  "allow",
			})
		}
		for _, pattern := range s.Permissions.Deny {
			ps.DenyRules = append(ps.DenyRules, canonical.PermissionRule{
				Pattern: pattern,
				Action:  "deny",
			})
		}
		cfg.Permissions = ps
	}

	// MCP servers
	for name, srv := range s.MCPServers {
		mcpType := "stdio"
		if srv.URL != "" {
			mcpType = "http"
		}
		if srv.Type == "sse" {
			mcpType = "sse"
		}
		cfg.MCPServers = append(cfg.MCPServers, canonical.MCPServerDef{
			Name:    name,
			Type:    mcpType,
			Command: srv.Command,
			Args:    srv.Args,
			URL:     srv.URL,
			Env:     srv.Env,
			Enabled: true,
		})
	}

	// Hooks
	for event, hookList := range s.Hooks {
		for _, h := range hookList {
			for _, action := range h.Hooks {
				cfg.Hooks = append(cfg.Hooks, canonical.HookDef{
					Event:   event,
					Matcher: h.Matcher,
					Command: action.Command,
				})
			}
		}
	}

	// Env → Settings.Extra
	if len(s.Env) > 0 {
		if cfg.Settings.Extra == nil {
			cfg.Settings.Extra = make(map[string]interface{})
		}
		cfg.Settings.Extra["env"] = s.Env
	}
}

// buildSettings constructs a Claude Code settings struct from the canonical model.
func (a *Adapter) buildSettings(cfg *canonical.ShifterConfig, warnings *[]canonical.LossWarning) *claudeSettings {
	s := &claudeSettings{}

	// Model
	if cfg.Settings.Model != "" {
		s.Model = cfg.Settings.Model
	}

	// Permissions
	if cfg.Permissions != nil {
		s.Permissions = &claudePermissions{
			DefaultMode: cfg.Permissions.DefaultMode,
		}
		for _, r := range cfg.Permissions.AllowRules {
			s.Permissions.Allow = append(s.Permissions.Allow, r.Pattern)
		}
		for _, r := range cfg.Permissions.DenyRules {
			s.Permissions.Deny = append(s.Permissions.Deny, r.Pattern)
		}
	}

	// MCP servers
	if len(cfg.MCPServers) > 0 {
		s.MCPServers = make(map[string]claudeMCPServer)
		for _, mcp := range cfg.MCPServers {
			s.MCPServers[mcp.Name] = claudeMCPServer{
				Command: mcp.Command,
				Args:    mcp.Args,
				Env:     mcp.Env,
				URL:     mcp.URL,
				Type:    mcp.Type,
			}
		}
	}

	// Hooks
	if len(cfg.Hooks) > 0 {
		s.Hooks = make(map[string][]claudeHook)
		for _, hook := range cfg.Hooks {
			entry := claudeHook{
				Matcher: hook.Matcher,
				Hooks: []claudeHookAction{
					{Type: "command", Command: hook.Command},
				},
			}
			s.Hooks[hook.Event] = append(s.Hooks[hook.Event], entry)
		}
	}

	// Extra env
	if cfg.Settings.Extra != nil {
		if envRaw, ok := cfg.Settings.Extra["env"]; ok {
			if envMap, ok := envRaw.(map[string]interface{}); ok {
				s.Env = make(map[string]string)
				for k, v := range envMap {
					if vs, ok := v.(string); ok {
						s.Env[k] = vs
					}
				}
			}
		}
	}

	return s
}

// readAgentFile reads a Claude Code agent .md file.
func readAgentFile(path string) (canonical.AgentDef, error) {
	agent := canonical.AgentDef{}
	content, err := os.ReadFile(path)
	if err != nil {
		return agent, err
	}

	fm, body, err := format.ParseFrontMatter(string(content))
	if err != nil {
		return agent, fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}

	agent.SystemPrompt = strings.TrimSpace(body)
	if name, ok := fm["name"].(string); ok {
		agent.Name = name
	} else {
		agent.Name = strings.TrimSuffix(filepath.Base(path), ".md")
	}
	if desc, ok := fm["description"].(string); ok {
		agent.Description = desc
	}
	if tools, ok := fm["tools"].(string); ok {
		for _, t := range strings.Split(tools, ",") {
			agent.Tools = append(agent.Tools, strings.TrimSpace(t))
		}
	}
	if model, ok := fm["model"].(string); ok {
		agent.Model = model
	}
	if skills, ok := fm["skills"].(string); ok {
		for _, s := range strings.Split(skills, ",") {
			agent.Skills = append(agent.Skills, strings.TrimSpace(s))
		}
	}
	if mcp, ok := fm["mcpServers"].(string); ok {
		for _, m := range strings.Split(mcp, ",") {
			agent.MCPServers = append(agent.MCPServers, strings.TrimSpace(m))
		}
	}
	if mode, ok := fm["mode"].(string); ok {
		agent.Mode = mode
	}
	if color, ok := fm["color"].(string); ok {
		agent.Color = color
	}

	return agent, nil
}

// readSkillDir reads a Claude Code skill directory.
func readSkillDir(dir string) (canonical.SkillDef, error) {
	skill := canonical.SkillDef{
		Scripts:    make(map[string]string),
		References: make(map[string]string),
		Examples:   make(map[string]string),
		Templates:  make(map[string]string),
	}

	skillMDPath := filepath.Join(dir, "SKILL.md")
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return skill, fmt.Errorf("read SKILL.md in %s: %w", dir, err)
	}

	fm, body, err := format.ParseFrontMatter(string(content))
	if err != nil {
		return skill, fmt.Errorf("parse frontmatter in %s: %w", skillMDPath, err)
	}

	skill.Markdown = strings.TrimSpace(body)
	if name, ok := fm["name"].(string); ok {
		skill.Name = name
	} else {
		skill.Name = filepath.Base(dir)
	}
	if desc, ok := fm["description"].(string); ok {
		skill.Description = desc
	}
	if tools, ok := fm["allowed_tools"].([]interface{}); ok {
		for _, t := range tools {
			if ts, ok := t.(string); ok {
				skill.AllowedTools = append(skill.AllowedTools, ts)
			}
		}
	}

	// Read supporting files
	readDirFiles := func(subdir string, target map[string]string) {
		fullPath := filepath.Join(dir, subdir)
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(fullPath, entry.Name()))
			if err == nil {
				target[entry.Name()] = string(data)
			}
		}
	}

	readDirFiles("scripts", skill.Scripts)
	readDirFiles(".", skill.References) // REFERENCE.md, EXAMPLES.md
	readDirFiles("templates", skill.Templates)

	// Also check for REFERENCE.md and EXAMPLES.md in root
	for _, name := range []string{"REFERENCE.md", "EXAMPLES.md"} {
		refPath := filepath.Join(dir, name)
		if data, err := os.ReadFile(refPath); err == nil {
			skill.References[name] = string(data)
		}
	}

	return skill, nil
}

// readCommandFile reads a Claude Code command .md file.
func readCommandFile(path string) (canonical.CommandDef, error) {
	cmd := canonical.CommandDef{
		Name: strings.TrimSuffix(filepath.Base(path), ".md"),
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return cmd, err
	}

	fm, body, err := format.ParseFrontMatter(string(content))
	if err != nil {
		// No frontmatter — use entire content as prompt
		cmd.Prompt = strings.TrimSpace(string(content))
		return cmd, nil
	}

	cmd.Prompt = strings.TrimSpace(body)
	if desc, ok := fm["description"].(string); ok {
		cmd.Description = desc
	}
	if hint, ok := fm["argument-hint"].(string); ok {
		cmd.ArgumentHint = hint
	}

	return cmd, nil
}

// formatAgentFile formats an agent definition as a Claude Code agent .md file.
func formatAgentFile(agent canonical.AgentDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", agent.Name))
	if agent.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", agent.Description))
	}
	if len(agent.Tools) > 0 {
		b.WriteString(fmt.Sprintf("tools: %s\n", strings.Join(agent.Tools, ", ")))
	}
	if agent.Model != "" {
		b.WriteString(fmt.Sprintf("model: %s\n", agent.Model))
	}
	if len(agent.Skills) > 0 {
		b.WriteString(fmt.Sprintf("skills: %s\n", strings.Join(agent.Skills, ", ")))
	}
	if len(agent.MCPServers) > 0 {
		b.WriteString(fmt.Sprintf("mcpServers: %s\n", strings.Join(agent.MCPServers, ", ")))
	}
	if agent.Mode != "" {
		b.WriteString(fmt.Sprintf("mode: %s\n", agent.Mode))
	}
	if agent.Color != "" {
		b.WriteString(fmt.Sprintf("color: %s\n", agent.Color))
	}
	b.WriteString("---\n\n")
	b.WriteString(agent.SystemPrompt)
	if !strings.HasSuffix(agent.SystemPrompt, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// formatSkillFile formats a skill definition as a Claude Code SKILL.md file.
func formatSkillFile(skill canonical.SkillDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", skill.Name))
	if skill.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", skill.Description))
	}
	if len(skill.AllowedTools) > 0 {
		b.WriteString("allowed_tools:\n")
		for _, t := range skill.AllowedTools {
			b.WriteString(fmt.Sprintf("  - %s\n", t))
		}
	}
	b.WriteString("---\n\n")
	b.WriteString(skill.Markdown)
	if !strings.HasSuffix(skill.Markdown, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// formatCommandFile formats a command definition as a Claude Code command .md file.
func formatCommandFile(cmd canonical.CommandDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	if cmd.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", cmd.Description))
	}
	if cmd.ArgumentHint != "" {
		b.WriteString(fmt.Sprintf("argument-hint: %s\n", cmd.ArgumentHint))
	}
	b.WriteString("---\n\n")
	b.WriteString(cmd.Prompt)
	if !strings.HasSuffix(cmd.Prompt, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}
