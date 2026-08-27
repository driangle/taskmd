package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
)

// completedInSibling builds the canonical fixture for view tests: local 001
// pending, 002 pending, and a sibling worktree where 001 is completed.
func completedInSibling(t *testing.T) (*taskRepo, []gitmeta.Worktree) {
	t.Helper()
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
		"002-beta.md":  overlayTaskMD("002", "Beta", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "completed"),
	})}
	return repo, siblings
}

func mustUnmarshal(t *testing.T, out string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(out), v); err != nil {
		t.Fatalf("parse output: %v\n%s", err, out)
	}
}

func TestBoardCommand_WorktreeOverlay_EffectiveStatus(t *testing.T) {
	repo, siblings := completedInSibling(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "board", "--format", "json")
	if res.Err != nil {
		t.Fatalf("board failed: %v", res.Err)
	}

	var groups []struct {
		Group string `json:"group"`
		Tasks []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"tasks"`
	}
	mustUnmarshal(t, res.Stdout, &groups)

	byGroup := map[string][]string{}
	for _, g := range groups {
		for _, task := range g.Tasks {
			byGroup[g.Group] = append(byGroup[g.Group], task.ID)
		}
	}
	if got := byGroup["completed"]; len(got) != 1 || got[0] != "001" {
		t.Errorf("completed group = %v, want [001] (completed in sibling)", got)
	}
	if got := byGroup["pending"]; len(got) != 1 || got[0] != "002" {
		t.Errorf("pending group = %v, want [002]", got)
	}
}

func TestStatsCommand_WorktreeOverlay_EffectiveStatusAndSiblingOnly(t *testing.T) {
	repo, siblings := completedInSibling(t)
	// A task that exists only in a sibling is counted too.
	siblings = append(siblings, newSiblingWorktree(t, "agent-c", "c", map[string]string{
		"003-gamma.md": overlayTaskMD("003", "Gamma", "in-progress"),
	}))

	res := repo.RunWith(stubSiblings(siblings, nil), "stats", "--format", "json")
	if res.Err != nil {
		t.Fatalf("stats failed: %v", res.Err)
	}

	var m struct {
		TotalTasks    int            `json:"total_tasks"`
		TasksByStatus map[string]int `json:"tasks_by_status"`
	}
	mustUnmarshal(t, res.Stdout, &m)

	if m.TotalTasks != 3 {
		t.Errorf("total_tasks = %d, want 3 (2 local + 1 sibling-only)", m.TotalTasks)
	}
	want := map[string]int{"completed": 1, "pending": 1, "in-progress": 1}
	for status, count := range want {
		if m.TasksByStatus[status] != count {
			t.Errorf("tasks_by_status[%s] = %d, want %d (full: %v)", status, m.TasksByStatus[status], count, m.TasksByStatus)
		}
	}
}

func TestGraphCommand_WorktreeOverlay_EffectiveStatus(t *testing.T) {
	repo, siblings := completedInSibling(t)

	// Default view excludes completed tasks: 001 is done per the overlay.
	res := repo.RunWith(stubSiblings(siblings, nil), "graph", "--format", "json")
	if res.Err != nil {
		t.Fatalf("graph failed: %v", res.Err)
	}
	var g struct {
		Nodes []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"nodes"`
	}
	mustUnmarshal(t, res.Stdout, &g)
	for _, n := range g.Nodes {
		if n.ID == "001" {
			t.Errorf("001 present in default graph, but it is completed in a sibling")
		}
	}

	// --all shows it with its effective status.
	res = repo.RunWith(stubSiblings(siblings, nil), "graph", "--format", "json", "--all")
	if res.Err != nil {
		t.Fatalf("graph --all failed: %v", res.Err)
	}
	mustUnmarshal(t, res.Stdout, &g)
	found := false
	for _, n := range g.Nodes {
		if n.ID == "001" {
			found = true
			if n.Status != "completed" {
				t.Errorf("001 status = %s, want completed (effective)", n.Status)
			}
		}
	}
	if !found {
		t.Error("001 missing from graph --all")
	}
}

func TestReportCommand_WorktreeOverlay_EffectiveStatus(t *testing.T) {
	repo, siblings := completedInSibling(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "report", "--format", "json")
	if res.Err != nil {
		t.Fatalf("report failed: %v", res.Err)
	}

	var report struct {
		Summary struct {
			TasksByStatus map[string]int `json:"tasks_by_status"`
		} `json:"summary"`
		Groups []struct {
			Group string `json:"group"`
			Tasks []struct {
				ID string `json:"id"`
			} `json:"tasks"`
		} `json:"groups"`
	}
	mustUnmarshal(t, res.Stdout, &report)

	if report.Summary.TasksByStatus["completed"] != 1 {
		t.Errorf("summary completed = %d, want 1", report.Summary.TasksByStatus["completed"])
	}
	for _, g := range report.Groups {
		if g.Group == "completed" {
			if len(g.Tasks) != 1 || g.Tasks[0].ID != "001" {
				t.Errorf("completed group = %+v, want [001]", g.Tasks)
			}
			return
		}
	}
	t.Error("no completed group in report")
}

func TestPhasesCommand_WorktreeOverlay_EffectiveStatus(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending", `phase: "p1"`),
		"002-beta.md":  overlayTaskMD("002", "Beta", "pending", `phase: "p1"`),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "completed", `phase: "p1"`),
	})}
	phasesConfig := []any{map[string]any{"id": "p1", "name": "Phase 1"}}

	res := repo.RunWith(stubSiblings(siblings, map[string]any{"phases": phasesConfig}), "phases", "--format", "json")
	if res.Err != nil {
		t.Fatalf("phases failed: %v", res.Err)
	}

	var summaries []struct {
		ID   string `json:"id"`
		Done int    `json:"done"`
		Tot  int    `json:"tasks"`
	}
	mustUnmarshal(t, res.Stdout, &summaries)
	if len(summaries) != 1 || summaries[0].ID != "p1" {
		t.Fatalf("summaries = %+v, want single p1", summaries)
	}
	if summaries[0].Done != 1 || summaries[0].Tot != 2 {
		t.Errorf("p1 = %d/%d done, want 1/2 (001 completed in sibling)", summaries[0].Done, summaries[0].Tot)
	}
}

func TestTracksCommand_WorktreeOverlay_SiblingCompletionNotActionable(t *testing.T) {
	repo, siblings := completedInSibling(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "tracks", "--format", "json")
	if res.Err != nil {
		t.Fatalf("tracks failed: %v", res.Err)
	}

	var result struct {
		Tracks []struct {
			Tasks []struct {
				ID string `json:"id"`
			} `json:"tasks"`
		} `json:"tracks"`
		Flexible []struct {
			ID string `json:"id"`
		} `json:"flexible"`
	}
	mustUnmarshal(t, res.Stdout, &result)

	var ids []string
	for _, track := range result.Tracks {
		for _, task := range track.Tasks {
			ids = append(ids, task.ID)
		}
	}
	for _, task := range result.Flexible {
		ids = append(ids, task.ID)
	}
	for _, id := range ids {
		if id == "001" {
			t.Errorf("001 is actionable in tracks, but it is completed in a sibling (ids: %v)", ids)
		}
	}
	if len(ids) != 1 || ids[0] != "002" {
		t.Errorf("actionable ids = %v, want [002]", ids)
	}
}

func TestGetCommand_WorktreeOverlay_WorktreesSection(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001/parser", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "completed", `owner: "alice"`),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "001")
	if res.Err != nil {
		t.Fatalf("get failed: %v", res.Err)
	}
	out := res.Stdout
	if !strings.Contains(out, "Worktrees:") {
		t.Fatalf("get output missing Worktrees section:\n%s", out)
	}
	if !strings.Contains(out, "this worktree: pending") {
		t.Errorf("Worktrees section missing local copy line:\n%s", out)
	}
	if !strings.Contains(out, "agent-b (branch dnc/001/parser): completed (owner: alice)") {
		t.Errorf("Worktrees section missing sibling copy line:\n%s", out)
	}
}

func TestGetCommand_WorktreeOverlay_JSONWorktrees(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/001", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress", `owner: "alice"`),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "001", "--format", "json")
	if res.Err != nil {
		t.Fatalf("get failed: %v", res.Err)
	}

	var out struct {
		Status    string              `json:"status"`
		Worktrees []worktreeCopyEntry `json:"worktrees"`
	}
	mustUnmarshal(t, res.Stdout, &out)

	if out.Status != "pending" {
		t.Errorf("status = %s, want the local pending copy", out.Status)
	}
	if len(out.Worktrees) != 2 {
		t.Fatalf("worktrees = %+v, want 2 entries", out.Worktrees)
	}
	local, sib := out.Worktrees[0], out.Worktrees[1]
	if !local.Local || local.Status != "pending" {
		t.Errorf("local entry = %+v, want local pending first", local)
	}
	if sib.Worktree != "agent-b" || sib.Branch != "dnc/001" || sib.Status != "in-progress" || sib.Owner != "alice" {
		t.Errorf("sibling entry = %+v, want agent-b/dnc/001 in-progress owned by alice", sib)
	}
}

func TestGetCommand_WorktreeOverlay_NoSectionWhenCopiesAgree(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "001")
	if res.Err != nil {
		t.Fatalf("get failed: %v", res.Err)
	}
	if strings.Contains(res.Stdout, "Worktrees:") {
		t.Errorf("Worktrees section rendered although all copies agree:\n%s", res.Stdout)
	}
}

// siblingOnlyTask builds the fixture for the sibling-only get tests: local 001
// only, with 002 existing solely in worktree agent-b on branch dnc/002.
func siblingOnlyTask(t *testing.T) (*taskRepo, []gitmeta.Worktree) {
	t.Helper()
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/002", map[string]string{
		"002-beta.md": overlayTaskMD("002", "Beta", "in-progress", `owner: "alice"`),
	})}
	return repo, siblings
}

func TestGetCommand_SiblingOnlyID_ResolvesWithProvenance(t *testing.T) {
	repo, siblings := siblingOnlyTask(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "002")
	if res.Err != nil {
		t.Fatalf("get on a sibling-only ID failed: %v", res.Err)
	}
	for _, want := range []string{
		"Beta",
		"in-progress",
		"agent-b (branch dnc/002)",
		"remote-only",
	} {
		if !strings.Contains(res.Stdout, want) {
			t.Errorf("get output missing %q:\n%s", want, res.Stdout)
		}
	}
}

func TestGetCommand_SiblingOnlyID_JSONProvenance(t *testing.T) {
	repo, siblings := siblingOnlyTask(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "002", "--format", "json")
	if res.Err != nil {
		t.Fatalf("get on a sibling-only ID failed: %v", res.Err)
	}

	var out struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Status     string `json:"status"`
		RemoteOnly bool   `json:"remote_only"`
		Worktree   string `json:"worktree"`
		Branch     string `json:"branch"`
	}
	mustUnmarshal(t, res.Stdout, &out)

	if out.ID != "002" || out.Title != "Beta" || out.Status != "in-progress" {
		t.Errorf("payload = %+v, want the sibling copy's content", out)
	}
	if !out.RemoteOnly || out.Worktree != "agent-b" || out.Branch != "dnc/002" {
		t.Errorf("provenance = %+v, want remote_only in agent-b/dnc/002", out)
	}
}

func TestGetCommand_SiblingOnlyID_YAMLProvenance(t *testing.T) {
	repo, siblings := siblingOnlyTask(t)

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "002", "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("get on a sibling-only ID failed: %v", res.Err)
	}
	for _, want := range []string{"remote_only: true", "worktree: agent-b", "branch: dnc/002"} {
		if !strings.Contains(res.Stdout, want) {
			t.Errorf("yaml output missing %q:\n%s", want, res.Stdout)
		}
	}
}

// A local task must not grow the remote-only provenance block just because the
// overlay is active — the fields are omitted from json entirely.
func TestGetCommand_LocalIDUnchangedByOverlay(t *testing.T) {
	repo, siblings := siblingOnlyTask(t)

	text := repo.RunWith(stubSiblings(siblings, nil), "get", "001")
	if text.Err != nil {
		t.Fatalf("get failed: %v", text.Err)
	}
	if strings.Contains(text.Stdout, "remote-only") || strings.Contains(text.Stdout, "Worktree:") {
		t.Errorf("local task rendered remote provenance:\n%s", text.Stdout)
	}

	jsonRes := repo.RunWith(stubSiblings(siblings, nil), "get", "001", "--format", "json")
	if jsonRes.Err != nil {
		t.Fatalf("get --format json failed: %v", jsonRes.Err)
	}
	var raw map[string]any
	mustUnmarshal(t, jsonRes.Stdout, &raw)
	for _, key := range []string{"remote_only", "worktree", "branch"} {
		if _, present := raw[key]; present {
			t.Errorf("local task payload carries %q: %v", key, raw[key])
		}
	}
}

// A query matching both a local and a sibling-only task must raise the same
// ambiguity error a purely local collision raises, rather than silently
// preferring one worktree.
func TestGetCommand_AmbiguousQueryAcrossWorktrees(t *testing.T) {
	// Same filename in both worktrees, and a query that matches neither an ID,
	// a title, nor a full file path — so resolution lands on the filename rule
	// with one candidate per worktree.
	repo := newTaskRepo(t, map[string]string{
		"shared.md": overlayTaskMD("001", "First", "pending"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "dnc/002", map[string]string{
		"shared.md": overlayTaskMD("002", "Second", "pending"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "shared")
	if res.Err == nil {
		t.Fatalf("get on an ambiguous filename succeeded:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Err.Error(), "ambiguous filename") {
		t.Errorf("error = %v, want an ambiguous filename error", res.Err)
	}
	for _, want := range []string{"001", "002"} {
		if !strings.Contains(res.Err.Error(), want) {
			t.Errorf("error %v does not name candidate %s", res.Err, want)
		}
	}
}

// With the overlay off, a sibling-only ID is simply not found — get gains no
// cross-worktree reach that worktree_scope: isolated did not ask for.
func TestGetCommand_SiblingOnlyID_IsolatedScopeNotFound(t *testing.T) {
	repo, siblings := siblingOnlyTask(t)

	configure := stubSiblings(siblings, map[string]any{"worktree_scope": worktreeScopeIsolated})
	res := repo.RunWith(configure, "get", "002", "--exact")
	if res.Err == nil {
		t.Fatalf("get resolved a sibling-only ID under isolated scope:\n%s", res.Stdout)
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("error = %v, want task not found", res.Err)
	}
}

// The worklog of a sibling-only task lives in the worktree that owns it; get
// must resolve it there rather than against the local tasks dir.
func TestGetCommand_SiblingOnlyID_WorklogFromOwningWorktree(t *testing.T) {
	repo, siblings := siblingOnlyTask(t)
	wlDir := filepath.Join(siblings[0].TasksDir, ".worklogs")
	if err := os.MkdirAll(wlDir, 0o755); err != nil {
		t.Fatalf("mkdir sibling worklogs: %v", err)
	}
	entry := "# Worklog\n\n## 2026-08-27T10:00:00Z\n\nClaimed in agent-b.\n"
	if err := os.WriteFile(filepath.Join(wlDir, "002.md"), []byte(entry), 0o644); err != nil {
		t.Fatalf("write sibling worklog: %v", err)
	}

	res := repo.RunWith(stubSiblings(siblings, nil), "get", "002")
	if res.Err != nil {
		t.Fatalf("get failed: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "Worklog: 1 entries") {
		t.Errorf("get output missing the sibling worklog:\n%s", res.Stdout)
	}
}

func TestValidateCommand_WorktreeOverlay_DivergentTerminalWarning(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "cancelled"),
	})
	siblings := []gitmeta.Worktree{newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "completed"),
	})}

	res := repo.RunWith(stubSiblings(siblings, nil), "validate", "--format", "json")
	if res.Err != nil {
		t.Fatalf("validate failed: %v", res.Err)
	}

	var result struct {
		Issues []struct {
			Level   string `json:"level"`
			TaskID  string `json:"task_id"`
			Message string `json:"message"`
		} `json:"issues"`
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	}
	mustUnmarshal(t, res.Stdout, &result)

	if result.Errors != 0 {
		t.Errorf("errors = %d, want 0", result.Errors)
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Level == "warning" && issue.TaskID == "001" &&
			strings.Contains(issue.Message, "completed in worktree agent-b but cancelled in this worktree") {
			found = true
		}
	}
	if !found {
		t.Errorf("no divergent-terminal-state warning in issues: %+v", result.Issues)
	}
}

// nestedSiblingCheckout lays a sibling worktree checkout *inside* the repo's
// scan root (non-hidden path), the §8 double-scan case, and returns its
// descriptor.
func nestedSiblingCheckout(t *testing.T, repo *taskRepo, name string, files map[string]string) gitmeta.Worktree {
	t.Helper()
	for fname, content := range files {
		repo.Write(filepath.Join(name, "tasks", fname), content)
	}
	root := filepath.Join(repo.Dir, name)
	return gitmeta.Worktree{Root: root, Branch: name, TasksDir: filepath.Join(root, "tasks")}
}

func TestValidateCommand_WorktreeOverlay_NestedCheckoutNotDuplicate(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	sibling := nestedSiblingCheckout(t, repo, "agent-b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
	})

	res := repo.RunWith(stubSiblings([]gitmeta.Worktree{sibling}, nil), "validate", "--format", "json")
	if res.Err != nil {
		t.Fatalf("validate failed: %v", res.Err)
	}
	var result struct {
		Errors int `json:"errors"`
	}
	mustUnmarshal(t, res.Stdout, &result)
	if result.Errors != 0 {
		t.Errorf("errors = %d, want 0 (nested checkout copies are not duplicates)\n%s", result.Errors, res.Stdout)
	}
}

func TestListCommand_WorktreeOverlay_NestedCheckoutAttributed(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "pending"),
	})
	sibling := nestedSiblingCheckout(t, repo, "agent-b", map[string]string{
		"001-alpha.md": overlayTaskMD("001", "Alpha", "in-progress"),
	})

	res := repo.RunWith(stubSiblings([]gitmeta.Worktree{sibling}, nil), "list", "--format", "json")
	if res.Err != nil {
		t.Fatalf("list failed: %v", res.Err)
	}
	if strings.Contains(res.Stderr, "duplicate task IDs") {
		t.Errorf("duplicate-ID warning fired for a nested sibling checkout:\n%s", res.Stderr)
	}

	var rows []map[string]any
	mustUnmarshal(t, res.Stdout, &rows)
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1 (single merged 001)\n%s", len(rows), res.Stdout)
	}
	if rows[0]["effective_status"] != "in-progress" || rows[0]["worktree"] != "agent-b" {
		t.Errorf("001 = %v, want effective in-progress attributed to agent-b", rows[0])
	}
}
