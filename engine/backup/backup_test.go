package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndList(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a test file to back up
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	archivePath, err := Create([]string{testFile})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if archivePath == "" {
		t.Fatal("no archive path returned")
	}
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("archive not created on disk")
	}

	backups, err := List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup, got %d", len(backups))
	}
}

func TestRestore(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create file to back up
	testFile := filepath.Join(tmpDir, "restore-test.txt")
	os.WriteFile(testFile, []byte("restore me"), 0644)

	archivePath, err := Create([]string{testFile})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Delete original
	os.Remove(testFile)

	// Restore
	if err := Restore(archivePath); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("restored file not found: %v", err)
	}
	if string(data) != "restore me" {
		t.Errorf("restored content: got %q, want 'restore me'", string(data))
	}
}

func TestCreate_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create with a non-existent file — should not error, just skip
	_, err := Create([]string{"/tmp/nonexistent-file-12345.xyz"})
	if err != nil {
		t.Fatalf("Create with nonexistent file should not error: %v", err)
	}
}

func TestLatest_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	path, err := Latest()
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty string, got %q", path)
	}
}

func TestRotate(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)

	backupDir := filepath.Join(tmpDir, ".shifter", "backups")
	os.MkdirAll(backupDir, 0755)
	os.Setenv("HOME", tmpDir)

	// Create MaxBackups + 5 fake backup files
	for i := 0; i < MaxBackups+5; i++ {
		f, _ := os.Create(filepath.Join(backupDir, "backup-"+string(rune('a'+i))+".tar.gz"))
		f.Close()
	}

	// Call rotate
	rotate(backupDir)

	// Verify only MaxBackups remain
	entries, _ := os.ReadDir(backupDir)
	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			count++
		}
	}
	if count > MaxBackups {
		t.Errorf("rotate: got %d backups, want <= %d", count, MaxBackups)
	}
}
