package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Shifter to the latest version",
	Long: `Download and install the latest Shifter release from GitHub.

This replaces the currently running binary with the newest version.

Examples:
  shifter upgrade              # upgrade to latest
  shifter upgrade -v 0.2.0     # upgrade to specific version`,
	RunE: runUpgrade,
}

var upgradeVersion string

func init() {
	upgradeCmd.Flags().StringVarP(&upgradeVersion, "version", "v", "", "Specific version (default: latest)")
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find current binary: %w", err)
	}

	platform := runtime.GOOS + "-" + runtime.GOARCH
	if strings.Contains(runtime.GOARCH, "amd64") {
		platform = runtime.GOOS + "-amd64"
	}

	targetVersion := upgradeVersion
	if targetVersion == "" {
		latest, err := fetchLatestVersion()
		if err != nil {
			return fmt.Errorf("cannot fetch latest version: %w\n  Use -v to specify version", err)
		}
		targetVersion = latest
	}

	// Strip "v" prefix if present for URL building
	verNum := strings.TrimPrefix(targetVersion, "v")

	downloadURL := fmt.Sprintf(
		"https://github.com/Moximxxx/shifter/releases/download/v%s/shifter_%s_%s.tar.gz",
		verNum, verNum, platform,
	)

	fmt.Printf("  Current:  %s\n", VersionString())
	fmt.Printf("  Platform: %s\n", platform)
	fmt.Printf("  Download: %s\n", downloadURL)

	// Download
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d — version %q may not exist for %s", resp.StatusCode, targetVersion, platform)
	}

	// Extract
	newBinary, err := extractBinary(resp.Body, platform)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	defer os.Remove(newBinary)

	// Install — replace current binary
	if err := os.Rename(newBinary, binaryPath); err != nil {
		// Cross-device — copy instead
		src, _ := os.Open(newBinary)
		if src == nil {
			return fmt.Errorf("cannot open new binary")
		}
		defer src.Close()

		// Write to a temp file on the same device, then rename
		tmpPath := binaryPath + ".new"
		dst, err := os.Create(tmpPath)
		if err != nil {
			return fmt.Errorf("cannot write to %s (try sudo): %w", filepath.Dir(binaryPath), err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			return fmt.Errorf("copy binary: %w", err)
		}
		dst.Close()
		os.Chmod(tmpPath, 0755)
		if err := os.Rename(tmpPath, binaryPath); err != nil {
			return fmt.Errorf("replace binary: %w", err)
		}
	}

	fmt.Printf("\n✓ Upgraded to v%s\n", verNum)
	fmt.Println("Run 'shifter --version' to verify.")
	return nil
}

// fetchLatestVersion gets the latest release tag from GitHub API.
func fetchLatestVersion() (string, error) {
	url := "https://api.github.com/repos/Moximxxx/shifter/releases/latest"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Use token for authenticated request (higher rate limit)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else {
		// Try gh CLI token
		home, _ := os.UserHomeDir()
		data, _ := os.ReadFile(home + "/.config/gh/hosts.yml")
		if data != nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.Contains(line, "oauth_token:") || strings.Contains(line, "token:") {
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						token := strings.TrimSpace(parts[1])
						req.Header.Set("Authorization", "Bearer "+token)
						break
					}
				}
			}
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned %d (rate limited? set GITHUB_TOKEN)", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no release found")
	}
	return release.TagName, nil
}

// extractBinary extracts the shifter binary from a tar.gz archive.
func extractBinary(r io.Reader, platform string) (string, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", fmt.Errorf("gunzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar: %w", err)
		}
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "shifter") {
			tmpFile, err := os.CreateTemp("", "shifter-upgrade-*")
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				return "", err
			}
			tmpFile.Close()
			os.Chmod(tmpFile.Name(), 0755)
			return tmpFile.Name(), nil
		}
	}
	return "", fmt.Errorf("shifter binary not found in archive")
}
