package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func createTestTaskFiles(t *testing.T) string {
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
tags: ["infra"]
created: 2026-01-01
---

# Setup project
`,
		"002-auth.md": `---
id: "002"
title: "Add authentication"
status: pending
priority: high
effort: medium
dependencies: ["001"]
tags: ["feature", "security"]
created: 2026-01-02
---

# Add authentication
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: pending
priority: medium
effort: large
dependencies: []
tags: ["feature"]
created: 2026-01-03
---

# Build UI components
`,
		"004-tests.md": `---
id: "004"
title: "Write tests"
status: in-progress
priority: low
effort: small
dependencies: ["001"]
tags: ["test"]
created: 2026-01-04
---

# Write tests
`,
	}

	for name, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644)
		if err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	return tmpDir
}

func setupTestServer(t *testing.T) *gomcp.ClientSession {
	t.Helper()

	ctx := context.Background()

	server := NewServer("test")
	client := gomcp.NewClient(&gomcp.Implementation{
		Name:    "test-client",
		Version: "1.0",
	}, nil)

	st, ct := gomcp.NewInMemoryTransports()

	_, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server connect failed: %v", err)
	}

	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect failed: %v", err)
	}

	return session
}

func callList(t *testing.T, session *gomcp.ClientSession, args map[string]any) []model.Task {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "list",
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

	var tasks []model.Task
	if err := json.Unmarshal([]byte(text.Text), &tasks); err != nil {
		t.Fatalf("failed to unmarshal tasks: %v", err)
	}

	return tasks
}

func TestListTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	tasks := callList(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(tasks))
	}
}

func TestListTool_FilterByStatus(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	tasks := callList(t, session, map[string]any{
		"task_dir": tmpDir,
		"filters":  []string{"status=pending"},
	})

	if len(tasks) != 2 {
		t.Fatalf("expected 2 pending tasks, got %d", len(tasks))
	}
	for _, task := range tasks {
		if task.Status != model.StatusPending {
			t.Errorf("expected status pending, got %s (task %s)", task.Status, task.ID)
		}
	}
}

func TestListTool_FilterByMultiple(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	tasks := callList(t, session, map[string]any{
		"task_dir": tmpDir,
		"filters":  []string{"status=pending", "priority=high"},
	})

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task (pending+high), got %d", len(tasks))
	}
	if tasks[0].ID != "002" {
		t.Errorf("expected task 002, got %s", tasks[0].ID)
	}
}

func TestListTool_FilterByTag(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	tasks := callList(t, session, map[string]any{
		"task_dir": tmpDir,
		"filters":  []string{"tag=security"},
	})

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task with security tag, got %d", len(tasks))
	}
	if tasks[0].ID != "002" {
		t.Errorf("expected task 002, got %s", tasks[0].ID)
	}
}

func TestListTool_Sort(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	tasks := callList(t, session, map[string]any{
		"task_dir": tmpDir,
		"sort":     "priority",
	})

	if len(tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(tasks))
	}

	// Priority order: high (001, 002), medium (003), low (004)
	if tasks[0].Priority != model.PriorityHigh {
		t.Errorf("first task should be high priority, got %s", tasks[0].Priority)
	}
	if tasks[len(tasks)-1].Priority != model.PriorityLow {
		t.Errorf("last task should be low priority, got %s", tasks[len(tasks)-1].Priority)
	}
}

func TestListTool_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	session := setupTestServer(t)

	tasks := callList(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestListTool_InvalidFilter(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "list",
		Arguments: map[string]any{
			"task_dir": tmpDir,
			"filters":  []string{"bad-filter-no-equals"},
		},
	})
	if err != nil {
		// Error returned at protocol level is also acceptable
		return
	}
	if !result.IsError {
		t.Fatal("expected error for invalid filter")
	}
}

func TestListTool_InvalidSort(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "list",
		Arguments: map[string]any{
			"task_dir": tmpDir,
			"sort":     "nonexistent",
		},
	})
	if err != nil {
		// Error returned at protocol level is also acceptable
		return
	}
	if !result.IsError {
		t.Fatal("expected error for invalid sort field")
	}
}

func TestListTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "list" {
			found = true
			if tool.Description == "" {
				t.Error("list tool should have a description")
			}
			if tool.InputSchema == nil {
				t.Error("list tool should have an input schema")
			}
			break
		}
	}
	if !found {
		t.Fatal("list tool not found in tools list")
	}
}
