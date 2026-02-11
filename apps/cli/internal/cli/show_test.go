package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
)

func createShowTestFiles(t *testing.T) string {
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

func resetShowFlags() {
	showFormat = "text"
	showExact = false
	showThreshold = 0.6
	dir = "."
}

func captureShowOutput(t *testing.T, query string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runShow(showCmd, []string{query})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("runShow failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestShow_ExactMatchByID(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	output := captureShowOutput(t, "001")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
	if !strings.Contains(output, "Title: Setup project") {
		t.Error("Expected output to contain task title")
	}
}

func TestShow_ExactMatchByTitle(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	output := captureShowOutput(t, "Setup project")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
}

func TestShow_ExactMatchByTitle_CaseInsensitive(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	output := captureShowOutput(t, "setup PROJECT")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected case-insensitive title match to find task 001")
	}
}

func TestShow_IDPrecedenceOverTitle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a task whose title matches another task's ID
	task1 := `---
id: "abc"
title: "First task"
status: pending
priority: low
dependencies: []
tags: []
created: 2026-02-08
---

# First task
`
	task2 := `---
id: "xyz"
title: "abc"
status: pending
priority: high
dependencies: []
tags: []
created: 2026-02-08
---

# abc task
`
	os.WriteFile(filepath.Join(tmpDir, "task1.md"), []byte(task1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "task2.md"), []byte(task2), 0644)

	resetShowFlags()
	dir = tmpDir

	output := captureShowOutput(t, "abc")

	// Should match by ID (task1), not by title (task2)
	if !strings.Contains(output, "Title: First task") {
		t.Error("Expected ID match to take precedence over title match")
	}
}

func TestShow_TaskNotFound_ExactMode(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir
	showExact = true

	err := runShow(showCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("Expected error for non-matching query in exact mode")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestShow_TaskNotFound_NoMatches(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir
	showThreshold = 0.99 // very high threshold so nothing matches

	err := runShow(showCmd, []string{"zzzzzzzzzzzzzzz"})
	if err == nil {
		t.Fatal("Expected error for garbage query")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestShow_TextFormat(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	output := captureShowOutput(t, "002")

	expected := []string{
		"Task: 002",
		"Title: Implement authentication",
		"Status: in-progress",
		"Priority: critical",
		"Effort: large",
		"Tags: backend, security",
		"Created: 2026-02-08",
		"File:",
		"Description:",
		"Add JWT-based auth with refresh tokens.",
		"Dependencies:",
		"Depends on: 001 (Setup project)",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q", exp)
		}
	}
}

func TestShow_JSONFormat(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir
	showFormat = "json"

	output := captureShowOutput(t, "002")

	var result showOutput
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
	if result.Content == "" {
		t.Error("Expected non-empty content in JSON output")
	}
	if len(result.Dependencies.DependsOn) != 1 || result.Dependencies.DependsOn[0].ID != "001" {
		t.Error("Expected depends_on to contain task 001")
	}
}

func TestShow_YAMLFormat(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir
	showFormat = "yaml"

	output := captureShowOutput(t, "001")

	expected := []string{"id: \"001\"", "title: Setup project", "status: completed"}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected YAML output to contain %q", exp)
		}
	}
}

func TestShow_UnsupportedFormat(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir
	showFormat = "csv"

	err := runShow(showCmd, []string{"001"})
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestShow_FuzzyMatch_Substring(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	// "auth" is a substring of "Implement authentication" — should fuzzy match
	// Simulate selecting option 1
	showStdinReader = strings.NewReader("1\n")
	defer func() { showStdinReader = os.Stdin }()

	output := captureShowOutput(t, "auth")

	if !strings.Contains(output, "Task: 002") {
		t.Error("Expected fuzzy substring match to find task 002")
	}
}

func TestShow_FuzzyMatch_Selection(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	// "ui" should fuzzy match "Build UI components"
	showStdinReader = strings.NewReader("1\n")
	defer func() { showStdinReader = os.Stdin }()

	output := captureShowOutput(t, "ui")

	if !strings.Contains(output, "Task: 003") {
		t.Error("Expected fuzzy match selection to return task 003")
	}
}

func TestShow_FuzzyMatch_Cancel(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	showStdinReader = strings.NewReader("0\n")
	defer func() { showStdinReader = os.Stdin }()

	err := runShow(showCmd, []string{"auth"})
	if err == nil {
		t.Fatal("Expected error when user cancels selection")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("Expected 'cancelled' error, got: %v", err)
	}
}

func TestShow_Threshold(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir
	showThreshold = 0.95 // very high threshold

	err := runShow(showCmd, []string{"aut"})
	if err == nil {
		t.Fatal("Expected error when threshold filters out matches")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestShow_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	resetShowFlags()
	dir = tmpDir

	err := runShow(showCmd, []string{"anything"})
	if err == nil {
		t.Fatal("Expected error for empty directory")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestShow_Dependencies(t *testing.T) {
	tmpDir := createShowTestFiles(t)
	resetShowFlags()
	dir = tmpDir

	// Task 003 depends on 002, and 002 blocks 003
	output := captureShowOutput(t, "003")

	if !strings.Contains(output, "Depends on: 002 (Implement authentication)") {
		t.Error("Expected depends-on info for task 003")
	}

	// Check that task 002 shows it blocks 003
	output = captureShowOutput(t, "002")
	if !strings.Contains(output, "Blocks: 003 (Build UI components)") {
		t.Error("Expected blocks info for task 002")
	}
}

// --- Unit tests for helper functions ---

func TestFindExactMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Setup project"},
		{ID: "002", Title: "Auth service"},
	}

	// Match by ID
	if task := findExactMatch("001", tasks); task == nil || task.ID != "001" {
		t.Error("Expected to find task 001 by ID")
	}

	// Match by title (case-insensitive)
	if task := findExactMatch("auth service", tasks); task == nil || task.ID != "002" {
		t.Error("Expected to find task 002 by title")
	}

	// No match
	if task := findExactMatch("nonexistent", tasks); task != nil {
		t.Error("Expected nil for non-matching query")
	}
}

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		query    string
		target   string
		minScore float64
		maxScore float64
	}{
		{"auth", "Implement authentication", 0.7, 1.0},  // substring
		{"setup", "Setup project", 0.7, 1.0},             // substring
		{"setup project", "Setup project", 1.0, 1.0},     // exact
		{"zzzzz", "Setup project", 0.0, 0.3},             // no relation
		{"seutp", "Setup project", 0.3, 0.8},             // typo
	}

	for _, tt := range tests {
		score := calculateSimilarity(tt.query, tt.target)
		if score < tt.minScore || score > tt.maxScore {
			t.Errorf("calculateSimilarity(%q, %q) = %.2f, expected [%.2f, %.2f]",
				tt.query, tt.target, score, tt.minScore, tt.maxScore)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		result := levenshtein(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("levenshtein(%q, %q) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestFuzzyMatchTasks(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Setup project"},
		{ID: "002", Title: "Implement authentication"},
		{ID: "003", Title: "Build UI components"},
	}

	// "auth" should match task 002 via substring
	matches := fuzzyMatchTasks("auth", tasks, 0.6)
	if len(matches) == 0 {
		t.Fatal("Expected at least one fuzzy match for 'auth'")
	}
	if matches[0].Task.ID != "002" {
		t.Errorf("Expected top match to be task 002, got %s", matches[0].Task.ID)
	}

	// Very high threshold should filter everything
	matches = fuzzyMatchTasks("auth", tasks, 0.99)
	if len(matches) != 0 {
		t.Errorf("Expected no matches with threshold 0.99, got %d", len(matches))
	}
}
