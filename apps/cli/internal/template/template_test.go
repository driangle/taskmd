package template

import (
	"strings"
	"testing"
)

const sampleTemplate = `---
_template:
  name: bug
  description: "Bug report with reproduction steps"
title: "{{title}}"
id: "{{id}}"
status: pending
priority: high
type: bug
tags: []
created: "{{date}}"
---

# {{title}}

## Steps to Reproduce
1. ...

## Expected Behavior

## Actual Behavior
`

func TestParseTemplate_Valid(t *testing.T) {
	tmpl, err := ParseTemplate(sampleTemplate, "built-in")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl.Name != "bug" {
		t.Errorf("expected name 'bug', got %q", tmpl.Name)
	}
	if tmpl.Description != "Bug report with reproduction steps" {
		t.Errorf("expected description 'Bug report with reproduction steps', got %q", tmpl.Description)
	}
	if tmpl.Source != "built-in" {
		t.Errorf("expected source 'built-in', got %q", tmpl.Source)
	}
}

func TestParseTemplate_MissingName(t *testing.T) {
	content := `---
_template:
  description: "No name"
title: test
---
`
	_, err := ParseTemplate(content, "test")
	if err == nil {
		t.Fatal("expected error for missing template name")
	}
	if !strings.Contains(err.Error(), "_template.name") {
		t.Errorf("error should mention _template.name, got: %v", err)
	}
}

func TestParseTemplate_NoTemplateBlock(t *testing.T) {
	content := `---
title: test
status: pending
---
`
	_, err := ParseTemplate(content, "test")
	if err == nil {
		t.Fatal("expected error for missing _template block")
	}
}

func TestParseTemplate_NoFrontmatter(t *testing.T) {
	content := "# Just a heading\nSome text\n"
	_, err := ParseTemplate(content, "test")
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestRenderTask_SubstitutesVariables(t *testing.T) {
	tmpl, err := ParseTemplate(sampleTemplate, "built-in")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	result := RenderTask(tmpl, map[string]string{
		"id":    "042",
		"title": "Login fails on Safari",
		"date":  "2026-02-20",
	})

	if strings.Contains(result, "{{id}}") {
		t.Error("{{id}} should have been replaced")
	}
	if strings.Contains(result, "{{title}}") {
		t.Error("{{title}} should have been replaced")
	}
	if strings.Contains(result, "{{date}}") {
		t.Error("{{date}} should have been replaced")
	}
	if !strings.Contains(result, `id: "042"`) {
		t.Error("expected id: \"042\" in output")
	}
	if !strings.Contains(result, `title: "Login fails on Safari"`) {
		t.Error("expected title in output")
	}
	if !strings.Contains(result, "# Login fails on Safari") {
		t.Error("expected heading with title in output")
	}
}

func TestRenderTask_StripsTemplateBlock(t *testing.T) {
	tmpl, err := ParseTemplate(sampleTemplate, "built-in")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	result := RenderTask(tmpl, map[string]string{
		"id":    "001",
		"title": "Test",
		"date":  "2026-01-01",
	})

	if strings.Contains(result, "_template:") {
		t.Error("_template: block should have been stripped")
	}
	if strings.Contains(result, "name: bug") {
		t.Error("_template.name should have been stripped")
	}
	if strings.Contains(result, "description: ") {
		t.Error("_template.description should have been stripped")
	}

	// Other frontmatter should remain
	if !strings.Contains(result, "status: pending") {
		t.Error("expected status: pending to remain")
	}
	if !strings.Contains(result, "priority: high") {
		t.Error("expected priority: high to remain")
	}
	if !strings.Contains(result, "type: bug") {
		t.Error("expected type: bug to remain")
	}
}

func TestApplyOverrides(t *testing.T) {
	content := `---
title: "Original"
status: pending
priority: high
tags: []
---

# Original
`
	result := ApplyOverrides(content, map[string]string{
		"priority": "critical",
		"status":   "in-progress",
	})

	if !strings.Contains(result, "priority: critical") {
		t.Error("expected priority override to critical")
	}
	if !strings.Contains(result, "status: in-progress") {
		t.Error("expected status override to in-progress")
	}
	// Title should be unchanged
	if !strings.Contains(result, `title: "Original"`) {
		t.Error("title should not have changed")
	}
}

func TestApplyOverrides_AppendsMissingKeys(t *testing.T) {
	// Mirrors the real bug: the `feature` template has no `phase:` (or `owner:`)
	// line, so `taskmd add --template feature --phase mvp` must still inject the
	// field rather than silently dropping it.
	content := `---
title: "Original"
status: pending
priority: high
tags: []
---

# Original
`
	result := ApplyOverrides(content, map[string]string{
		"phase":    "mvp",
		"priority": "critical",
		"owner":    "alice",
	})

	// Existing key is still replaced in place.
	if !strings.Contains(result, "priority: critical") {
		t.Error("expected priority override to critical")
	}
	// Missing keys are appended inside the frontmatter block.
	if !strings.Contains(result, "phase: mvp") {
		t.Error("expected missing phase key to be appended")
	}
	if !strings.Contains(result, "owner: alice") {
		t.Error("expected missing owner key to be appended")
	}

	// Appended keys must land inside the frontmatter, before the closing "---".
	afterOpen := strings.Index(result, "\n") // past the opening "---"
	closeIdx := strings.Index(result[afterOpen:], "\n---")
	if closeIdx < 0 {
		t.Fatalf("could not find closing frontmatter delimiter in:\n%s", result)
	}
	closeIdx += afterOpen
	phaseIdx := strings.Index(result, "phase: mvp")
	if phaseIdx < 0 || phaseIdx > closeIdx {
		t.Errorf("phase key must be inside frontmatter (before closing ---), got:\n%s", result)
	}
}

func TestApplyOverrides_EmptyOverrides(t *testing.T) {
	content := `---
status: pending
---
`
	result := ApplyOverrides(content, nil)
	if result != content {
		t.Error("empty overrides should return content unchanged")
	}
}

func TestApplyOverrides_NoFrontmatter(t *testing.T) {
	content := "# Just a heading\n"
	result := ApplyOverrides(content, map[string]string{"status": "done"})
	if result != content {
		t.Error("no frontmatter should return content unchanged")
	}
}

func TestSubstituteVars(t *testing.T) {
	input := "Hello {{name}}, your task {{id}} is ready."
	result := substituteVars(input, map[string]string{
		"name": "Alice",
		"id":   "007",
	})
	expected := "Hello Alice, your task 007 is ready."
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestSubstituteVars_NoVars(t *testing.T) {
	input := "No variables here."
	result := substituteVars(input, map[string]string{"id": "001"})
	if result != input {
		t.Error("content without matching vars should be unchanged")
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{"hello", "hello"},
		{`""`, ""},
		{`"a"`, "a"},
	}
	for _, tt := range tests {
		got := unquote(tt.input)
		if got != tt.expected {
			t.Errorf("unquote(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestStripTemplateBlock(t *testing.T) {
	lines := strings.Split(sampleTemplate, "\n")
	openIdx, closeIdx := findFrontmatterBounds(lines)
	result := stripTemplateBlock(lines, openIdx, closeIdx)
	joined := strings.Join(result, "\n")

	if strings.Contains(joined, "_template:") {
		t.Error("_template: line should have been stripped")
	}
	if strings.Contains(joined, "  name: bug") {
		t.Error("name: bug line should have been stripped")
	}

	// Other fields should remain
	if !strings.Contains(joined, `title: "{{title}}"`) {
		t.Error("title should remain")
	}
	if !strings.Contains(joined, "status: pending") {
		t.Error("status should remain")
	}
}

func TestExtractTemplateMetadata(t *testing.T) {
	lines := strings.Split(sampleTemplate, "\n")
	openIdx, closeIdx := findFrontmatterBounds(lines)
	name, desc := extractTemplateMetadata(lines, openIdx, closeIdx)

	if name != "bug" {
		t.Errorf("expected name 'bug', got %q", name)
	}
	if desc != "Bug report with reproduction steps" {
		t.Errorf("expected description, got %q", desc)
	}
}

func TestExtractTemplateMetadata_UnquotedValues(t *testing.T) {
	content := `---
_template:
  name: feature
  description: A feature template
title: test
---
`
	lines := strings.Split(content, "\n")
	openIdx, closeIdx := findFrontmatterBounds(lines)
	name, desc := extractTemplateMetadata(lines, openIdx, closeIdx)

	if name != "feature" {
		t.Errorf("expected name 'feature', got %q", name)
	}
	if desc != "A feature template" {
		t.Errorf("expected description 'A feature template', got %q", desc)
	}
}
