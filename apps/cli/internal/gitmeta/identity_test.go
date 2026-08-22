package gitmeta

import (
	"path/filepath"
	"testing"
)

func TestParseRevParse_PrimaryWorktree(t *testing.T) {
	out := "/repo/.git\n/repo/.git\n/repo\n"

	id, err := parseRevParse(out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.CommonDir != filepath.Clean("/repo/.git") {
		t.Errorf("CommonDir = %q, want /repo/.git", id.CommonDir)
	}
	if id.WorktreeRoot != filepath.Clean("/repo") {
		t.Errorf("WorktreeRoot = %q, want /repo", id.WorktreeRoot)
	}
	if id.IsLinked {
		t.Error("IsLinked = true for primary worktree, want false")
	}
}

func TestParseRevParse_LinkedWorktree(t *testing.T) {
	out := "/repo/.git\n/repo/.git/worktrees/agent-b\n/checkouts/agent-b\n"

	id, err := parseRevParse(out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !id.IsLinked {
		t.Error("IsLinked = false for linked worktree, want true")
	}
	if id.CommonDir != filepath.Clean("/repo/.git") {
		t.Errorf("CommonDir = %q, want /repo/.git", id.CommonDir)
	}
	if id.WorktreeRoot != filepath.Clean("/checkouts/agent-b") {
		t.Errorf("WorktreeRoot = %q, want /checkouts/agent-b", id.WorktreeRoot)
	}
}

func TestParseRevParse_MalformedOutput(t *testing.T) {
	for _, out := range []string{"", "\n", "/only/one/line\n", "/a\n/b\n"} {
		if _, err := parseRevParse(out); err == nil {
			t.Errorf("parseRevParse(%q) returned nil error, want error", out)
		}
	}
}

func TestResolve_NonGitDirectory(t *testing.T) {
	id, err := Resolve(t.TempDir())
	if err != nil {
		t.Fatalf("Resolve on non-git dir returned error %v, want nil", err)
	}
	if id != nil {
		t.Errorf("Resolve on non-git dir returned %+v, want nil identity", id)
	}
}

func TestResolve_GitMissingFromPath(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	id, err := Resolve(t.TempDir())
	if err != nil {
		t.Fatalf("Resolve without git returned error %v, want nil", err)
	}
	if id != nil {
		t.Errorf("Resolve without git returned %+v, want nil identity", id)
	}
}
