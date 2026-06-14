package logo

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	result := Render("x")
	if result == "" {
		t.Error("Render returned empty string")
	}
	if len(result) < 100 {
		t.Errorf("Render output too short: %d chars", len(result))
	}
	// Should start with ANSI reset code
	if !strings.HasPrefix(result, "\033[0m") {
		t.Error("Render should start with ANSI reset")
	}
}

func TestSmall(t *testing.T) {
	result := Small()
	if result == "" {
		t.Error("Small returned empty string")
	}
	if !strings.Contains(result, "Shifter") {
		t.Error("Small should contain Shifter")
	}
}
