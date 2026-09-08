package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
)

// writeTasksUnder creates dir and writes the given task files into it.
func writeTasksUnder(t *testing.T, dir string, files map[string]string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	for fname, content := range files {
		if err := os.WriteFile(filepath.Join(dir, fname), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", fname, err)
		}
	}
	return dir
}

// TestAttributeNestedSiblingCopies_AncestorSiblingKeepsLocalTasks covers the
// layout where the local checkout lives *under* the primary worktree's root
// (repo/.claude/worktrees/x). The primary is a sibling whose root contains
// every local task path, but it never scanned them — attribution must not drop
// them, or every read view inside the worktree reports zero tasks.
func TestAttributeNestedSiblingCopies_AncestorSiblingKeepsLocalTasks(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	primaryTasks := writeTasksUnder(t, filepath.Join(repo, "tasks"), map[string]string{
		"a.md": taskMD("a", "A", "pending"),
	})
	localTasksDir := writeTasksUnder(t,
		filepath.Join(repo, ".claude", "worktrees", "wt-a", "tasks"),
		map[string]string{
			"a.md": taskMD("a", "A", "pending"),
			"b.md": taskMD("b", "B", "pending"),
		})

	local := scanDirTasks(t, localTasksDir)
	if len(local) != 2 {
		t.Fatalf("fixture: scanned %d local tasks, want 2", len(local))
	}

	siblings := []gitmeta.Worktree{{Root: repo, Branch: "main", TasksDir: primaryTasks}}
	got := AttributeNestedSiblingCopies(local, localTasksDir, siblings)

	if len(got) != 2 {
		t.Fatalf("attributed %d local tasks, want 2 — an ancestor sibling must not "+
			"swallow the local task list", len(got))
	}
}

// TestAttributeNestedSiblingCopies_NestedSiblingDropsDoubleScannedCopies is the
// case attribution exists for (spec §8): a checkout nested inside the scan root
// is double-scanned, and those copies belong to that worktree.
func TestAttributeNestedSiblingCopies_NestedSiblingDropsDoubleScannedCopies(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	writeTasksUnder(t, repo, map[string]string{"a.md": taskMD("a", "A", "pending")})
	nestedRoot := filepath.Join(repo, "sub-wt")
	nestedTasks := writeTasksUnder(t, filepath.Join(nestedRoot, "tasks"), map[string]string{
		"a.md": taskMD("a", "A", "in-progress"),
	})

	local := scanDirTasks(t, repo)
	if len(local) != 2 {
		t.Fatalf("fixture: scanned %d local tasks, want 2 (the nested copy is double-scanned)", len(local))
	}

	siblings := []gitmeta.Worktree{{Root: nestedRoot, Branch: "agent-a", TasksDir: nestedTasks}}
	got := AttributeNestedSiblingCopies(local, repo, siblings)

	if len(got) != 1 {
		t.Fatalf("attributed %d local tasks, want 1", len(got))
	}
	if dir := filepath.Dir(got[0].FilePath); dir != filepath.Clean(repo) {
		t.Errorf("kept copy from %s, want the one directly under the scan root %s", dir, repo)
	}
}

// TestAttributeNestedSiblingCopies_UnrelatedSiblingIsNoOp keeps the common
// layout honest: siblings parked beside the repo share no path prefix, so
// nothing is attributed away.
func TestAttributeNestedSiblingCopies_UnrelatedSiblingIsNoOp(t *testing.T) {
	t.Parallel()
	localDir := writeTaskDir(t, map[string]string{"a.md": taskMD("a", "A", "pending")})
	local := scanDirTasks(t, localDir)
	sibling := newSiblingWorktree(t, "agent-b", "b", map[string]string{
		"a.md": taskMD("a", "A", "completed"),
	})

	if got := AttributeNestedSiblingCopies(local, localDir, []gitmeta.Worktree{sibling}); len(got) != 1 {
		t.Fatalf("attributed %d local tasks, want 1", len(got))
	}
}
