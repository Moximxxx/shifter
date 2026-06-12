// Package detect implements filesystem scanning for configured coding agents.
package detect

import (
	"sync"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/registry"
)

// Result holds detection info for a single agent.
type Result struct {
	ID      string                  `json:"id"`
	Name    string                  `json:"name"`
	Found   bool                    `json:"found"`
	Paths   []string                `json:"paths,omitempty"`
	Summary map[string]int          `json:"summary,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// ScanAll concurrently detects all registered coding agents.
func ScanAll() []Result {
	ids := registry.ListIDs()

	var results []Result
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range ids {
		wg.Add(1)
		go func(agentID string) {
			defer wg.Done()

			a, err := registry.Get(agentID)
			if err != nil {
				mu.Lock()
				results = append(results, Result{
					ID:    agentID,
					Name:  agentID,
					Found: false,
					Error: err.Error(),
				})
				mu.Unlock()
				return
			}

			dr, err := a.Detect()
			mu.Lock()
			defer mu.Unlock()

			r := Result{
				ID:      agentID,
				Name:    a.Name(),
				Found:   dr.Found,
				Summary: dr.Summary,
			}
			r.Paths = append(r.Paths, dr.GlobalPaths...)
			if dr.ProjectPath != "" {
				r.Paths = append(r.Paths, dr.ProjectPath)
			}
			if err != nil {
				r.Error = err.Error()
			}
			results = append(results, r)
		}(id)
	}

	wg.Wait()

	// Sort results by ID for consistent output
	sortResults(results)
	return results
}

func sortResults(results []Result) {
	// Sort by ID for deterministic output
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].ID > results[j].ID {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// Ensure unused imports are resolved
var _ = adapter.DetectionResult{}
