package env

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	os.Setenv("SHIFTER_TEST_GET", "hello")
	defer os.Unsetenv("SHIFTER_TEST_GET")

	if v := Get("SHIFTER_TEST_GET", "default"); v != "hello" {
		t.Errorf("Get: got %q, want hello", v)
	}
	if v := Get("SHIFTER_NONEXISTENT", "fallback"); v != "fallback" {
		t.Errorf("Get default: got %q, want fallback", v)
	}
}

func TestBool(t *testing.T) {
	tests := []struct{ val string; want bool }{
		{"1", true}, {"true", true}, {"TRUE", true}, {"yes", true}, {"on", true},
		{"0", false}, {"false", false}, {"no", false}, {"off", false}, {"", false},
	}
	for _, tc := range tests {
		os.Setenv("SHIFTER_TEST_BOOL", tc.val)
		if got := Bool("SHIFTER_TEST_BOOL"); got != tc.want {
			t.Errorf("Bool(%q) = %v, want %v", tc.val, got, tc.want)
		}
	}
	os.Unsetenv("SHIFTER_TEST_BOOL")
}

func TestSource(t *testing.T) {
	os.Setenv("SHIFTER_SOURCE", "claude-code")
	defer os.Unsetenv("SHIFTER_SOURCE")
	if s := Source(); s != "claude-code" {
		t.Errorf("Source: got %q", s)
	}
}

func TestTarget(t *testing.T) {
	os.Setenv("SHIFTER_TARGET", "codex")
	defer os.Unsetenv("SHIFTER_TARGET")
	if s := Target(); s != "codex" {
		t.Errorf("Target: got %q", s)
	}
}

func TestScope(t *testing.T) {
	os.Setenv("SHIFTER_SCOPE", "global")
	defer os.Unsetenv("SHIFTER_SCOPE")
	if s := Scope(); s != "global" {
		t.Errorf("Scope: got %q", s)
	}
}

func TestProjectRoot(t *testing.T) {
	os.Setenv("SHIFTER_PROJECT_ROOT", "/tmp/test")
	defer os.Unsetenv("SHIFTER_PROJECT_ROOT")
	if s := ProjectRoot(); s != "/tmp/test" {
		t.Errorf("ProjectRoot: got %q", s)
	}
}

func TestDryRun(t *testing.T) {
	os.Setenv("SHIFTER_DRY_RUN", "1")
	defer os.Unsetenv("SHIFTER_DRY_RUN")
	if !DryRun() {
		t.Error("DryRun should be true")
	}
}

func TestNoBackup(t *testing.T) {
	os.Setenv("SHIFTER_NO_BACKUP", "true")
	defer os.Unsetenv("SHIFTER_NO_BACKUP")
	if !NoBackup() {
		t.Error("NoBackup should be true")
	}
}

func TestForce(t *testing.T) {
	os.Setenv("SHIFTER_FORCE", "yes")
	defer os.Unsetenv("SHIFTER_FORCE")
	if !Force() {
		t.Error("Force should be true")
	}
}

func TestAspects(t *testing.T) {
	os.Setenv("SHIFTER_ASPECTS", "agents, skills, mcp")
	defer os.Unsetenv("SHIFTER_ASPECTS")
	aspects := Aspects()
	if len(aspects) != 3 {
		t.Errorf("Aspects: got %d items, want 3: %v", len(aspects), aspects)
	}
	if aspects[0] != "agents" || aspects[1] != "skills" || aspects[2] != "mcp" {
		t.Errorf("Aspects parsing wrong: %v", aspects)
	}
}

func TestAspects_Empty(t *testing.T) {
	os.Unsetenv("SHIFTER_ASPECTS")
	if aspects := Aspects(); aspects != nil {
		t.Errorf("empty aspects should be nil, got %v", aspects)
	}
}

func TestProfile(t *testing.T) {
	os.Setenv("SHIFTER_PROFILE", "my-workflow")
	defer os.Unsetenv("SHIFTER_PROFILE")
	if s := Profile(); s != "my-workflow" {
		t.Errorf("Profile: got %q", s)
	}
}

func TestStrategy(t *testing.T) {
	os.Setenv("SHIFTER_STRATEGY", "merge")
	defer os.Unsetenv("SHIFTER_STRATEGY")
	if s := Strategy(); s != "merge" {
		t.Errorf("Strategy: got %q", s)
	}
}

func TestPretty(t *testing.T) {
	tests := []struct{ val string; want bool }{
		{"", true}, {"1", true}, {"0", false}, {"false", false}, {"off", false},
	}
	for _, tc := range tests {
		if tc.val == "" {
			os.Unsetenv("SHIFTER_PRETTY")
		} else {
			os.Setenv("SHIFTER_PRETTY", tc.val)
		}
		if got := Pretty(); got != tc.want {
			t.Errorf("Pretty(%q) = %v, want %v", tc.val, got, tc.want)
		}
	}
	os.Unsetenv("SHIFTER_PRETTY")
}

func TestConfig(t *testing.T) {
	os.Setenv("SHIFTER_SOURCE", "claude-code")
	os.Setenv("SHIFTER_DRY_RUN", "1")
	defer os.Unsetenv("SHIFTER_SOURCE")
	defer os.Unsetenv("SHIFTER_DRY_RUN")

	config := Config()
	if config["SHIFTER_SOURCE"] != "claude-code" {
		t.Errorf("Config source: got %q", config["SHIFTER_SOURCE"])
	}
	if config["SHIFTER_DRY_RUN"] != "1" {
		t.Errorf("Config dry_run: got %q", config["SHIFTER_DRY_RUN"])
	}
}

func TestConfig_Empty(t *testing.T) {
	oldSource := os.Getenv("SHIFTER_SOURCE")
	oldTarget := os.Getenv("SHIFTER_TARGET")
	os.Unsetenv("SHIFTER_SOURCE")
	os.Unsetenv("SHIFTER_TARGET")
	defer os.Setenv("SHIFTER_SOURCE", oldSource)
	defer os.Setenv("SHIFTER_TARGET", oldTarget)

	config := Config()
	if len(config) != 0 {
		t.Errorf("empty config should be empty, got %d entries", len(config))
	}
}
