// Package adapter defines the AgentAdapter interface that every coding agent
// must implement, plus the global registry for discovering adapters.

// AgentAdapter is the interface every coding agent adapter must implement.
// Each adapter reads native config into the canonical model and writes it back.

package adapter

import (
	"context"

	"github.com/moximxxx/shifter/canonical"
)

// AgentAdapter is the interface every coding agent adapter must implement.
type AgentAdapter interface {
	// Identity
	ID() string    // stable machine name: "claude-code", "codex", "opencode"
	Name() string  // human-readable: "Claude Code"

	// SearchPaths returns directories where this agent's config may live
	// (both global/user and project-level).
	SearchPaths() []string

	// Detect checks whether this agent is configured on the system.
	Detect() (DetectionResult, error)

	// Read converts native config to the canonical model.
	Read(ctx context.Context, opts ReadOptions) (*canonical.ShifterConfig, error)

	// Write converts the canonical model back to native config and writes it.
	Write(ctx context.Context, cfg *canonical.ShifterConfig, opts WriteOptions) (*WriteResult, error)

	// Capabilities reports what this agent supports.
	Capabilities() canonical.CapabilityMatrix
}

// DetectionResult describes whether and where an agent's config was found.
type DetectionResult struct {
	Found       bool
	GlobalPaths []string          // e.g., ~/.claude/, ~/.codex/
	ProjectPath string            // e.g., .claude/, .codex/
	Summary     map[string]int    // e.g., {"agents": 3, "skills": 5}
}

// ReadOptions controls how an adapter reads configuration.
type ReadOptions struct {
	Scope       string // "global" | "project" | "all"
	ProjectRoot string // path to project root
}

// WriteOptions controls how an adapter writes configuration.
type WriteOptions struct {
	Scope       string // "global" | "project"
	ProjectRoot string
	DryRun      bool // compute diff but don't write
	Backup      bool // create backup before write (default: true)
	Force       bool // skip confirmation prompts
	Merge       bool // merge with existing instead of replace
}

// WriteResult describes what happened during a write operation.
type WriteResult struct {
	FilesWritten []string
	FilesSkipped []string
	LossWarnings []canonical.LossWarning
	BackupPath   string // path to backup archive if created
}

// DiffResult describes proposed changes without writing.
type DiffResult struct {
	FilesAdded    []FileChange
	FilesModified []FileChange
	FilesDeleted  []FileChange
	Unchanged     []string
	LossWarnings  []canonical.LossWarning
}

// FileChange describes a single file change in a diff.
type FileChange struct {
	Path    string
	OldHash string
	NewHash string
	Diff    string // unified diff
	Action  string // "create" | "modify" | "delete"
}
