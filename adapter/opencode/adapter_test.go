package opencode

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadRealProject(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "opencode-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	cfg, err := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(cfg.Agents) != 7 {
		t.Fatalf("agents: got %d, want 7", len(cfg.Agents))
	}
	if len(cfg.Skills) != 16 {
		t.Fatalf("skills: got %d, want 16", len(cfg.Skills))
	}
	if len(cfg.Commands) != 6 {
		t.Fatalf("commands: got %d, want 6", len(cfg.Commands))
	}
	if len(cfg.Instructions) < 19 {
		t.Fatalf("instructions: got %d, want at least 19", len(cfg.Instructions))
	}
	if len(cfg.Hooks) != 4 {
		t.Fatalf("hooks: got %d, want 4", len(cfg.Hooks))
	}

	// Check agent details
	foundCoordinator := false
	for _, ag := range cfg.Agents {
		if ag.Name == "coordinator" {
			foundCoordinator = true
			if ag.Mode != "primary" {
				t.Errorf("coordinator mode: got %q, want primary", ag.Mode)
			}
			if ag.Temperature != 0.2 {
				t.Errorf("coordinator temp: got %v, want 0.2", ag.Temperature)
			}
		}
	}
	if !foundCoordinator {
		t.Error("coordinator agent not found")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "opencode-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	cfg, _ := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	tmpDir := t.TempDir()

	_, err := a.Write(context.Background(), cfg, adapter.WriteOptions{Scope: "project", ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Verify key files
	configPath := filepath.Join(tmpDir, "opencode.jsonc")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("opencode.jsonc not created")
	}

	agentsDir := filepath.Join(tmpDir, ".opencode", "agents")
	if entries, _ := os.ReadDir(agentsDir); len(entries) != 7 {
		t.Errorf("agents dir: got %d entries, want 7", len(entries))
	}
}

func TestDetect(t *testing.T) {
	a := New()
	result, err := a.Detect()
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	_ = result.Found
}

func TestCapabilities(t *testing.T) {
	a := New()
	caps := a.Capabilities()
	if !caps.SupportsAgents {
		t.Error("should support agents")
	}
	if !caps.SupportsSkills {
		t.Error("should support skills")
	}
	if !caps.SupportsCommands {
		t.Error("should support commands")
	}
	if !caps.SupportsMCP {
		t.Error("should support MCP")
	}
	if caps.SupportsHooks {
		t.Error("opencode should not natively support hooks (uses plugins)")
	}
}

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "opencode" {
		t.Errorf("ID: got %q", a.ID())
	}
}
