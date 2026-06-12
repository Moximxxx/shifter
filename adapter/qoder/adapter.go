package qoder

import (
	"context"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string             { return "qoder" }
func (a *Adapter) Name() string           { return "Qoder" }
func (a *Adapter) SearchPaths() []string  { return []string{"~/.qoder", ".qoder"} }
func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	return adapter.DetectionResult{}, nil
}
func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsAgents: true,
		SupportsSkills: true,
		SupportsMCP:    true,
	}
}
func (a *Adapter) Read(ctx context.Context, opts adapter.ReadOptions) (*canonical.ShifterConfig, error) {
	return &canonical.ShifterConfig{}, nil
}
func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	return &adapter.WriteResult{}, nil
}
func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}
