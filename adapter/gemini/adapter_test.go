package gemini

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "gemini-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	cfg, err := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(cfg.MCPServers) != 3 {
		t.Fatalf("mcp servers: got %d, want 3", len(cfg.MCPServers))
	}
	if cfg.Permissions == nil || len(cfg.Permissions.AllowRules) == 0 {
		t.Error("permissions should have allow rules")
	}
	if len(cfg.Instructions) != 1 {
		t.Errorf("instructions: got %d, want 1", len(cfg.Instructions))
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "gemini-test")
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

	settingsPath := filepath.Join(tmpDir, ".gemini", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("settings.json not created")
	}

	cfg2, _ := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: tmpDir})
	if len(cfg2.MCPServers) != len(cfg.MCPServers) {
		t.Errorf("round-trip mcp: got %d, want %d", len(cfg2.MCPServers), len(cfg.MCPServers))
	}
}

func TestCapabilities(t *testing.T) {
	a := New()
	caps := a.Capabilities()
	if !caps.SupportsMCP {
		t.Error("should support MCP")
	}
	if !caps.SupportsPermissions {
		t.Error("should support permissions")
	}
	if caps.SupportsAgents {
		t.Error("gemini should not support agents")
	}
}

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "gemini-cli" {
		t.Errorf("ID: got %q", a.ID())
	}
}
