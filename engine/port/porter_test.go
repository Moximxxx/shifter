package port

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPortClaudeCodeToCodex(t *testing.T) {
	// This is an integration test that requires fixtures
	fixtureDir := filepath.Join("..", "..", "fixtures", "claude-project")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	tmpDir := t.TempDir()

	// Copy the entire .claude directory from fixtures to the temp project
	// Use cp -r for simplicity in the test
	srcClaudeDir := filepath.Join(fixtureDir, ".claude")
	dstClaudeDir := filepath.Join(tmpDir, ".claude")

	copyDir := func(src, dst string) error {
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
	if err := copyDir(srcClaudeDir, dstClaudeDir); err != nil {
		t.Fatalf("copy fixtures: %v", err)
	}

	// Copy CLAUDE.md
	claudeMDBytes, _ := os.ReadFile(filepath.Join(fixtureDir, "CLAUDE.md"))
	os.WriteFile(filepath.Join(tmpDir, "CLAUDE.md"), claudeMDBytes, 0644)

	ctx := context.Background()

	result, err := Port(ctx, Options{
		Source:      "claude-code",
		Target:      "codex",
		Scope:       "project",
		ProjectRoot: tmpDir,
		Backup:      false,
	})
	if err != nil {
		t.Fatalf("Port failed: %v", err)
	}

	// Verify Codex files were created
	if len(result.FilesWritten) == 0 {
		t.Error("no files were written")
	}

	configPath := filepath.Join(tmpDir, ".codex", "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config.toml was not created at %s", configPath)
	} else {
		data, _ := os.ReadFile(configPath)
		t.Logf("config.toml:\n%s", string(data))
	}

	codexMDPath := filepath.Join(tmpDir, ".codex", "codex.md")
	if _, err := os.Stat(codexMDPath); os.IsNotExist(err) {
		t.Errorf("codex.md was not created at %s", codexMDPath)
	} else {
		data, _ := os.ReadFile(codexMDPath)
		t.Logf("codex.md:\n%s", string(data))
	}

	t.Logf("Port result: %d files written, %d loss warnings",
		len(result.FilesWritten), len(result.LossWarnings))
	for _, w := range result.LossWarnings {
		t.Logf("  [%s] %s: %s", w.Severity, w.Feature, w.Reason)
	}
}
