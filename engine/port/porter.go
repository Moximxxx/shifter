// Package port implements the one-directional config transfer engine.
//
// The port engine orchestrates the pipeline:
//
//	Source Adapter.Read() → Canonical Model → Target Adapter.Write()
//
// It handles loss warning accumulation, dry-run previews, and backup creation.
package port

import (
	"context"
	"fmt"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/registry"
)

// Options controls the port operation.
type Options struct {
	Source      string // source adapter ID
	Target      string // target adapter ID
	Scope       string // "global" | "project" | "all"
	ProjectRoot string
	DryRun      bool
	Backup      bool
	Force       bool
	Aspects     []string // empty = all; otherwise filter by aspect name
}

// Result describes the outcome of a port operation.
type Result struct {
	Source       string
	Target       string
	FilesWritten []string
	FilesSkipped []string
	LossWarnings []canonical.LossWarning
	Config       *canonical.ShifterConfig // the canonical config (for inspection)
}

// Port executes a one-directional config transfer from source to target.
func Port(ctx context.Context, opts Options) (*Result, error) {
	result := &Result{
		Source: opts.Source,
		Target: opts.Target,
	}

	// Get adapters
	srcAdapter, err := registry.Get(opts.Source)
	if err != nil {
		return nil, fmt.Errorf("source adapter: %w", err)
	}

	tgtAdapter, err := registry.Get(opts.Target)
	if err != nil {
		return nil, fmt.Errorf("target adapter: %w", err)
	}

	// Step 1: Read from source
	readOpts := adapter.ReadOptions{
		Scope:       opts.Scope,
		ProjectRoot: opts.ProjectRoot,
	}
	if opts.Scope == "" {
		readOpts.Scope = "all"
	}

	cfg, err := srcAdapter.Read(ctx, readOpts)
	if err != nil {
		return nil, fmt.Errorf("read from %s: %w", opts.Source, err)
	}
	result.Config = cfg

	// Step 2: Filter aspects if specified
	if len(opts.Aspects) > 0 {
		cfg = filterAspects(cfg, opts.Aspects)
	}

	// Step 3: Update metadata for target
	cfg.Meta.SourceAdapter = opts.Source

	// Step 4: Write to target
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

	writeResult, err := tgtAdapter.Write(ctx, cfg, writeOpts)
	if err != nil {
		return nil, fmt.Errorf("write to %s: %w", opts.Target, err)
	}

	result.FilesWritten = writeResult.FilesWritten
	result.FilesSkipped = writeResult.FilesSkipped
	result.LossWarnings = append(cfg.LossWarnings, writeResult.LossWarnings...)

	return result, nil
}

// Preview runs a port in dry-run mode and returns the diff.
func Preview(ctx context.Context, opts Options) (*adapter.DiffResult, error) {
	opts.DryRun = true

	srcAdapter, err := registry.Get(opts.Source)
	if err != nil {
		return nil, fmt.Errorf("source adapter: %w", err)
	}

	tgtAdapter, err := registry.Get(opts.Target)
	if err != nil {
		return nil, fmt.Errorf("target adapter: %w", err)
	}

	readOpts := adapter.ReadOptions{
		Scope:       opts.Scope,
		ProjectRoot: opts.ProjectRoot,
	}
	if opts.Scope == "" {
		readOpts.Scope = "all"
	}

	cfg, err := srcAdapter.Read(ctx, readOpts)
	if err != nil {
		return nil, fmt.Errorf("read from %s: %w", opts.Source, err)
	}

	if len(opts.Aspects) > 0 {
		cfg = filterAspects(cfg, opts.Aspects)
	}

	return tgtAdapter.Preview(ctx, cfg)
}

// filterAspects returns a new ShifterConfig containing only the requested aspects.
func filterAspects(cfg *canonical.ShifterConfig, aspects []string) *canonical.ShifterConfig {
	include := make(map[string]bool)
	for _, a := range aspects {
		include[a] = true
	}

	filtered := &canonical.ShifterConfig{
		Meta: cfg.Meta,
	}

	if !include["instructions"] {
		filtered.Instructions = nil
	} else {
		filtered.Instructions = cfg.Instructions
	}
	if !include["agents"] {
		filtered.Agents = nil
	} else {
		filtered.Agents = cfg.Agents
	}
	if !include["skills"] {
		filtered.Skills = nil
	} else {
		filtered.Skills = cfg.Skills
	}
	if !include["commands"] {
		filtered.Commands = nil
	} else {
		filtered.Commands = cfg.Commands
	}
	if !include["mcp"] {
		filtered.MCPServers = nil
	} else {
		filtered.MCPServers = cfg.MCPServers
	}
	if !include["permissions"] {
		filtered.Permissions = nil
	} else {
		filtered.Permissions = cfg.Permissions
	}
	if !include["hooks"] {
		filtered.Hooks = nil
	} else {
		filtered.Hooks = cfg.Hooks
	}
	if !include["settings"] {
		filtered.Settings = canonical.SettingsMap{}
	} else {
		filtered.Settings = cfg.Settings
	}

	return filtered
}
