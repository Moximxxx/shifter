package registry

import (
	"testing"
)

func TestGet_AllAdapters(t *testing.T) {
	for _, id := range ListIDs() {
		a, err := Get(id)
		if err != nil {
			t.Errorf("Get(%q): %v", id, err)
			continue
		}
		if a.ID() != id {
			t.Errorf("Get(%q): ID() returned %q", id, a.ID())
		}
		if a.Name() == "" {
			t.Errorf("Get(%q): Name() empty", id)
		}
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Error("expected error for unknown adapter")
	}
}

func TestListIDs(t *testing.T) {
	ids := ListIDs()
	if len(ids) == 0 {
		t.Fatal("expected at least 1 adapter")
	}
	// Verify no duplicates
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID: %q", id)
		}
		seen[id] = true
	}
	// Verify sorted
	for i := 1; i < len(ids); i++ {
		if ids[i] < ids[i-1] {
			t.Errorf("not sorted: %v", ids)
		}
	}
}

func TestMustGet(t *testing.T) {
	a := MustGet("claude-code")
	if a == nil {
		t.Fatal("MustGet returned nil")
	}
	if a.ID() != "claude-code" {
		t.Errorf("MustGet: got %q", a.ID())
	}
}

func TestMustGet_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown adapter")
		}
	}()
	MustGet("nonexistent")
}
