package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear cached settings
	mu.Lock()
	cached = nil
	mu.Unlock()

	// Point to a non-existent path
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/tmp/shifter-test-nonexistent")
	defer os.Setenv("HOME", oldHome)

	s, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if s.Lang != "en" {
		t.Errorf("default lang: got %q, want en", s.Lang)
	}
	if !s.FirstRun {
		t.Error("default first_run should be true")
	}
}

func TestSave_Load_RoundTrip(t *testing.T) {
	mu.Lock()
	cached = nil
	mu.Unlock()

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".shifter")
	os.MkdirAll(configDir, 0755)

	// Override home for test
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	s := &UserSettings{Lang: "zh", FirstRun: false}
	if err := Save(s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(configDir, "settings.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("settings.json not created")
	}

	// Clear cache to force re-read
	mu.Lock()
	cached = nil
	mu.Unlock()

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Lang != "zh" {
		t.Errorf("round-trip lang: got %q, want zh", loaded.Lang)
	}
	if loaded.FirstRun {
		t.Error("round-trip first_run should be false")
	}
}

func TestIsFirstRun(t *testing.T) {
	mu.Lock()
	cached = nil
	mu.Unlock()

	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	if !IsFirstRun() {
		t.Error("should be first run in empty dir")
	}

	Save(&UserSettings{Lang: "en", FirstRun: false})

	mu.Lock()
	cached = nil
	mu.Unlock()

	if IsFirstRun() {
		t.Error("should not be first run after save with FirstRun=false")
	}
}

func TestSave_InvalidPath(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "") // Clear home to cause error
	defer os.Setenv("HOME", oldHome)

	mu.Lock()
	cached = nil
	mu.Unlock()

	err := Save(&UserSettings{Lang: "en"})
	if err == nil {
		t.Error("expected error when home dir is empty")
	}
}

func TestDefaults(t *testing.T) {
	s := defaults()
	if s.Lang != "en" {
		t.Errorf("default lang: got %q", s.Lang)
	}
	if !s.FirstRun {
		t.Error("default first_run should be true")
	}
}

func TestCaching(t *testing.T) {
	mu.Lock()
	cached = nil
	mu.Unlock()

	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	Save(&UserSettings{Lang: "zh", FirstRun: false})

	// First load
	s1, _ := Load()
	// Second load should use cache
	s2, _ := Load()

	if s1.Lang != s2.Lang {
		t.Error("cached and uncached loads differ")
	}
}
