//go:build e2e

package gitmeta

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runGit runs a git command in dir, failing the test on error.
func runGit(t *testing.T, dir string, args ...string) {
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

// initRepo creates a git repository whose initial commit contains a
// .taskmd.yaml and a tasks dir, returning its symlink-resolved root.
func initRepo(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".taskmd.yaml"), []byte("dir: ./tasks\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Git tracks files, not directories — .gitkeep makes tasks/ exist in
	// every worktree checked out from this commit.
	if err := os.MkdirAll(filepath.Join(root, "tasks"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tasks", ".gitkeep"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init", "--initial-branch=main")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "init")
	return root
}

// addWorktree creates a linked worktree of repo on a new branch and returns
// its symlink-resolved root.
func addWorktree(t *testing.T, repo, name string) string {
	t.Helper()
	path := filepath.Join(filepath.Dir(repo), name)
	runGit(t, repo, "worktree", "add", path, "-b", name)
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func TestE2E_Resolve_PrimaryVsLinked(t *testing.T) {
	repo := initRepo(t)
	linked := addWorktree(t, repo, "agent-b")

	primary, err := Resolve(repo)
	if err != nil || primary == nil {
		t.Fatalf("Resolve(primary) = %+v, %v", primary, err)
	}
	if primary.IsLinked {
		t.Error("primary worktree reported IsLinked = true")
	}
	if primary.WorktreeRoot != repo {
		t.Errorf("primary WorktreeRoot = %q, want %q", primary.WorktreeRoot, repo)
	}

	// Resolve from a subdirectory of the linked worktree.
	sub, err := Resolve(filepath.Join(linked, "tasks"))
	if err != nil || sub == nil {
		t.Fatalf("Resolve(linked/tasks) = %+v, %v", sub, err)
	}
	if !sub.IsLinked {
		t.Error("linked worktree reported IsLinked = false")
	}
	if sub.WorktreeRoot != linked {
		t.Errorf("linked WorktreeRoot = %q, want %q", sub.WorktreeRoot, linked)
	}
	if sub.CommonDir != primary.CommonDir {
		t.Errorf("CommonDir differs across worktrees: %q vs %q", sub.CommonDir, primary.CommonDir)
	}
}

func TestE2E_ListWorktrees_DiscoveryAndFiltering(t *testing.T) {
	repo := initRepo(t)
	sibling := addWorktree(t, repo, "agent-b")
	optedOut := addWorktree(t, repo, "opted-out")
	deleted := addWorktree(t, repo, "deleted")

	// opted-out drops its config; deleted disappears from disk entirely.
	if err := os.Remove(filepath.Join(optedOut, ".taskmd.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(deleted); err != nil {
		t.Fatal(err)
	}

	id, err := Resolve(sibling)
	if err != nil || id == nil {
		t.Fatalf("Resolve(sibling) = %+v, %v", id, err)
	}
	worktrees, err := ListWorktrees(id)
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}

	byRoot := map[string]Worktree{}
	for _, w := range worktrees {
		byRoot[w.Root] = w
	}
	if len(worktrees) != 2 {
		t.Fatalf("got %d worktrees, want 2 (primary + agent-b): %+v", len(worktrees), worktrees)
	}
	if _, ok := byRoot[optedOut]; ok {
		t.Error("worktree without .taskmd.yaml appeared in results")
	}
	if _, ok := byRoot[deleted]; ok {
		t.Error("deleted worktree appeared in results")
	}

	prim, ok := byRoot[repo]
	if !ok {
		t.Fatalf("primary %q missing from results", repo)
	}
	if prim.IsLocal || prim.Branch != "main" {
		t.Errorf("primary = %+v, want IsLocal=false Branch=main", prim)
	}

	sib, ok := byRoot[sibling]
	if !ok {
		t.Fatalf("sibling %q missing from results", sibling)
	}
	if !sib.IsLocal || sib.Branch != "agent-b" {
		t.Errorf("sibling = %+v, want IsLocal=true Branch=agent-b", sib)
	}
	if sib.TasksDir != filepath.Join(sibling, "tasks") {
		t.Errorf("sibling TasksDir = %q, want %q", sib.TasksDir, filepath.Join(sibling, "tasks"))
	}
}

func TestE2E_Degradation_NoRepoAndNoGit(t *testing.T) {
	if id, err := Resolve(t.TempDir()); id != nil || err != nil {
		t.Errorf("Resolve(non-repo) = %+v, %v, want nil, nil", id, err)
	}

	t.Setenv("PATH", t.TempDir())
	if id, err := Resolve(t.TempDir()); id != nil || err != nil {
		t.Errorf("Resolve without git = %+v, %v, want nil, nil", id, err)
	}
}
