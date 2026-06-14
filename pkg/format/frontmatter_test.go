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

func TestStripBOM_WithBOM(t *testing.T) {
	input := "\xEF\xBB\xBF" + `{"key": "value"}`
	result := StripBOM(input)
	if len(result) != len(input)-3 {
		t.Errorf("BOM not stripped: got len %d, want %d", len(result), len(input)-3)
	}
	if result[0] != '{' {
		t.Error("first char should be { after BOM removal")
	}
}

func TestStripBOM_WithoutBOM(t *testing.T) {
	input := `{"key": "value"}`
	result := StripBOM(input)
	if result != input {
		t.Error("should be unchanged without BOM")
	}
}

func TestStripJSONComments(t *testing.T) {
	input := `{
  "key": "value",
  // this is a comment
  "other": true // trailing comment
}
// full line comment
`
	result := StripJSONComments(input)
	if strings.Contains(result, "//") {
		t.Error("comments should be stripped")
	}
	if !strings.Contains(result, `"key"`) {
		t.Error("content should be preserved")
	}
	if !strings.Contains(result, `"other"`) {
		t.Error("content after trailing comment should be preserved")
	}
}

func TestStripJSONComments_NoComments(t *testing.T) {
	input := `{"key": "value"}`
	result := StripJSONComments(input)
	if result != input+"\n" {
		t.Errorf("unchanged: got %q, want %q", result, input)
	}
}

func TestParseFrontMatter_EmptyFile(t *testing.T) {
	fm, body, err := ParseFrontMatter("")
	if err != nil {
		t.Errorf("empty file should not error: %v", err)
	}
	if fm != nil {
		t.Error("empty file should return nil frontmatter")
	}
	if body != "" {
		t.Error("empty file body should be empty")
	}
}

func TestParseFrontMatter_OnlyDelimiters(t *testing.T) {
	input := "---\nkey: value\n---\n"
	fm, body, err := ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("only delimiters: %v", err)
	}
	if body != "" {
		t.Errorf("body should be empty, got %q", body)
	}
	if fm == nil || fm["key"] != "value" {
		t.Errorf("frontmatter not parsed: %v", fm)
	}
}

func TestParseFrontMatter_ChineseContent(t *testing.T) {
	input := "---\nname: 代码审查专家\ndescription: 审查代码质量和安全性\n---\n你是一位资深代码审查专家。"
	fm, body, err := ParseFrontMatter(input)
	if err != nil {
		t.Fatalf("chinese frontmatter: %v", err)
	}
	if fm["name"] != "代码审查专家" {
		t.Errorf("chinese name: got %q", fm["name"])
	}
	if !strings.Contains(body, "资深代码审查专家") {
		t.Error("chinese body not preserved")
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

func TestGetFMString(t *testing.T) {
	fm := map[string]interface{}{
		"name": "test-agent",
		"desc": "A description",
	}
	if s := GetFMString(fm, "name"); s != "test-agent" {
		t.Errorf("GetFMString: got %q", s)
	}
	if s := GetFMString(fm, "missing"); s != "" {
		t.Errorf("GetFMString missing key: got %q", s)
	}
}

func TestGetFMStringSlice(t *testing.T) {
	fm := map[string]interface{}{
		"tools": []interface{}{"Read", "Grep"},
		"tags":  "a, b, c",
	}
	tools := GetFMStringSlice(fm, "tools")
	if len(tools) != 2 || tools[0] != "Read" {
		t.Errorf("tools: got %v", tools)
	}
	tags := GetFMStringSlice(fm, "tags")
	if len(tags) != 3 || tags[0] != "a" {
		t.Errorf("tags: got %v", tags)
	}
	empty := GetFMStringSlice(fm, "missing")
	if empty != nil {
		t.Errorf("missing: got %v", empty)
	}
}

func TestFormatYAMLFrontmatter(t *testing.T) {
	fm := map[string]interface{}{
		"name":        "test-agent",
		"description": "A test agent",
	}
	result := FormatYAMLFrontmatter(fm)
	if !strings.HasPrefix(result, "---\n") {
		t.Error("should start with ---")
	}
	if !strings.Contains(result, "name: test-agent") {
		t.Error("should contain name field")
	}
}
