//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitWT runs a git command in dir, failing the test on error.
func gitWT(t *testing.T, dir string, args ...string) {
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

// initWorktreeRepo creates a git repo whose initial commit carries a
// .taskmd.yaml (dir: ./tasks) and the given task files, so every worktree
// checked out from it participates in the overlay. Returns the resolved root.
func initWorktreeRepo(t *testing.T, tasks map[string]taskSpec) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".taskmd.yaml"), []byte("dir: ./tasks\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for filename, spec := range tasks {
		writeTask(t, filepath.Join(root, "tasks"), filename, spec.id, spec.title, spec.status, spec.deps)
	}
	gitWT(t, root, "init", "--initial-branch=main")
	gitWT(t, root, "add", ".")
	gitWT(t, root, "commit", "-m", "init")
	return root
}

type taskSpec struct {
	id     string
	title  string
	status string
	deps   []string
}

// addLinkedWorktree checks out a linked worktree on a new branch named after
// name, returning its resolved root.
func addLinkedWorktree(t *testing.T, repo, name string) string {
	t.Helper()
	path := filepath.Join(filepath.Dir(repo), name)
	gitWT(t, repo, "worktree", "add", path, "-b", name)
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

// runWithEnv executes taskmd like run, with extra environment variables.
func runWithEnv(t *testing.T, dir string, extraEnv []string, args ...string) runResult {
	t.Helper()
	cmd := buildCmd(dir, args...)
	homeDir := t.TempDir()
	cmd.Env = append([]string{
		"HOME=" + homeDir,
		"NO_COLOR=1",
		"PATH=" + os.Getenv("PATH"),
	}, extraEnv...)
	return execCmd(t, cmd, args)
}

// nextIDs parses `next --format json` output into recommendation IDs.
func nextIDs(t *testing.T, res runResult) []string {
	t.Helper()
	var recs []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("parse next output: %v\n%s", err, res.Stdout)
	}
	ids := make([]string, len(recs))
	for i, r := range recs {
		ids[i] = r.ID
	}
	return ids
}

func TestWorktree_DoubleAssignmentPrevented(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md":  {id: "001", title: "First task", status: "pending"},
		"002-second.md": {id: "002", title: "Second task", status: "pending"},
	})
	agentB := addLinkedWorktree(t, repo, "agent-b")

	// Claim 001 in the primary worktree.
	mustRun(t, repo, "set", "001", "--status", "in-progress")

	// next in the sibling must not recommend 001.
	ids := nextIDs(t, mustRun(t, agentB, "next", "--format", "json"))
	for _, id := range ids {
		if id == "001" {
			t.Fatalf("next in sibling worktree recommended 001, which is in-progress in %s; ids = %v",
				filepath.Base(repo), ids)
		}
	}
	if len(ids) != 1 || ids[0] != "002" {
		t.Errorf("next ids = %v, want [002]", ids)
	}

	// list in the sibling shows 001 with effective status and provenance.
	listRes := mustRun(t, agentB, "list")
	var row001 string
	for _, line := range strings.Split(listRes.Stdout, "\n") {
		if strings.Contains(line, "001") {
			row001 = line
			break
		}
	}
	if !strings.Contains(row001, "in-progress") || !strings.Contains(row001, filepath.Base(repo)) {
		t.Errorf("list row for 001 should show in-progress with worktree %q, got: %q\nfull output:\n%s",
			filepath.Base(repo), row001, listRes.Stdout)
	}
}

func TestWorktree_SetGuardOnSiblingOnlyTask(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md": {id: "001", title: "First task", status: "pending"},
	})
	agentB := addLinkedWorktree(t, repo, "agent-b")

	// 099 exists only in agent-b (uncommitted, branch-local work).
	writeTask(t, filepath.Join(agentB, "tasks"), "099-sibling.md", "099", "Sibling only", "pending", nil)
	before, err := os.ReadFile(filepath.Join(agentB, "tasks", "099-sibling.md"))
	if err != nil {
		t.Fatal(err)
	}

	res := run(t, repo, "set", "099", "--status", "completed")
	if res.ExitCode == 0 {
		t.Fatalf("set on a sibling-only ID should fail\nstdout: %s", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "exists only in worktree") ||
		!strings.Contains(res.Stderr, "agent-b") ||
		!strings.Contains(res.Stderr, "run taskmd there") {
		t.Errorf("guard message missing worktree provenance: %q", res.Stderr)
	}

	after, err := os.ReadFile(filepath.Join(agentB, "tasks", "099-sibling.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("set modified a file outside the current worktree")
	}
}

func TestWorktree_OptOutRestoresLocalView(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md": {id: "001", title: "First task", status: "pending"},
	})
	agentB := addLinkedWorktree(t, repo, "agent-b")
	mustRun(t, repo, "set", "001", "--status", "in-progress")

	// Overlay active: 001 is suppressed in the sibling.
	if ids := nextIDs(t, mustRun(t, agentB, "next", "--format", "json")); len(ids) != 0 {
		t.Fatalf("overlay should suppress 001, got %v", ids)
	}

	// --worktrees=false restores the purely local view.
	ids := nextIDs(t, mustRun(t, agentB, "--worktrees", "false", "next", "--format", "json"))
	if len(ids) != 1 || ids[0] != "001" {
		t.Errorf("--worktrees=false: ids = %v, want [001]", ids)
	}

	// TASKMD_WORKTREES=false does the same.
	res := runWithEnv(t, agentB, []string{"TASKMD_WORKTREES=false"}, "next", "--format", "json")
	if res.ExitCode != 0 {
		t.Fatalf("next failed: %s", res.Stderr)
	}
	if ids := nextIDs(t, res); len(ids) != 1 || ids[0] != "001" {
		t.Errorf("TASKMD_WORKTREES=false: ids = %v, want [001]", ids)
	}
}

func TestWorktree_SingleWorktreeBehaviorUnchanged(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md":  {id: "001", title: "First task", status: "in-progress"},
		"002-second.md": {id: "002", title: "Second task", status: "pending", deps: []string{"001"}},
	})

	for _, args := range [][]string{
		{"next", "--format", "json"},
		{"list"},
		{"list", "--format", "json"},
	} {
		auto := mustRun(t, repo, args...)
		off := mustRun(t, repo, append([]string{"--worktrees", "false"}, args...)...)
		if auto.Stdout != off.Stdout {
			t.Errorf("taskmd %v: single-worktree output differs between auto and false:\n--- auto ---\n%s\n--- false ---\n%s",
				args, auto.Stdout, off.Stdout)
		}
	}
}
