// Package flowhub provides the GitHub-backed workflow marketplace.
// All data is stored in Moximxxx/flowhub on GitHub — zero server cost.
package flowhub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

const (
	// RepoURL is the FlowHub registry repository.
	RepoURL = "https://github.com/Moximxxx/flowhub"
	// IndexURL is the raw index.json URL.
	IndexURL = "https://raw.githubusercontent.com/Moximxxx/flowhub/main/index.json"
	// WorkflowURL formats a raw workflow download URL.
	WorkflowURL = "https://raw.githubusercontent.com/Moximxxx/flowhub/main/workflows/%s/workflow.shifter.json"
	// SkillURL formats a raw SKILL.md download URL.
	SkillURL = "https://raw.githubusercontent.com/Moximxxx/flowhub/main/workflows/%s/SKILL.md"
)

// Workflow represents a published workflow in the FlowHub registry.
type Workflow struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	Agent       string   `json:"agent"`
	Tags        []string `json:"tags"`
	Category    string   `json:"category,omitempty"`
	Description string   `json:"description"`
	Downloads   int      `json:"downloads"`
	Updated     string   `json:"updated"`
}

// Index is the registry index file.
type Index struct {
	Workflows []Workflow `json:"workflows"`
}

// FetchIndex downloads the FlowHub registry index.
// Uses GitHub API to avoid CDN caching issues.
func FetchIndex() (*Index, error) {
	// Use GitHub API for fresh content
	apiURL := "https://api.github.com/repos/Moximxxx/flowhub/contents/index.json"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Fallback to raw URL
		resp, err = http.Get(IndexURL)
		if err != nil {
			return nil, fmt.Errorf("fetch index: %w", err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index not available (HTTP %d)", resp.StatusCode)
	}

	var idx Index
	if err := json.NewDecoder(resp.Body).Decode(&idx); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}
	return &idx, nil
}

// Search queries the FlowHub registry for matching workflows.
func Search(query string) ([]Workflow, error) {
	idx, err := FetchIndex()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []Workflow
	for _, w := range idx.Workflows {
		if query == "" || matchWorkflow(w, query) {
			results = append(results, w)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Downloads > results[j].Downloads
	})
	return results, nil
}

func matchWorkflow(w Workflow, q string) bool {
	if strings.Contains(strings.ToLower(w.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(w.Description), q) {
		return true
	}
	for _, tag := range w.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

// Download fetches a workflow's canonical JSON from FlowHub.
func Download(name string) ([]byte, error) {
	url := fmt.Sprintf(WorkflowURL, name)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("workflow %q not found in FlowHub", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed (HTTP %d)", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Metadata represents workflow metadata for publishing.
type Metadata struct {
	Name        string   `json:"name"`
	Version     string   `json:"version,omitempty"`
	Author      string   `json:"author,omitempty"`
	Agent       string   `json:"agent"`
	Tags        []string `json:"tags,omitempty"`
	Category    string   `json:"category,omitempty"`
	Description string   `json:"description,omitempty"`
}

// GenerateMetadata creates a metadata.json from workflow info.
func GenerateMetadata(name, agent, desc string, tags []string) Metadata {
	return Metadata{
		Name:        name,
		Version:     "0.1.0",
		Author:      "",
		Agent:       agent,
		Tags:        tags,
		Description: desc,
	}
}
