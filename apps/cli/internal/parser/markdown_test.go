package parser

import (
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestParseTaskContent_ValidTask(t *testing.T) {
	content := []byte(`---
id: "001"
title: "Test Task"
status: pending
priority: high
effort: medium
dependencies: ["002", "003"]
tags:
  - test
  - cli
group: backend
created: 2026-02-08
---

# Test Task

This is the task body.

- Item 1
- Item 2
`)

	task, err := ParseTaskContent("test.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "001" {
		t.Errorf("expected ID '001', got '%s'", task.ID)
	}

	if task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got '%s'", task.Title)
	}

	if task.Status != model.StatusPending {
		t.Errorf("expected status 'pending', got '%s'", task.Status)
	}

	if task.Priority != model.PriorityHigh {
		t.Errorf("expected priority 'high', got '%s'", task.Priority)
	}

	if task.Effort != model.EffortMedium {
		t.Errorf("expected effort 'medium', got '%s'", task.Effort)
	}

	if len(task.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(task.Dependencies))
	}

	if len(task.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(task.Tags))
	}

	if task.Group != "backend" {
		t.Errorf("expected group 'backend', got '%s'", task.Group)
	}

	expectedDate := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	if !task.Created.Equal(expectedDate) {
		t.Errorf("expected created date %v, got %v", expectedDate, task.Created)
	}

	if task.Body == "" {
		t.Error("expected body to be parsed")
	}

	if task.FilePath != "test.md" {
		t.Errorf("expected FilePath 'test.md', got '%s'", task.FilePath)
	}
}

func TestParseTaskContent_MinimalTask(t *testing.T) {
	content := []byte(`---
id: "002"
title: "Minimal Task"
---

Body content here.
`)

	task, err := ParseTaskContent("minimal.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "002" {
		t.Errorf("expected ID '002', got '%s'", task.ID)
	}

	if task.Title != "Minimal Task" {
		t.Errorf("expected title 'Minimal Task', got '%s'", task.Title)
	}

	if task.Body != "Body content here." {
		t.Errorf("expected body 'Body content here.', got '%s'", task.Body)
	}
}

func TestParseTaskContent_EmptyFile(t *testing.T) {
	content := []byte("")

	_, err := ParseTaskContent("empty.md", content)
	if err == nil {
		t.Error("expected error for empty file")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if parseErr.FilePath != "empty.md" {
		t.Errorf("expected FilePath 'empty.md', got '%s'", parseErr.FilePath)
	}
}

func TestParseTaskContent_MissingFrontmatter(t *testing.T) {
	content := []byte(`# No Frontmatter

This file has no frontmatter.
`)

	_, err := ParseTaskContent("no-frontmatter.md", content)
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseTaskContent_MalformedYAML(t *testing.T) {
	content := []byte(`---
id: "003"
title: "Bad YAML"
invalid: [unclosed
---

Body
`)

	_, err := ParseTaskContent("bad-yaml.md", content)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "YAML") {
		t.Errorf("expected error message to mention YAML, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_UnclosedFrontmatter(t *testing.T) {
	content := []byte(`---
id: "004"
title: "Unclosed"

This has no closing delimiter
`)

	_, err := ParseTaskContent("unclosed.md", content)
	if err == nil {
		t.Error("expected error for unclosed frontmatter")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "frontmatter") {
		t.Errorf("expected error message to mention frontmatter, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_MissingID(t *testing.T) {
	content := []byte(`---
title: "No ID"
---

Body
`)

	_, err := ParseTaskContent("no-id.md", content)
	if err == nil {
		t.Error("expected error for missing ID")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "required fields") {
		t.Errorf("expected error message to mention required fields, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_MissingTitle(t *testing.T) {
	content := []byte(`---
id: "005"
---

Body
`)

	_, err := ParseTaskContent("no-title.md", content)
	if err == nil {
		t.Error("expected error for missing title")
	}

	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %T", err)
	} else if !contains(parseErr.Message, "required fields") {
		t.Errorf("expected error message to mention required fields, got: %s", parseErr.Message)
	}
}

func TestParseTaskContent_EmptyBody(t *testing.T) {
	content := []byte(`---
id: "006"
title: "Empty Body"
---
`)

	task, err := ParseTaskContent("empty-body.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.Body != "" {
		t.Errorf("expected empty body, got '%s'", task.Body)
	}
}

func TestExtractFrontmatter_NoFrontmatter(t *testing.T) {
	content := []byte("Just plain content")

	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(frontmatter) != 0 {
		t.Error("expected empty frontmatter")
	}

	if body != "Just plain content" {
		t.Errorf("expected body to be full content, got '%s'", body)
	}
}

func TestExtractFrontmatter_WithWhitespace(t *testing.T) {
	content := []byte(`---
id: "007"
title: "Whitespace Test"
---

Body with leading/trailing whitespace

`)

	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(frontmatter) == 0 {
		t.Error("expected frontmatter")
	}

	if body != "Body with leading/trailing whitespace" {
		t.Errorf("expected trimmed body, got '%s'", body)
	}
}

func TestParseTaskContent_DeriveIDFromFilename(t *testing.T) {
	content := []byte(`---
title: "My Task"
status: pending
---

Body
`)

	task, err := ParseTaskContent("009-add-feature.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "009" {
		t.Errorf("expected ID '009' derived from filename, got '%s'", task.ID)
	}

	if task.Title != "My Task" {
		t.Errorf("expected title from frontmatter 'My Task', got '%s'", task.Title)
	}
}

func TestParseTaskContent_DeriveTitleFromFilename(t *testing.T) {
	content := []byte(`---
status: pending
---

Body
`)

	task, err := ParseTaskContent("012-setup-database.md", content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if task.ID != "012" {
		t.Errorf("expected ID '012', got '%s'", task.ID)
	}

	if task.Title != "setup database" {
		t.Errorf("expected title 'setup database' derived from filename, got '%s'", task.Title)
	}
}

func TestParseTaskContent_NonNumericFilenameNoDerivation(t *testing.T) {
	content := []byte(`---
title: "Some Task"
---

Body
`)

	_, err := ParseTaskContent("readme.md", content)
	if err == nil {
		t.Error("expected error for non-numeric filename with missing ID")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
