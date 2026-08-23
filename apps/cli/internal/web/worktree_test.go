package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
	"github.com/driangle/taskmd/apps/cli/internal/watcher"
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
)

// effortScaleForTest returns the default effort vocabulary.
func effortScaleForTest() effort.Scale { return effort.Default() }

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
		Mode:     worktree.ModeAuto,
		Discover: func(string) ([]gitmeta.Worktree, error) { return siblings, nil },
	}
}

// overlayFixtureDP builds a DataProvider over a local dir with one task
// claimed in a sibling worktree (001), a free local task (002), and a
// sibling-only task (099).
func overlayFixtureDP(t *testing.T) (*DataProvider, string, gitmeta.Worktree) {
	t.Helper()
	localDir := t.TempDir()
	for name, content := range map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "pending"),
		"002-free.md":    overlayTaskMD("002", "Free", "pending"),
	} {
		if err := os.WriteFile(filepath.Join(localDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sibling := newSiblingWorktree(t, "agent-b", "dnc/001/parser", map[string]string{
		"001-claimed.md": overlayTaskMD("001", "Claimed elsewhere", "in-progress", `owner: "agent-b"`),
		"099-sibling.md": overlayTaskMD("099", "Sibling only", "pending"),
	})
	dp := NewDataProviderWithWorktrees(localDir, false, overlayBuilder(sibling))
	return dp, localDir, sibling
}

func getJSON(t *testing.T, handler http.HandlerFunc, url string, out any) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code == http.StatusOK && out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("unmarshal %s: %v\n%s", url, err, rec.Body.String())
		}
	}
	return rec
}

func TestHandleTasks_WorktreeOverlay_AdditiveFields(t *testing.T) {
	dp, _, _ := overlayFixtureDP(t)

	var rows []map[string]any
	getJSON(t, handleTasks(dp), "/api/tasks", &rows)

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (2 local + 1 sibling-only), got %d", len(rows))
	}
	byID := map[string]map[string]any{}
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}
	claimed := byID["001"]
	if claimed["status"] != "pending" || claimed["effective_status"] != "in-progress" {
		t.Errorf("001 = status %v effective %v, want pending/in-progress", claimed["status"], claimed["effective_status"])
	}
	if claimed["worktree"] != "agent-b" || claimed["branch"] != "dnc/001/parser" {
		t.Errorf("001 missing provenance: %v", claimed)
	}
	if sib := byID["099"]; sib["remote_only"] != true {
		t.Errorf("099 should be remote_only: %v", sib)
	}
	if free := byID["002"]; free["effective_status"] != "pending" || free["worktree"] != nil {
		t.Errorf("002 unexpected overlay fields: %v", free)
	}
}

func TestHandleTasks_OverlayInactive_ShapeUnchanged(t *testing.T) {
	dir := createTestTaskDir(t)
	dp := NewDataProvider(dir, false)

	var rows []map[string]any
	getJSON(t, handleTasks(dp), "/api/tasks", &rows)

	for _, row := range rows {
		if _, ok := row["effective_status"]; ok {
			t.Errorf("effective_status present with overlay inactive: %v", row)
		}
	}
}

func TestHandleTaskByID_WorktreeOverlay_ProvenanceAndCopies(t *testing.T) {
	dp, _, _ := overlayFixtureDP(t)

	var detail map[string]any
	getJSON(t, handleTaskByID(dp), "/api/tasks/001", nil)
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/001", nil)
	req.SetPathValue("id", "001")
	rec := httptest.NewRecorder()
	handleTaskByID(dp)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}

	if detail["status"] != "pending" || detail["effective_status"] != "in-progress" {
		t.Errorf("detail 001 = status %v effective %v", detail["status"], detail["effective_status"])
	}
	copies, ok := detail["worktrees"].([]any)
	if !ok || len(copies) != 2 {
		t.Errorf("detail 001 worktrees = %v, want 2 diverging copies", detail["worktrees"])
	}
}

func TestHandleTaskByID_WorktreeOverlay_RemoteOnlyResolves(t *testing.T) {
	dp, _, _ := overlayFixtureDP(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/099", nil)
	req.SetPathValue("id", "099")
	rec := httptest.NewRecorder()
	handleTaskByID(dp)(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("remote-only detail status = %d, want 200", rec.Code)
	}
	var detail map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatal(err)
	}
	if detail["id"] != "099" || detail["remote_only"] != true {
		t.Errorf("detail 099 = %v, want remote_only sibling task", detail)
	}
}

func TestHandleBoard_WorktreeOverlay_GroupsByEffectiveStatus(t *testing.T) {
	dp, _, _ := overlayFixtureDP(t)

	var board []struct {
		Group string           `json:"group"`
		Tasks []map[string]any `json:"tasks"`
	}
	getJSON(t, handleBoard(dp, nil, effortScaleForTest()), "/api/board?groupBy=status", &board)

	found := false
	for _, group := range board {
		for _, task := range group.Tasks {
			if task["id"] == "001" && group.Group == "in-progress" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("001 should appear under in-progress (effective status); board = %+v", board)
	}
}

func TestHandleNext_WorktreeOverlay_ExcludesSiblingClaims(t *testing.T) {
	dp, _, _ := overlayFixtureDP(t)

	var recs []map[string]any
	getJSON(t, handleNext(dp, effortScaleForTest()), "/api/next", &recs)

	if len(recs) == 0 {
		t.Fatal("expected a recommendation for 002")
	}
	for _, rec := range recs {
		if rec["id"] == "001" || rec["id"] == "099" {
			t.Errorf("next recommended %v, which is claimed/only in a sibling worktree", rec["id"])
		}
	}
}

func TestHandleUpdateTask_WorktreeOverlay_SiblingGuard(t *testing.T) {
	dp, _, sibling := overlayFixtureDP(t)

	body := strings.NewReader(`{"status": "completed"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/tasks/099", body)
	req.SetPathValue("id", "099")
	rec := httptest.NewRecorder()
	handleUpdateTask(dp, false, effortScaleForTest())(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 Conflict guard", rec.Code)
	}
	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Error, "exists only in worktree") || !strings.Contains(resp.Error, "run taskmd there") {
		t.Errorf("guard error = %q", resp.Error)
	}

	// The sibling's file must never be written.
	data, err := os.ReadFile(filepath.Join(sibling.TasksDir, "099-sibling.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "completed") {
		t.Error("sibling worktree file was modified by the web mutation")
	}
}

func TestHandleValidate_WorktreeOverlay_DivergenceWarning(t *testing.T) {
	localDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(localDir, "001-a.md"),
		[]byte(overlayTaskMD("001", "Alpha", "cancelled")), 0o644); err != nil {
		t.Fatal(err)
	}
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"001-a.md": overlayTaskMD("001", "Alpha", "completed"),
	})
	dp := NewDataProviderWithWorktrees(localDir, false, overlayBuilder(sibling))

	var result struct {
		Issues []struct {
			Level   string `json:"level"`
			Message string `json:"message"`
		} `json:"issues"`
	}
	getJSON(t, handleValidate(dp, effortScaleForTest()), "/api/validate", &result)

	found := false
	for _, issue := range result.Issues {
		if issue.Level == "warning" && strings.Contains(issue.Message, "completed in worktree agent-b") {
			found = true
		}
	}
	if !found {
		t.Errorf("validate should warn on divergent terminal states, got %+v", result.Issues)
	}
}

func TestDataProvider_Overlay_InvalidationRefreshesSiblingState(t *testing.T) {
	dp, _, sibling := overlayFixtureDP(t)

	tasks, err := dp.GetEffectiveTasks()
	if err != nil {
		t.Fatal(err)
	}
	if status := statusOf(tasks, "002"); status != "pending" {
		t.Fatalf("002 = %s, want pending before the sibling claim", status)
	}

	// Claim 002 in the sibling worktree; the cache must not see it until
	// invalidated, and must see it after.
	claim := overlayTaskMD("002", "Free", "in-progress")
	if err := os.WriteFile(filepath.Join(sibling.TasksDir, "002-free.md"), []byte(claim), 0o644); err != nil {
		t.Fatal(err)
	}

	tasks, _ = dp.GetEffectiveTasks()
	if status := statusOf(tasks, "002"); status != "pending" {
		t.Fatalf("002 = %s, cache should serve the pre-claim state until invalidated", status)
	}

	dp.Invalidate()
	tasks, _ = dp.GetEffectiveTasks()
	if status := statusOf(tasks, "002"); status != "in-progress" {
		t.Fatalf("002 = %s after invalidation, want in-progress from the sibling claim", status)
	}
}

func TestLiveRefresh_SiblingEditInvalidatesAndBroadcasts(t *testing.T) {
	dp, _, sibling := overlayFixtureDP(t)
	broker := NewSSEBroker()

	// Mirror NewServer's wiring: change → invalidate + broadcast + re-sync.
	var w *watcher.Watcher
	w = watcher.New(dp.ScanDir(), func() {
		dp.Invalidate()
		broker.Broadcast()
		w.SetDirs(dp.WatchDirs(), dp.WatchMetaDirs())
	}, 30*time.Millisecond)
	w.SetDirs(dp.WatchDirs(), dp.WatchMetaDirs())

	go func() { _ = w.Start() }()
	defer w.Stop()
	time.Sleep(100 * time.Millisecond) // let the watcher register its dirs

	// Subscribe a fake SSE client.
	ch := make(chan struct{}, 1)
	broker.mu.Lock()
	broker.clients[ch] = struct{}{}
	broker.mu.Unlock()

	// Prime the cache, then claim a task in the sibling worktree.
	if _, err := dp.GetEffectiveTasks(); err != nil {
		t.Fatal(err)
	}
	claim := overlayTaskMD("002", "Free", "in-progress")
	if err := os.WriteFile(filepath.Join(sibling.TasksDir, "002-free.md"), []byte(claim), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ch:
		// SSE broadcast reached the client.
	case <-time.After(3 * time.Second):
		t.Fatal("no SSE broadcast after editing a task in a sibling worktree")
	}

	tasks, err := dp.GetEffectiveTasks()
	if err != nil {
		t.Fatal(err)
	}
	if status := statusOf(tasks, "002"); status != "in-progress" {
		t.Fatalf("002 = %s in the next payload, want the sibling claim", status)
	}
}

func TestDataProvider_WatchDirs_CoverSiblings(t *testing.T) {
	dp, localDir, sibling := overlayFixtureDP(t)

	dirs := dp.WatchDirs()
	if len(dirs) != 2 || dirs[0] != localDir || dirs[1] != sibling.TasksDir {
		t.Errorf("WatchDirs = %v, want [%s %s]", dirs, localDir, sibling.TasksDir)
	}

	// Overlay disabled: only the scan dir.
	plain := NewDataProvider(localDir, false)
	if dirs := plain.WatchDirs(); len(dirs) != 1 || dirs[0] != localDir {
		t.Errorf("WatchDirs (disabled) = %v, want only the scan dir", dirs)
	}
}

func TestExport_WorktreeOverlay_BakesEffectiveState(t *testing.T) {
	dp, localDir, sibling := overlayFixtureDP(t)
	_ = dp
	outDir := filepath.Join(t.TempDir(), "export")

	err := exportWithMockFS(t, ExportConfig{
		OutputDir: outDir,
		ScanDir:   localDir,
		BasePath:  "/",
		Version:   "test",
		Worktrees: overlayBuilder(sibling),
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(outDir, "api", "tasks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("tasks.json rows = %d, want 3 (including sibling-only)", len(rows))
	}
	byID := map[string]map[string]any{}
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}
	if byID["001"]["effective_status"] != "in-progress" || byID["001"]["worktree"] != "agent-b" {
		t.Errorf("tasks.json 001 missing baked provenance: %v", byID["001"])
	}

	// Board bakes effective status: 001 must sit under in-progress.
	rawBoard, err := os.ReadFile(filepath.Join(outDir, "api", "board", "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rawBoard), "in-progress") {
		t.Errorf("board/status.json should contain an in-progress group:\n%s", rawBoard)
	}

	// The sibling-only task gets a detail file and SPA route.
	if _, err := os.Stat(filepath.Join(outDir, "api", "tasks", "099.json")); err != nil {
		t.Errorf("missing detail file for sibling-only task: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "tasks", "099", "index.html")); err != nil {
		t.Errorf("missing SPA fallback for sibling-only task: %v", err)
	}
}

// statusOf returns the status of the task with the given id, or "".
func statusOf(tasks []*model.Task, id string) string {
	for _, t := range tasks {
		if t.ID == id {
			return string(t.Status)
		}
	}
	return ""
}
