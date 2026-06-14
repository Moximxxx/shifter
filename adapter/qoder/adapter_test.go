package qoder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "qoder-test")
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
	for _, ag := range cfg.Agents {
		if ag.Name == "" || ag.SystemPrompt == "" {
			t.Errorf("agent %q has missing fields", ag.Name)
		}
	}

	if len(cfg.Skills) != 2 {
		t.Fatalf("skills: got %d, want 2", len(cfg.Skills))
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "qoder-test")
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

	// Verify agent files
	for _, ag := range cfg.Agents {
		agentPath := filepath.Join(tmpDir, ".qoder", "agents", ag.Name+".md")
		if _, err := os.Stat(agentPath); os.IsNotExist(err) {
			t.Errorf("agent file not created: %s", agentPath)
		}
	}

	// Verify skill files
	for _, sk := range cfg.Skills {
		skillPath := filepath.Join(tmpDir, ".qoder", "skills", sk.Name, "SKILL.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("skill file not created: %s", skillPath)
		}
	}

	if len(result.FilesWritten) == 0 {
		t.Error("no files written")
	}

	// Read back
	cfg2, err := a.Read(ctx, adapter.ReadOptions{Scope: "project", ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("Read-back: %v", err)
	}
	if len(cfg2.Agents) != len(cfg.Agents) {
		t.Errorf("round-trip agents: got %d, want %d", len(cfg2.Agents), len(cfg.Agents))
	}
	if len(cfg2.Skills) != len(cfg.Skills) {
		t.Errorf("round-trip skills: got %d, want %d", len(cfg2.Skills), len(cfg.Skills))
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
	if !caps.SupportsMCP {
		t.Error("should support MCP")
	}
}

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "qoder" {
		t.Errorf("ID: got %q", a.ID())
	}
}
