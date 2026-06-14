package profile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/moximxxx/shifter/canonical"
)

func TestSave_Load(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	p := Profile{
		Name:        "test-profile",
		Description: "A test profile",
		SourceAgent: "claude-code",
		CreatedAt:   time.Now(),
		Config: &canonical.ShifterConfig{
			Meta: canonical.ConfigMeta{SourceAdapter: "claude-code"},
			Agents: []canonical.AgentDef{
				{Name: "code-reviewer", SystemPrompt: "You are a reviewer."},
			},
		},
	}

	if err := Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	profilePath := filepath.Join(tmpDir, ".shifter", "profiles", "test-profile.json")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Fatal("profile file not created")
	}

	loaded, err := Load("test-profile")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Name != "test-profile" {
		t.Errorf("name: got %q", loaded.Name)
	}
	if loaded.SourceAgent != "claude-code" {
		t.Errorf("source: got %q", loaded.SourceAgent)
	}
	if len(loaded.Config.Agents) != 1 {
		t.Errorf("agents: got %d", len(loaded.Config.Agents))
	}
}

func TestLoad_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	_, err := Load("nonexistent-profile")
	if err == nil {
		t.Error("expected error for nonexistent profile")
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	for _, name := range []string{"profile-a", "profile-b"} {
		Save(Profile{Name: name, SourceAgent: "claude-code", Config: &canonical.ShifterConfig{}})
	}

	profiles, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(profiles) < 2 {
		t.Errorf("expected at least 2 profiles, got %d", len(profiles))
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	Save(Profile{Name: "to-delete", Config: &canonical.ShifterConfig{}})

	if err := Delete("to-delete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := Load("to-delete")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestSummary(t *testing.T) {
	p := Profile{
		Name: "test",
		Config: &canonical.ShifterConfig{
			Agents:     []canonical.AgentDef{{}, {}, {}},
			Skills:     []canonical.SkillDef{{}},
			MCPServers: []canonical.MCPServerDef{{}, {}},
			Hooks:      []canonical.HookDef{{}, {}, {}, {}},
		},
	}
	summary := p.Summary()
	if summary == "" || summary == "empty" {
		t.Errorf("expected non-empty summary, got %q", summary)
	}
}

func TestSummary_Empty(t *testing.T) {
	p := Profile{Name: "empty"}
	summary := p.Summary()
	if summary != "empty" {
		t.Errorf("expected 'empty', got %q", summary)
	}
}

func TestSanitizeName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"My Profile", "my-profile"},
		{"Test 123", "test-123"},
		{"Hello_World", "hello_world"},
		{"  spaces  ", "spaces"},
		{"", "unnamed"},
	}
	for _, tc := range tests {
		got := sanitizeName(tc.in)
		if got != tc.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
