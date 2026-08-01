//go:build e2e

package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// gitExec runs a git command in dir with a deterministic author/committer,
// failing the test on error. Used to build task history for feed tests.
func gitExec(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=Tester",
		"GIT_AUTHOR_EMAIL=tester@example.com",
		"GIT_COMMITTER_NAME=Tester",
		"GIT_COMMITTER_EMAIL=tester@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// setupTaskGitRepo creates a task project under git with one task that has been
// created, moved to in-progress, then completed — three commits of history.
func setupTaskGitRepo(t *testing.T) string {
	t.Helper()
	dir := setupTaskDir(t)

	gitExec(t, dir, "init")
	writeTask(t, dir, "001-setup.md", "001", "Setup", "pending", nil)
	gitExec(t, dir, "add", "-A")
	gitExec(t, dir, "commit", "-m", "add task 001")

	mustRun(t, dir, "set", "001", "--status", "in-progress")
	gitExec(t, dir, "add", "-A")
	gitExec(t, dir, "commit", "-m", "start task 001")

	mustRun(t, dir, "set", "001", "--status", "completed")
	gitExec(t, dir, "add", "-A")
	gitExec(t, dir, "commit", "-m", "complete task 001")

	return dir
}

func TestFeed_SingleTask_Timeline(t *testing.T) {
	dir := setupTaskGitRepo(t)

	result := mustRun(t, dir, "feed", "001")

	if !strings.Contains(result.Stdout, "[Added]") {
		t.Errorf("expected creation event, got:\n%s", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "status pending → in-progress") {
		t.Errorf("expected pending→in-progress transition, got:\n%s", result.Stdout)
	}
	if !strings.Contains(result.Stdout, "status in-progress → completed") {
		t.Errorf("expected in-progress→completed transition, got:\n%s", result.Stdout)
	}
}

func TestFeed_SingleTask_JSON(t *testing.T) {
	dir := setupTaskGitRepo(t)

	result := mustRun(t, dir, "feed", "001", "--format", "json")

	if !strings.Contains(result.Stdout, `"field": "status"`) {
		t.Errorf("expected status field change in JSON, got:\n%s", result.Stdout)
	}
	if !strings.Contains(result.Stdout, `"newValue": "completed"`) {
		t.Errorf("expected completed transition in JSON, got:\n%s", result.Stdout)
	}
}

func TestFeed_SingleTask_FieldPivot(t *testing.T) {
	dir := setupTaskGitRepo(t)

	// No priority changes were made, so only the structural creation event
	// should remain — no status transition lines.
	result := mustRun(t, dir, "feed", "001", "--field", "priority")

	if !strings.Contains(result.Stdout, "[Added]") {
		t.Errorf("expected creation event to remain, got:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "→ completed") {
		t.Errorf("did not expect status transitions when pivoting to priority, got:\n%s", result.Stdout)
	}
}

func TestFeed_SingleTask_UnknownID(t *testing.T) {
	dir := setupTaskGitRepo(t)

	result := run(t, dir, "feed", "does-not-exist")

	if result.ExitCode == 0 {
		t.Fatal("expected non-zero exit code for unknown task id")
	}
	if !strings.Contains(result.Stdout+result.Stderr, "not found") {
		t.Errorf("expected 'not found' error, got:\nstdout: %s\nstderr: %s", result.Stdout, result.Stderr)
	}
}
