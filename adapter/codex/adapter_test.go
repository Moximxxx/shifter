package codex

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "codex-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	ctx := context.Background()
	cfg, err := a.Read(ctx, adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(cfg.Agents) != 3 {
		t.Fatalf("agents: got %d, want 3", len(cfg.Agents))
	}
	names := map[string]bool{}
	for _, ag := range cfg.Agents {
		names[ag.Name] = true
		if ag.SystemPrompt == "" {
			t.Errorf("agent %s has empty system prompt", ag.Name)
		}
	}
	for _, want := range []string{"code-reviewer", "test-writer", "architect"} {
		if !names[want] {
			t.Errorf("missing agent: %s", want)
		}
	}

	if len(cfg.MCPServers) != 4 {
		t.Fatalf("mcp servers: got %d, want 4", len(cfg.MCPServers))
	}

	if len(cfg.Hooks) != 5 {
		t.Fatalf("hooks: got %d, want 5", len(cfg.Hooks))
	}

	if len(cfg.Instructions) != 1 {
		t.Errorf("instructions: got %d, want 1", len(cfg.Instructions))
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "codex-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	ctx := context.Background()
	cfg, err := a.Read(ctx, adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	tmpDir := t.TempDir()
	result, err := a.Write(ctx, cfg, adapter.WriteOptions{Scope: "project", ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.toml not created")
	}

	codexMDPath := filepath.Join(tmpDir, ".codex", "codex.md")
	if _, err := os.Stat(codexMDPath); os.IsNotExist(err) {
		t.Error("codex.md not created")
	}

	if len(result.FilesWritten) != 2 {
		t.Errorf("files: got %d, want 2", len(result.FilesWritten))
	}

	// Read back
	cfg2, err := a.Read(ctx, adapter.ReadOptions{Scope: "project", ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("Read-back: %v", err)
	}
	if len(cfg2.Agents) != len(cfg.Agents) {
		t.Errorf("round-trip agents: got %d, want %d", len(cfg2.Agents), len(cfg.Agents))
	}
	if len(cfg2.MCPServers) != len(cfg.MCPServers) {
		t.Errorf("round-trip mcp: got %d, want %d", len(cfg2.MCPServers), len(cfg.MCPServers))
	}
}

func TestDetect(t *testing.T) {
	a := New()
	result, err := a.Detect()
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	_ = result.Found // may or may not be configured
}

func TestCapabilities(t *testing.T) {
	a := New()
	caps := a.Capabilities()
	if !caps.SupportsAgents {
		t.Error("should support agents")
	}
	if !caps.SupportsMCP {
		t.Error("should support MCP")
	}
	if !caps.SupportsHooks {
		t.Error("should support hooks")
	}
	if caps.SupportsCommands {
		t.Error("codex should not natively support commands")
	}
}

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "codex" {
		t.Errorf("ID: got %q", a.ID())
	}
	if a.Name() == "" {
		t.Error("Name is empty")
	}
}
