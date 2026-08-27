//go:build e2e

package e2e

import (
	"strings"
	"testing"
)

// runWithoutGit executes the taskmd binary with PATH pointed at an empty
// directory, so nothing the CLI shells out to — git in particular — can be
// found. The binary itself is invoked by absolute path, so it still runs.
func runWithoutGit(t *testing.T, dir string, args ...string) runResult {
	t.Helper()
	cmd := buildCmd(dir, args...)
	cmd.Env = []string{
		"HOME=" + t.TempDir(),
		"NO_COLOR=1",
		"PATH=" + t.TempDir(),
	}
	return execCmd(t, cmd, args)
}

// assertCleanRun fails when the command did not exit 0 or wrote anything to
// stderr. A missing git binary must be invisible to the user: no warnings, no
// diagnostics, no non-zero exit.
func assertCleanRun(t *testing.T, res runResult, args []string) {
	t.Helper()
	if res.ExitCode != 0 {
		t.Fatalf("taskmd %v without git exited %d\nstdout: %s\nstderr: %s",
			args, res.ExitCode, res.Stdout, res.Stderr)
	}
	if strings.TrimSpace(res.Stderr) != "" {
		t.Errorf("taskmd %v without git wrote to stderr: %q", args, res.Stderr)
	}
}

// listRow returns the first line of table-formatted list output mentioning id.
func listRow(t *testing.T, res runResult, id string) string {
	t.Helper()
	for _, line := range strings.Split(res.Stdout, "\n") {
		if strings.Contains(line, id) {
			return line
		}
	}
	t.Fatalf("no row for %s in list output:\n%s", id, res.Stdout)
	return ""
}

// TestNoGit_PlainDirDegradesGracefully covers the spec's "graceful no-git
// behavior by scrubbing PATH" against the real binary, in the simplest case:
// a task directory that is not a git repository at all.
func TestNoGit_PlainDirDegradesGracefully(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-first.md", "001", "First task", "pending", nil)
	writeTask(t, dir, "002-second.md", "002", "Second task", "pending", []string{"001"})

	listArgs := []string{"list", "--format", "json"}
	res := runWithoutGit(t, dir, listArgs...)
	assertCleanRun(t, res, listArgs)
	statuses := listStatuses(t, res)
	if statuses["001"] != "pending" || statuses["002"] != "pending" {
		t.Errorf("list without git = %v, want both tasks pending", statuses)
	}

	nextArgs := []string{"next", "--format", "json"}
	res = runWithoutGit(t, dir, nextArgs...)
	assertCleanRun(t, res, nextArgs)
	if ids := nextIDs(t, res); len(ids) != 1 || ids[0] != "001" {
		t.Errorf("next without git = %v, want [001]", ids)
	}
}

// TestNoGit_WorktreeOverlayDegradesToLocalOnly is the interesting half: a real
// repo with a sibling worktree that has claimed a task. With git on PATH the
// overlay suppresses the claimed task; with git scrubbed the CLI must fall
// back to exactly the local-only view rather than erroring or warning.
func TestNoGit_WorktreeOverlayDegradesToLocalOnly(t *testing.T) {
	requireGit(t)
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md":  {id: "001", title: "First task", status: "pending"},
		"002-second.md": {id: "002", title: "Second task", status: "pending"},
	})
	agentB := addLinkedWorktree(t, repo, "agent-b")
	mustRun(t, agentB, "set", "001", "--status", "in-progress")

	// Baseline: with git available the overlay suppresses the claimed task and
	// shows its provenance in list.
	for _, id := range nextIDs(t, mustRun(t, repo, "next", "--format", "json")) {
		if id == "001" {
			t.Fatalf("precondition: overlay should suppress 001, but next offered it")
		}
	}
	if row := listRow(t, mustRun(t, repo, "list"), "001"); !strings.Contains(row, "agent-b") {
		t.Fatalf("precondition: list should show agent-b provenance for 001, got %q", row)
	}

	// Scrubbed PATH: no overlay, no noise, no failure.
	listArgs := []string{"list"}
	res := runWithoutGit(t, repo, listArgs...)
	assertCleanRun(t, res, listArgs)
	row := listRow(t, res, "001")
	if !strings.Contains(row, "pending") {
		t.Errorf("list without git: 001 row = %q, want the local-only pending", row)
	}
	if strings.Contains(row, "agent-b") {
		t.Errorf("list without git leaked worktree provenance: %q", row)
	}

	nextArgs := []string{"next", "--format", "json"}
	res = runWithoutGit(t, repo, nextArgs...)
	assertCleanRun(t, res, nextArgs)
	ids := nextIDs(t, res)
	var sawClaimed bool
	for _, id := range ids {
		if id == "001" {
			sawClaimed = true
		}
	}
	if !sawClaimed {
		t.Errorf("next without git = %v, want 001 offered again (overlay inert)", ids)
	}
}

// TestNoGit_WorktreeScopeFlagStillAccepted checks that explicitly asking for
// the overlay without git present is still a clean local-only run, not an
// error — the flag configures intent, git availability decides the outcome.
func TestNoGit_WorktreeScopeFlagStillAccepted(t *testing.T) {
	dir := setupTaskDir(t)
	writeTask(t, dir, "001-first.md", "001", "First task", "pending", nil)

	for _, scope := range []string{"unified", "isolated"} {
		args := []string{"--worktree-scope", scope, "list", "--format", "json"}
		res := runWithoutGit(t, dir, args...)
		assertCleanRun(t, res, args)
		if statuses := listStatuses(t, res); statuses["001"] != "pending" {
			t.Errorf("--worktree-scope %s without git = %v", scope, statuses)
		}
	}
}
