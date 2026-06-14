// Package format provides utilities for parsing and writing configuration formats.
package format

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// StripBOM removes a UTF-8 BOM (byte order mark) from the start of content.
func StripBOM(content string) string {
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		return content[3:]
	}
	return content
}

// StripJSONComments removes // line comments from JSON content.
// This allows reading JSONC (JSON with Comments) files.
// Also handles UTF-8 BOM removal.
func StripJSONComments(content string) string {
	content = StripBOM(content)
	var result strings.Builder
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Remove // comments, but not inside strings
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		// Handle trailing comments (naive — good enough for config files)
		if idx := strings.Index(line, " //"); idx >= 0 {
			line = line[:idx]
		}
		result.WriteString(line)
		result.WriteString("\n")
	}
	return result.String()
}

// ParseFrontMatter extracts YAML frontmatter and body from Markdown content.
// Returns the frontmatter as a map, the body text, and any error.
func ParseFrontMatter(content string) (map[string]interface{}, string, error) {
	// Frontmatter is delimited by --- lines
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return nil, content, nil
	}

	// Find closing ---
	rest := content[3:] // skip opening ---
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	} else if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		// No closing delimiter — not frontmatter
		return nil, content, nil
	}

	yamlContent := rest[:endIdx]
	body := rest[endIdx+4:] // skip \n---

	// Strip leading newline from body
	body = strings.TrimLeft(body, "\n\r")

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, content, err
	}

	if fm == nil {
		fm = make(map[string]interface{})
	}

	return fm, body, nil
}

// GetFMString extracts a string field from a frontmatter map.
func GetFMString(fm map[string]interface{}, key string) string {
	if v, ok := fm[key].(string); ok {
		return v
	}
	return ""
}

// GetFMStringSlice extracts a string slice from a frontmatter map,
// splitting on commas if the value is a string instead of a list.
func GetFMStringSlice(fm map[string]interface{}, key string) []string {
	switch v := fm[key].(type) {
	case []interface{}:
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case string:
		var result []string
		for _, s := range strings.Split(v, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// ReadFrontmatterFile reads a Markdown file and parses its YAML frontmatter.
func ReadFrontmatterFile(path string) (map[string]interface{}, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return ParseFrontMatter(string(data))
}

// FormatYAMLFrontmatter marshals a map to YAML wrapped in frontmatter delimiters.
func FormatYAMLFrontmatter(fields map[string]interface{}) string {
	fmBytes, err := yaml.Marshal(fields)
	if err != nil {
		return "---\n---\n"
	}
	return "---\n" + string(fmBytes) + "---\n"
}
