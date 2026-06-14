// Package e2e provides end-to-end tests for the Shifter CLI.
// These tests invoke the actual shifter binary.
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildBinary compiles shifter and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "shifter")
	cmd := exec.Command("go", "build", "-o", bin, "github.com/moximxxx/shifter")
	cmd.Dir = filepath.Join("..", "..")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestE2E_Version(t *testing.T) {
	bin := buildBinary(t)
	out, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if len(out) == 0 {
		t.Error("no output from --version")
	}
}

func TestE2E_Detect(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "detect", "--project-root", filepath.Join("..", ".."))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	// Should find at least claude-code agent (our fixture)
	output := string(out)
	if len(output) == 0 {
		t.Error("no output from detect")
	}
}

func TestE2E_Detect_JSON(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "detect", "--json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect --json: %v", err)
	}
	if len(out) == 0 {
		t.Error("no output from detect --json")
	}
}

func TestE2E_Status(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "status", "--project-root", filepath.Join("..", ".."))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(out) == 0 {
		t.Error("no output from status")
	}
}

func TestE2E_Config(t *testing.T) {
	bin := buildBinary(t)
	// Set a temp HOME to avoid modifying real settings
	tmpHome := t.TempDir()
	cmd := exec.Command(bin, "config")
	cmd.Env = append(os.Environ(), "HOME="+tmpHome)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if len(out) == 0 {
		t.Error("no output from config")
	}
}

func TestE2E_Env(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "env")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("env: %v", err)
	}
	if len(out) == 0 {
		t.Error("no output from env")
	}
}

func TestE2E_Export(t *testing.T) {
	bin := buildBinary(t)
	fixtureDir := filepath.Join("..", "..", "test", "claude-code-test")
	if _, err := os.Stat(fixtureDir); os.IsNotExist(err) {
		t.Skip("fixtures not found")
	}

	cmd := exec.Command(bin, "export", "claude-code", "--project-root", fixtureDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("export: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Error("no output from export")
	}
}
