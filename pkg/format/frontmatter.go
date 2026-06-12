// Package format provides utilities for parsing and writing configuration formats.
package format

import (
	"strings"

	"gopkg.in/yaml.v3"
)

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
