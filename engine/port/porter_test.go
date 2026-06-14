package port

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moximxxx/shifter/canonical"
)

func TestPortClaudeCodeToCodex(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "fixtures", "claude-project")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	tmpDir := t.TempDir()
	srcClaudeDir := filepath.Join(fixtureDir, ".claude")
	dstClaudeDir := filepath.Join(tmpDir, ".claude")
	copyDir(srcClaudeDir, dstClaudeDir)
	os.WriteFile(filepath.Join(tmpDir, "CLAUDE.md"), readFixture(fixtureDir, "CLAUDE.md"), 0644)

	ctx := context.Background()
	result, err := Port(ctx, Options{Source: "claude-code", Target: "codex", Scope: "project", ProjectRoot: tmpDir, Backup: false})
	if err != nil {
		t.Fatalf("Port: %v", err)
	}

	if len(result.FilesWritten) == 0 {
		t.Error("no files written")
	}
	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config.toml not created at %s", configPath)
	}
	for _, w := range result.LossWarnings {
		t.Logf("  [%s] %s: %s", w.Severity, w.Feature, w.Reason)
	}
}

func TestPreview(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "fixtures", "claude-project")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	tmpDir := t.TempDir()
	copyDir(filepath.Join(fixtureDir, ".claude"), filepath.Join(tmpDir, ".claude"))

	ctx := context.Background()
	diff, err := Preview(ctx, Options{Source: "claude-code", Target: "codex", Scope: "project", ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if diff == nil {
		t.Fatal("diff should not be nil")
	}
}

func TestFilterAspects(t *testing.T) {
	cfg := &canonical.ShifterConfig{
		Agents:       []canonical.AgentDef{{Name: "test"}},
		Skills:       []canonical.SkillDef{{Name: "skill"}},
		Commands:     []canonical.CommandDef{{Name: "cmd"}},
		MCPServers:   []canonical.MCPServerDef{{Name: "mcp"}},
		Hooks:        []canonical.HookDef{{Event: "test"}},
		Instructions: []canonical.Instruction{{Path: "test.md"}},
		Permissions:  &canonical.PermissionSet{DefaultMode: "edit"},
		Settings:     canonical.SettingsMap{Model: "sonnet"},
	}

	tests := []struct {
		name     string
		aspects  []string
		check    func(*canonical.ShifterConfig) bool
	}{
		{"agents only", []string{"agents"}, func(c *canonical.ShifterConfig) bool { return len(c.Agents) == 1 && len(c.Skills) == 0 }},
		{"skills only", []string{"skills"}, func(c *canonical.ShifterConfig) bool { return len(c.Agents) == 0 && len(c.Skills) == 1 }},
		{"mcp only", []string{"mcp"}, func(c *canonical.ShifterConfig) bool { return len(c.MCPServers) == 1 && len(c.Agents) == 0 }},
		{"all", []string{"agents", "skills", "commands", "mcp", "hooks", "instructions", "permissions", "settings"}, func(c *canonical.ShifterConfig) bool { return len(c.Agents) == 1 && len(c.Skills) == 1 }},
		{"multiple", []string{"agents", "skills"}, func(c *canonical.ShifterConfig) bool { return len(c.Agents) == 1 && len(c.Skills) == 1 && len(c.Commands) == 0 }},
		{"empty", nil, func(c *canonical.ShifterConfig) bool { return len(c.Agents) == 0 }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filtered := filterAspects(cfg, tc.aspects)
			if !tc.check(filtered) {
				t.Errorf("filterAspects(%v) failed check", tc.aspects)
			}
		})
	}
}

func TestPort_InvalidSource(t *testing.T) {
	ctx := context.Background()
	_, err := Port(ctx, Options{Source: "nonexistent", Target: "codex", Backup: false})
	if err == nil {
		t.Error("expected error for invalid source")
	}
	if err != nil && !strings.Contains(err.Error(), "source") {
		t.Logf("error: %v", err)
	}
}

func TestPort_InvalidTarget(t *testing.T) {
	ctx := context.Background()
	_, err := Port(ctx, Options{Source: "aider", Target: "nonexistent", Backup: false})
	if err == nil {
		t.Error("expected error for invalid target")
	}
}

func TestPort_DryRun(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "fixtures", "claude-project")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	tmpDir := t.TempDir()
	copyDir(filepath.Join(fixtureDir, ".claude"), filepath.Join(tmpDir, ".claude"))

	ctx := context.Background()
	result, err := Port(ctx, Options{Source: "claude-code", Target: "codex", Scope: "project", ProjectRoot: tmpDir, DryRun: true, Backup: false})
	if err != nil {
		t.Fatalf("Port dry-run: %v", err)
	}
	// Dry run should not write files
	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		t.Error("dry-run should not create files")
	}
	_ = result
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func readFixture(dir, name string) []byte {
	data, _ := os.ReadFile(filepath.Join(dir, name))
	return data
}
