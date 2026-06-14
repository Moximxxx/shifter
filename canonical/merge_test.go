package canonical

import (
	"strings"
	"testing"
)

func TestMergeAgents_Newer(t *testing.T) {
	a := []AgentDef{
		{Name: "coordinator", Description: "Old desc", SystemPrompt: "Old prompt"},
		{Name: "analyzer", Description: "Analyzer A", SystemPrompt: "Prompt A"},
	}
	b := []AgentDef{
		{Name: "coordinator", Description: "New desc", SystemPrompt: "New prompt"},
		{Name: "builder", Description: "Builder B", SystemPrompt: "Prompt B"},
	}

	result := MergeConfigs(
		&ShifterConfig{Agents: a},
		&ShifterConfig{Agents: b},
		MergeNewer,
	)

	if len(result.Config.Agents) != 3 {
		t.Fatalf("expected 3 agents (union), got %d", len(result.Config.Agents))
	}

	// Find coordinator — should use side B (newer)
	for _, agent := range result.Config.Agents {
		if agent.Name == "coordinator" {
			if agent.Description != "New desc" {
				t.Errorf("newer strategy: expected 'New desc', got %q", agent.Description)
			}
		}
	}
}

func TestMergeAgents_Source(t *testing.T) {
	a := []AgentDef{{Name: "coordinator", Description: "Source desc"}}
	b := []AgentDef{{Name: "coordinator", Description: "Target desc"}}

	result := MergeConfigs(
		&ShifterConfig{Agents: a},
		&ShifterConfig{Agents: b},
		MergePreferSource,
	)

	for _, agent := range result.Config.Agents {
		if agent.Name == "coordinator" && agent.Description != "Source desc" {
			t.Errorf("source strategy: expected 'Source desc', got %q", agent.Description)
		}
	}
}

func TestMergeInstructions(t *testing.T) {
	a := []Instruction{
		{Path: "CLAUDE.md", Content: "Content A", Scope: "root"},
	}
	b := []Instruction{
		{Path: "CLAUDE.md", Content: "Content B", Scope: "root"},
		{Path: "GEMINI.md", Content: "Gemini content", Scope: "root"},
	}

	result := MergeConfigs(
		&ShifterConfig{Instructions: a},
		&ShifterConfig{Instructions: b},
		MergeUnion,
	)

	if len(result.Config.Instructions) != 2 {
		t.Fatalf("expected 2 instructions, got %d", len(result.Config.Instructions))
	}
	// CLAUDE.md should be merged
	for _, inst := range result.Config.Instructions {
		if inst.Path == "CLAUDE.md" && inst.Content == "Content A" {
			t.Error("union strategy: CLAUDE.md should have merged content")
		}
	}
}

func TestMergeMCPServers(t *testing.T) {
	a := []MCPServerDef{
		{Name: "github", Command: "npx", Args: []string{"-y", "old"}, Env: map[string]string{"OLD": "1"}},
	}
	b := []MCPServerDef{
		{Name: "github", Command: "npx", Args: []string{"-y", "new"}, Env: map[string]string{"NEW": "1"}},
		{Name: "postgres", Command: "docker", Args: []string{"run", "postgres"}},
	}

	result := MergeConfigs(
		&ShifterConfig{MCPServers: a},
		&ShifterConfig{MCPServers: b},
		MergeNewer,
	)

	if len(result.Config.MCPServers) != 2 {
		t.Fatalf("expected 2 MCP servers, got %d", len(result.Config.MCPServers))
	}
	for _, mcp := range result.Config.MCPServers {
		if mcp.Name == "github" && mcp.Args[1] != "new" {
			t.Errorf("newer strategy: expected 'new' arg, got %q", mcp.Args[1])
		}
	}
}

func TestMergePermissions(t *testing.T) {
	a := &PermissionSet{
		DefaultMode: "edit",
		AllowRules:  []PermissionRule{{Pattern: "Read(*)", Action: "allow"}},
		DenyRules:   []PermissionRule{{Pattern: "Bash(rm *)", Action: "deny"}},
	}
	b := &PermissionSet{
		DefaultMode: "ask",
		AllowRules:  []PermissionRule{{Pattern: "Bash(npm:*)", Action: "allow"}},
	}

	result := MergeConfigs(
		&ShifterConfig{Permissions: a},
		&ShifterConfig{Permissions: b},
		MergeNewer,
	)

	merged := result.Config.Permissions
	if merged.DefaultMode != "ask" {
		t.Errorf("newer default mode: got %q, want %q", merged.DefaultMode, "ask")
	}
	if len(merged.AllowRules) != 2 {
		t.Errorf("expected 2 allow rules (union), got %d", len(merged.AllowRules))
	}
	if len(merged.DenyRules) != 1 {
		t.Errorf("expected 1 deny rule, got %d", len(merged.DenyRules))
	}
}

func TestMergeHooks(t *testing.T) {
	a := []HookDef{
		{Event: "PostToolUse", Matcher: "Write", Command: "prettier"},
	}
	b := []HookDef{
		{Event: "PostToolUse", Matcher: "Write", Command: "prettier"}, // duplicate
		{Event: "PreToolUse", Matcher: "Bash", Command: "guard"},
	}

	result := MergeConfigs(
		&ShifterConfig{Hooks: a},
		&ShifterConfig{Hooks: b},
		MergeUnion,
	)

	if len(result.Config.Hooks) != 2 {
		t.Fatalf("expected 2 hooks (deduplicated), got %d", len(result.Config.Hooks))
	}
}

func TestMergeSettings(t *testing.T) {
	a := SettingsMap{Model: "sonnet", ApprovalMode: "edit"}
	b := SettingsMap{Model: "opus", SandboxMode: "workspace-write", Extra: map[string]interface{}{"lsp": true}}

	result := MergeConfigs(
		&ShifterConfig{Settings: a},
		&ShifterConfig{Settings: b},
		MergeNewer,
	)

	s := result.Config.Settings
	if s.Model != "opus" {
		t.Errorf("model: got %q, want %q", s.Model, "opus")
	}
	if s.ApprovalMode != "edit" {
		t.Errorf("approval mode should be preserved from a: got %q", s.ApprovalMode)
	}
	if s.SandboxMode != "workspace-write" {
		t.Errorf("sandbox mode: got %q", s.SandboxMode)
	}
	if s.Extra == nil || s.Extra["lsp"] != true {
		t.Errorf("extra fields not merged: %v", s.Extra)
	}
}

func TestMergeNilPermissions(t *testing.T) {
	b := &PermissionSet{DefaultMode: "ask", AllowRules: []PermissionRule{{Pattern: "*", Action: "allow"}}}

	result := MergeConfigs(
		&ShifterConfig{Permissions: nil},
		&ShifterConfig{Permissions: b},
		MergeNewer,
	)
	if result.Config.Permissions == nil || result.Config.Permissions.DefaultMode != "ask" {
		t.Error("merge with nil a: should use side B")
	}

	result2 := MergeConfigs(
		&ShifterConfig{Permissions: b},
		&ShifterConfig{Permissions: nil},
		MergeNewer,
	)
	if result2.Config.Permissions == nil || result2.Config.Permissions.DefaultMode != "ask" {
		t.Error("merge with nil b: should use side A")
	}
}

func TestMergeEmptyConfigs(t *testing.T) {
	result := MergeConfigs(
		&ShifterConfig{Meta: ConfigMeta{SourceAdapter: "claude-code"}},
		&ShifterConfig{Meta: ConfigMeta{SourceAdapter: "codex"}},
		MergeNewer,
	)
	if result.Config.Meta.SourceAdapter != "merged" {
		t.Errorf("meta source adapter: got %q", result.Config.Meta.SourceAdapter)
	}
	if len(result.Config.Agents) != 0 {
		t.Error("empty merge should have no agents")
	}
	if len(result.Config.MCPServers) != 0 {
		t.Error("empty merge should have no MCP servers")
	}
}

func TestMergeSkills(t *testing.T) {
	a := []SkillDef{
		{Name: "old-skill", Description: "Old", Markdown: "Old body"},
	}
	b := []SkillDef{
		{Name: "old-skill", Description: "New", Markdown: "New body"},
		{Name: "new-skill", Description: "New skill", Markdown: "New skill body"},
	}

	result := MergeConfigs(
		&ShifterConfig{Skills: a},
		&ShifterConfig{Skills: b},
		MergeNewer,
	)

	if len(result.Config.Skills) != 2 {
		t.Fatalf("skills: got %d, want 2", len(result.Config.Skills))
	}
	for _, sk := range result.Config.Skills {
		if sk.Name == "old-skill" && sk.Description != "New" {
			t.Errorf("newer strategy: skill should be updated, got %q", sk.Description)
		}
	}
}

func TestMergeCommands(t *testing.T) {
	a := []CommandDef{
		{Name: "review", Description: "Old desc", Prompt: "Old prompt"},
	}
	b := []CommandDef{
		{Name: "review", Description: "New desc", Prompt: "New prompt"},
		{Name: "deploy", Description: "Deploy command", Prompt: "Deploy prompt"},
	}

	result := MergeConfigs(
		&ShifterConfig{Commands: a},
		&ShifterConfig{Commands: b},
		MergeNewer,
	)

	if len(result.Config.Commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(result.Config.Commands))
	}
	for _, cmd := range result.Config.Commands {
		if cmd.Name == "review" && cmd.Description != "New desc" {
			t.Errorf("newer strategy: command should be updated, got %q", cmd.Description)
		}
	}
}

func TestMergeUnion_Instructions(t *testing.T) {
	a := []Instruction{
		{Path: "shared.md", Content: "Content from A", Scope: "root"},
	}
	b := []Instruction{
		{Path: "shared.md", Content: "Content from B", Scope: "root"},
	}

	result := MergeConfigs(
		&ShifterConfig{Instructions: a},
		&ShifterConfig{Instructions: b},
		MergeUnion,
	)

	for _, inst := range result.Config.Instructions {
		if inst.Path == "shared.md" {
			if !strings.Contains(inst.Content, "Content from A") || !strings.Contains(inst.Content, "Content from B") {
				t.Error("union strategy: should concatenate contents")
			}
		}
	}
}

func TestMergeStringSlices(t *testing.T) {
	result := mergeStringSlices([]string{"a", "b", "c"}, []string{"b", "c", "d"})
	if len(result) != 4 {
		t.Errorf("expected 4 unique, got %d: %v", len(result), result)
	}
	// Should be sorted
	for i := 1; i < len(result); i++ {
		if result[i] < result[i-1] {
			t.Errorf("not sorted: %v", result)
		}
	}
}
