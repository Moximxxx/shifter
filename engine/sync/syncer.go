// Package sync implements bidirectional config synchronization between agents.
package sync

import (
	"context"
	"fmt"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/registry"
)

// Options controls the sync operation.
type Options struct {
	AgentA      string
	AgentB      string
	Strategy    canonical.MergeStrategy
	Scope       string
	ProjectRoot string
	DryRun      bool
	Backup      bool
	Force       bool
}

// Result describes the outcome of a sync operation.
type Result struct {
	AgentA        string
	AgentB        string
	FilesWritten  []string
	LossWarnings  []canonical.LossWarning
	Conflicts     []canonical.MergeConflict
	MergedConfig  *canonical.ShifterConfig
}

// Sync performs bidirectional synchronization between two agents.
//
// Process:
//  1. Read config from AgentA
//  2. Read config from AgentB
//  3. Merge into unified canonical config
//  4. Write merged config to AgentA
//  5. Write merged config to AgentB
func Sync(ctx context.Context, opts Options) (*Result, error) {
	result := &Result{
		AgentA: opts.AgentA,
		AgentB: opts.AgentB,
	}

	if opts.Strategy == "" {
		opts.Strategy = canonical.MergeNewer
	}

	adapterA, err := registry.Get(opts.AgentA)
	if err != nil {
		return nil, fmt.Errorf("agent A (%s): %w", opts.AgentA, err)
	}
	adapterB, err := registry.Get(opts.AgentB)
	if err != nil {
		return nil, fmt.Errorf("agent B (%s): %w", opts.AgentB, err)
	}

	readOpts := adapter.ReadOptions{
		Scope:       opts.Scope,
		ProjectRoot: opts.ProjectRoot,
	}
	if opts.Scope == "" {
		readOpts.Scope = "all"
	}

	// Read both sides
	cfgA, err := adapterA.Read(ctx, readOpts)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", opts.AgentA, err)
	}

	cfgB, err := adapterB.Read(ctx, readOpts)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", opts.AgentB, err)
	}

	// Merge
	mergeResult := canonical.MergeConfigs(cfgA, cfgB, opts.Strategy)
	result.MergedConfig = mergeResult.Config
	result.Conflicts = mergeResult.Conflicts

	// Set metadata
	mergeResult.Config.Meta.SourceAdapter = "sync:" + opts.AgentA + "+" + opts.AgentB

	// Write back to both
	writeOpts := adapter.WriteOptions{
		Scope:       opts.Scope,
		ProjectRoot: opts.ProjectRoot,
		DryRun:      opts.DryRun,
		Backup:      opts.Backup,
		Force:       opts.Force,
	}
	if opts.Scope == "all" {
		writeOpts.Scope = "project"
	}

	writeA, err := adapterA.Write(ctx, mergeResult.Config, writeOpts)
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", opts.AgentA, err)
	}
	result.FilesWritten = append(result.FilesWritten, writeA.FilesWritten...)
	result.LossWarnings = append(result.LossWarnings, writeA.LossWarnings...)

	writeB, err := adapterB.Write(ctx, mergeResult.Config, writeOpts)
	if err != nil {
		return nil, fmt.Errorf("write %s: %w", opts.AgentB, err)
	}
	result.FilesWritten = append(result.FilesWritten, writeB.FilesWritten...)
	result.LossWarnings = append(result.LossWarnings, writeB.LossWarnings...)

	return result, nil
}
