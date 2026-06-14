// Package opencode implements the AgentAdapter for OpenCode.
//
// OpenCode stores configuration in:
//   - opencode.jsonc / .opencode/opencode.jsonc (JSON with comments)
//   - .opencode/agents/*.md (YAML frontmatter + Markdown)
//   - .opencode/commands/*.md (slash commands)
//   - .opencode/rules/*.md (numbered rules)
//   - .opencode/plugins/*.ts (TypeScript plugins)
//   - .opencode/hooks/*.sh (shell hook scripts)
//   - .opencode/skills/*/SKILL.md (skills)
//   - .opencode/tools/*.ts (custom tools)
//   - .opencode/constraints/*.md (constraint docs)
//   - AGENTS.md (project instructions, auto-loaded)
//
// Reference: https://opencode.ai/docs
package opencode

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/convert"
	"github.com/moximxxx/shifter/pkg/format"
	"github.com/moximxxx/shifter/pkg/paths"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string   { return "opencode" }
func (a *Adapter) Name() string { return "OpenCode" }

func (a *Adapter) SearchPaths() []string {
	home := paths.MustHomeDir()
	return []string{
		filepath.Join(home, ".config", "opencode"),
		".opencode",
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
		SupportsSettings:     true,
		Notes: map[string]string{
			"hooks":    "OpenCode uses shell scripts in .opencode/hooks/ with RESULT: PASS/BLOCK/WARN protocol",
			"plugins":  "OpenCode supports TypeScript plugins using @opencode-ai/plugin SDK",
			"rules":    "OpenCode supports numbered rules in .opencode/rules/ with groups and severity",
			"tools":    "OpenCode supports custom tools defined with the plugin SDK",
			"commands": "Commands support agent scoping and inline exec with ! syntax",
		},
	}
}

// === OpenCode JSON Config Structures ===

type opencodeConfig struct {
	Schema       string                       `json:"$schema,omitempty"`
	Model        string                       `json:"model,omitempty"`
	SmallModel   string                       `json:"small_model,omitempty"`
	DefaultAgent string                       `json:"default_agent,omitempty"`
	Autoupdate   string                       `json:"autoupdate,omitempty"`
	LSP          bool                         `json:"lsp,omitempty"`
	Formatter    *opencodeFormatter           `json:"formatter,omitempty"`
	Watcher      *opencodeWatcher             `json:"watcher,omitempty"`
	Compaction   *opencodeCompaction          `json:"compaction,omitempty"`
	Instructions []string                     `json:"instructions,omitempty"`
	Permission   map[string]interface{}       `json:"permission,omitempty"`
	Agent        map[string]opencodeAgentDef  `json:"agent,omitempty"`
	MCP          map[string]opencodeMCPServer `json:"mcp,omitempty"`
	Plugin       []string                     `json:"plugin,omitempty"`
	Skills       *opencodeSkillsConfig        `json:"skills,omitempty"`
	Command      map[string]opencodeCmdDef    `json:"command,omitempty"`
	Provider     map[string]interface{}       `json:"provider,omitempty"`
	// Old format keys (fallback)
	MCPServers map[string]opencodeMCPServer `json:"mcpServers,omitempty"`
}

type opencodeFormatter struct {
	Prettier *struct {
		Extensions []string `json:"extensions,omitempty"`
	} `json:"prettier,omitempty"`
}

type opencodeWatcher struct {
	Ignore []string `json:"ignore,omitempty"`
}

type opencodeCompaction struct {
	Auto     bool `json:"auto,omitempty"`
	Prune    bool `json:"prune,omitempty"`
	Reserved int  `json:"reserved,omitempty"`
}

type opencodeAgentDef struct {
	Description string                 `json:"description,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
	Hidden      bool                   `json:"hidden,omitempty"`
	Color       string                 `json:"color,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	Steps       int                    `json:"steps,omitempty"`
	Prompt      string                 `json:"prompt,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Tools       map[string]interface{} `json:"tools,omitempty"`
	Permission  map[string]interface{} `json:"permission,omitempty"`
}

type opencodeMCPServer struct {
	Type        string            `json:"type,omitempty"`
	Command     []string          `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	URL         string            `json:"url,omitempty"`
	Timeout     int               `json:"timeout,omitempty"`
}

type opencodeSkillsConfig struct {
	Paths  []string `json:"paths,omitempty"`
	Global []string `json:"global,omitempty"`
}

type opencodeCmdDef struct {
	Description string `json:"description,omitempty"`
	Agent       string `json:"agent,omitempty"`
	UserInput   string `json:"userInput,omitempty"`
	Subtask     *bool  `json:"subtask,omitempty"`
	Template    *struct {
		File string `json:"file,omitempty"`
	} `json:"template,omitempty"`
}

// ============================================================================
// Detect
// ============================================================================

func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}
	home := paths.MustHomeDir()

	// Check for opencode.jsonc in various locations
	globalCandidates := []string{
		filepath.Join(home, ".config", "opencode", "opencode.jsonc"),
		filepath.Join(home, ".config", "opencode", "opencode.json"),
	}
	projectCandidates := []string{
		"opencode.jsonc",
		"opencode.json",
		filepath.Join(".opencode", "opencode.jsonc"),
		filepath.Join(".opencode", "opencode.json"),
	}

	for _, p := range globalCandidates {
		if _, err := os.Stat(p); err == nil {
			result.Found = true
			result.GlobalPaths = append(result.GlobalPaths, filepath.Dir(p))
			result.Summary["config_file"] = 1
			break
		}
	}

	for _, p := range projectCandidates {
		if _, err := os.Stat(p); err == nil {
			result.Found = true
			result.ProjectPath = filepath.Dir(p)
			if result.Summary["config_file"] == 0 {
				result.Summary["config_file"] = 1
			}
			break
		}
	}

	// Count .opencode/ subdirectories
	dirs := map[string]string{
		"agents":      ".opencode/agents",
		"commands":    ".opencode/commands",
		"rules":       ".opencode/rules",
		"plugins":     ".opencode/plugins",
		"hooks":       ".opencode/hooks",
		"skills":      ".opencode/skills",
		"constraints": ".opencode/constraints",
	}
	for key, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			entries, _ := os.ReadDir(dir)
			count := 0
			for _, e := range entries {
				if e.IsDir() || !strings.HasPrefix(e.Name(), ".") {
					count++
				}
			}
			if count > 0 {
				result.Summary[key] = count
			}
		}
	}

	return result, nil
}

// ============================================================================
// Read
// ============================================================================

func (a *Adapter) Read(ctx context.Context, opts adapter.ReadOptions) (*canonical.ShifterConfig, error) {
	cfg := &canonical.ShifterConfig{
		Meta: canonical.ConfigMeta{
			SourceAdapter: a.ID(),
			Version:       "1.0.0",
		},
	}

	// Find config file
	configPath := a.findConfigFile(opts)
	if configPath == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, nil
	}

	// Strip comments (JSONC support)
	jsonData := format.StripJSONComments(string(data))

	var oc opencodeConfig
	if err := json.Unmarshal([]byte(jsonData), &oc); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", configPath, err)
	}
	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, configPath)

	// --- Model & Settings ---
	if oc.Model != "" {
		cfg.Settings.Model = oc.Model
	}
	if oc.SmallModel != "" {
		cfg.Settings.SmallModel = oc.SmallModel
	}
	if oc.DefaultAgent != "" {
		cfg.Settings.DefaultAgent = oc.DefaultAgent
	}
	if oc.Autoupdate != "" {
		cfg.Settings.AutoUpdate = oc.Autoupdate != "off"
	}

	// Store rich settings in Extra for round-trip preservation
	cfg.Settings.Extra = make(map[string]interface{})
	if oc.LSP {
		cfg.Settings.Extra["lsp"] = true
	}
	if oc.Compaction != nil {
		cfg.Settings.Extra["compaction"] = map[string]interface{}{
			"auto":     oc.Compaction.Auto,
			"prune":    oc.Compaction.Prune,
			"reserved": oc.Compaction.Reserved,
		}
	}
	if oc.Formatter != nil {
		cfg.Settings.Extra["formatter"] = oc.Formatter
	}
	if oc.Watcher != nil {
		cfg.Settings.Extra["watcher"] = map[string]interface{}{
			"ignore": oc.Watcher.Ignore,
		}
	}

	// --- Instructions ---
	// AGENTS.md is the root instruction in OpenCode (auto-loaded into every context)
	hasRootInstruction := false
	for _, path := range oc.Instructions {
		// Support glob patterns — read individual files
		matches, _ := filepath.Glob(path)
		if len(matches) == 0 {
			matches = []string{path}
		}
		for _, match := range matches {
			data, err := os.ReadFile(match)
			if err != nil {
				continue
			}
			// AGENTS.md is the root instruction
			scope := "named"
			if filepath.Base(match) == "AGENTS.md" || filepath.Base(match) == "AGENTS.md" {
				scope = "root"
				hasRootInstruction = true
			}
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    match,
				Content: string(data),
				Scope:   scope,
			})
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, match)
		}
	}

	// If AGENTS.md wasn't in the instructions array, still load it as root
	if !hasRootInstruction {
		agentsMDPath := filepath.Join(opts.ProjectRoot, "AGENTS.md")
		if data, err := os.ReadFile(agentsMDPath); err == nil {
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    "AGENTS.md",
				Content: string(data),
				Scope:   "root",
			})
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, agentsMDPath)
		}
	}

	// --- Agents (from JSON config) ---
	for name, agent := range oc.Agent {
		prompt := agent.Prompt

		// Resolve {file:...} references
		if strings.HasPrefix(agent.Prompt, "{file:") && strings.HasSuffix(agent.Prompt, "}") {
			fileRef := agent.Prompt[6 : len(agent.Prompt)-1]
			fullPath := filepath.Join(opts.ProjectRoot, fileRef)
			if data, err := os.ReadFile(fullPath); err == nil {
				// Strip frontmatter from .md file — JSON fields take precedence
				_, body, _ := format.ParseFrontMatter(string(data))
				if strings.TrimSpace(body) != "" {
					prompt = body
				} else {
					prompt = string(data)
				}
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, fullPath)
			}
		}

		canonAgent := canonical.AgentDef{
			Name:         name,
			Description:  agent.Description,
			Model:        agent.Model,
			SystemPrompt: prompt,
			Color:        agent.Color,
			Mode:         agent.Mode,
			Hidden:       agent.Hidden,
			Temperature:  agent.Temperature,
			MaxSteps:     agent.Steps,
		}

		// Parse tools
		if agent.Tools != nil {
			for tool, val := range agent.Tools {
				if enabled, ok := val.(bool); ok && enabled {
					canonAgent.Tools = append(canonAgent.Tools, tool)
				}
			}
		}

		// Parse agent-level permission overrides
		if agent.Permission != nil {
			canonAgent.PermissionOverrides = parseOpenCodePermissions(agent.Permission)
		}

		cfg.Agents = append(cfg.Agents, canonAgent)
		cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, configPath)
	}

	// --- Agents (from .opencode/agents/*.md files) ---
	agentsDir := filepath.Join(opts.ProjectRoot, ".opencode", "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			agentPath := filepath.Join(agentsDir, entry.Name())
			// Skip if already loaded from JSON config (by name)
			agentName := strings.TrimSuffix(entry.Name(), ".md")
			alreadyLoaded := false
			for _, a := range cfg.Agents {
				if a.Name == agentName {
					alreadyLoaded = true
					break
				}
			}
			if alreadyLoaded {
				continue
			}

			data, err := os.ReadFile(agentPath)
			if err != nil {
				continue
			}
			fm, body, err := format.ParseFrontMatter(string(data))
			if err != nil {
				continue
			}

			canonAgent := canonical.AgentDef{
				Name:         agentName,
				SystemPrompt: strings.TrimSpace(body),
			}
			if name, ok := fm["name"].(string); ok {
				canonAgent.Name = name
			}
			if desc, ok := fm["description"].(string); ok {
				canonAgent.Description = desc
			}
			if mode, ok := fm["mode"].(string); ok {
				canonAgent.Mode = mode
			}
			if temp, ok := fm["temperature"].(float64); ok {
				canonAgent.Temperature = temp
			}
			if color, ok := fm["color"].(string); ok {
				canonAgent.Color = color
			}
			if steps, ok := fm["steps"].(int); ok {
				canonAgent.MaxSteps = steps
			}
			if hidden, ok := fm["hidden"].(bool); ok {
				canonAgent.Hidden = hidden
			}
			if model, ok := fm["model"].(string); ok {
				canonAgent.Model = model
			}

			cfg.Agents = append(cfg.Agents, canonAgent)
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, agentPath)
		}
	}

	// --- MCP Servers ---
	mcpServers := oc.MCP
	if mcpServers == nil {
		mcpServers = oc.MCPServers
	}
	for name, srv := range mcpServers {
		mcpType := srv.Type
		if mcpType == "" {
			mcpType = "stdio"
		}
		cmd := ""
		if len(srv.Command) > 0 {
			cmd = srv.Command[0]
		}
		var args []string
		if len(srv.Command) > 1 {
			args = srv.Command[1:]
		}
		if len(srv.Args) > 0 {
			args = append(args, srv.Args...)
		}
		cfg.MCPServers = append(cfg.MCPServers, canonical.MCPServerDef{
			Name:    name,
			Type:    mcpType,
			Command: cmd,
			Args:    args,
			Env:     srv.Environment,
			URL:     srv.URL,
			Timeout: srv.Timeout,
			Enabled: true,
		})
	}

	// --- Permissions ---
	cfg.Permissions = parseOpenCodePermissions(oc.Permission)

	// --- Commands (from JSON) ---
	for name, cmd := range oc.Command {
		canonCmd := canonical.CommandDef{
			Name:        name,
			Description: cmd.Description,
		}
		if cmd.Agent != "" {
			canonCmd.ArgumentHint = "agent:" + cmd.Agent
		}
		cfg.Commands = append(cfg.Commands, canonCmd)
	}

	// --- Commands (from .opencode/commands/*.md) ---
	commandsDir := filepath.Join(opts.ProjectRoot, ".opencode", "commands")
	if entries, err := os.ReadDir(commandsDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			cmdPath := filepath.Join(commandsDir, entry.Name())
			data, err := os.ReadFile(cmdPath)
			if err != nil {
				continue
			}
			fm, body, err := format.ParseFrontMatter(string(data))
			if err != nil {
				// No frontmatter — use entire content as prompt
				cmdName := strings.TrimSuffix(entry.Name(), ".md")
				// Check if already loaded from JSON
				exists := false
				for _, c := range cfg.Commands {
					if c.Name == cmdName {
						exists = true
						break
					}
				}
				if !exists {
					cfg.Commands = append(cfg.Commands, canonical.CommandDef{
						Name:   cmdName,
						Prompt: strings.TrimSpace(string(data)),
					})
				}
				continue
			}

			cmdName := strings.TrimSuffix(entry.Name(), ".md")
			canonCmd := canonical.CommandDef{
				Name:   cmdName,
				Prompt: strings.TrimSpace(body),
			}
			if desc, ok := fm["description"].(string); ok {
				canonCmd.Description = desc
			}
			if agent, ok := fm["agent"].(string); ok {
				canonCmd.ArgumentHint = "agent:" + agent
			}

			// Merge with JSON-defined command if exists
			found := false
			for i, c := range cfg.Commands {
				if c.Name == cmdName {
					if canonCmd.Description != "" {
						cfg.Commands[i].Description = canonCmd.Description
					}
					if canonCmd.Prompt != "" {
						cfg.Commands[i].Prompt = canonCmd.Prompt
					}
					found = true
					break
				}
			}
			if !found {
				cfg.Commands = append(cfg.Commands, canonCmd)
			}
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, cmdPath)
		}
	}

	// --- Plugins ---
	if len(oc.Plugin) > 0 {
		if cfg.Settings.Extra == nil {
			cfg.Settings.Extra = make(map[string]interface{})
		}
		cfg.Settings.Extra["plugins"] = oc.Plugin
	}

	// --- Skills ---
	skillsDir := filepath.Join(opts.ProjectRoot, ".opencode", "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(skillsDir, entry.Name())
			skill := readOpenCodeSkill(skillPath)
			if skill.Name != "" {
				cfg.Skills = append(cfg.Skills, skill)
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, skillPath)
			}
		}
	}

	// --- Rules ---
	rulesDir := filepath.Join(opts.ProjectRoot, ".opencode", "rules")
	if entries, err := os.ReadDir(rulesDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			rulePath := filepath.Join(rulesDir, entry.Name())
			if data, err := os.ReadFile(rulePath); err == nil {
				cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
					Path:    filepath.Join(".opencode", "rules", entry.Name()),
					Content: string(data),
					Scope:   "named",
				})
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, rulePath)
			}
		}
	}

	// --- Hooks ---
	hooksDir := filepath.Join(opts.ProjectRoot, ".opencode", "hooks")
	if entries, err := os.ReadDir(hooksDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".sh") {
				continue
			}
			hookPath := filepath.Join(hooksDir, entry.Name())
			if _, err := os.Stat(hookPath); err == nil {
				name := strings.TrimSuffix(entry.Name(), ".sh")
				cfg.Hooks = append(cfg.Hooks, canonical.HookDef{
					Event:   "PostToolUse",
					Matcher: name,
					Command: hookPath,
				})
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, hookPath)
			}
		}
	}

	return cfg, nil
}

// ============================================================================
// Write
// ============================================================================

func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	oc := opencodeConfig{
		Schema: "https://opencode.ai/config.json",
	}

	// Model
	if cfg.Settings.Model != "" {
		oc.Model = cfg.Settings.Model
	}
	if cfg.Settings.SmallModel != "" {
		oc.SmallModel = cfg.Settings.SmallModel
	}
	if cfg.Settings.DefaultAgent != "" {
		oc.DefaultAgent = cfg.Settings.DefaultAgent
	}

	// Instructions → file paths
	for _, inst := range cfg.Instructions {
		oc.Instructions = append(oc.Instructions, inst.Path)
	}

	// Agents → JSON config
	if len(cfg.Agents) > 0 {
		oc.Agent = make(map[string]opencodeAgentDef)
		for _, agent := range cfg.Agents {
			oad := opencodeAgentDef{
				Description: agent.Description,
				Mode:        agent.Mode,
				Color:       agent.Color,
				Temperature: agent.Temperature,
				Steps:       agent.MaxSteps,
				Model:       agent.Model,
				Hidden:      agent.Hidden,
			}

			// Reference the prompt file if stored externally, or inline it
			if agent.SystemPrompt != "" {
				oad.Prompt = fmt.Sprintf("{file:./.opencode/agents/%s.md}", agent.Name)
			}

			// Tools
			if len(agent.Tools) > 0 {
				oad.Tools = make(map[string]interface{})
				for _, t := range agent.Tools {
					oad.Tools[strings.ToLower(t)] = true
				}
			}

			// Agent-level permissions
			if agent.PermissionOverrides != nil {
				oad.Permission = writeOpenCodePermissions(agent.PermissionOverrides)
			}

			oc.Agent[agent.Name] = oad
		}
	}

	// MCP servers
	if len(cfg.MCPServers) > 0 {
		oc.MCP = make(map[string]opencodeMCPServer)
		for _, mcp := range cfg.MCPServers {
			command := []string{mcp.Command}
			command = append(command, mcp.Args...)
			oc.MCP[mcp.Name] = opencodeMCPServer{
				Type:        mcp.Type,
				Command:     command,
				Environment: mcp.Env,
				URL:         mcp.URL,
				Timeout:     mcp.Timeout,
			}
		}
	}

	// Permissions
	if cfg.Permissions != nil {
		oc.Permission = writeOpenCodePermissions(cfg.Permissions)
	}

	// Plugins from Extra
	if cfg.Settings.Extra != nil {
		if plugins, ok := cfg.Settings.Extra["plugins"].([]interface{}); ok {
			for _, p := range plugins {
				if ps, ok := p.(string); ok {
					oc.Plugin = append(oc.Plugin, ps)
				}
			}
		}
		// Restore LSP, compaction, formatter, watcher
		if lsp, ok := cfg.Settings.Extra["lsp"].(bool); ok {
			oc.LSP = lsp
		}
		if compaction, ok := cfg.Settings.Extra["compaction"].(map[string]interface{}); ok {
			oc.Compaction = &opencodeCompaction{}
			if auto, ok := compaction["auto"].(bool); ok {
				oc.Compaction.Auto = auto
			}
			if prune, ok := compaction["prune"].(bool); ok {
				oc.Compaction.Prune = prune
			}
			if reserved, ok := compaction["reserved"].(float64); ok {
				oc.Compaction.Reserved = int(reserved)
			}
		}
	}

	// Write config file
	configPath := filepath.Join(opts.ProjectRoot, "opencode.jsonc")
	if !opts.DryRun {
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return result, fmt.Errorf("create dir: %w", err)
		}
		data, err := json.MarshalIndent(oc, "", "  ")
		if err != nil {
			return result, fmt.Errorf("marshal config: %w", err)
		}
		// Write as .jsonc with schema comment
		content := "{\n  \"$schema\": \"https://opencode.ai/config.json\",\n"
		content += string(data)[1:] // append rest of JSON
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			return result, fmt.Errorf("write config: %w", err)
		}
	}
	result.FilesWritten = append(result.FilesWritten, configPath)

	// Write agent files
	if len(cfg.Agents) > 0 {
		agentsDir := filepath.Join(opts.ProjectRoot, ".opencode", "agents")
		if !opts.DryRun {
			os.MkdirAll(agentsDir, 0755)
		}
		for _, agent := range cfg.Agents {
			agentPath := filepath.Join(agentsDir, agent.Name+".md")
			content := formatOpenCodeAgentFile(agent)
			if !opts.DryRun {
				os.WriteFile(agentPath, []byte(content), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, agentPath)
		}
	}

	// Write command files
	if len(cfg.Commands) > 0 {
		commandsDir := filepath.Join(opts.ProjectRoot, ".opencode", "commands")
		if !opts.DryRun {
			os.MkdirAll(commandsDir, 0755)
		}
		for _, cmd := range cfg.Commands {
			cmdPath := filepath.Join(commandsDir, cmd.Name+".md")
			content := formatOpenCodeCommandFile(cmd)
			if !opts.DryRun {
				os.WriteFile(cmdPath, []byte(content), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, cmdPath)
		}
	}

	// Write instruction files
	for _, inst := range cfg.Instructions {
		p := filepath.Join(opts.ProjectRoot, inst.Path)
		if !opts.DryRun {
			dir := filepath.Dir(p)
			os.MkdirAll(dir, 0755)
			os.WriteFile(p, []byte(inst.Content), 0644)
		}
		result.FilesWritten = append(result.FilesWritten, p)
	}

	// Loss warnings
	if len(cfg.Hooks) > 0 {
		result.LossWarnings = append(result.LossWarnings, convert.NewLossWarning("hooks", cfg.Meta.SourceAdapter, a.ID(), "OpenCode uses shell script hooks with RESULT protocol; hooks preserved as files in .opencode/hooks/", "warning"))

	}

	return result, nil
}

func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}

// ============================================================================
// Helpers
// ============================================================================

func (a *Adapter) findConfigFile(opts adapter.ReadOptions) string {
	home := paths.MustHomeDir()

	candidates := []string{
		filepath.Join(opts.ProjectRoot, "opencode.jsonc"),
		filepath.Join(opts.ProjectRoot, "opencode.json"),
		filepath.Join(opts.ProjectRoot, ".opencode", "opencode.jsonc"),
		filepath.Join(opts.ProjectRoot, ".opencode", "opencode.json"),
	}

	if opts.Scope == "global" || opts.Scope == "all" {
		candidates = append(candidates,
			filepath.Join(home, ".config", "opencode", "opencode.jsonc"),
			filepath.Join(home, ".config", "opencode", "opencode.json"),
		)
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// parseOpenCodePermissions converts OpenCode's nested permission map to canonical format.
func parseOpenCodePermissions(perm map[string]interface{}) *canonical.PermissionSet {
	if perm == nil {
		return nil
	}

	ps := &canonical.PermissionSet{}

	for key, val := range perm {
		switch v := val.(type) {
		case string:
			// Simple: "tool": "allow" | "ask" | "deny"
			if v == "allow" {
				ps.AllowRules = append(ps.AllowRules, canonical.PermissionRule{
					Pattern: key,
					Action:  "allow",
				})
			} else if v == "deny" {
				ps.DenyRules = append(ps.DenyRules, canonical.PermissionRule{
					Pattern: key,
					Action:  "deny",
				})
			} else {
				ps.AllowRules = append(ps.AllowRules, canonical.PermissionRule{
					Pattern: key,
					Action:  "ask",
				})
			}

		case map[string]interface{}:
			// Nested: "tool": { "*.env": "deny", "*.env.example": "allow" }
			for pattern, pval := range v {
				action := "allow"
				if ps, ok := pval.(string); ok {
					action = ps
				}
				rule := canonical.PermissionRule{
					Pattern: pattern,
					Tool:    key,
					Action:  action,
				}
				if action == "deny" {
					ps.DenyRules = append(ps.DenyRules, rule)
				} else {
					ps.AllowRules = append(ps.AllowRules, rule)
				}
			}
		}
	}

	return ps
}

// writeOpenCodePermissions converts canonical permissions to OpenCode format.
func writeOpenCodePermissions(ps *canonical.PermissionSet) map[string]interface{} {
	result := make(map[string]interface{})

	// Collect by tool
	toolPerms := make(map[string]map[string]string)
	for _, r := range ps.AllowRules {
		if r.Tool != "" {
			if toolPerms[r.Tool] == nil {
				toolPerms[r.Tool] = make(map[string]string)
			}
			toolPerms[r.Tool][r.Pattern] = "allow"
		}
	}
	for _, r := range ps.DenyRules {
		if r.Tool != "" {
			if toolPerms[r.Tool] == nil {
				toolPerms[r.Tool] = make(map[string]string)
			}
			toolPerms[r.Tool][r.Pattern] = "deny"
		}
	}

	for tool, patterns := range toolPerms {
		if len(patterns) == 1 && patterns["*"] != "" {
			result[tool] = patterns["*"]
		} else {
			nested := make(map[string]interface{})
			for p, a := range patterns {
				nested[p] = a
			}
			result[tool] = nested
		}
	}

	return result
}

func readOpenCodeSkill(dir string) canonical.SkillDef {
	skill := canonical.SkillDef{
		Scripts:    make(map[string]string),
		References: make(map[string]string),
	}

	skillMDPath := filepath.Join(dir, "SKILL.md")
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return skill
	}

	fm, body, err := format.ParseFrontMatter(string(content))
	if err == nil {
		if name, ok := fm["name"].(string); ok {
			skill.Name = name
		}
		if desc, ok := fm["description"].(string); ok {
			skill.Description = desc
		}
		skill.Markdown = strings.TrimSpace(body)
	}

	if skill.Name == "" {
		skill.Name = filepath.Base(dir)
	}
	if skill.Markdown == "" {
		skill.Markdown = strings.TrimSpace(string(content))
	}

	// Read supporting files
	for _, name := range []string{"REFERENCE.md", "EXAMPLES.md"} {
		if data, err := os.ReadFile(filepath.Join(dir, name)); err == nil {
			skill.References[name] = string(data)
		}
	}

	return skill
}

func formatOpenCodeAgentFile(agent canonical.AgentDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	if agent.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", agent.Description))
	}
	if agent.Mode != "" {
		b.WriteString(fmt.Sprintf("mode: %s\n", agent.Mode))
	}
	if agent.Temperature != 0 {
		b.WriteString(fmt.Sprintf("temperature: %.1f\n", agent.Temperature))
	}
	if agent.Color != "" {
		b.WriteString(fmt.Sprintf("color: %q\n", agent.Color))
	}
	if agent.MaxSteps > 0 {
		b.WriteString(fmt.Sprintf("steps: %d\n", agent.MaxSteps))
	}
	if agent.Hidden {
		b.WriteString("hidden: true\n")
	}
	if agent.Model != "" {
		b.WriteString(fmt.Sprintf("model: %s\n", agent.Model))
	}
	b.WriteString("---\n\n")
	b.WriteString(agent.SystemPrompt)
	if !strings.HasSuffix(agent.SystemPrompt, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func formatOpenCodeCommandFile(cmd canonical.CommandDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	if cmd.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", cmd.Description))
	}
	// Extract agent scope from argument hint
	if strings.HasPrefix(cmd.ArgumentHint, "agent:") {
		b.WriteString(fmt.Sprintf("agent: %s\n", strings.TrimPrefix(cmd.ArgumentHint, "agent:")))
	}
	b.WriteString("---\n\n")
	b.WriteString(cmd.Prompt)
	if !strings.HasSuffix(cmd.Prompt, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}
