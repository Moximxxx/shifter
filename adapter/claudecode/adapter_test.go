package claudecode

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadFixtures(t *testing.T) {
	// Find the fixtures directory relative to the project root
	fixtureDir := filepath.Join("..", "..", "fixtures", "claude-project")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	ctx := context.Background()

	cfg, err := a.Read(ctx, adapter.ReadOptions{
		Scope:       "project",
		ProjectRoot: fixtureDir,
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Verify agents
	if len(cfg.Agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(cfg.Agents))
	}
	agent := cfg.Agents[0]
	if agent.Name != "code-reviewer" {
		t.Errorf("agent name = %q, want %q", agent.Name, "code-reviewer")
	}
	if agent.Description != "Review code for quality, security, and best practices" {
		t.Errorf("agent description mismatch")
	}
	if len(agent.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d: %v", len(agent.Tools), agent.Tools)
	}
	if agent.Model != "sonnet" {
		t.Errorf("agent model = %q, want %q", agent.Model, "sonnet")
	}

	// Verify skills
	if len(cfg.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(cfg.Skills))
	}
	skill := cfg.Skills[0]
	if skill.Name != "pdf-processor" {
		t.Errorf("skill name = %q, want %q", skill.Name, "pdf-processor")
	}

	// Verify commands
	if len(cfg.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cfg.Commands))
	}
	cmd := cfg.Commands[0]
	if cmd.Name != "review" {
		t.Errorf("command name = %q, want %q", cmd.Name, "review")
	}
	if cmd.Description != "Review the current changes for quality and correctness" {
		t.Errorf("command description mismatch")
	}

	// Verify MCP servers
	if len(cfg.MCPServers) != 2 {
		t.Fatalf("expected 2 MCP servers, got %d", len(cfg.MCPServers))
	}
	mcpNames := make(map[string]bool)
	for _, m := range cfg.MCPServers {
		mcpNames[m.Name] = true
		if m.Command == "" {
			t.Errorf("MCP server %s has no command", m.Name)
		}
	}
	if !mcpNames["github"] || !mcpNames["filesystem"] {
		t.Errorf("expected MCP servers 'github' and 'filesystem', got %v", mcpNames)
	}

	// Verify permissions
	if cfg.Permissions == nil {
		t.Fatal("permissions should not be nil")
	}
	if len(cfg.Permissions.AllowRules) != 4 {
		t.Errorf("expected 4 allow rules, got %d", len(cfg.Permissions.AllowRules))
	}
	if len(cfg.Permissions.DenyRules) != 3 {
		t.Errorf("expected 3 deny rules, got %d", len(cfg.Permissions.DenyRules))
	}

	// Verify hooks
	if len(cfg.Hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(cfg.Hooks))
	}

	// Verify instructions (CLAUDE.md)
	if len(cfg.Instructions) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(cfg.Instructions))
	}
	if cfg.Instructions[0].Content == "" {
		t.Error("instruction content should not be empty")
	}

	// Verify settings
	if cfg.Settings.Model != "sonnet" {
		t.Errorf("settings model = %q, want %q", cfg.Settings.Model, "sonnet")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "fixtures", "claude-project")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	ctx := context.Background()

	// Read from fixture
	cfg, err := a.Read(ctx, adapter.ReadOptions{
		Scope:       "project",
		ProjectRoot: fixtureDir,
	})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Write to temp dir
	tmpDir := t.TempDir()
	result, err := a.Write(ctx, cfg, adapter.WriteOptions{
		Scope:       "project",
		ProjectRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify files were written
	expectedFiles := []string{
		filepath.Join(tmpDir, ".claude", "settings.json"),
		filepath.Join(tmpDir, "CLAUDE.md"),
		filepath.Join(tmpDir, ".claude", "agents", "code-reviewer.md"),
		filepath.Join(tmpDir, ".claude", "skills", "pdf-processor", "SKILL.md"),
		filepath.Join(tmpDir, ".claude", "commands", "review.md"),
	}
	for _, f := range expectedFiles {
		found := false
		for _, wf := range result.FilesWritten {
			if wf == f {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected file %q to be written", f)
		}
	}

	// Read back and verify
	cfg2, err := a.Read(ctx, adapter.ReadOptions{
		Scope:       "project",
		ProjectRoot: tmpDir,
	})
	if err != nil {
		t.Fatalf("Read-back failed: %v", err)
	}

	if len(cfg2.Agents) != len(cfg.Agents) {
		t.Errorf("agent count mismatch: read %d, wrote %d", len(cfg2.Agents), len(cfg.Agents))
	}
	if len(cfg2.MCPServers) != len(cfg.MCPServers) {
		t.Errorf("MCP server count mismatch: read %d, wrote %d", len(cfg2.MCPServers), len(cfg.MCPServers))
	}
	if len(cfg2.Skills) != len(cfg.Skills) {
		t.Errorf("skill count mismatch: read %d, wrote %d", len(cfg2.Skills), len(cfg.Skills))
	}
	if len(cfg2.Commands) != len(cfg.Commands) {
		t.Errorf("command count mismatch: read %d, wrote %d", len(cfg2.Commands), len(cfg.Commands))
	}
}

func TestDetect(t *testing.T) {
	a := New()
	result, err := a.Detect()
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	// May or may not be found depending on system, just ensure no panic
	t.Logf("Detect result: found=%v, summary=%v", result.Found, result.Summary)
}
