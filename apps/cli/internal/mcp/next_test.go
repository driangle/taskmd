package mcp

import (
	"context"
	"encoding/json"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/next"
)

func callNext(t *testing.T, session *gomcp.ClientSession, args map[string]any) []next.Recommendation {
	t.Helper()

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      "next",
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

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(text.Text), &recs); err != nil {
		t.Fatalf("failed to unmarshal recommendations: %v", err)
	}
	return recs
}

func TestNextTool_HappyPath(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	recs := callNext(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	// We have 4 tasks: 001 completed, 002 pending (dep on 001=completed), 003 pending (no deps), 004 in-progress (dep on 001=completed)
	// All of 002, 003, 004 are actionable
	if len(recs) == 0 {
		t.Fatal("expected at least one recommendation")
	}
	if len(recs) > 5 {
		t.Errorf("expected at most 5 recommendations (default limit), got %d", len(recs))
	}

	// Each recommendation should have a rank
	for i, rec := range recs {
		if rec.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, rec.Rank)
		}
	}
}

func TestNextTool_WithLimit(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	recs := callNext(t, session, map[string]any{
		"task_dir": tmpDir,
		"limit":    1,
	})

	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation with limit=1, got %d", len(recs))
	}
}

func TestNextTool_WithFilters(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	recs := callNext(t, session, map[string]any{
		"task_dir": tmpDir,
		"filters":  []string{"priority=medium"},
	})

	for _, rec := range recs {
		if rec.Priority != "medium" {
			t.Errorf("expected priority medium, got %s (task %s)", rec.Priority, rec.ID)
		}
	}
}

func TestNextTool_QuickWins(t *testing.T) {
	tmpDir := createTestTaskFiles(t)
	session := setupTestServer(t)

	recs := callNext(t, session, map[string]any{
		"task_dir":   tmpDir,
		"quick_wins": true,
	})

	for _, rec := range recs {
		if rec.Effort != "small" {
			t.Errorf("expected small effort for quick wins, got %s (task %s)", rec.Effort, rec.ID)
		}
	}
}

func TestNextTool_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	session := setupTestServer(t)

	recs := callNext(t, session, map[string]any{
		"task_dir": tmpDir,
	})

	if len(recs) != 0 {
		t.Fatalf("expected 0 recommendations for empty dir, got %d", len(recs))
	}
}

func TestNextTool_Discoverable(t *testing.T) {
	session := setupTestServer(t)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	found := false
	for _, tool := range result.Tools {
		if tool.Name == "next" {
			found = true
			if tool.Description == "" {
				t.Error("next tool should have a description")
			}
			break
		}
	}
	if !found {
		t.Fatal("next tool not found in tools list")
	}
}
