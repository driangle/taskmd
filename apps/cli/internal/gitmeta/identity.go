// Package gitmeta is the single helper through which core learns about the
// git repository containing the task files (ADR 0004). It provides repo
// identity (the shared git common directory) and worktree discovery, always
// degrading to inert behavior when git is absent or the directory is not a
// repository.
package gitmeta

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Identity describes the repository containing a directory, if any.
// Two paths with the same CommonDir belong to the same repository.
type Identity struct {
	CommonDir    string // absolute path to the shared .git directory
	WorktreeRoot string // top level of the worktree containing the directory
	IsLinked     bool   // true when the directory is inside a linked worktree
}

// Resolve returns the repo identity for dir, backed by a single git
// invocation. It returns (nil, nil) when git is missing from PATH, dir is not
// inside a git repository, or git fails for any reason — callers treat a nil
// identity as "worktree features inert".
func Resolve(dir string) (*Identity, error) {
	if _, err := exec.LookPath("git"); err != nil {
		Debugf("gitmeta: git not found in PATH: %v", err)
		return nil, nil
	}

	cmd := exec.Command("git", "-C", dir, "rev-parse",
		"--path-format=absolute", "--git-common-dir", "--git-dir", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		Debugf("gitmeta: rev-parse failed in %s: %v", dir, err)
		return nil, nil
	}

	id, err := parseRevParse(string(out))
	if err != nil {
		Debugf("gitmeta: %v", err)
		return nil, nil
	}
	return id, nil
}

// parseRevParse builds an Identity from the three-line output of
// `git rev-parse --git-common-dir --git-dir --show-toplevel`.
func parseRevParse(out string) (*Identity, error) {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected rev-parse output: %q", out)
	}
	commonDir := filepath.Clean(lines[0])
	gitDir := filepath.Clean(lines[1])
	return &Identity{
		CommonDir:    commonDir,
		WorktreeRoot: filepath.Clean(lines[2]),
		IsLinked:     gitDir != commonDir,
	}, nil
}
