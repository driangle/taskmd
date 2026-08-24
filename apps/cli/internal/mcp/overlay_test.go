package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
)

// overlayTaskMD renders a minimal task file for overlay tests.
func overlayTaskMD(id, title, status string, extraFrontmatter ...string) string {
	extra := ""
	if len(extraFrontmatter) > 0 {
		extra = strings.Join(extraFrontmatter, "\n") + "\n"
	}
	return fmt.Sprintf(`---
id: %q
title: %q
status: %s
priority: medium
created: 2026-01-01
%s---

# %s
`, id, title, status, extra, title)
}

// newSiblingWorktree lays out a fake sibling worktree and returns its
// descriptor for the injected discovery seam.
func newSiblingWorktree(t *testing.T, name, branch string, files map[string]string) gitmeta.Worktree {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	tasksDir := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("mkdir sibling tasks dir: %v", err)
	}
	for fname, content := range files {
		if err := os.WriteFile(filepath.Join(tasksDir, fname), []byte(content), 0o644); err != nil {
			t.Fatalf("write sibling task %s: %v", fname, err)
		}
	}
	return gitmeta.Worktree{Root: root, Branch: branch, TasksDir: tasksDir}
}

// overlayBuilder returns a Builder whose discovery is stubbed to the given
// siblings — no git involved.
func overlayBuilder(siblings ...gitmeta.Worktree) worktree.Builder {
	return worktree.Builder{
		Enabled:  true,
		Discover: func(string) ([]gitmeta.Worktree, error) { return siblings, nil },
	}
}

// overlayFixture is a local task dir with one task claimed in a sibling
// worktree (001) and one sibling-only task (099).
func overlayFixture(t *testing.T) (string, worktree.Builder) {
	t.Helper()
	localDir := t.TempDir()
	files := map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(localDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write local task %s: %v", name, err)
		}
	}
	sibling := newSiblingWorktree(t, "agent-b", "dnc/001/parser", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress", `owner: "agent-b"`),
		"099-sibling.md": overlayTaskMD("099", "Sibling only", "pending"),
	})
	return localDir, overlayBuilder(sibling)
}

// callToolJSON invokes a tool and unmarshals its text content into out.
func callToolJSON(t *testing.T, session *gomcp.ClientSession, name string, args map[string]any, out any) {
	t.Helper()
	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
	if err != nil {
		t.Fatalf("CallTool %s failed: %v", name, err)
	}
	if result.IsError {
		t.Fatalf("%s returned error: %+v", name, result.Content)
	}
	text := result.Content[0].(*gomcp.TextContent)
	if err := json.Unmarshal([]byte(text.Text), out); err != nil {
		t.Fatalf("unmarshal %s output: %v\n%s", name, err, text.Text)
	}
}

func TestListTool_WorktreeOverlay_AdditiveFields(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	var rows []map[string]any
	callToolJSON(t, session, "list", map[string]any{"task_dir": localDir}, &rows)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (2 local + 1 sibling-only), got %d", len(rows))
	}
	byID := map[string]map[string]any{}
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}

	claimed := byID["001"]
	if claimed["status"] != "pending" {
		t.Errorf("001 status = %v, want local pending (additive-only shape)", claimed["status"])
	}
	if claimed["effective_status"] != "in-progress" || claimed["effective_owner"] != "agent-b" {
		t.Errorf("001 missing effective fields: %v", claimed)
	}
	if claimed["worktree"] != "agent-b" || claimed["branch"] != "dnc/001/parser" {
		t.Errorf("001 missing provenance: %v", claimed)
	}

	if free := byID["002"]; free["worktree"] != nil || free["remote_only"] != nil {
		t.Errorf("002 should carry no provenance: %v", free)
	}
	if sib := byID["099"]; sib["remote_only"] != true {
		t.Errorf("099 should be remote_only: %v", sib)
	}
}

func TestListTool_WorktreeOverlay_StatusFiltersEffective(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	var rows []map[string]any
	callToolJSON(t, session, "list", map[string]any{
		"task_dir": localDir,
		"filters":  []string{"status=in-progress"},
	}, &rows)

	if len(rows) != 1 || rows[0]["id"] != "001" {
		t.Fatalf("rows = %v, want only 001 (in-progress by effective status)", rows)
	}
}

func TestListTool_WorktreeOverlay_InactiveShapeUnchanged(t *testing.T) {
	localDir, _ := overlayFixture(t)
	session := setupTestServer(t) // zero-value builder: overlay disabled

	var rows []map[string]any
	callToolJSON(t, session, "list", map[string]any{"task_dir": localDir}, &rows)

	if len(rows) != 2 {
		t.Fatalf("expected only the 2 local tasks, got %d", len(rows))
	}
	for _, row := range rows {
		if _, ok := row["effective_status"]; ok {
			t.Errorf("effective_status present with overlay inactive: %v", row)
		}
	}
}

func TestGetTool_WorktreeOverlay_ProvenanceAndCopies(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	var out map[string]any
	callToolJSON(t, session, "get", map[string]any{"task_dir": localDir, "task_id": "001"}, &out)

	if out["status"] != "pending" || out["effective_status"] != "in-progress" {
		t.Errorf("get 001 = status %v effective %v, want pending/in-progress", out["status"], out["effective_status"])
	}
	if out["worktree"] != "agent-b" {
		t.Errorf("get 001 missing worktree provenance: %v", out["worktree"])
	}
	copies, ok := out["worktrees"].([]any)
	if !ok || len(copies) != 2 {
		t.Fatalf("get 001 worktrees section = %v, want 2 diverging copies", out["worktrees"])
	}
}

func TestGetTool_WorktreeOverlay_RemoteOnlyResolves(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	var out map[string]any
	callToolJSON(t, session, "get", map[string]any{"task_dir": localDir, "task_id": "099"}, &out)

	if out["id"] != "099" || out["remote_only"] != true {
		t.Errorf("get 099 should resolve the sibling-only task with remote_only, got %v", out)
	}
}

func TestStatusTool_WorktreeOverlay_EffectiveFields(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	var out map[string]any
	callToolJSON(t, session, "status", map[string]any{"task_dir": localDir, "task_id": "001"}, &out)

	if out["status"] != "pending" || out["effective_status"] != "in-progress" || out["worktree"] != "agent-b" {
		t.Errorf("status 001 missing overlay fields: %v", out)
	}
}

func TestNextTool_WorktreeOverlay_ExcludesSiblingClaims(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	var recs []map[string]any
	callToolJSON(t, session, "next", map[string]any{"task_dir": localDir}, &recs)

	for _, rec := range recs {
		id := rec["id"]
		if id == nil {
			id = rec["task_id"]
		}
		if id == "001" || id == "099" {
			t.Errorf("next recommended %v, which is claimed/only in a sibling worktree", id)
		}
	}
	if len(recs) == 0 {
		t.Fatal("expected 002 to be recommended")
	}
}

func TestSetTool_WorktreeOverlay_SiblingOnlyGuard(t *testing.T) {
	localDir, wt := overlayFixture(t)
	session := setupTestServerWith(t, wt)

	result, err := session.CallTool(context.Background(), &gomcp.CallToolParams{
		Name: "set",
		Arguments: map[string]any{
			"task_dir": localDir,
			"task_id":  "099",
			"status":   "completed",
		},
	})
	if err == nil && !result.IsError {
		t.Fatal("set on a sibling-only ID should fail with the guard error")
	}
	var msg string
	if err != nil {
		msg = err.Error()
	} else if len(result.Content) > 0 {
		if text, ok := result.Content[0].(*gomcp.TextContent); ok {
			msg = text.Text
		}
	}
	if !strings.Contains(msg, "exists only in worktree") || !strings.Contains(msg, "run taskmd there") {
		t.Errorf("guard message = %q, want sibling-only guidance", msg)
	}
}

func TestValidateTool_WorktreeOverlay_DivergenceWarning(t *testing.T) {
	localDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(localDir, "001-a.md"),
		[]byte(overlayTaskMD("001", "Alpha", "cancelled")), 0o644); err != nil {
		t.Fatal(err)
	}
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-a.md": overlayTaskMD("001", "Alpha", "completed"),
	})
	session := setupTestServerWith(t, overlayBuilder(sibling))

	var out struct {
		Warnings int `json:"warnings"`
		Issues   []struct {
			Level   string `json:"level"`
			Message string `json:"message"`
		} `json:"issues"`
	}
	callToolJSON(t, session, "validate", map[string]any{"task_dir": localDir}, &out)

	found := false
	for _, issue := range out.Issues {
		if issue.Level == "warning" && strings.Contains(issue.Message, "completed in worktree agent-b") {
			found = true
		}
	}
	if !found {
		t.Errorf("validate should warn about divergent terminal states, got %+v", out.Issues)
	}
}
