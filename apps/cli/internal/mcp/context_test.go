package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/taskcontext"
)

// createTestTaskFilesWithContext creates task files that have touches and context fields.
func createTestTaskFilesWithContext(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Create some source files that scopes can point to
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "helper.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to write helper.go: %v", err)
	}

	// Create tasks directory
	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	tasks := map[string]string{
		"001-with-touches.md": `---
id: "001"
title: "Task with touches"
status: pending
priority: high
touches: ["backend"]
created: 2026-01-01
---

# Task with touches
`,
		"002-with-context.md": `---
id: "002"
title: "Task with explicit context"
status: pending
priority: medium
context: ["src/main.go"]
created: 2026-01-02
---

# Task with explicit context
`,
		"003-no-context.md": `---
id: "003"
title: "Task without context"
status: pending
priority: low
created: 2026-01-03
---

# Task without context
`,
	}

	for name, content := range tasks {
		err := os.WriteFile(filepath.Join(tasksDir, name), []byte(content), 0o644)
		if err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	return tmpDir
}

func callContext(t *testing.T, session *gomcp.ClientSession, args map[string]any) taskcontext.Result {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "context",
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %+v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	text, ok := result.Content[0].(*gomcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var out taskcontext.Result
	if err := json.Unmarshal([]byte(text.Text), &out); err != nil {
		t.Fatalf("failed to unmarshal context result: %v", err)
	}
	return out
}

func TestContextTool_WithScopes(t *testing.T) {
	tmpDir := createTestTaskFilesWithContext(t)
	session := setupTestServer(t)

	out := callContext(t, session, map[string]any{
		"task_dir":     filepath.Join(tmpDir, "tasks"),
		"task_id":      "001",
		"project_root": tmpDir,
		"scopes": map[string]any{
			"backend": []any{"src/main.go", "src/helper.go"},
		},
	})

	if out.TaskID != "001" {
		t.Errorf("expected task_id 001, got %s", out.TaskID)
	}
	if len(out.Files) != 2 {
		t.Fatalf("expected 2 files from scope, got %d", len(out.Files))
	}

	// All files should exist
	for _, f := range out.Files {
		if !f.Exists {
			t.Errorf("expected file %s to exist", f.Path)
		}
	}
}

func TestContextTool_WithExplicitContext(t *testing.T) {
	tmpDir := createTestTaskFilesWithContext(t)
	session := setupTestServer(t)

	out := callContext(t, session, map[string]any{
		"task_dir":     filepath.Join(tmpDir, "tasks"),
		"task_id":      "002",
		"project_root": tmpDir,
	})

	if out.TaskID != "002" {
		t.Errorf("expected task_id 002, got %s", out.TaskID)
	}
	if len(out.Files) != 1 {
		t.Fatalf("expected 1 explicit file, got %d", len(out.Files))
	}
	if out.Files[0].Path != "src/main.go" {
		t.Errorf("expected path src/main.go, got %s", out.Files[0].Path)
	}
	if out.Files[0].Source != "explicit" {
		t.Errorf("expected source explicit, got %s", out.Files[0].Source)
	}
}

func TestContextTool_NoContextFiles(t *testing.T) {
	tmpDir := createTestTaskFilesWithContext(t)
	session := setupTestServer(t)

	out := callContext(t, session, map[string]any{
		"task_dir":     filepath.Join(tmpDir, "tasks"),
		"task_id":      "003",
		"project_root": tmpDir,
	})

	if out.TaskID != "003" {
		t.Errorf("expected task_id 003, got %s", out.TaskID)
	}
	if len(out.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(out.Files))
	}
}

func TestContextTool_NotFound(t *testing.T) {
	tmpDir := createTestTaskFilesWithContext(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "context",
		Arguments: map[string]any{
			"task_dir": filepath.Join(tmpDir, "tasks"),
			"task_id":  "999",
		},
	})
	if err != nil {
		return
	}
	if !result.IsError {
		t.Fatal("expected error for non-existent task")
	}
}

func TestContextTool_MissingTaskID(t *testing.T) {
	tmpDir := createTestTaskFilesWithContext(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "context",
		Arguments: map[string]any{
			"task_dir": filepath.Join(tmpDir, "tasks"),
		},
	})
	if err != nil {
		return
	}
	if !result.IsError {
		t.Fatal("expected error for missing task_id")
	}
}

func TestContextTool_MaxFiles(t *testing.T) {
	tmpDir := createTestTaskFilesWithContext(t)
	session := setupTestServer(t)

	out := callContext(t, session, map[string]any{
		"task_dir":     filepath.Join(tmpDir, "tasks"),
		"task_id":      "001",
		"project_root": tmpDir,
		"max_files":    1,
		"scopes": map[string]any{
			"backend": []any{"src/main.go", "src/helper.go"},
		},
	})

	if len(out.Files) != 1 {
		t.Fatalf("expected 1 file with max_files=1, got %d", len(out.Files))
	}
}

func TestContextTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "context" {
			found = true
			if tool.Description == "" {
				t.Error("context tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("context tool not found in tools list")
	}
}
