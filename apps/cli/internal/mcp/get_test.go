package mcp

import (
	"context"
	"encoding/json"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func callGet(t *testing.T, session *gomcp.ClientSession, args map[string]any) getOutput {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "get",
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

	var out getOutput
	if err := json.Unmarshal([]byte(text.Text), &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	return out
}

func TestGetTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
	})

	if out.ID != "002" {
		t.Errorf("expected ID 002, got %s", out.ID)
	}
	if out.Title != "Add authentication" {
		t.Errorf("expected title 'Add authentication', got %s", out.Title)
	}
	if out.Status != "pending" {
		t.Errorf("expected status pending, got %s", out.Status)
	}
	if out.Priority != "high" {
		t.Errorf("expected priority high, got %s", out.Priority)
	}
}

func TestGetTool_IncludesBody(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
	})

	if out.Content == "" {
		t.Error("expected non-empty content (body)")
	}
}

func TestGetTool_IncludesDependencies(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callGet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
	})

	if len(out.DependsOn) != 1 {
		t.Fatalf("expected 1 dependency, got %d", len(out.DependsOn))
	}
	if out.DependsOn[0].ID != "001" {
		t.Errorf("expected dependency on 001, got %s", out.DependsOn[0].ID)
	}
	if out.DependsOn[0].Title != "Setup project" {
		t.Errorf("expected dependency title 'Setup project', got %s", out.DependsOn[0].Title)
	}
}

func TestGetTool_IncludesBlocks(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	// Task 001 is depended on by 002 and 004, so it blocks them
	out := callGet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
	})

	if len(out.Blocks) < 1 {
		t.Fatalf("expected at least 1 block, got %d", len(out.Blocks))
	}

	blockIDs := make(map[string]bool)
	for _, b := range out.Blocks {
		blockIDs[b.ID] = true
	}
	if !blockIDs["002"] {
		t.Error("expected 001 to block 002")
	}
	if !blockIDs["004"] {
		t.Error("expected 001 to block 004")
	}
}

func TestGetTool_NotFound(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "get",
		Arguments: map[string]any{
			"task_dir": tmpDir,
			"task_id":  "999",
		},
	})
	if err != nil {
		// Error at protocol level is acceptable
		return
	}
	if !result.IsError {
		t.Fatal("expected error for non-existent task")
	}
}

func TestGetTool_MissingTaskID(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "get",
		Arguments: map[string]any{
			"task_dir": tmpDir,
		},
	})
	if err != nil {
		return
	}
	if !result.IsError {
		t.Fatal("expected error for missing task_id")
	}
}

func TestGetTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "get" {
			found = true
			if tool.Description == "" {
				t.Error("get tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("get tool not found in tools list")
	}
}
