package gitmeta

import (
	"os"
	"path/filepath"
	"testing"
)

const porcelainFixture = `worktree /repo
HEAD 62004dbd51f6f52ba99a6b98b6b09f312bc10e10
branch refs/heads/main

worktree /checkouts/agent-b
HEAD c272559a2b3fd9d61f5eb843c68d9c1e2ab7c4b1
branch refs/heads/dnc/042/parser

worktree /checkouts/detached
HEAD c92f235bfabd51e6c9256b423ac0a8f00cf6dd08
detached

worktree /checkouts/gone
HEAD a376555fabd51e6c9256b423ac0a8f00cf6dd08a
branch refs/heads/gone
prunable gitdir file points to non-existent location

worktree /repos/bare.git
bare
`

func TestParsePorcelain_Fixture(t *testing.T) {
	entries := parsePorcelain(porcelainFixture)
	if len(entries) != 5 {
		t.Fatalf("parsed %d entries, want 5: %+v", len(entries), entries)
	}

	want := []porcelainEntry{
		{root: "/repo", branch: "main"},
		{root: "/checkouts/agent-b", branch: "dnc/042/parser"},
		{root: "/checkouts/detached", detached: true},
		{root: "/checkouts/gone", branch: "gone", prunable: true},
		{root: "/repos/bare.git", bare: true},
	}
	for i, w := range want {
		if entries[i] != w {
			t.Errorf("entry %d = %+v, want %+v", i, entries[i], w)
		}
	}
}

func TestParsePorcelain_CRLFAndEmpty(t *testing.T) {
	if entries := parsePorcelain(""); len(entries) != 0 {
		t.Errorf("empty input parsed to %d entries, want 0", len(entries))
	}

	entries := parsePorcelain("worktree /repo\r\nHEAD abc\r\nbranch refs/heads/main\r\n")
	if len(entries) != 1 || entries[0].root != "/repo" || entries[0].branch != "main" {
		t.Errorf("CRLF input parsed to %+v", entries)
	}
}

func TestParsePorcelain_AttributeLineOutsideStanza(t *testing.T) {
	entries := parsePorcelain("branch refs/heads/stray\n\nworktree /repo\nbranch refs/heads/main\n")
	if len(entries) != 1 || entries[0].root != "/repo" || entries[0].branch != "main" {
		t.Errorf("parsed %+v, want single /repo entry on branch main", entries)
	}
}

// makeWorktreeDir creates a directory that looks like a taskmd-enabled
// worktree root, returning its path.
func makeWorktreeDir(t *testing.T, config string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, configFilename), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestFilterWorktrees(t *testing.T) {
	local := makeWorktreeDir(t, "dir: ./tasks\n")
	sibling := makeWorktreeDir(t, "dir: ./tasks\n")
	optedOut := t.TempDir() // exists but has no .taskmd.yaml

	entries := []porcelainEntry{
		{root: local, branch: "main"},
		{root: sibling, branch: "agent-b"},
		{root: optedOut, branch: "no-config"},
		{root: filepath.Join(local, "does-not-exist"), branch: "missing"},
		{root: sibling, branch: "prunable-copy", prunable: true},
		{root: sibling, branch: "bare-copy", bare: true},
	}

	got := filterWorktrees(entries, local)
	if len(got) != 2 {
		t.Fatalf("filtered to %d worktrees, want 2: %+v", len(got), got)
	}

	if !got[0].IsLocal || got[0].Root != local || got[0].Branch != "main" {
		t.Errorf("local worktree = %+v", got[0])
	}
	if got[1].IsLocal || got[1].Root != sibling || got[1].Branch != "agent-b" {
		t.Errorf("sibling worktree = %+v", got[1])
	}
	for _, w := range got {
		if w.TasksDir != filepath.Join(w.Root, "tasks") {
			t.Errorf("TasksDir = %q, want %q", w.TasksDir, filepath.Join(w.Root, "tasks"))
		}
	}
}

func TestFilterWorktrees_RootIsFileNotDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	got := filterWorktrees([]porcelainEntry{{root: file}}, dir)
	if len(got) != 0 {
		t.Errorf("file root passed the filter: %+v", got)
	}
}

func TestListWorktrees_NilIdentity(t *testing.T) {
	worktrees, err := ListWorktrees(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if worktrees != nil {
		t.Errorf("ListWorktrees(nil) = %+v, want nil", worktrees)
	}
}
