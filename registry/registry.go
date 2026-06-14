// Package registry provides the global adapter registry.
// It imports all adapter packages and maps agent IDs to constructors.
package registry

import (
"fmt"
	"sort"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/adapter/aider"
	"github.com/moximxxx/shifter/adapter/claudecode"
	"github.com/moximxxx/shifter/adapter/cline"
	"github.com/moximxxx/shifter/adapter/codex"
	"github.com/moximxxx/shifter/adapter/gemini"
	"github.com/moximxxx/shifter/adapter/opencode"
	"github.com/moximxxx/shifter/adapter/qoder"
)

// adapterMap maps agent IDs to constructor functions.
var adapterMap = map[string]func() adapter.AgentAdapter{
	"claude-code": func() adapter.AgentAdapter { return claudecode.New() },
	"codex":       func() adapter.AgentAdapter { return codex.New() },
	"opencode":    func() adapter.AgentAdapter { return opencode.New() },
	"gemini-cli":  func() adapter.AgentAdapter { return gemini.New() },
	"qoder":       func() adapter.AgentAdapter { return qoder.New() },
	"cline":       func() adapter.AgentAdapter { return cline.New() },
	"aider":       func() adapter.AgentAdapter { return aider.New() },
}

// Get returns an adapter by ID, or an error if not found.
func Get(id string) (adapter.AgentAdapter, error) {
	ctor, ok := adapterMap[id]
	if !ok {
		return nil, fmt.Errorf("unknown adapter: %q (available: %v)", id, ListIDs())
	}
	return ctor(), nil
}

// ListIDs returns all registered adapter IDs in sorted order.
func ListIDs() []string {
	ids := make([]string, 0, len(adapterMap))
	for id := range adapterMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// MustGet returns an adapter by ID or panics. For tests only.
func MustGet(id string) adapter.AgentAdapter {
	a, err := Get(id)
	if err != nil {
		panic(err)
	}
	return a
}

