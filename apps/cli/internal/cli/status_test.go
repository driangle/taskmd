package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createStatusTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	tasks := map[string]string{
		"001-setup.md": `---
id: "001"
title: "Setup project"
status: completed
priority: high
effort: small
dependencies: []
tags: ["infra", "setup"]
created: 2026-02-08
---

# Setup project

Initial project setup with build tooling.
`,
		"002-auth.md": `---
id: "002"
title: "Implement authentication"
status: in-progress
priority: critical
effort: large
dependencies: ["001"]
tags: ["backend", "security"]
owner: "alice"
created: 2026-02-08
---

# Implement authentication

Add JWT-based auth with refresh tokens.
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: pending
priority: medium
effort: medium
dependencies: ["002"]
tags: ["frontend"]
parent: "001"
created: 2026-02-08
---

# Build UI components

Create reusable component library.
`,
	}

	for filename, content := range tasks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return tmpDir
}

func resetStatusFlags() {
	statusFormat = "text"
	statusExact = false
	statusThreshold = 0.6
	taskDir = "."
}

func captureStatusOutput(t *testing.T, query string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runStatus(statusCmd, []string{query})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("runStatus failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestStatus_ExactMatchByID(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir

	output := captureStatusOutput(t, "001")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
	if !strings.Contains(output, "Title: Setup project") {
		t.Error("Expected output to contain task title")
	}
}

func TestStatus_TextFormat(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir

	output := captureStatusOutput(t, "002")

	expected := []string{
		"Task: 002",
		"Title: Implement authentication",
		"Status: in-progress",
		"Priority: critical",
		"Effort: large",
		"Tags: backend, security",
		"Owner: alice",
		"Created: 2026-02-08",
		"Dependencies: 001",
		"File:",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q", exp)
		}
	}

	// Verify no body content is present
	if strings.Contains(output, "Description:") {
		t.Error("Status output should not contain Description section")
	}
	if strings.Contains(output, "Add JWT-based auth") {
		t.Error("Status output should not contain body content")
	}
}

func TestStatus_TextFormat_ParentField(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir

	output := captureStatusOutput(t, "003")

	if !strings.Contains(output, "Parent: 001") {
		t.Error("Expected output to contain 'Parent: 001'")
	}
}

func TestStatus_JSONFormat(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir
	statusFormat = "json"

	output := captureStatusOutput(t, "002")

	var result statusOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if result.ID != "002" {
		t.Errorf("Expected ID '002', got %q", result.ID)
	}
	if result.Title != "Implement authentication" {
		t.Errorf("Expected title 'Implement authentication', got %q", result.Title)
	}
	if result.Status != "in-progress" {
		t.Errorf("Expected status 'in-progress', got %q", result.Status)
	}
	if result.Priority != "critical" {
		t.Errorf("Expected priority 'critical', got %q", result.Priority)
	}
	if result.Effort != "large" {
		t.Errorf("Expected effort 'large', got %q", result.Effort)
	}
	if result.Owner != "alice" {
		t.Errorf("Expected owner 'alice', got %q", result.Owner)
	}
	if len(result.Dependencies) != 1 || result.Dependencies[0] != "001" {
		t.Errorf("Expected dependencies [001], got %v", result.Dependencies)
	}

	// Verify no content/body field in JSON
	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse raw JSON: %v", err)
	}
	if _, ok := raw["content"]; ok {
		t.Error("JSON output should not contain 'content' key")
	}
	if _, ok := raw["body"]; ok {
		t.Error("JSON output should not contain 'body' key")
	}
}

func TestStatus_YAMLFormat(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir
	statusFormat = "yaml"

	output := captureStatusOutput(t, "001")

	expected := []string{"id: \"001\"", "title: Setup project", "status: completed"}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected YAML output to contain %q", exp)
		}
	}

	// Verify no content field
	if strings.Contains(output, "content:") {
		t.Error("YAML output should not contain 'content' field")
	}
}

func TestStatus_UnsupportedFormat(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir
	statusFormat = "csv"

	err := runStatus(statusCmd, []string{"001"})
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestStatus_TaskNotFound_ExactMode(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir
	statusExact = true

	err := runStatus(statusCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("Expected error for non-matching query in exact mode")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestStatus_FuzzyMatch(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir

	// "auth" is a substring of "Implement authentication"
	statusStdinReader = strings.NewReader("1\n")
	defer func() { statusStdinReader = os.Stdin }()

	output := captureStatusOutput(t, "auth")

	if !strings.Contains(output, "Task: 002") {
		t.Error("Expected fuzzy match to find task 002")
	}
}

func TestStatus_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	resetStatusFlags()
	taskDir = tmpDir

	err := runStatus(statusCmd, []string{"anything"})
	if err == nil {
		t.Fatal("Expected error for empty directory")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestStatus_NoBodyInOutput(t *testing.T) {
	tmpDir := createStatusTestFiles(t)
	resetStatusFlags()
	taskDir = tmpDir

	// Text format
	output := captureStatusOutput(t, "001")
	if strings.Contains(output, "Initial project setup") {
		t.Error("Text output should not contain task body")
	}

	// JSON format
	resetStatusFlags()
	taskDir = tmpDir
	statusFormat = "json"
	output = captureStatusOutput(t, "001")

	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if _, ok := raw["content"]; ok {
		t.Error("JSON should not have 'content' field")
	}
}
