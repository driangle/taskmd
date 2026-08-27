//go:build e2e

package web

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/watcher"
	"github.com/driangle/taskmd/apps/cli/internal/worktree"
)

// These tests drive real `git worktree add` / `git worktree remove` against a
// real repository. The watcher unit tests simulate membership changes with
// os.Mkdir; this exercises the layout git actually produces, and asserts the
// served payload — not just that onChange fired — follows along.

// requireGit skips when git is unavailable.
func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
}

// runGitWeb runs a git command in dir, failing the test on error.
func runGitWeb(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=e2e", "GIT_AUTHOR_EMAIL=e2e@test",
		"GIT_COMMITTER_NAME=e2e", "GIT_COMMITTER_EMAIL=e2e@test",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// initGitRepoWithTasks creates a repo whose initial commit carries a
// .taskmd.yaml and tasks/, so every worktree checked out from it participates
// in the overlay. Returns the symlink-resolved root.
func initGitRepoWithTasks(t *testing.T, files map[string]string) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".taskmd.yaml"), []byte("dir: ./tasks\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tasksDir := filepath.Join(root, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tasksDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runGitWeb(t, root, "init", "--initial-branch=main")
	runGitWeb(t, root, "add", ".")
	runGitWeb(t, root, "commit", "-m", "init")
	return root
}

// waitFor polls cond until it holds or the timeout expires. Because the data
// provider only rescans after the watcher invalidates it, a condition that
// depends on fresh data can only become true if live refresh actually ran.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// liveRefreshHarness wires a DataProvider, SSE broker and watcher exactly the
// way NewServer does, and returns the provider plus a channel that receives
// one token per broadcast.
func liveRefreshHarness(t *testing.T, scanDir string) (*DataProvider, chan struct{}) {
	t.Helper()
	dp := NewDataProviderWithWorktrees(scanDir, false, worktree.Builder{Enabled: true})
	broker := NewSSEBroker()

	var w *watcher.Watcher
	w = watcher.New(scanDir, func() {
		dp.Invalidate()
		broker.Broadcast()
		w.SetDirs(dp.WatchDirs(), dp.WatchMetaDirs())
	}, 30*time.Millisecond)
	w.SetDirs(dp.WatchDirs(), dp.WatchMetaDirs())

	go func() { _ = w.Start() }()
	t.Cleanup(w.Stop)
	time.Sleep(100 * time.Millisecond) // let the watcher register its dirs

	broadcasts := make(chan struct{}, 64)
	broker.mu.Lock()
	broker.clients[broadcasts] = struct{}{}
	broker.mu.Unlock()

	return dp, broadcasts
}

// tasksByID fetches /api/tasks through the real handler and indexes the rows.
func tasksByID(t *testing.T, dp *DataProvider) map[string]map[string]any {
	t.Helper()
	var rows []map[string]any
	getJSON(t, handleTasks(dp), "/api/tasks", &rows)
	byID := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		byID[row["id"].(string)] = row
	}
	return byID
}

// siblingRoots returns the sibling worktree roots the provider currently
// enumerates, derived from its watch dirs (scan dir first, siblings after).
func siblingRoots(t *testing.T, dp *DataProvider) []string {
	t.Helper()
	dirs := dp.WatchDirs()
	if len(dirs) == 0 {
		t.Fatal("WatchDirs returned nothing, expected at least the scan dir")
	}
	return dirs[1:]
}

func TestLiveRefresh_RealWorktreeAddRemoveReenumerates(t *testing.T) {
	requireGit(t)
	repo := initGitRepoWithTasks(t, map[string]string{
		"001-first.md":  overlayTaskMD("001", "First", "pending"),
		"002-second.md": overlayTaskMD("002", "Second", "pending"),
	})
	scanDir := filepath.Join(repo, "tasks")
	dp, broadcasts := liveRefreshHarness(t, scanDir)

	// Baseline: no siblings, no overlay fields on the served payload.
	if roots := siblingRoots(t, dp); len(roots) != 0 {
		t.Fatalf("expected no siblings before `git worktree add`, got %v", roots)
	}
	if row := tasksByID(t, dp)["001"]; row["effective_status"] != nil {
		t.Fatalf("overlay fields present with a single worktree: %v", row)
	}

	// --- git worktree add -------------------------------------------------
	sibling := filepath.Join(filepath.Dir(repo), "agent-b")
	runGitWeb(t, repo, "worktree", "add", sibling, "-b", "agent-b")
	resolvedSibling, err := filepath.EvalSymlinks(sibling)
	if err != nil {
		t.Fatal(err)
	}
	wantSiblingTasks := filepath.Join(resolvedSibling, "tasks")

	waitFor(t, "the new worktree to be enumerated", func() bool {
		roots := siblingRoots(t, dp)
		return len(roots) == 1 && roots[0] == wantSiblingTasks
	})
	select {
	case <-broadcasts:
	default:
		t.Error("`git worktree add` produced no SSE broadcast")
	}

	// A claim in the brand-new worktree must reach the served payload.
	claim := overlayTaskMD("001", "First", "in-progress")
	if err := os.WriteFile(filepath.Join(wantSiblingTasks, "001-first.md"), []byte(claim), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the sibling claim in the served payload", func() bool {
		row := tasksByID(t, dp)["001"]
		return row["effective_status"] == "in-progress" && row["worktree"] == "agent-b"
	})

	// --- git worktree remove ----------------------------------------------
	drain(broadcasts)
	runGitWeb(t, repo, "worktree", "remove", "--force", resolvedSibling)

	waitFor(t, "the removed worktree to drop out of enumeration", func() bool {
		return len(siblingRoots(t, dp)) == 0
	})
	select {
	case <-broadcasts:
	default:
		t.Error("`git worktree remove` produced no SSE broadcast")
	}

	// The payload is back to the local-only shape: no overlay fields at all.
	row := tasksByID(t, dp)["001"]
	if row["status"] != "pending" {
		t.Errorf("001 status after removal = %v, want the local pending", row["status"])
	}
	if row["effective_status"] != nil || row["worktree"] != nil {
		t.Errorf("overlay fields survived `git worktree remove`: %v", row)
	}
}

func TestWatchMetaDirs_RealRepoAndDegradation(t *testing.T) {
	requireGit(t)
	repo := initGitRepoWithTasks(t, map[string]string{
		"001-first.md": overlayTaskMD("001", "First", "pending"),
	})
	scanDir := filepath.Join(repo, "tasks")

	// Enabled, inside a repo: the git common dir's worktrees registry.
	dp := NewDataProviderWithWorktrees(scanDir, false, worktree.Builder{Enabled: true})
	want := filepath.Join(repo, ".git", "worktrees")
	dirs := dp.WatchMetaDirs()
	if len(dirs) != 1 || dirs[0] != want {
		t.Errorf("WatchMetaDirs = %v, want [%s]", dirs, want)
	}

	// From a linked worktree: the same shared common dir, not the per-worktree
	// git dir — both checkouts watch one registry.
	linked := filepath.Join(filepath.Dir(repo), "agent-b")
	runGitWeb(t, repo, "worktree", "add", linked, "-b", "agent-b")
	resolvedLinked, err := filepath.EvalSymlinks(linked)
	if err != nil {
		t.Fatal(err)
	}
	linkedDP := NewDataProviderWithWorktrees(filepath.Join(resolvedLinked, "tasks"), false, worktree.Builder{Enabled: true})
	if dirs := linkedDP.WatchMetaDirs(); len(dirs) != 1 || dirs[0] != want {
		t.Errorf("WatchMetaDirs from linked worktree = %v, want [%s]", dirs, want)
	}

	// Overlay disabled: nothing to watch.
	if dirs := NewDataProvider(scanDir, false).WatchMetaDirs(); dirs != nil {
		t.Errorf("WatchMetaDirs (overlay disabled) = %v, want nil", dirs)
	}

	// Outside a git repo: nothing to watch.
	outside := NewDataProviderWithWorktrees(t.TempDir(), false, worktree.Builder{Enabled: true})
	if dirs := outside.WatchMetaDirs(); dirs != nil {
		t.Errorf("WatchMetaDirs (non-repo) = %v, want nil", dirs)
	}
}

// drain empties a buffered channel without blocking.
func drain(ch chan struct{}) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
