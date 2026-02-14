package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func createGetTestFiles(t *testing.T) string {
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

func resetGetFlags() {
	getFormat = "text"
	getExact = false
	getThreshold = 0.6
	taskDir = "."
}

func captureGetOutput(t *testing.T, query string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGet(getCmd, []string{query})
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("runGet failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestGet_ExactMatchByID(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "001")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
	if !strings.Contains(output, "Title: Setup project") {
		t.Error("Expected output to contain task title")
	}
}

func TestGet_ExactMatchByTitle(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "Setup project")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
}

func TestGet_ExactMatchByTitle_CaseInsensitive(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "setup PROJECT")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected case-insensitive title match to find task 001")
	}
}

func TestGet_IDPrecedenceOverTitle(t *testing.T) {
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

	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "abc")

	// Should match by ID (task1), not by title (task2)
	if !strings.Contains(output, "Title: First task") {
		t.Error("Expected ID match to take precedence over title match")
	}
}

func TestGet_TaskNotFound_ExactMode(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getExact = true

	err := runGet(getCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("Expected error for non-matching query in exact mode")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestGet_TaskNotFound_NoMatches(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getThreshold = 0.99 // very high threshold so nothing matches

	err := runGet(getCmd, []string{"zzzzzzzzzzzzzzz"})
	if err == nil {
		t.Fatal("Expected error for garbage query")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestGet_TextFormat(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "002")

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

func TestGet_JSONFormat(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getFormat = "json"

	output := captureGetOutput(t, "002")

	var result getOutput
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

func TestGet_YAMLFormat(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getFormat = "yaml"

	output := captureGetOutput(t, "001")

	expected := []string{"id: \"001\"", "title: Setup project", "status: completed"}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected YAML output to contain %q", exp)
		}
	}
}

func TestGet_UnsupportedFormat(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getFormat = "csv"

	err := runGet(getCmd, []string{"001"})
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestGet_FuzzyMatch_Substring(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	// "auth" is a substring of "Implement authentication" — should fuzzy match
	// Simulate selecting option 1
	getStdinReader = strings.NewReader("1\n")
	defer func() { getStdinReader = os.Stdin }()

	output := captureGetOutput(t, "auth")

	if !strings.Contains(output, "Task: 002") {
		t.Error("Expected fuzzy substring match to find task 002")
	}
}

func TestGet_FuzzyMatch_Selection(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	// "ui" should fuzzy match "Build UI components"
	getStdinReader = strings.NewReader("1\n")
	defer func() { getStdinReader = os.Stdin }()

	output := captureGetOutput(t, "ui")

	if !strings.Contains(output, "Task: 003") {
		t.Error("Expected fuzzy match selection to return task 003")
	}
}

func TestGet_FuzzyMatch_Cancel(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	getStdinReader = strings.NewReader("0\n")
	defer func() { getStdinReader = os.Stdin }()

	err := runGet(getCmd, []string{"auth"})
	if err == nil {
		t.Fatal("Expected error when user cancels selection")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("Expected 'cancelled' error, got: %v", err)
	}
}

func TestGet_Threshold(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getThreshold = 0.95 // very high threshold

	err := runGet(getCmd, []string{"aut"})
	if err == nil {
		t.Fatal("Expected error when threshold filters out matches")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestGet_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	resetGetFlags()
	taskDir = tmpDir

	err := runGet(getCmd, []string{"anything"})
	if err == nil {
		t.Fatal("Expected error for empty directory")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestGet_Dependencies(t *testing.T) {
	tmpDir := createGetTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	// Task 003 depends on 002, and 002 blocks 003
	output := captureGetOutput(t, "003")

	if !strings.Contains(output, "Depends on: 002 (Implement authentication)") {
		t.Error("Expected depends-on info for task 003")
	}

	// Check that task 002 shows it blocks 003
	output = captureGetOutput(t, "002")
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
		{"auth", "Implement authentication", 0.7, 1.0}, // substring
		{"setup", "Setup project", 0.7, 1.0},           // substring
		{"setup project", "Setup project", 1.0, 1.0},   // exact
		{"zzzzz", "Setup project", 0.0, 0.3},           // no relation
		{"seutp", "Setup project", 0.3, 0.8},           // typo
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

// --- File path matching tests ---

func createGetTestFilesWithSubdirs(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	dirs := []string{"cli", "backend"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	files := map[string]string{
		"cli/042-task.md": `---
id: "cli-042"
title: "CLI task"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-02-08
---

# CLI task
`,
		"backend/055-api.md": `---
id: "backend-055"
title: "API task"
status: pending
priority: high
dependencies: []
tags: []
created: 2026-02-08
---

# API task
`,
		"cli/055-api.md": `---
id: "cli-055"
title: "CLI API task"
status: pending
priority: low
dependencies: []
tags: []
created: 2026-02-08
---

# CLI API task
`,
	}

	for relPath, content := range files {
		path := filepath.Join(tmpDir, relPath)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", relPath, err)
		}
	}

	return tmpDir
}

func TestGet_FilePathMatch_FullRelativePath(t *testing.T) {
	tmpDir := createGetTestFilesWithSubdirs(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "cli/042-task.md")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected full relative path to match task cli-042")
	}
}

func TestGet_FilePathMatch_FilenameWithExtension(t *testing.T) {
	tmpDir := createGetTestFilesWithSubdirs(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "042-task.md")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected filename with extension to match task cli-042")
	}
}

func TestGet_FilePathMatch_FilenameWithoutExtension(t *testing.T) {
	tmpDir := createGetTestFilesWithSubdirs(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "042-task")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected filename without extension to match task cli-042")
	}
}

func TestGet_FilePathMatch_AmbiguousFilename(t *testing.T) {
	tmpDir := createGetTestFilesWithSubdirs(t)
	resetGetFlags()
	taskDir = tmpDir

	// "055-api.md" exists in both cli/ and backend/ — should be ambiguous
	err := runGet(getCmd, []string{"055-api.md"})
	if err == nil {
		t.Fatal("Expected error for ambiguous filename")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("Expected 'ambiguous' error, got: %v", err)
	}
}

func TestGet_FilePathMatch_ExactPathResolvesAmbiguity(t *testing.T) {
	tmpDir := createGetTestFilesWithSubdirs(t)
	resetGetFlags()
	taskDir = tmpDir

	// Full relative path should resolve ambiguity
	output := captureGetOutput(t, "backend/055-api.md")

	if !strings.Contains(output, "Task: backend-055") {
		t.Error("Expected exact path to resolve ambiguity and match backend-055")
	}
}

func TestGet_FilePathMatch_IDStillTakesPriority(t *testing.T) {
	tmpDir := createGetTestFilesWithSubdirs(t)
	resetGetFlags()
	taskDir = tmpDir

	// "cli-042" is a task ID — should match by ID, not filepath
	output := captureGetOutput(t, "cli-042")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected ID match to still work")
	}
	if !strings.Contains(output, "Title: CLI task") {
		t.Error("Expected ID match to return CLI task")
	}
}

func TestFindFilePathMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", FilePath: "cli/001-setup.md"},
		{ID: "002", Title: "Task B", FilePath: "backend/002-api.md"},
		{ID: "003", Title: "Task C", FilePath: "cli/003-shared.md"},
		{ID: "004", Title: "Task D", FilePath: "backend/003-shared.md"},
	}

	// Exact full path match
	task, err := findFilePathMatch("cli/001-setup.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "001" {
		t.Error("Expected exact path match to find task 001")
	}

	// Filename with extension (unique)
	task, err = findFilePathMatch("002-api.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "002" {
		t.Error("Expected filename match to find task 002")
	}

	// Filename without extension (unique)
	task, err = findFilePathMatch("001-setup", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "001" {
		t.Error("Expected filename without extension to find task 001")
	}

	// Ambiguous filename
	task, err = findFilePathMatch("003-shared.md", tasks)
	if err == nil {
		t.Fatal("Expected error for ambiguous filename")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("Expected 'ambiguous' error, got: %v", err)
	}
	if task != nil {
		t.Error("Expected nil task for ambiguous match")
	}

	// No match
	task, err = findFilePathMatch("nonexistent.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task != nil {
		t.Error("Expected nil for no match")
	}

	// Exact path takes priority over ambiguous filename
	task, err = findFilePathMatch("cli/003-shared.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "003" {
		t.Error("Expected exact path to resolve ambiguity")
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

func createParentTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	tasks := map[string]string{
		"010-parent.md": `---
id: "010"
title: "Parent task"
status: pending
priority: high
dependencies: []
tags: []
created: 2026-02-08
---

# Parent task
`,
		"011-child.md": `---
id: "011"
title: "Child task"
status: pending
priority: medium
parent: "010"
dependencies: []
tags: []
created: 2026-02-08
---

# Child task
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

func TestGet_ParentDisplay(t *testing.T) {
	tmpDir := createParentTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "011")

	if !strings.Contains(output, "Parent:") {
		t.Error("Expected output to contain 'Parent:'")
	}
	if !strings.Contains(output, "010") {
		t.Error("Expected output to contain parent ID '010'")
	}
}

func TestGet_ChildrenDisplay(t *testing.T) {
	tmpDir := createParentTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir

	output := captureGetOutput(t, "010")

	if !strings.Contains(output, "Children:") {
		t.Error("Expected output to contain 'Children:'")
	}
	if !strings.Contains(output, "011") {
		t.Error("Expected output to contain child ID '011'")
	}
}

func TestGet_ParentJSON(t *testing.T) {
	tmpDir := createParentTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getFormat = "json"

	output := captureGetOutput(t, "011")

	var result getOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if result.Parent == nil {
		t.Fatal("Expected parent in JSON output")
	}
	if result.Parent.ID != "010" {
		t.Errorf("Expected parent ID '010', got %q", result.Parent.ID)
	}
}

func TestGet_ChildrenJSON(t *testing.T) {
	tmpDir := createParentTestFiles(t)
	resetGetFlags()
	taskDir = tmpDir
	getFormat = "json"

	output := captureGetOutput(t, "010")

	var result getOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(result.Children))
	}
	if result.Children[0].ID != "011" {
		t.Errorf("Expected child ID '011', got %q", result.Children[0].ID)
	}
}
