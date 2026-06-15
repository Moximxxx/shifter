// Package golden provides golden file tests for all adapters.
// Run with: go test ./test/golden/... -update
package golden

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/registry"
)

var update = flag.Bool("update", false, "update golden files")

func goldenTest(t *testing.T, agentID, fixtureDir, goldenDir string) {
	t.Helper()
	a, err := registry.Get(agentID)
	if err != nil {
		t.Fatalf("get %s: %v", agentID, err)
	}
	ctx := context.Background()
	cfg, err := a.Read(ctx, adapter.ReadOptions{Scope: "project", ProjectRoot: fixtureDir})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	tmpDir := t.TempDir()
	_, err = a.Write(ctx, cfg, adapter.WriteOptions{Scope: "project", ProjectRoot: tmpDir})
	if err != nil {
		t.Fatalf("write: %v", err)
	}

	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(tmpDir, path)
		content, _ := os.ReadFile(path)
		goldenPath := filepath.Join(goldenDir, rel)
		if *update {
			os.MkdirAll(filepath.Dir(goldenPath), 0755)
			os.WriteFile(goldenPath, content, 0644)
			return nil
		}
		golden, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Errorf("no golden: %s (run with -update)", rel)
			return nil
		}
		if string(content) != string(golden) {
			t.Errorf("mismatch: %s (run with -update)", rel)
		}
		return nil
	})
}

func TestGolden_ClaudeCode(t *testing.T) {
	goldenTest(t, "claude-code",
		filepath.Join("..", "..", "test", "claude-code-test"),
		filepath.Join("..", "..", "adapter", "claudecode", "testdata"),
	)
}

func TestGolden_Codex(t *testing.T) {
	goldenTest(t, "codex",
		filepath.Join("..", "..", "test", "codex-test"),
		filepath.Join("..", "..", "adapter", "codex", "testdata"),
	)
}

func TestGolden_OpenCode(t *testing.T) {
	goldenTest(t, "opencode",
		filepath.Join("..", "..", "test", "opencode-test"),
		filepath.Join("..", "..", "adapter", "opencode", "testdata"),
	)
}

func TestGolden_Qoder(t *testing.T) {
	goldenTest(t, "qoder",
		filepath.Join("..", "..", "test", "qoder-test"),
		filepath.Join("..", "..", "adapter", "qoder", "testdata"),
	)
}

func TestGolden_Gemini(t *testing.T) {
	goldenTest(t, "gemini-cli",
		filepath.Join("..", "..", "test", "gemini-test"),
		filepath.Join("..", "..", "adapter", "gemini", "testdata"),
	)
}

func TestGolden_Cline(t *testing.T) {
	goldenTest(t, "cline",
		filepath.Join("..", "..", "test", "cline-test"),
		filepath.Join("..", "..", "adapter", "cline", "testdata"),
	)
}

func TestGolden_Aider(t *testing.T) {
	goldenTest(t, "aider",
		filepath.Join("..", "..", "test", "aider-test"),
		filepath.Join("..", "..", "adapter", "aider", "testdata"),
	)
}
