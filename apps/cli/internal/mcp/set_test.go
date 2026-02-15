package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func callSet(t *testing.T, session *gomcp.ClientSession, args map[string]any) setOutput {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "set",
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

	var out setOutput
	if err := json.Unmarshal([]byte(text.Text), &out); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	return out
}

func callSetExpectError(t *testing.T, session *gomcp.ClientSession, args map[string]any) {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "set",
		Arguments: args,
	})
	if err != nil {
		// Error at protocol level is acceptable
		return
	}
	if !result.IsError {
		t.Fatal("expected error but tool succeeded")
	}
}

func readFileContent(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(data)
}

func TestSetTool_UpdateStatus(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
		"status":   "in-progress",
	})

	if out.TaskID != "002" {
		t.Errorf("expected task_id 002, got %s", out.TaskID)
	}
	if out.Updated["status"] != "in-progress" {
		t.Errorf("expected updated status in-progress, got %s", out.Updated["status"])
	}

	// Verify the file was actually updated
	content := readFileContent(t, filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(content, "status: in-progress") {
		t.Error("expected file to contain 'status: in-progress'")
	}
}

func TestSetTool_UpdatePriority(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "003",
		"priority": "critical",
	})

	if out.Updated["priority"] != "critical" {
		t.Errorf("expected updated priority critical, got %s", out.Updated["priority"])
	}

	content := readFileContent(t, filepath.Join(tmpDir, "003-ui.md"))
	if !strings.Contains(content, "priority: critical") {
		t.Error("expected file to contain 'priority: critical'")
	}
}

func TestSetTool_UpdateEffort(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "003",
		"effort":   "small",
	})

	if out.Updated["effort"] != "small" {
		t.Errorf("expected updated effort small, got %s", out.Updated["effort"])
	}

	content := readFileContent(t, filepath.Join(tmpDir, "003-ui.md"))
	if !strings.Contains(content, "effort: small") {
		t.Error("expected file to contain 'effort: small'")
	}
}

func TestSetTool_UpdateOwner(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
		"owner":    "alice",
	})

	if out.Updated["owner"] != "alice" {
		t.Errorf("expected updated owner alice, got %s", out.Updated["owner"])
	}

	content := readFileContent(t, filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(content, "owner: alice") {
		t.Error("expected file to contain 'owner: alice'")
	}
}

func TestSetTool_UpdateMultipleFields(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
		"status":   "in-progress",
		"priority": "critical",
		"effort":   "large",
	})

	if out.Updated["status"] != "in-progress" {
		t.Errorf("expected status in-progress, got %s", out.Updated["status"])
	}
	if out.Updated["priority"] != "critical" {
		t.Errorf("expected priority critical, got %s", out.Updated["priority"])
	}
	if out.Updated["effort"] != "large" {
		t.Errorf("expected effort large, got %s", out.Updated["effort"])
	}

	content := readFileContent(t, filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(content, "status: in-progress") {
		t.Error("expected file to contain 'status: in-progress'")
	}
	if !strings.Contains(content, "priority: critical") {
		t.Error("expected file to contain 'priority: critical'")
	}
	if !strings.Contains(content, "effort: large") {
		t.Error("expected file to contain 'effort: large'")
	}
}

func TestSetTool_AddTags(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	out := callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
		"add_tags": []string{"urgent"},
	})

	if out.Updated["add_tags"] != "urgent" {
		t.Errorf("expected add_tags urgent, got %s", out.Updated["add_tags"])
	}

	content := readFileContent(t, filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(content, "urgent") {
		t.Error("expected file to contain 'urgent' tag")
	}
	// Original tag should still be present
	if !strings.Contains(content, "infra") {
		t.Error("expected file to still contain 'infra' tag")
	}
}

func TestSetTool_RemoveTags(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
		"rem_tags": []string{"security"},
	})

	content := readFileContent(t, filepath.Join(tmpDir, "002-auth.md"))
	if strings.Contains(content, "security") {
		t.Error("expected 'security' tag to be removed")
	}
	if !strings.Contains(content, "feature") {
		t.Error("expected 'feature' tag to still be present")
	}
}

func TestSetTool_ReplaceTags(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSet(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "002",
		"tags":     []string{"new-tag"},
	})

	content := readFileContent(t, filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(content, "new-tag") {
		t.Error("expected file to contain 'new-tag'")
	}
	if strings.Contains(content, "feature") {
		t.Error("expected 'feature' tag to be replaced")
	}
	if strings.Contains(content, "security") {
		t.Error("expected 'security' tag to be replaced")
	}
}

func TestSetTool_InvalidStatus(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSetExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
		"status":   "invalid-status",
	})
}

func TestSetTool_InvalidPriority(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSetExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
		"priority": "super-high",
	})
}

func TestSetTool_InvalidEffort(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSetExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
		"effort":   "huge",
	})
}

func TestSetTool_MissingTaskID(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSetExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"status":   "pending",
	})
}

func TestSetTool_TaskNotFound(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSetExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "999",
		"status":   "pending",
	})
}

func TestSetTool_NoFieldsToUpdate(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	callSetExpectError(t, session, map[string]any{
		"task_dir": tmpDir,
		"task_id":  "001",
	})
}

func TestSetTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "set" {
			found = true
			if tool.Description == "" {
				t.Error("set tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("set tool not found in tools list")
	}
}
