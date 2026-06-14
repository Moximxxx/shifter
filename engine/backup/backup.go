// Package backup provides automatic backup creation before config writes.
//
// Every write operation creates a timestamped tar archive by default.
// Backups are stored in ~/.shifter/backups/ and rotated to prevent unbounded growth.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sort"
	"time"

	"github.com/moximxxx/shifter/pkg/paths"
)

const (
	// DefaultBackupDir is where backup archives are stored.
	DefaultBackupDir = ".shifter/backups"
	// MaxBackups is the maximum number of backups to retain.
	MaxBackups = 20
)

// Create creates a timestamped tar.gz backup of the specified files.
// Returns the path to the created archive.
func Create(files []string) (string, error) {
	home := paths.MustHomeDir()
backupDir := filepath.Join(home, DefaultBackupDir)
	if err := os.MkdirAll(backupDir, paths.DirPerm); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02T150405")
	archiveName := fmt.Sprintf("shifter-backup-%s.tar.gz", timestamp)
	archivePath := filepath.Join(backupDir, archiveName)

	// Create the archive
	f, err := os.Create(archivePath)
	if err != nil {
		return "", fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		if err := addFileToTar(tw, file); err != nil {
			// If file doesn't exist, skip it (nothing to back up)
			if os.IsNotExist(err) {
				continue
			}
			return "", fmt.Errorf("add %s to archive: %w", file, err)
		}
	}

	// Rotate old backups
	rotate(backupDir)

	return archivePath, nil
}

// Restore restores files from a backup archive to their original locations.
func Restore(archivePath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gunzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		// Path traversal protection
		cleanName := filepath.Clean(header.Name)
		if strings.Contains(cleanName, "..") {
			return fmt.Errorf("unsafe path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(cleanName, paths.DirPerm); err != nil {
				return fmt.Errorf("create dir %s: %w", header.Name, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(cleanName), paths.DirPerm); err != nil {
				return fmt.Errorf("create parent dir for %s: %w", header.Name, err)
			}
			outFile, err := os.Create(cleanName)
			if err != nil {
				return fmt.Errorf("create file %s: %w", header.Name, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("write file %s: %w", header.Name, err)
			}
			outFile.Close()
		}
	}

	return nil
}

// List returns all backup archives sorted by time (newest first).
func List() ([]string, error) {
	home := paths.MustHomeDir()
backupDir := filepath.Join(home, DefaultBackupDir)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	// Sort by name (which includes timestamp), newest first
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))
	return backups, nil
}

// Latest returns the path to the most recent backup, or empty string if none.
func Latest() (string, error) {
	backups, err := List()
	if err != nil {
		return "", err
	}
	if len(backups) == 0 {
		return "", nil
	}
	return backups[0], nil
}

func addFileToTar(tw *tar.Writer, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Walk the directory and add all files
		return filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			return addSingleFileToTar(tw, path, info)
		})
	}

	return addSingleFileToTar(tw, filePath, info)
}

func addSingleFileToTar(tw *tar.Writer, filePath string, info os.FileInfo) error {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filePath // store with original path

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

func rotate(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			backups = append(backups, entry)
		}
	}

	// Sort by name (timestamp), oldest first
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Name() < backups[j].Name()
	})

	// Delete oldest if over limit
	for len(backups) > MaxBackups {
		oldest := filepath.Join(dir, backups[0].Name())
		os.Remove(oldest)
		backups = backups[1:]
	}
}
