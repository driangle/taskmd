package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/sdk/go/model"
)

// Merge-level tests (status ladder, provenance, mtime tie-break, remote-only,
// divergence warnings) live in internal/worktree, next to the merge layer.
// This file covers the CLI wiring: activation, next/list behavior, and the
// mutation guard.

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
created_at: 2026-01-01
%s---

# %s
`, id, title, status, extra, title)
}

// newSiblingWorktree lays out a fake sibling worktree (root + tasks dir with
// the given files) and returns its descriptor for the discovery stub.
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

// stubSiblings returns a RunWith configure hook that injects the given sibling
// worktrees into discovery (and optionally seeds viper keys).
func stubSiblings(siblings []gitmeta.Worktree, viperSets map[string]any) func() {
	return func() {
		discoverSiblingWorktrees = func(string) ([]gitmeta.Worktree, error) {
			return siblings, nil
		}
		for k, v := range viperSets {
			viper.Set(k, v)
		}
	}
}

// scanDirTasks scans a directory into model tasks for direct overlay tests.
func scanDirTasks(t *testing.T, dir string) []*model.Task {
	t.Helper()
	result, err := newTaskScanner(dir, GlobalFlags{}).Scan()
	if err != nil {
		t.Fatalf("scan %s: %v", dir, err)
	}
	return result.Tasks
}

func TestBuildWorktreeOverlay_ActivationMatrix(t *testing.T) {
	sibling := func(t *testing.T) []gitmeta.Worktree {
		return []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
			"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
		})}
	}

	cases := []struct {
		name       string
		mode       any  // viper "worktrees" value; nil = unset
		siblings   bool // discovery returns a sibling
		wantActive bool
		wantErr    bool
	}{
		{name: "default auto with siblings", mode: nil, siblings: true, wantActive: true},
		{name: "default auto without siblings", mode: nil, siblings: false},
		{name: "auto with siblings", mode: "auto", siblings: true, wantActive: true},
		{name: "true with siblings", mode: true, siblings: true, wantActive: true},
		{name: "true without siblings", mode: true, siblings: false},
		{name: "false with siblings", mode: "false", siblings: true},
		{name: "invalid value", mode: "sometimes", siblings: true, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetCLIState()
			if tc.mode != nil {
				viper.Set("worktrees", tc.mode)
			}
			if tc.siblings {
				sibs := sibling(t)
				discoverSiblingWorktrees = func(string) ([]gitmeta.Worktree, error) { return sibs, nil }
			}

			repo := newTaskRepo(t, map[string]string{"001-alpha.md": overlayTaskMD("001", "Alpha", "pending")})
			overlay, err := buildWorktreeOverlay(repo.Dir, scanDirTasks(t, repo.Dir), GlobalFlags{Quiet: true})

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error for invalid worktrees value")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (overlay != nil) != tc.wantActive {
				t.Errorf("overlay active = %v, want %v", overlay != nil, tc.wantActive)
			}
		})
	}
	resetCLIState()
}

func TestNextCommand_WorktreeOverlay_ExcludesSiblingClaims(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "next", "--format", "json")
	if res.Err != nil {
		t.Fatalf("next failed: %v", res.Err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("parse output: %v\n%s", err, res.Stdout)
	}
	ids := make([]string, len(recs))
	for i, r := range recs {
		ids[i] = r.ID
	}
	if len(recs) != 1 || recs[0].ID != "002" {
		t.Errorf("recommendations = %v, want only 002 (001 is claimed in a sibling)", ids)
	}
}

func TestNextCommand_WorktreeOverlay_ExplainNamesWorktree(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "next", "--explain")
	if res.Err != nil {
		t.Fatalf("next --explain failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "Excluded by sibling worktrees:") {
		t.Errorf("--explain output missing exclusion section:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "in-progress in worktree agent-b (branch dnc/001)") {
		t.Errorf("--explain output does not name the excluding worktree:\n%s", res.Stdout)
	}
}

func TestNextCommand_WorktreeOverlay_SiblingCompletionUnblocksDependent(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-dep.md":       overlayTaskMD("001", "Dep", "pending"),
		"002-dependent.md": overlayTaskMD("002", "Dependent", "pending", `dependencies: ["001"]`),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-dep.md": overlayTaskMD("001", "Dep", "completed"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "next", "--format", "json")
	if res.Err != nil {
		t.Fatalf("next failed: %v", res.Err)
	}
	var recs []Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("parse output: %v\n%s", err, res.Stdout)
	}
	if len(recs) != 1 || recs[0].ID != "002" {
		t.Errorf("recommendations = %+v, want only 002 (001 completed in sibling unblocks it)", recs)
	}
}

func TestListCommand_WorktreeOverlay_ColumnAndEffectiveStatus(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
		"003-sibling.md": overlayTaskMD("003", "Sibling only", "pending"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "list")
	if res.Err != nil {
		t.Fatalf("list failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "worktree") {
		t.Errorf("list output missing worktree column:\n%s", res.Stdout)
	}
	claimedRow := lineContaining(res.Stdout, "001")
	if !strings.Contains(claimedRow, "in-progress") || !strings.Contains(claimedRow, "agent-b") {
		t.Errorf("001 row should show effective status and provenance, got: %s", claimedRow)
	}
	siblingRow := lineContaining(res.Stdout, "003")
	if !strings.Contains(siblingRow, "agent-b*") {
		t.Errorf("sibling-only 003 should be marked with agent-b*, got: %s", siblingRow)
	}
}

func TestListCommand_WorktreeOverlay_StatusFiltersEffective(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "list", "--status", "in-progress", "--format", "json")
	if res.Err != nil {
		t.Fatalf("list failed: %v", res.Err)
	}
	var rows []map[string]any
	if err := json.Unmarshal([]byte(res.Stdout), &rows); err != nil {
		t.Fatalf("parse output: %v\n%s", err, res.Stdout)
	}
	if len(rows) != 1 || rows[0]["id"] != "001" {
		t.Fatalf("rows = %v, want only 001 (in-progress by effective status)", rows)
	}
	if rows[0]["effective_status"] != "in-progress" || rows[0]["worktree"] != "agent-b" {
		t.Errorf("row missing overlay fields: %v", rows[0])
	}
	if rows[0]["status"] != "pending" {
		t.Errorf("base status = %v, want the local pending copy", rows[0]["status"])
	}
}

func TestListCommand_WorktreeOverlay_NoAnnotationsNoColumn(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
	})
	// Sibling has the identical (older) copy: local wins, nothing annotated.
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "list")
	if res.Err != nil {
		t.Fatalf("list failed: %v", res.Err)
	}
	if strings.Contains(res.Stdout, "worktree") {
		t.Errorf("worktree column rendered with no annotated tasks:\n%s", res.Stdout)
	}
}

func TestNextCommand_WorktreesFalseFlagRestoresTodayBehavior(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "next", "--worktrees", "false", "--format", "json")
	if res.Err != nil {
		t.Fatalf("next failed: %v", res.Err)
	}
	var recs []Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("parse output: %v\n%s", err, res.Stdout)
	}
	if len(recs) != 1 || recs[0].ID != "001" {
		t.Errorf("with --worktrees=false, 001 should be recommended as today; got %+v", recs)
	}
}

func TestSetCommand_SiblingOnlyID_GuardFails(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-local.md": overlayTaskMD("001", "Local", "pending"),
	})
	siblingContent := overlayTaskMD("099", "Sibling only", "pending")
	sibling := newSiblingWorktree(t, "agent-b", "dnc/099/parser", map[string]string{
		"099-sibling.md": siblingContent,
	})

	res := repo.RunWith(stubSiblings([]gitmeta.Worktree{sibling}, nil), "set", "099", "--status", "completed")
	if res.Err == nil {
		t.Fatal("set on a sibling-only ID should fail")
	}
	msg := res.Err.Error()
	if !strings.Contains(msg, "exists only in worktree") ||
		!strings.Contains(msg, "agent-b") ||
		!strings.Contains(msg, "dnc/099/parser") ||
		!strings.Contains(msg, "run taskmd there") {
		t.Errorf("guard message = %q, want worktree + branch + guidance", msg)
	}

	// The sibling's file must never be written.
	data, err := os.ReadFile(filepath.Join(sibling.TasksDir, "099-sibling.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != siblingContent {
		t.Error("sibling worktree file was modified by set")
	}
}

func TestSetCommand_LocalTaskUnaffectedByOverlay(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-local.md": overlayTaskMD("001", "Local", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-local.md": overlayTaskMD("001", "Local", "in-progress"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "set", "001", "--status", "in-progress")
	if res.Err != nil {
		t.Fatalf("set on a local task should succeed with overlay active: %v", res.Err)
	}
	data, err := os.ReadFile(repo.Path("001-local.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "status: in-progress") {
		t.Error("local task file was not updated")
	}
}

// lineContaining returns the first output line containing substr.
func lineContaining(out, substr string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, substr) {
			return line
		}
	}
	return ""
}
