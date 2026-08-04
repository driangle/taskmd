package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/taskcontext"
)

// contextFiles is the context-specific fixture: task files exercising
// touches/context/deps permutations, the referenced files they point at, and a
// .taskmd.yaml defining the `cli` scope. Kept inline because the scope config
// and referenced-file layout are the subject of these tests.
func contextFiles() map[string]string {
	return map[string]string{
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
		"docs/readme.md": "# README\n",
		"docs/notes.md":  "# Notes\n",
		"src/main.go":    "package main\n",
		"src/util.go":    "package main\n",
		".taskmd.yaml": `scopes:
  cli:
    paths:
      - "src/main.go"
      - "src/util.go"
`,
	}
}

// contextWithConfig runs `context <args...>` against repo with the repo's
// .taskmd.yaml scope config loaded into viper. The harness deliberately skips
// config discovery, so the config is seeded through RunWith (mirroring the old
// setupTestConfig) so scope paths and the project root resolve correctly.
func contextWithConfig(t *testing.T, repo *taskRepo, args ...string) cliResult {
	t.Helper()
	return repo.RunWith(func() {
		cfgFile = repo.Path(".taskmd.yaml")
		initConfig()
	}, append([]string{"context"}, args...)...)
}

func TestContext_TouchesOnly(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "001")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

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
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "002")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "docs/readme.md") {
		t.Errorf("expected docs/readme.md in output, got: %s", output)
	}
	if !strings.Contains(output, "Explicit files") {
		t.Errorf("expected 'Explicit files' heading, got: %s", output)
	}
}

func TestContext_BothSources(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "003")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "Scope files") {
		t.Errorf("expected scope files section, got: %s", output)
	}
	if !strings.Contains(output, "Explicit files") {
		t.Errorf("expected explicit files section, got: %s", output)
	}
}

func TestContext_EmptyTask(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "005")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "No context files found") {
		t.Errorf("expected 'No context files found', got: %s", output)
	}
}

func TestContext_TaskNotFound(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := repo.Run("context", "--task-id", "999")
	if res.Err == nil {
		t.Fatal("expected error for non-existent task")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("expected 'task not found', got: %v", res.Err)
	}
}

func TestContext_MissingFiles(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "006")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "does/not/exist.go") {
		t.Errorf("expected missing file path in output, got: %s", output)
	}
	if !strings.Contains(output, "missing") {
		t.Errorf("expected (missing) annotation, got: %s", output)
	}
}

func TestContext_JSON(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "003", "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

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
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "002", "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "task_id:") {
		t.Errorf("expected YAML task_id field, got: %s", output)
	}
	if !strings.Contains(output, "docs/readme.md") {
		t.Errorf("expected docs/readme.md in YAML, got: %s", output)
	}
}

func TestContext_IncludeContent(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "002", "--format", "json", "--include-content")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

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
	repo := newTaskRepo(t, contextFiles())

	// default format is text
	res := contextWithConfig(t, repo, "--task-id", "002", "--include-content")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

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
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "004", "--format", "json", "--include-deps")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

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
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "003", "--format", "json", "--max-files", "1")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	var result taskcontext.Result
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result.Files) != 1 {
		t.Errorf("expected 1 file (capped), got %d", len(result.Files))
	}
}

func TestContext_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := repo.Run("context", "--task-id", "001", "--format", "csv")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestContext_Dependencies(t *testing.T) {
	repo := newTaskRepo(t, contextFiles())

	res := contextWithConfig(t, repo, "--task-id", "004")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "Dependencies") {
		t.Errorf("expected Dependencies section, got: %s", output)
	}
	if !strings.Contains(output, "001") {
		t.Errorf("expected dependency ID 001, got: %s", output)
	}
}
