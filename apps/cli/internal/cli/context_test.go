package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/taskcontext"
)

func createContextTestFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	tasks := map[string]string{
		"001-touches.md": `---
id: "001"
title: "Task with touches"
status: pending
touches:
  - cli
created: 2026-02-14
---

# Task with touches
`,
		"002-context.md": `---
id: "002"
title: "Task with context"
status: pending
context:
  - "docs/readme.md"
created: 2026-02-14
---

# Task with context

Some body text here.
`,
		"003-both.md": `---
id: "003"
title: "Task with both"
status: pending
touches:
  - cli
context:
  - "docs/readme.md"
created: 2026-02-14
---

# Task with both
`,
		"004-deps.md": `---
id: "004"
title: "Task with deps"
status: pending
dependencies: ["001"]
context:
  - "docs/notes.md"
created: 2026-02-14
---

# Task with deps
`,
		"005-empty.md": `---
id: "005"
title: "Task with nothing"
status: pending
created: 2026-02-14
---

# Task with nothing
`,
		"006-missing.md": `---
id: "006"
title: "Task with missing files"
status: pending
context:
  - "does/not/exist.go"
created: 2026-02-14
---

# Missing files
`,
	}

	for filename, content := range tasks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Create referenced files
	createDirAndFile(t, tmpDir, "docs/readme.md", "# README\n")
	createDirAndFile(t, tmpDir, "docs/notes.md", "# Notes\n")
	createDirAndFile(t, tmpDir, "src/main.go", "package main\n")
	createDirAndFile(t, tmpDir, "src/util.go", "package main\n")

	// Create a .taskmd.yaml with scope definitions
	configContent := `scopes:
  cli:
    paths:
      - "src/main.go"
      - "src/util.go"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".taskmd.yaml"), []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	return tmpDir
}

func createDirAndFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func resetContextFlags() {
	ctxTaskID = ""
	ctxFormat = "text"
	ctxResolve = false
	ctxIncludeContent = false
	ctxIncludeDeps = false
	ctxMaxFiles = 0
	taskDir = "."
}

func captureContextOutput(t *testing.T) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runContext(contextCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestContext_TouchesOnly(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir

	// Override config path to use tmpDir's .taskmd.yaml
	setupTestConfig(t, tmpDir)

	ctxTaskID = "001"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Context for task") {
		t.Errorf("expected header, got: %s", output)
	}
	if !strings.Contains(output, "src/main.go") {
		t.Errorf("expected src/main.go in output, got: %s", output)
	}
	if !strings.Contains(output, "src/util.go") {
		t.Errorf("expected src/util.go in output, got: %s", output)
	}
}

func TestContext_ExplicitOnly(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "002"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "docs/readme.md") {
		t.Errorf("expected docs/readme.md in output, got: %s", output)
	}
	if !strings.Contains(output, "Explicit files") {
		t.Errorf("expected 'Explicit files' heading, got: %s", output)
	}
}

func TestContext_BothSources(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "003"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Scope files") {
		t.Errorf("expected scope files section, got: %s", output)
	}
	if !strings.Contains(output, "Explicit files") {
		t.Errorf("expected explicit files section, got: %s", output)
	}
}

func TestContext_EmptyTask(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "005"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "No context files found") {
		t.Errorf("expected 'No context files found', got: %s", output)
	}
}

func TestContext_TaskNotFound(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir

	ctxTaskID = "999"
	_, err := captureContextOutput(t)
	if err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("expected 'task not found', got: %v", err)
	}
}

func TestContext_MissingFiles(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "006"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "does/not/exist.go") {
		t.Errorf("expected missing file path in output, got: %s", output)
	}
	if !strings.Contains(output, "missing") {
		t.Errorf("expected (missing) annotation, got: %s", output)
	}
}

func TestContext_JSON(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "003"
	ctxFormat = "json"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result taskcontext.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if result.TaskID != "003" {
		t.Errorf("expected task_id 003, got %s", result.TaskID)
	}
	if len(result.Files) < 2 {
		t.Errorf("expected at least 2 files, got %d", len(result.Files))
	}

	// Check source tagging
	foundScope := false
	foundExplicit := false
	for _, f := range result.Files {
		if strings.HasPrefix(f.Source, "scope:") {
			foundScope = true
		}
		if f.Source == "explicit" {
			foundExplicit = true
		}
	}
	if !foundScope {
		t.Error("expected at least one scope-sourced file")
	}
	if !foundExplicit {
		t.Error("expected at least one explicit-sourced file")
	}
}

func TestContext_YAML(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "002"
	ctxFormat = "yaml"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "task_id:") {
		t.Errorf("expected YAML task_id field, got: %s", output)
	}
	if !strings.Contains(output, "docs/readme.md") {
		t.Errorf("expected docs/readme.md in YAML, got: %s", output)
	}
}

func TestContext_IncludeContent(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "002"
	ctxFormat = "json"
	ctxIncludeContent = true
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result taskcontext.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result.TaskBody == "" {
		t.Error("expected task_body with --include-content")
	}

	for _, f := range result.Files {
		if f.Exists && f.Content == "" {
			t.Errorf("expected content for existing file %s", f.Path)
		}
	}
}

func TestContext_IncludeContentText(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "002"
	ctxIncludeContent = true
	// default format is text
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task body should appear
	if !strings.Contains(output, "Some body text here.") {
		t.Errorf("expected task body in text output, got: %s", output)
	}

	// File content should appear
	if !strings.Contains(output, "# README") {
		t.Errorf("expected file content (# README) in text output, got: %s", output)
	}

	// Line count should appear
	if !strings.Contains(output, "lines") {
		t.Errorf("expected line count in text output, got: %s", output)
	}
}

func TestContext_IncludeDeps(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "004"
	ctxFormat = "json"
	ctxIncludeDeps = true
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result taskcontext.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Should have files from task 004 (docs/notes.md) and dep 001 (cli scope: src/main.go, src/util.go)
	if len(result.Files) < 2 {
		t.Errorf("expected files from both task and dependency, got %d files", len(result.Files))
	}

	// Should have dependency entries
	if len(result.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(result.Dependencies))
	}
	if len(result.Dependencies) > 0 && result.Dependencies[0].ID != "001" {
		t.Errorf("expected dependency ID 001, got %s", result.Dependencies[0].ID)
	}
}

func TestContext_MaxFiles(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "003"
	ctxFormat = "json"
	ctxMaxFiles = 1
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result taskcontext.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 file (capped), got %d", len(result.Files))
	}
}

func TestContext_InvalidFormat(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir

	ctxTaskID = "001"
	ctxFormat = "csv"
	_, err := captureContextOutput(t)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", err)
	}
}

func TestContext_Dependencies(t *testing.T) {
	tmpDir := createContextTestFiles(t)
	resetContextFlags()
	taskDir = tmpDir
	setupTestConfig(t, tmpDir)

	ctxTaskID = "004"
	output, err := captureContextOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Dependencies") {
		t.Errorf("expected Dependencies section, got: %s", output)
	}
	if !strings.Contains(output, "001") {
		t.Errorf("expected dependency ID 001, got: %s", output)
	}
}

// setupTestConfig sets viper to read the .taskmd.yaml from the test directory.
func setupTestConfig(t *testing.T, dir string) {
	t.Helper()
	configPath := filepath.Join(dir, ".taskmd.yaml")
	oldConfigFile := cfgFile
	t.Cleanup(func() {
		cfgFile = oldConfigFile
		initConfig()
	})
	cfgFile = configPath
	initConfig()
	// Ensure taskDir still points to tmpDir (config's "dir" is relative to the config file)
	taskDir = dir
}
