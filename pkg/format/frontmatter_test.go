package format

import (
	"strings"
	"testing"
)

func TestParseFrontMatter(t *testing.T) {
	input := `---
name: test-skill
description: A test skill
tools:
  - Read
  - Write
---
This is the body content.`

	fm, body, err := ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("ParseFrontMatter failed: %v", err)
	}

	if fm["name"] != "test-skill" {
		t.Errorf("name = %q, want %q", fm["name"], "test-skill")
	}
	if fm["description"] != "A test skill" {
		t.Errorf("description mismatch")
	}

	tools, ok := fm["tools"].([]interface{})
	if !ok {
		t.Fatal("tools should be a list")
	}
	if len(tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(tools))
	}

	expectedBody := "This is the body content."
	if strings.TrimSpace(body) != expectedBody {
		t.Errorf("body = %q, want %q", body, expectedBody)
	}
}

func TestParseFrontMatterNoFrontmatter(t *testing.T) {
	input := "Just some markdown content without frontmatter."

	fm, body, err := ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("ParseFrontMatter failed: %v", err)
	}

	if fm != nil {
		t.Errorf("expected nil frontmatter, got %v", fm)
	}
	if body != input {
		t.Errorf("body should be unchanged")
	}
}

func TestParseFrontMatterSimpleFrontmatter(t *testing.T) {
	input := `---
name: code-reviewer
description: Review code for quality
tools: Read, Glob, Grep
model: sonnet
---
You are a code review specialist...`

	fm, body, err := ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("ParseFrontMatter failed: %v", err)
	}

	if fm["name"] != "code-reviewer" {
		t.Errorf("name = %q", fm["name"])
	}
	if fm["model"] != "sonnet" {
		t.Errorf("model = %q", fm["model"])
	}
	if fm["tools"] != "Read, Glob, Grep" {
		t.Errorf("tools = %q", fm["tools"])
	}
	if !strings.Contains(body, "code review specialist") {
		t.Errorf("body should contain the prompt text")
	}
}
