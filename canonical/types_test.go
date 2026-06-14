package canonical

import (
	"encoding/json"
	"testing"
	"time"
)

func TestShifterConfig_JSON_RoundTrip(t *testing.T) {
	cfg := &ShifterConfig{
		Meta: ConfigMeta{
			SourceAdapter: "claude-code",
			GeneratedAt:   time.Now(),
			Version:       "1.0.0",
			SourcePaths:   []string{".claude/settings.json"},
		},
		Instructions: []Instruction{
			{Path: "CLAUDE.md", Content: "# Test", Scope: "root"},
		},
		Agents: []AgentDef{
			{
				Name: "code-reviewer", Description: "Review code",
				Model: "sonnet", Tools: []string{"Read", "Grep"},
				SystemPrompt: "You are a reviewer.", Color: "#FF0000",
				Mode: "subagent", Temperature: 0.2, MaxSteps: 15, Hidden: false,
			},
		},
		Skills: []SkillDef{
			{Name: "test-skill", Description: "A test skill", Markdown: "## Skill"},
		},
		Commands: []CommandDef{
			{Name: "review", Description: "Review changes", Prompt: "Review: $ARGUMENTS", ArgumentHint: "[file]"},
		},
		MCPServers: []MCPServerDef{
			{Name: "github", Type: "stdio", Command: "npx", Args: []string{"-y", "mcp-server"}, Enabled: true},
		},
		Permissions: &PermissionSet{
			DefaultMode: "edit",
			AllowRules:  []PermissionRule{{Pattern: "Bash(npm:*)", Action: "allow"}},
			DenyRules:   []PermissionRule{{Pattern: "Bash(rm *)", Action: "deny"}},
		},
		Hooks: []HookDef{
			{Event: "PostToolUse", Matcher: "Write", Command: "prettier --write"},
		},
		Settings: SettingsMap{
			Model: "sonnet", ApprovalMode: "edit", SandboxMode: "workspace-write",
			DefaultAgent: "coordinator", SmallModel: "haiku",
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded ShifterConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Agents) != 1 || decoded.Agents[0].Name != "code-reviewer" {
		t.Errorf("agent mismatch after round-trip")
	}
	if len(decoded.MCPServers) != 1 || decoded.MCPServers[0].Name != "github" {
		t.Errorf("mcp mismatch after round-trip")
	}
	if decoded.Permissions == nil || len(decoded.Permissions.AllowRules) != 1 {
		t.Errorf("permissions mismatch after round-trip")
	}
	if decoded.Settings.DefaultAgent != "coordinator" {
		t.Errorf("settings mismatch: %+v", decoded.Settings)
	}
}

func TestAgentDef_AllFields(t *testing.T) {
	agent := AgentDef{
		Name: "coordinator", Description: "Primary coordinator",
		Model: "sonnet", Tools: []string{"Read", "Write", "Bash"},
		SystemPrompt: "You are the coordinator.", Color: "#7C3AED",
		Mode: "primary", Hidden: false, Temperature: 0.2, MaxSteps: 0,
		Skills: []string{"git-commit", "code-review"},
		MCPServers: []string{"github", "postgres"},
		PermissionOverrides: &PermissionSet{
			DefaultMode: "edit",
			AllowRules:  []PermissionRule{{Pattern: "*.md", Tool: "edit", Action: "allow"}},
		},
	}

	data, _ := json.Marshal(agent)
	var decoded AgentDef
	json.Unmarshal(data, &decoded)

	if decoded.Name != agent.Name {
		t.Errorf("name: got %q, want %q", decoded.Name, agent.Name)
	}
	if decoded.Temperature != agent.Temperature {
		t.Errorf("temperature: got %v, want %v", decoded.Temperature, agent.Temperature)
	}
	if decoded.MaxSteps != agent.MaxSteps {
		t.Errorf("maxSteps: got %d, want %d", decoded.MaxSteps, agent.MaxSteps)
	}
	if decoded.Hidden != agent.Hidden {
		t.Errorf("hidden: got %v, want %v", decoded.Hidden, agent.Hidden)
	}
	if len(decoded.Skills) != 2 {
		t.Errorf("skills: got %d, want 2", len(decoded.Skills))
	}
	if decoded.PermissionOverrides == nil {
		t.Fatal("permission overrides should not be nil")
	}
}

func TestInstruction_AllFields(t *testing.T) {
	inst := Instruction{Path: "CLAUDE.md", Content: "# Test\n\nContent.", Scope: "root"}
	data, _ := json.Marshal(inst)
	var decoded Instruction
	json.Unmarshal(data, &decoded)
	if decoded.Path != "CLAUDE.md" || decoded.Scope != "root" || decoded.Content == "" {
		t.Errorf("instruction round-trip failed: %+v", decoded)
	}
}

func TestSkillDef_WithSupportingFiles(t *testing.T) {
	skill := SkillDef{
		Name: "test-skill", Description: "Test",
		AllowedTools: []string{"Read", "Bash"},
		Markdown: "## Skill Body",
		Scripts:  map[string]string{"run.sh": "#!/bin/bash\necho hi"},
		References: map[string]string{"REFERENCE.md": "# Ref"},
		Examples:   map[string]string{"EXAMPLES.md": "# Examples"},
		Templates:  map[string]string{"tmpl.txt": "Hello {name}"},
	}
	data, _ := json.Marshal(skill)
	var decoded SkillDef
	json.Unmarshal(data, &decoded)
	if decoded.Name != "test-skill" {
		t.Errorf("name mismatch")
	}
	if len(decoded.Scripts) != 1 || decoded.Scripts["run.sh"] == "" {
		t.Errorf("scripts mismatch: %v", decoded.Scripts)
	}
	if len(decoded.References) != 1 {
		t.Errorf("references mismatch: %v", decoded.References)
	}
}

func TestConfigMeta(t *testing.T) {
	now := time.Now()
	meta := ConfigMeta{
		SourceAdapter: "opencode",
		GeneratedAt:   now,
		Version:       "1.0.0",
		SourcePaths:   []string{"a.jsonc", "b.jsonc"},
	}
	data, _ := json.Marshal(meta)
	var decoded ConfigMeta
	json.Unmarshal(data, &decoded)
	if decoded.SourceAdapter != "opencode" {
		t.Errorf("source adapter mismatch")
	}
	if len(decoded.SourcePaths) != 2 {
		t.Errorf("source paths mismatch: %d", len(decoded.SourcePaths))
	}
}

func TestLossWarning(t *testing.T) {
	w := LossWarning{
		Feature: "hooks", Field: "PostToolUse",
		SourceAgent: "claude-code", TargetAgent: "qoder",
		Reason: "Qoder has no hook system", Severity: "warning",
	}
	if w.Severity != "warning" {
		t.Errorf("severity: got %q, want %q", w.Severity, "warning")
	}
}

func TestAddLoss(t *testing.T) {
	cfg := &ShifterConfig{}
	cfg.AddLoss(LossWarning{Feature: "test", Severity: "info"})
	cfg.AddLosses([]LossWarning{
		{Feature: "test2", Severity: "warning"},
		{Feature: "test3", Severity: "critical"},
	})
	if len(cfg.LossWarnings) != 3 {
		t.Errorf("expected 3 loss warnings, got %d", len(cfg.LossWarnings))
	}
}

func TestCapabilityMatrix(t *testing.T) {
	m := CapabilityMatrix{
		SupportsInstructions: true,
		SupportsAgents:       true,
		SupportsMCP:          true,
		Notes: map[string]string{"hooks": "Use plugins instead"},
	}
	if !m.SupportsInstructions || !m.SupportsAgents {
		t.Error("capability flags not set correctly")
	}
	if m.SupportsSkills {
		t.Error("SupportsSkills should default to false")
	}
	if m.Notes["hooks"] == "" {
		t.Error("notes not preserved")
	}
}

func TestSettingsMap_Extra(t *testing.T) {
	s := SettingsMap{
		Model: "sonnet", SmallModel: "haiku",
		MaxTurns: 50, ReasoningEffort: "high",
		ApprovalMode: "edit", SandboxMode: "workspace-write",
		Theme: "dark", VimMode: true, DarkMode: true,
		AutoUpdate: true, AutoCommit: false, Personality: "pragmatic",
		WebSearch: "cached", DefaultAgent: "coordinator",
		Extra: map[string]interface{}{
			"lsp":        true,
			"plugins":    []string{"plugin-a.ts", "plugin-b.ts"},
			"compaction": map[string]interface{}{"auto": true, "prune": false},
		},
	}
	data, _ := json.Marshal(s)
	var decoded SettingsMap
	json.Unmarshal(data, &decoded)
	if decoded.Model != "sonnet" || decoded.SmallModel != "haiku" {
		t.Errorf("model fields mismatch")
	}
	if decoded.VimMode != true || decoded.DarkMode != true {
		t.Errorf("bool fields mismatch")
	}
	if decoded.Extra == nil || decoded.Extra["lsp"] != true {
		t.Errorf("extra fields mismatch: %v", decoded.Extra)
	}
}
