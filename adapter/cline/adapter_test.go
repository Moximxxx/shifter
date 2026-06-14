package cline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
)

func TestReadFixtures(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "cline-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	a := New()
	cfg, err := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(cfg.Instructions) != 6 {
		t.Fatalf("instructions: got %d, want 6", len(cfg.Instructions))
	}
	for _, inst := range cfg.Instructions {
		if inst.Content == "" {
			t.Errorf("instruction %s is empty", inst.Path)
		}
	}
}

func TestWriteRoundTrip(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "test", "cline-test")
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

	// Verify .clinerules/ directory
	rulesDir := filepath.Join(tmpDir, ".clinerules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error(".clinerules/ not created")
	}

	cfg2, _ := a.Read(context.Background(), adapter.ReadOptions{Scope: "project", ProjectRoot: tmpDir})
	if len(cfg2.Instructions) != len(cfg.Instructions) {
		t.Errorf("round-trip instructions: got %d, want %d", len(cfg2.Instructions), len(cfg.Instructions))
	}
}

func TestCapabilities(t *testing.T) {
	a := New()
	caps := a.Capabilities()
	if !caps.SupportsInstructions {
		t.Error("should support instructions")
	}
	if caps.SupportsAgents {
		t.Error("cline should not support agents natively")
	}
}

func TestID(t *testing.T) {
	a := New()
	if a.ID() != "cline" {
		t.Errorf("ID: got %q", a.ID())
	}
}
