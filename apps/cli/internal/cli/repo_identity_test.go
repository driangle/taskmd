package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
)

// gitIdentity builds an Identity for a worktree of the repo whose primary
// checkout is at primary.
func gitIdentity(primary, worktreeRoot string) *gitmeta.Identity {
	return &gitmeta.Identity{
		CommonDir:    filepath.Join(primary, ".git"),
		WorktreeRoot: worktreeRoot,
		IsLinked:     primary != worktreeRoot,
	}
}

// stubRepoIdentities installs a resolveRepoIdentity stub that answers by
// containment: a path maps to the identity of the worktree root that contains
// it. resetCLIState restores the default "not a git repo" stub afterwards.
func stubRepoIdentities(worktrees map[string]*gitmeta.Identity) {
	resolveRepoIdentity = func(dir string) *gitmeta.Identity {
		for root, id := range worktrees {
			if dir == root || strings.HasPrefix(dir, root+string(filepath.Separator)) {
				return id
			}
		}
		return nil
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
}

func TestCanonicalRegisterPath(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := filepath.Join(t.TempDir(), "repo")
	worktree := filepath.Join(t.TempDir(), "repo-wt")
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary:  gitIdentity(primary, primary),
		worktree: gitIdentity(primary, worktree),
	})

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{"non-git passthrough", "/not/a/repo", "/not/a/repo"},
		{"primary root unchanged", primary, primary},
		{"primary subdir unchanged", filepath.Join(primary, "apps", "cli"), filepath.Join(primary, "apps", "cli")},
		{"linked worktree root maps to primary", worktree, primary},
		{"linked worktree subdir keeps subpath", filepath.Join(worktree, "apps", "cli"), filepath.Join(primary, "apps", "cli")},
	}
	for _, tt := range tests {
		if got := canonicalRegisterPath(tt.target); got != tt.want {
			t.Errorf("%s: canonicalRegisterPath(%q) = %q, want %q", tt.name, tt.target, got, tt.want)
		}
	}
}

func TestCanonicalRegisterPath_BareRepoPassthrough(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	worktree := filepath.Join(t.TempDir(), "wt")
	resolveRepoIdentity = func(string) *gitmeta.Identity {
		return &gitmeta.Identity{CommonDir: "/srv/repo.git", WorktreeRoot: worktree, IsLinked: true}
	}

	if got := canonicalRegisterPath(worktree); got != worktree {
		t.Errorf("canonicalRegisterPath(%q) = %q, want passthrough", worktree, got)
	}
}

func TestProjectRegister_LinkedWorktreeStoresPrimary(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := createProjectDir(t)
	worktree := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary:  gitIdentity(primary, primary),
		worktree: gitIdentity(primary, worktree),
	})

	projectRegisterPath = worktree
	if err := runProjectRegister(projectRegisterCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != primary {
		t.Errorf("expected primary path %q, got %q", primary, entries[0].Path)
	}
	if want := filepath.Base(primary); entries[0].ID != want {
		t.Errorf("expected ID %q (primary basename), got %q", want, entries[0].ID)
	}
}

func TestProjectRegister_SecondWorktreeIsNoOp(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := createProjectDir(t)
	worktree := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary:  gitIdentity(primary, primary),
		worktree: gitIdentity(primary, worktree),
	})

	projectRegisterPath = primary
	projectRegisterID = "myrepo"
	if err := runProjectRegister(projectRegisterCmd, nil); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Registering a second worktree of the same repo must not add an entry,
	// even under a different ID.
	projectRegisterPath = worktree
	projectRegisterID = "other-id"
	stdout, _ := captureOutput(t, func() {
		if err := runProjectRegister(projectRegisterCmd, nil); err != nil {
			t.Errorf("duplicate register should be a no-op, got error: %v", err)
		}
	})
	if !strings.Contains(stdout, `Already registered as "myrepo"`) {
		t.Errorf("expected friendly no-op message, got: %q", stdout)
	}

	entries, err := LoadGlobalRegistry()
	if err != nil {
		t.Fatalf("failed to load registry: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after duplicate register, got %d", len(entries))
	}
}

func TestProjectRegister_NonGitDuplicatePathBehavesAsToday(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	projDir := createProjectDir(t)
	globalCfg := filepath.Join(t.TempDir(), ".taskmd.yaml")
	t.Setenv("TASKMD_HOME_CONFIG", globalCfg)
	// Default stub: not a git repo.

	projectRegisterPath = projDir
	if err := runProjectRegister(projectRegisterCmd, nil); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	// Same path, same derived ID: still the ID-duplicate error, not a no-op.
	projectRegisterPath = projDir
	err := runProjectRegister(projectRegisterCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected duplicate-ID error, got: %v", err)
	}
}

func TestResolveProjectDir_ScopesToCurrentWorktree(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := createProjectDir(t)
	worktree := createProjectDir(t)
	setupProjectFlagRegistry(t, "  - id: proj\n    path: "+primary+"\n")
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary:  gitIdentity(primary, primary),
		worktree: gitIdentity(primary, worktree),
	})

	chdir(t, worktree)
	dir, err := resolveProjectDir("proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != worktree {
		t.Errorf("expected current worktree dir %q, got %q", worktree, dir)
	}
}

func TestResolveProjectDir_OutsideRepoUsesRegisteredPath(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := createProjectDir(t)
	setupProjectFlagRegistry(t, "  - id: proj\n    path: "+primary+"\n")
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary: gitIdentity(primary, primary),
	})

	elsewhere, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	chdir(t, elsewhere)
	dir, err := resolveProjectDir("proj")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != primary {
		t.Errorf("expected registered path %q, got %q", primary, dir)
	}
}

func TestDedupeRepoEntries_CollapsesSameRepo(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := createProjectDir(t)
	worktree := createProjectDir(t)
	nonGit := createProjectDir(t)
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary:  gitIdentity(primary, primary),
		worktree: gitIdentity(primary, worktree),
	})

	entries := []GlobalProjectEntry{
		{ID: "a", Path: primary},
		{ID: "b", Path: worktree}, // legacy entry: same repo, different worktree
		{ID: "c", Path: nonGit},
	}

	elsewhere, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	chdir(t, elsewhere)
	deduped := dedupeRepoEntries(entries)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 entries after dedupe, got %d: %+v", len(deduped), deduped)
	}
	if deduped[0].ID != "a" || deduped[0].Path != primary {
		t.Errorf("expected first entry a@%q, got %+v", primary, deduped[0])
	}
	if deduped[1].ID != "c" {
		t.Errorf("expected non-git entry c kept, got %+v", deduped[1])
	}
}

func TestDedupeRepoEntries_RetargetsToCurrentWorktree(t *testing.T) {
	resetCLIState()
	defer resetCLIState()

	primary := createProjectDir(t)
	worktree := createProjectDir(t)
	stubRepoIdentities(map[string]*gitmeta.Identity{
		primary:  gitIdentity(primary, primary),
		worktree: gitIdentity(primary, worktree),
	})

	chdir(t, worktree)
	deduped := dedupeRepoEntries([]GlobalProjectEntry{{ID: "a", Path: primary}})
	if len(deduped) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(deduped))
	}
	if deduped[0].Path != worktree {
		t.Errorf("expected entry retargeted to current worktree %q, got %q", worktree, deduped[0].Path)
	}
}
