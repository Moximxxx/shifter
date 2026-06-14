package convert

import "testing"

func TestNewLossWarning(t *testing.T) {
	w := NewLossWarning("hooks", "claude-code", "qoder", "Qoder has no hook system", "warning")
	if w.Feature != "hooks" {
		t.Errorf("Feature: got %q", w.Feature)
	}
	if w.Field != "hooks" {
		t.Errorf("Field: got %q", w.Field)
	}
	if w.SourceAgent != "claude-code" {
		t.Errorf("SourceAgent: got %q", w.SourceAgent)
	}
	if w.TargetAgent != "qoder" {
		t.Errorf("TargetAgent: got %q", w.TargetAgent)
	}
	if w.Reason != "Qoder has no hook system" {
		t.Errorf("Reason: got %q", w.Reason)
	}
	if w.Severity != "warning" {
		t.Errorf("Severity: got %q", w.Severity)
	}
}
