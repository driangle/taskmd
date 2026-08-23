package cli

import (
	"os"
	"path/filepath"

	"github.com/driangle/taskmd/apps/cli/internal/gitmeta"
)

// Repo-identity resolution for the projects registry (ADR 0005 §1): all
// worktrees of one repository resolve to a single registered project. Project
// identity is the git common directory, never the checkout path.

// resolveRepoIdentity is a seam over gitmeta.Resolve so registry tests can
// inject identities without git. Returns nil when dir is not in a git repo.
var resolveRepoIdentity = func(dir string) *gitmeta.Identity {
	id, _ := gitmeta.Resolve(dir)
	return id
}

// sameRepo reports whether two identities refer to the same repository.
func sameRepo(a, b *gitmeta.Identity) bool {
	return a != nil && b != nil && a.CommonDir == b.CommonDir
}

// primaryWorktreeRoot derives the primary worktree's root from the repo's
// common dir (<primary>/.git). Returns "" when there is no primary worktree
// to derive (bare repos, or no identity).
func primaryWorktreeRoot(id *gitmeta.Identity) string {
	if id == nil || filepath.Base(id.CommonDir) != ".git" {
		return ""
	}
	return filepath.Dir(id.CommonDir)
}

// canonicalRegisterPath maps a register target inside a linked worktree to
// the corresponding directory under the primary worktree, preserving any
// subpath below the worktree root (monorepo subprojects). Targets outside a
// git repo, in bare repos, or already in the primary pass through unchanged.
func canonicalRegisterPath(target string) string {
	id := resolveRepoIdentity(target)
	primary := primaryWorktreeRoot(id)
	if primary == "" {
		return target
	}
	rel, err := filepath.Rel(id.WorktreeRoot, target)
	if err != nil {
		return target
	}
	return filepath.Join(primary, rel)
}

// localWorktreePathFor returns the current worktree's directory corresponding
// to registeredPath when the cwd is inside a different worktree of the same
// repository, else "". This is what makes --project <id> scope to the local
// checkout instead of the registered primary.
func localWorktreePathFor(registeredPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	cwdID := resolveRepoIdentity(cwd)
	if cwdID == nil {
		return ""
	}
	regID := resolveRepoIdentity(registeredPath)
	if !sameRepo(cwdID, regID) || cwdID.WorktreeRoot == regID.WorktreeRoot {
		return ""
	}
	return remapAcrossWorktrees(registeredPath, regID.WorktreeRoot, cwdID.WorktreeRoot)
}

// dedupeRepoEntries collapses registry entries that resolve to the same
// repository (keeping the first) and retargets an entry to the current
// worktree when the command runs inside that repo. Entries outside git repos
// pass through untouched.
func dedupeRepoEntries(entries []GlobalProjectEntry) []GlobalProjectEntry {
	var cwdID *gitmeta.Identity
	if cwd, err := os.Getwd(); err == nil {
		cwdID = resolveRepoIdentity(cwd)
	}

	seen := make(map[string]string, len(entries))
	deduped := make([]GlobalProjectEntry, 0, len(entries))
	for _, entry := range entries {
		id := resolveRepoIdentity(entry.Path)
		if id == nil {
			deduped = append(deduped, entry)
			continue
		}
		if firstID, ok := seen[id.CommonDir]; ok {
			debugLog("all-projects: skipping %q (same repository as %q)", entry.ID, firstID)
			continue
		}
		seen[id.CommonDir] = entry.ID
		if sameRepo(id, cwdID) && id.WorktreeRoot != cwdID.WorktreeRoot {
			if local := remapAcrossWorktrees(entry.Path, id.WorktreeRoot, cwdID.WorktreeRoot); local != "" {
				entry.Path = local
			}
		}
		deduped = append(deduped, entry)
	}
	return deduped
}

// remapAcrossWorktrees translates path from one worktree root to the same
// relative location under another. Returns "" when path is not under fromRoot.
func remapAcrossWorktrees(path, fromRoot, toRoot string) string {
	rel, err := filepath.Rel(fromRoot, path)
	if err != nil {
		return ""
	}
	return filepath.Join(toRoot, rel)
}
