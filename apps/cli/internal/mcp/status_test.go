package mcp

import (
	"context"
	"encoding/json"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func callStatus(t *testing.T, session *gomcp.ClientSession, args map[string]any) statusOutput {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "status",
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

	var out statusOutput
	if err := json.Unmarshal([]byte(text.Text), &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	return out
}

func TestStatusTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callStatus(t, session, map[string]any{
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

func TestStatusTool_NoContentField(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "status",
		Arguments: map[string]any{
			"task_dir": tmpDir,
			"task_id":  "002",
		},
	})
	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	text := result.Content[0].(*gomcp.TextContent)

	var raw map[string]any
	if err := json.Unmarshal([]byte(text.Text), &raw); err != nil {
		t.Fatalf("failed to unmarshal raw JSON: %v", err)
	}

	if _, ok := raw["content"]; ok {
		t.Error("status output should not contain 'content' key")
	}
	if _, ok := raw["depends_on"]; ok {
		t.Error("status output should not contain 'depends_on' key (resolved deps)")
	}
	if _, ok := raw["blocks"]; ok {
		t.Error("status output should not contain 'blocks' key (resolved deps)")
	}
}

func TestStatusTool_NotFound(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "status",
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

func TestStatusTool_MissingTaskID(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "status",
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

func TestStatusTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "status" {
			found = true
			if tool.Description == "" {
				t.Error("status tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("status tool not found in tools list")
	}
}
