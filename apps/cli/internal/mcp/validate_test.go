package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func callValidate(t *testing.T, session *gomcp.ClientSession, args map[string]any) validateOutput {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "validate",
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

	var out validateOutput
	if err := json.Unmarshal([]byte(text.Text), &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	return out
}

func TestValidateTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if !out.Valid {
		t.Errorf("expected valid, got invalid with %d errors", out.Errors)
		for _, issue := range out.Issues {
			t.Logf("  %s: %s (task %s)", issue.Level, issue.Message, issue.TaskID)
		}
	}
	if out.TaskCount != 4 {
		t.Errorf("expected 4 tasks, got %d", out.TaskCount)
	}
}

func TestValidateTool_DetectsInvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "001-bad.md"), []byte(`---
id: "001"
title: "Bad task"
status: invalid-status
priority: high
effort: small
dependencies: []
tags: []
created: 2026-01-01
---

# Bad task
`), 0o644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session := setupTestServer(t)
	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if out.Valid {
		t.Error("expected invalid result for bad status")
	}
	if out.Errors < 1 {
		t.Error("expected at least 1 error")
	}
}

func TestValidateTool_DetectsMissingDependency(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "001-task.md"), []byte(`---
id: "001"
title: "Task with missing dep"
status: pending
priority: high
effort: small
dependencies: ["999"]
tags: []
created: 2026-01-01
---

# Task with missing dependency
`), 0o644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session := setupTestServer(t)
	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if out.Valid {
		t.Error("expected invalid result for missing dependency")
	}
}

func TestValidateTool_DetectsDuplicateIDs(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"001-first.md", "001-second.md"} {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(`---
id: "001"
title: "Duplicate"
status: pending
priority: high
effort: small
dependencies: []
tags: []
created: 2026-01-01
---

# Duplicate
`), 0o644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
	}

	session := setupTestServer(t)
	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if out.Valid {
		t.Error("expected invalid result for duplicate IDs")
	}
}

func TestValidateTool_StrictMode(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
		"strict":   true,
	})

	// Strict mode may produce additional warnings
	if out.TaskCount != 4 {
		t.Errorf("expected 4 tasks, got %d", out.TaskCount)
	}
}

func TestValidateTool_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	session := setupTestServer(t)

	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if !out.Valid {
		t.Error("expected valid for empty directory")
	}
	if out.TaskCount != 0 {
		t.Errorf("expected 0 tasks, got %d", out.TaskCount)
	}
}

func TestValidateTool_IssuesHaveSeverityLevels(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "001-bad.md"), []byte(`---
id: "001"
title: "Bad task"
status: invalid
priority: high
effort: small
dependencies: []
tags: []
created: 2026-01-01
---

# Bad task
`), 0o644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session := setupTestServer(t)
	out := callValidate(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(out.Issues) == 0 {
		t.Fatal("expected at least one issue")
	}

	for _, issue := range out.Issues {
		if issue.Level != "error" && issue.Level != "warning" {
			t.Errorf("expected level 'error' or 'warning', got %q", issue.Level)
		}
		if issue.Message == "" {
			t.Error("expected non-empty message")
		}
	}
}

func TestValidateTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "validate" {
			found = true
			if tool.Description == "" {
				t.Error("validate tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("validate tool not found in tools list")
	}
}
