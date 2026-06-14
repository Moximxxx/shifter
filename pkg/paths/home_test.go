package paths

import (
	"os"
	"testing"
)

func TestMustHomeDir(t *testing.T) {
	home := MustHomeDir()
	if home == "" {
		t.Error("MustHomeDir returned empty string")
	}
}

func TestMustHomeDir_Fallback(t *testing.T) {
	oldHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	defer os.Setenv("HOME", oldHome)

	home := MustHomeDir()
	if home != "/tmp" {
		t.Errorf("HOME unset: got %q, want /tmp", home)
	}
}

func TestDirPerm(t *testing.T) {
	if DirPerm != 0755 {
		t.Errorf("DirPerm: got %o, want 0755", DirPerm)
	}
}

func TestFilePerm(t *testing.T) {
	if FilePerm != 0644 {
		t.Errorf("FilePerm: got %o, want 0644", FilePerm)
	}
}
