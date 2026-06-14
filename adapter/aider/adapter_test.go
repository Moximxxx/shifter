package aider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "aider-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	cfg, err := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if cfg.Settings.Model == "" {
		t.Error("model should not be empty")
	}
	if len(cfg.Instructions) == 0 {
		t.Error("should have at least CONVENTIONS.md as instruction")
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "aider-test")
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

	configPath := filepath.Join(tmpDir, ".aider.conf.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error(".aider.conf.yml not created")
	}

	cfg2, _ := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: tmpDir})
	if cfg2.Settings.Model != cfg.Settings.Model {
		t.Errorf("round-trip model: got %q, want %q", cfg2.Settings.Model, cfg.Settings.Model)
	}
}

func TestCapabilities(t *testing.T) {
	a := New()
	caps := a.Capabilities()
	if !caps.SupportsSettings {
		t.Error("should support settings")
	}
	if caps.SupportsAgents {
		t.Error("aider should not support agents")
	}
}

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "aider" {
		t.Errorf("ID: got %q", a.ID())
	}
}
