package gitmeta

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree is one checkout of the repository that participates in taskmd
// coordination (it exists on disk and has a .taskmd.yaml at its root).
type Worktree struct {
	Root     string // worktree top-level path
	Branch   string // e.g. "dnc/042/parser" ("" when detached)
	IsLocal  bool   // true for the worktree the command runs in
	TasksDir string // absolute tasks directory, from the worktree's own .taskmd.yaml
}

// ListWorktrees enumerates the repository's worktrees via
// `git worktree list --porcelain`, filtered to checkouts taskmd can use:
// prunable worktrees, roots missing on disk, and worktrees without a
// .taskmd.yaml are skipped. A nil identity yields no worktrees.
func ListWorktrees(id *Identity) ([]Worktree, error) {
	if id == nil {
		return nil, nil
	}

	cmd := exec.Command("git", "-C", id.WorktreeRoot, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	return filterWorktrees(parsePorcelain(string(out)), id.WorktreeRoot), nil
}

// porcelainEntry is one raw stanza of `git worktree list --porcelain`.
type porcelainEntry struct {
	root     string
	branch   string
	bare     bool
	detached bool
	prunable bool
}

// parsePorcelain parses `git worktree list --porcelain` output. Stanzas are
// separated by blank lines; each starts with a "worktree <path>" line
// followed by attribute lines.
func parsePorcelain(out string) []porcelainEntry {
	var entries []porcelainEntry
	var cur *porcelainEntry

	flush := func() {
		if cur != nil {
			entries = append(entries, *cur)
			cur = nil
		}
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "worktree "):
			flush()
			cur = &porcelainEntry{root: strings.TrimPrefix(line, "worktree ")}
		case cur == nil:
			// Attribute line outside a stanza; ignore.
		case strings.HasPrefix(line, "branch "):
			cur.branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case line == "bare":
			cur.bare = true
		case line == "detached":
			cur.detached = true
		case line == "prunable" || strings.HasPrefix(line, "prunable "):
			cur.prunable = true
		}
	}
	flush()
	return entries
}

// filterWorktrees turns raw porcelain entries into usable Worktrees, dropping
// bare and prunable entries, roots missing on disk, and worktrees that have
// no .taskmd.yaml at their root (they opted out of taskmd or predate it).
func filterWorktrees(entries []porcelainEntry, localRoot string) []Worktree {
	var worktrees []Worktree
	for _, e := range entries {
		if e.bare || e.prunable {
			Debugf("gitmeta: skipping worktree %s (bare or prunable)", e.root)
			continue
		}
		root := filepath.Clean(e.root)
		if info, err := os.Stat(root); err != nil || !info.IsDir() {
			Debugf("gitmeta: skipping worktree %s (root missing on disk)", root)
			continue
		}
		if _, err := os.Stat(filepath.Join(root, configFilename)); err != nil {
			Warnf("gitmeta: skipping worktree %s (no %s)", root, configFilename)
			continue
		}
		worktrees = append(worktrees, Worktree{
			Root:     root,
			Branch:   e.branch,
			IsLocal:  root == filepath.Clean(localRoot),
			TasksDir: resolveTasksDir(root),
		})
	}
	return worktrees
}
