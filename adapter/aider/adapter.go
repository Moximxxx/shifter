package aider

import (
	"context"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string             { return "aider" }
func (a *Adapter) Name() string           { return "Aider" }
func (a *Adapter) SearchPaths() []string  { return []string{"~/.aider.conf.yml", ".aider.conf.yml"} }
func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	return adapter.DetectionResult{}, nil
}
func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		SupportsPermissions:  true,
		SupportsSettings:     true,
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
