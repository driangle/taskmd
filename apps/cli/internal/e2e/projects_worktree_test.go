//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeGlobalRegistry writes a global registry file with the given project
// entries and returns the env var slice pointing taskmd at it.
func writeGlobalRegistry(t *testing.T, entries ...[2]string) []string {
	t.Helper()
	var b strings.Builder
	b.WriteString("projects:\n")
	for _, e := range entries {
		b.WriteString("  - id: " + e[0] + "\n    path: " + e[1] + "\n")
	}
	globalCfg := filepath.Join(t.TempDir(), "global.yaml")
	if err := os.WriteFile(globalCfg, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return []string{"TASKMD_HOME_CONFIG=" + globalCfg}
}

// listStatuses parses `list --format json` output into id → status.
func listStatuses(t *testing.T, res runResult) map[string]string {
	t.Helper()
	var tasks []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &tasks); err != nil {
		t.Fatalf("parse list output: %v\n%s", err, res.Stdout)
	}
	statuses := make(map[string]string, len(tasks))
	for _, task := range tasks {
		statuses[task.ID] = task.Status
	}
	return statuses
}

func TestProjectsRegister_FromLinkedWorktreeStoresPrimary(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md": {id: "001", title: "First task", status: "pending"},
	})
	wt := addLinkedWorktree(t, repo, "register-wt")

	globalCfg := filepath.Join(t.TempDir(), "global.yaml")
	env := []string{"TASKMD_HOME_CONFIG=" + globalCfg}

	res := runWithEnv(t, wt, env, "projects", "register", "--id", "myrepo")
	if res.ExitCode != 0 {
		t.Fatalf("register from linked worktree failed (%d): %s", res.ExitCode, res.Stderr)
	}

	data, err := os.ReadFile(globalCfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), repo) {
		t.Errorf("registry should store primary path %q, got:\n%s", repo, data)
	}
	if strings.Contains(string(data), wt) {
		t.Errorf("registry should not store worktree path %q, got:\n%s", wt, data)
	}

	// Registering the primary of the already-registered repo is a friendly no-op.
	res = runWithEnv(t, repo, env, "projects", "register")
	if res.ExitCode != 0 {
		t.Fatalf("duplicate register should succeed as no-op (%d): %s", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stdout, `Already registered as "myrepo"`) {
		t.Errorf("expected friendly no-op message, got: %q", res.Stdout)
	}

	// taskmd projects lists the repo exactly once.
	res = runWithEnv(t, repo, env, "projects", "--format", "json")
	if res.ExitCode != 0 {
		t.Fatalf("projects failed (%d): %s", res.ExitCode, res.Stderr)
	}
	var summaries []struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("parse projects output: %v\n%s", err, res.Stdout)
	}
	if len(summaries) != 1 || summaries[0].ID != "myrepo" {
		t.Errorf("expected exactly one project myrepo, got %+v", summaries)
	}
}

func TestProjectFlag_FromWorktreeScansCurrentCheckout(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md": {id: "001", title: "First task", status: "pending"},
	})
	wt := addLinkedWorktree(t, repo, "scope-wt")
	env := writeGlobalRegistry(t, [2]string{"proj", repo})

	// Claim 001 in the worktree only; the primary still has it pending.
	mustRun(t, wt, "set", "001", "--status", "in-progress")

	// --project from inside the worktree scans the worktree's own files.
	res := runWithEnv(t, wt, env, "list", "--project", "proj", "--worktree-scope", "isolated", "--format", "json")
	if res.ExitCode != 0 {
		t.Fatalf("list --project from worktree failed (%d): %s", res.ExitCode, res.Stderr)
	}
	if statuses := listStatuses(t, res); statuses["001"] != "in-progress" {
		t.Errorf("expected worktree's status in-progress, got %q", statuses["001"])
	}

	// From outside the repo, --project scans the registered primary.
	neutral := t.TempDir()
	res = runWithEnv(t, neutral, env, "list", "--project", "proj", "--worktree-scope", "isolated", "--format", "json")
	if res.ExitCode != 0 {
		t.Fatalf("list --project from outside failed (%d): %s", res.ExitCode, res.Stderr)
	}
	if statuses := listStatuses(t, res); statuses["001"] != "pending" {
		t.Errorf("expected primary's status pending, got %q", statuses["001"])
	}
}

func TestAllProjects_CountsEachRepoOnce(t *testing.T) {
	repo := initWorktreeRepo(t, map[string]taskSpec{
		"001-first.md": {id: "001", title: "First task", status: "pending"},
	})
	wt := addLinkedWorktree(t, repo, "dedupe-wt")
	// Legacy registry shape: both worktrees registered as separate projects.
	env := writeGlobalRegistry(t, [2]string{"a", repo}, [2]string{"b", wt})

	countTask001 := func(res runResult) (count int, status string) {
		t.Helper()
		var tasks []struct {
			Project string `json:"project"`
			ID      string `json:"id"`
			Status  string `json:"status"`
		}
		if err := json.Unmarshal([]byte(res.Stdout), &tasks); err != nil {
			t.Fatalf("parse all-projects output: %v\n%s", err, res.Stdout)
		}
		for _, task := range tasks {
			if task.ID == "001" {
				count++
				status = task.Status
			}
		}
		return count, status
	}

	// From outside the repo: each repo scanned once, at the registered primary.
	neutral := t.TempDir()
	res := runWithEnv(t, neutral, env, "list", "--all-projects", "--format", "json")
	if res.ExitCode != 0 {
		t.Fatalf("list --all-projects failed (%d): %s", res.ExitCode, res.Stderr)
	}
	if count, _ := countTask001(res); count != 1 {
		t.Errorf("expected task 001 exactly once from outside, got %d\n%s", count, res.Stdout)
	}

	// From inside a worktree: still once, scanned from the current checkout.
	mustRun(t, wt, "set", "001", "--status", "in-progress")
	res = runWithEnv(t, wt, env, "list", "--all-projects", "--format", "json")
	if res.ExitCode != 0 {
		t.Fatalf("list --all-projects in worktree failed (%d): %s", res.ExitCode, res.Stderr)
	}
	count, status := countTask001(res)
	if count != 1 {
		t.Errorf("expected task 001 exactly once from worktree, got %d\n%s", count, res.Stdout)
	}
	if status != "in-progress" {
		t.Errorf("expected current worktree's status in-progress, got %q", status)
	}
}
