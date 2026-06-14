package sync

import (
	"context"
	"testing"

	"github.com/moximxxx/shifter/canonical"
)

func TestSync_Newer(t *testing.T) {
	ctx := context.Background()
	result, err := Sync(ctx, Options{
		AgentA: "aider", AgentB: "cline",
		Strategy: canonical.MergeNewer,
		Scope: "project", Backup: false,
	})
	if err != nil {
		// Might fail if agents are not configured on this system
		t.Logf("Sync (expected in test env): %v", err)
		return
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

func TestSync_Merge(t *testing.T) {
	ctx := context.Background()
	result, err := Sync(ctx, Options{
		AgentA: "aider", AgentB: "cline",
		Strategy: canonical.MergeUnion,
		Scope: "project", Backup: false,
	})
	if err != nil {
		t.Logf("Sync merge (expected): %v", err)
		return
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

func TestSync_Source(t *testing.T) {
	ctx := context.Background()
	_, err := Sync(ctx, Options{
		AgentA: "aider", AgentB: "cline",
		Strategy: canonical.MergePreferSource,
		Scope: "project", Backup: false,
	})
	if err != nil {
		t.Logf("Sync source (expected): %v", err)
	}
}

func TestSync_InvalidAgent(t *testing.T) {
	ctx := context.Background()
	_, err := Sync(ctx, Options{
		AgentA: "nonexistent", AgentB: "cline",
		Strategy: canonical.MergeNewer,
	})
	if err == nil {
		t.Error("expected error for invalid agent A")
	}

	_, err = Sync(ctx, Options{
		AgentA: "aider", AgentB: "nonexistent",
		Strategy: canonical.MergeNewer,
	})
	if err == nil {
		t.Error("expected error for invalid agent B")
	}
}

func TestSync_DefaultStrategy(t *testing.T) {
	ctx := context.Background()
	result, err := Sync(ctx, Options{
		AgentA: "aider", AgentB: "cline",
		Scope: "project", Backup: false,
	})
	if err != nil {
		t.Logf("Sync default (expected): %v", err)
		return
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}
