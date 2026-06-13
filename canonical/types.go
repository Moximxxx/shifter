// Package canonical defines the universal intermediate representation
// that bridges configuration between different coding agents.
//
// The canonical model is a SUPERSET of all agent features — each adapter
// projects down to what it supports. Loss tracking is first-class:
// every adapter operation accumulates LossWarnings so users always know
// what could not be perfectly ported.
//
// This model is INTERNAL-ONLY. Users never write to it directly.
package canonical

import "time"

// ShifterConfig is the universal configuration model.
type ShifterConfig struct {
	Meta         ConfigMeta    `json:"_meta"`
	Instructions []Instruction `json:"instructions,omitempty"`
	Agents       []AgentDef    `json:"agents,omitempty"`
	Skills       []SkillDef    `json:"skills,omitempty"`
	Commands     []CommandDef  `json:"commands,omitempty"`
	MCPServers   []MCPServerDef `json:"mcp_servers,omitempty"`
	Permissions  *PermissionSet `json:"permissions,omitempty"`
	Hooks        []HookDef     `json:"hooks,omitempty"`
	Settings     SettingsMap   `json:"settings,omitempty"`
	Files        []FileDef     `json:"files,omitempty"`
	LossWarnings []LossWarning `json:"_loss_warnings,omitempty"`
}

// AddLoss appends a loss warning.
func (c *ShifterConfig) AddLoss(w LossWarning) {
	c.LossWarnings = append(c.LossWarnings, w)
}

// AddLosses appends multiple loss warnings.
func (c *ShifterConfig) AddLosses(ws []LossWarning) {
	c.LossWarnings = append(c.LossWarnings, ws...)
}

// ConfigMeta holds metadata about the configuration source.
type ConfigMeta struct {
	SourceAdapter string    `json:"source_adapter"`
	GeneratedAt   time.Time `json:"generated_at"`
	Version       string    `json:"version"`
	SourcePaths   []string  `json:"source_paths"`
}

// Instruction represents a project instruction file
// (e.g., CLAUDE.md, GEMINI.md, .clinerules).
type Instruction struct {
	Path    string `json:"path"`    // relative path like "CLAUDE.md"
	Content string `json:"content"` // markdown body
	Scope   string `json:"scope"`   // "root" | "named" — root is always loaded
}

// AgentDef represents a custom subagent definition.
type AgentDef struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        string   `json:"model,omitempty"`
	Tools        []string `json:"tools,omitempty"`
	SystemPrompt string   `json:"system_prompt"`
	Color        string   `json:"color,omitempty"`
	Mode         string   `json:"mode,omitempty"`         // "subagent" | "primary" | "all"
	Hidden       bool     `json:"hidden,omitempty"`       // hide from @ mention menu
	Temperature  float64  `json:"temperature,omitempty"`  // LLM temperature
	MaxSteps     int      `json:"max_steps,omitempty"`    // max steps before return
	Skills       []string `json:"skills,omitempty"`       // names of skills to load
	MCPServers   []string `json:"mcp_servers,omitempty"`  // names of MCP servers
	PermissionOverrides *PermissionSet `json:"permission_overrides,omitempty"` // agent-level overrides
}

// SkillDef represents a packaged skill (SKILL.md + supporting files).
type SkillDef struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	AllowedTools []string          `json:"allowed_tools,omitempty"`
	DisallowedTools []string       `json:"disallowed_tools,omitempty"`
	Markdown     string            `json:"markdown"`       // SKILL.md body
	Scripts      map[string]string `json:"scripts,omitempty"`    // filename → content
	References   map[string]string `json:"references,omitempty"` // filename → content
	Examples     map[string]string `json:"examples,omitempty"`   // filename → content
	Templates    map[string]string `json:"templates,omitempty"`  // filename → content
}

// CommandDef represents a user-defined slash command.
type CommandDef struct {
	Name         string       `json:"name"`
	Description  string       `json:"description,omitempty"`
	Prompt       string       `json:"prompt"`        // template/instructions
	ArgumentHint string       `json:"argument_hint,omitempty"`
	AllowedTools []string     `json:"allowed_tools,omitempty"`
	Args         []CommandArg `json:"args,omitempty"`
}

// CommandArg defines a parameter for a slash command.
type CommandArg struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// MCPServerDef represents an MCP server connection configuration.
type MCPServerDef struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"` // "stdio" | "http" | "sse"
	Command        string            `json:"command,omitempty"`
	Args           []string          `json:"args,omitempty"`
	URL            string            `json:"url,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	Timeout        int               `json:"timeout,omitempty"` // seconds
	Enabled        bool              `json:"enabled"`
	EnabledTools   []string          `json:"enabled_tools,omitempty"`
	DisabledTools  []string          `json:"disabled_tools,omitempty"`
	Trust          *bool             `json:"trust,omitempty"`
}

// PermissionSet groups allow and deny rules.
type PermissionSet struct {
	AllowRules []PermissionRule `json:"allow,omitempty"`
	DenyRules  []PermissionRule `json:"deny,omitempty"`
	DefaultMode string          `json:"default_mode,omitempty"` // "edit" | "ask" | "auto_edit"
}

// PermissionRule is a single allow/deny entry.
type PermissionRule struct {
	Pattern string `json:"pattern"` // glob or regex pattern
	Tool    string `json:"tool,omitempty"`  // "Bash", "Write", "Read", etc.
	Action  string `json:"action"` // "allow" | "deny" | "ask"
}

// HookDef represents a lifecycle event hook.
type HookDef struct {
	Event     string   `json:"event"`   // "PreToolUse" | "PostToolUse" | "SessionStart" | etc.
	Matcher   string   `json:"matcher,omitempty"` // tool name or regex
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Timeout   int      `json:"timeout,omitempty"` // seconds
	StatusMsg string   `json:"status_msg,omitempty"`
}

// SettingsMap holds general agent settings that don't fit in other categories.
type SettingsMap struct {
	// Model
	Model           string `json:"model,omitempty"`
	SmallModel      string `json:"small_model,omitempty"`
	MaxTurns        int    `json:"max_turns,omitempty"`
	ReasoningEffort string `json:"reasoning_effort,omitempty"` // "low" | "medium" | "high"

	// Safety & Approval
	ApprovalMode string `json:"approval_mode,omitempty"` // "default" | "auto_edit" | "on_request" | "never"
	SandboxMode  string `json:"sandbox_mode,omitempty"`  // "read-only" | "workspace-write" | "danger-full-access"

	// UI
	Theme    string `json:"theme,omitempty"`
	VimMode  bool   `json:"vim_mode,omitempty"`
	DarkMode bool   `json:"dark_mode,omitempty"`

	// Behavior
	AutoUpdate   bool   `json:"auto_update,omitempty"`
	AutoCommit   bool   `json:"auto_commit,omitempty"`
	Personality  string `json:"personality,omitempty"` // "friendly" | "pragmatic" | "none"
	DefaultAgent string `json:"default_agent,omitempty"` // default primary agent name

	// Web Search
	WebSearch string `json:"web_search,omitempty"` // "cached" | "live" | "disabled"

	// Agent-specific extras (escape hatch)
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// FileDef represents a raw config file that doesn't fit structured categories.
// This is the escape hatch — if an agent has a config file with no canonical
// equivalent, it's preserved here with annotations.
type FileDef struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Format  string `json:"format"` // "json" | "toml" | "yaml" | "markdown" | "text"
}

// LossWarning describes a feature that could not be perfectly ported.
type LossWarning struct {
	Feature     string `json:"feature"`      // what was lost
	Field       string `json:"field"`        // which specific field
	SourceAgent string `json:"source_agent"`
	TargetAgent string `json:"target_agent"`
	Reason      string `json:"reason"`       // human explanation
	Severity    string `json:"severity"`     // "info" | "warning" | "critical"
}

// CapabilityMatrix describes what an agent adapter supports.
type CapabilityMatrix struct {
	SupportsInstructions bool
	SupportsAgents       bool
	SupportsSkills       bool
	SupportsCommands     bool
	SupportsMCP          bool
	SupportsPermissions  bool
	SupportsHooks        bool
	SupportsSettings     bool
	Notes                map[string]string // feature → limitation note
}
