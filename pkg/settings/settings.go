// Package settings manages user settings stored in ~/.shifter/settings.json.
package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/moximxxx/shifter/pkg/paths"
)

// UserSettings holds all configurable options.
type UserSettings struct {
	Lang      string `json:"lang"`       // "en" or "zh"
	FirstRun  bool   `json:"first_run"`  // false after initial setup
}

const (
	settingsDir  = ".shifter"
	settingsFile = "settings.json"
)

var (
	mu     sync.RWMutex
	cached *UserSettings
)

// Load reads settings from disk, or returns defaults if not found.
func Load() (*UserSettings, error) {
	mu.RLock()
	if cached != nil {
		mu.RUnlock()
		return cached, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	path, err := settingsPath()
	if err != nil {
		return defaults(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// First run — return defaults with first_run=true
		s := defaults()
		s.FirstRun = true
		return s, nil
	}

	var s UserSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return defaults(), nil
	}

	cached = &s
	return &s, nil
}

// Save persists settings to disk.
func Save(s *UserSettings) error {
	mu.Lock()
	defer mu.Unlock()

	path, err := settingsPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create settings dir: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	cached = s
	return nil
}

// IsFirstRun returns true if this is the first time Shifter has been run.
func IsFirstRun() bool {
	s, err := Load()
	if err != nil {
		return true
	}
	return s.FirstRun
}

// defaults returns default settings (en, first_run=true).
func defaults() *UserSettings {
	return &UserSettings{
		Lang:     "en",
		FirstRun: true,
	}
}

func settingsPath() (string, error) {
	home := paths.MustHomeDir()
return filepath.Join(home, settingsDir, settingsFile), nil
}
