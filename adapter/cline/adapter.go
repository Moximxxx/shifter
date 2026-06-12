package cline

import (
	"context"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string             { return "cline" }
func (a *Adapter) Name() string           { return "Cline" }
func (a *Adapter) SearchPaths() []string  { return []string{".clinerules"} }
func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	return adapter.DetectionResult{}, nil
}
func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
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
