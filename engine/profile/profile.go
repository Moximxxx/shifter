// Package profile provides profile management for Shifter.
//
// Profiles are saved canonical configs stored in ~/.shifter/profiles/.
// They allow users to capture a project's agent configuration and
// apply it later in another project.
package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/moximxxx/shifter/canonical"
)

const profilesDir = ".shifter/profiles"

// Profile is a saved configuration snapshot with metadata.
type Profile struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	SourceAgent string                   `json:"source_agent"`
	Config      *canonical.ShifterConfig `json:"config"`
}

// Save saves a profile to ~/.shifter/profiles/<name>.json.
func Save(p Profile) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	dir := filepath.Join(home, profilesDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create profiles dir: %w", err)
	}

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	p.UpdatedAt = time.Now()

	filename := sanitizeName(p.Name) + ".json"
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Load loads a profile by name from ~/.shifter/profiles/.
func Load(name string) (*Profile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	path := filepath.Join(home, profilesDir, sanitizeName(name)+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile %q not found", name)
		}
		return nil, fmt.Errorf("read profile: %w", err)
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}

	return &p, nil
}

// List returns all saved profiles sorted by update time (newest first).
func List() ([]Profile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, profilesDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var profiles []Profile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var p Profile
		if err := json.Unmarshal(data, &p); err != nil {
			continue
		}
		profiles = append(profiles, p)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].UpdatedAt.After(profiles[j].UpdatedAt)
	})

	return profiles, nil
}

// Delete removes a profile by name.
func Delete(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, profilesDir, sanitizeName(name)+".json")
	return os.Remove(path)
}

// Summary returns a one-line description of a profile's contents.
func (p Profile) Summary() string {
	cfg := p.Config
	if cfg == nil {
		return "empty"
	}
	var parts []string
	if len(cfg.Agents) > 0 {
		parts = append(parts, fmt.Sprintf("%d agents", len(cfg.Agents)))
	}
	if len(cfg.Skills) > 0 {
		parts = append(parts, fmt.Sprintf("%d skills", len(cfg.Skills)))
	}
	if len(cfg.Commands) > 0 {
		parts = append(parts, fmt.Sprintf("%d commands", len(cfg.Commands)))
	}
	if len(cfg.MCPServers) > 0 {
		parts = append(parts, fmt.Sprintf("%d MCP", len(cfg.MCPServers)))
	}
	if len(cfg.Hooks) > 0 {
		parts = append(parts, fmt.Sprintf("%d hooks", len(cfg.Hooks)))
	}
	if len(parts) == 0 {
		return "instructions only"
	}
	return strings.Join(parts, ", ")
}

func sanitizeName(name string) string {
	// Replace spaces and special chars with hyphens
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '-'
		}
		return -1
	}, name)
	if name == "" {
		name = "unnamed"
	}
	return name
}
