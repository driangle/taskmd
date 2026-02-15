package mcp

import (
	"context"
	"encoding/json"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/search"
)

func callSearch(t *testing.T, session *gomcp.ClientSession, args map[string]any) []search.Result {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "search",
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

	var results []search.Result
	if err := json.Unmarshal([]byte(text.Text), &results); err != nil {
		t.Fatalf("failed to unmarshal search results: %v", err)
	}
	return results
}

func TestSearchTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	results := callSearch(t, session, map[string]any{
		"task_dir": tmpDir,
		"query":    "authentication",
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'authentication', got %d", len(results))
	}
	if results[0].ID != "002" {
		t.Errorf("expected task 002, got %s", results[0].ID)
	}
}

func TestSearchTool_CaseInsensitive(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	results := callSearch(t, session, map[string]any{
		"task_dir": tmpDir,
		"query":    "SETUP",
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'SETUP', got %d", len(results))
	}
	if results[0].ID != "001" {
		t.Errorf("expected task 001, got %s", results[0].ID)
	}
}

func TestSearchTool_MultipleMatches(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	// "Build" matches "Build UI components" title
	// "project" matches "Setup project" title
	// Both have bodies starting with "# <title>" so query on common word
	results := callSearch(t, session, map[string]any{
		"task_dir": tmpDir,
		"query":    "ui",
	})

	if len(results) < 1 {
		t.Fatal("expected at least 1 result for 'ui'")
	}

	foundUI := false
	for _, r := range results {
		if r.ID == "003" {
			foundUI = true
		}
	}
	if !foundUI {
		t.Error("expected task 003 in results")
	}
}

func TestSearchTool_NoMatches(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	results := callSearch(t, session, map[string]any{
		"task_dir": tmpDir,
		"query":    "zzzznonexistent",
	})

	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearchTool_MissingQuery(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "search",
		Arguments: map[string]any{
			"task_dir": tmpDir,
		},
	})
	if err != nil {
		return
	}
	if !result.IsError {
		t.Fatal("expected error for missing query")
	}
}

func TestSearchTool_ResultFields(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	results := callSearch(t, session, map[string]any{
		"task_dir": tmpDir,
		"query":    "authentication",
	})

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	r := results[0]
	if r.ID == "" {
		t.Error("expected non-empty ID")
	}
	if r.Title == "" {
		t.Error("expected non-empty title")
	}
	if r.Status == "" {
		t.Error("expected non-empty status")
	}
	if r.MatchLocation == "" {
		t.Error("expected non-empty match_location")
	}
}

func TestSearchTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "search" {
			found = true
			if tool.Description == "" {
				t.Error("search tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("search tool not found in tools list")
	}
}
