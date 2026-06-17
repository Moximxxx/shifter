package flowhub

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const apiBase = "https://api.github.com"

// PublishRequest contains all files to publish.
type PublishRequest struct {
	Name     string
	Files    map[string][]byte // relative path → content
	Message  string
	Token    string // GitHub personal access token
}

// Publish submits a workflow to FlowHub via the GitHub API.
// Creates a new branch and opens a Pull Request automatically.
func Publish(req PublishRequest) (string, error) {
	if req.Token == "" {
		req.Token = os.Getenv("GITHUB_TOKEN")
	}
	if req.Token == "" {
		return "", fmt.Errorf("GITHUB_TOKEN not set. Create one at https://github.com/settings/tokens")
	}

	// 1. Get default branch SHA
	defaultBranch, sha, err := getDefaultBranch(req.Token)
	if err != nil {
		return "", fmt.Errorf("get branch: %w", err)
	}

	// 2. Create new branch
	branchName := fmt.Sprintf("workflow/%s", req.Name)
	if err := createBranch(req.Token, branchName, sha); err != nil {
		// Branch may already exist — try with timestamp suffix
		branchName = fmt.Sprintf("workflow/%s-%d", req.Name, sha[:6])
		if err := createBranch(req.Token, branchName, sha); err != nil {
			return "", fmt.Errorf("create branch: %w", err)
		}
	}

	// 3. Commit files
	for path, content := range req.Files {
		fullPath := fmt.Sprintf("workflows/%s/%s", req.Name, path)
		if err := createOrUpdateFile(req.Token, fullPath, content, req.Message, branchName); err != nil {
			return "", fmt.Errorf("upload %s: %w", path, err)
		}
	}

	// 4. Create Pull Request
	prURL, err := createPR(req.Token, branchName, defaultBranch, req.Name, req.Message)
	if err != nil {
		return "", fmt.Errorf("create PR: %w", err)
	}

	return prURL, nil
}

func getDefaultBranch(token string) (string, string, error) {
	url := fmt.Sprintf("%s/repos/Moximxxx/flowhub/git/refs/heads/main", apiBase)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var ref struct {
		Ref    string `json:"ref"`
		Object struct {
			Sha string `json:"sha"`
		} `json:"object"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ref); err != nil {
		return "", "", err
	}
	return "main", ref.Object.Sha, nil
}

func createBranch(token, branch, sha string) error {
	url := fmt.Sprintf("%s/repos/Moximxxx/flowhub/git/refs", apiBase)
	body := fmt.Sprintf(`{"ref":"refs/heads/%s","sha":"%s"}`, branch, sha)
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 && resp.StatusCode != 422 { // 422 = branch exists
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

func createOrUpdateFile(token, path string, content []byte, message, branch string) error {
	url := fmt.Sprintf("%s/repos/Moximxxx/flowhub/contents/%s", apiBase, path)
	encoded := base64.StdEncoding.EncodeToString(content)
	body := fmt.Sprintf(`{"message":"%s","content":"%s","branch":"%s"}`, message, encoded, branch)
	req, _ := http.NewRequest("PUT", url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		// File may exist — try to update
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, buf.String())
	}
	return nil
}

func createPR(token, head, base, name, msg string) (string, error) {
	url := fmt.Sprintf("%s/repos/Moximxxx/flowhub/pulls", apiBase)
	title := fmt.Sprintf("Add workflow: %s", name)
	body := fmt.Sprintf(`{
		"title": "%s",
		"head": "%s",
		"base": "%s",
		"body": "%s"
	}`, escape(title), head, base, escape(msg))

	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, buf.String())
	}

	var pr struct {
		HTMLURL string `json:"html_url"`
	}
	json.NewDecoder(resp.Body).Decode(&pr)
	return pr.HTMLURL, nil
}

func escape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
