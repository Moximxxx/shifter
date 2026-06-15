package test

import (
	"context"
	"testing"
	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/registry"
)

func BenchmarkDetect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		detect.ScanAll()
	}
}

func BenchmarkPort(b *testing.B) {
	a, _ := registry.Get("claude-code")
	cfg, _ := a.Read(context.Background(), adapter.ReadOptions{
		Scope: "project",
		ProjectRoot: "test/claude-code-test",
	})
	for i := 0; i < b.N; i++ {
		a.Write(context.Background(), cfg, adapter.WriteOptions{
			Scope: "project", ProjectRoot: b.TempDir(),
		})
	}
}
