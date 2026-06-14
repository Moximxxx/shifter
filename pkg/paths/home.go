// Package paths provides common path utilities for Shifter.
package paths

import "os"

// DirPerm is the default permission for directories created by Shifter.
const DirPerm = 0755

// FilePerm is the default permission for files created by Shifter.
const FilePerm = 0644

// MustHomeDir returns the user's home directory, or "/tmp" as fallback.
func MustHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "/tmp"
	}
	return home
}
