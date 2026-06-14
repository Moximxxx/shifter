package detect

import (
	"testing"
)

func TestScanAll(t *testing.T) {
	results := ScanAll()
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	// Verify no duplicate IDs
	seen := make(map[string]bool)
	for _, r := range results {
		if seen[r.ID] {
			t.Errorf("duplicate ID: %s", r.ID)
		}
		seen[r.ID] = true
		if r.Name == "" {
			t.Errorf("empty name for %s", r.ID)
		}
	}
	// Should have all 7 agents
	if len(results) != 7 {
		t.Errorf("expected 7 agents, got %d", len(results))
	}
}

func TestScanAll_Sorted(t *testing.T) {
	results := ScanAll()
	for i := 1; i < len(results); i++ {
		if results[i].ID < results[i-1].ID {
			t.Errorf("not sorted: %s before %s", results[i-1].ID, results[i].ID)
		}
	}
}

func TestResult_Fields(t *testing.T) {
	r := Result{
		ID: "test-agent", Name: "Test Agent",
		Found: true, Summary: map[string]int{"agents": 3},
		Paths: []string{"/tmp/.test-agent", ".test-agent"},
	}
	if r.ID != "test-agent" || !r.Found {
		t.Error("result fields not set correctly")
	}
	if r.Summary["agents"] != 3 {
		t.Error("summary not correct")
	}
}
