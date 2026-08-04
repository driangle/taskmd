package cli

import (
	"os"
	"strings"
	"testing"
)

const rmTaskPending = `---
id: "001"
title: "Setup project"
status: pending
priority: high
effort: small
dependencies: []
tags: ["infra"]
created: 2026-02-08
---

# Setup project
`

const rmTaskCompleted = `---
id: "002"
title: "Old feature"
status: completed
priority: low
effort: medium
dependencies: []
tags: ["backend"]
created: 2026-02-08
---

# Old feature
`

// rmRepo seeds a repo with a pending task (001) and a completed task (002).
func rmRepo(t *testing.T) *taskRepo {
	t.Helper()
	return newTaskRepo(t, map[string]string{
		"001-setup.md": rmTaskPending,
		"002-old.md":   rmTaskCompleted,
	})
}

func TestRm_WithForce(t *testing.T) {
	repo := rmRepo(t)

	res := repo.Run("rm", "001", "--force")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Deleted 1 task") {
		t.Errorf("expected delete confirmation, got: %s", res.Stdout)
	}

	// File should be gone
	if _, err := os.Stat(repo.Path("001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}

	// Other file should remain
	if _, err := os.Stat(repo.Path("002-old.md")); err != nil {
		t.Error("expected other file to remain")
	}
}

func TestRm_DryRun(t *testing.T) {
	repo := rmRepo(t)

	res := repo.Run("rm", "001", "--dry-run")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Dry run") {
		t.Errorf("expected dry run message, got: %s", res.Stdout)
	}

	if !strings.Contains(res.Stdout, "Delete 1 task") {
		t.Errorf("expected preview of task, got: %s", res.Stdout)
	}

	// File should NOT be deleted
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected file to remain after dry run")
	}
}

func TestRm_TaskNotFound(t *testing.T) {
	repo := rmRepo(t)

	res := repo.Run("rm", "999", "--force")
	if res.Err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("expected 'task not found' error, got: %v", res.Err)
	}
}

func TestRm_InteractiveConfirmYes(t *testing.T) {
	repo := rmRepo(t)

	// Simulate user typing "y"
	oldStdin := rmStdinReader
	rmStdinReader = strings.NewReader("y\n")
	defer func() { rmStdinReader = oldStdin }()

	res := repo.Run("rm", "001")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Deleted 1 task") {
		t.Errorf("expected delete confirmation, got: %s", res.Stdout)
	}

	// File should be gone
	if _, err := os.Stat(repo.Path("001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected file to be deleted after confirming")
	}
}

func TestRm_InteractiveConfirmNo(t *testing.T) {
	repo := rmRepo(t)

	// Simulate user typing "n"
	oldStdin := rmStdinReader
	rmStdinReader = strings.NewReader("n\n")
	defer func() { rmStdinReader = oldStdin }()

	res := repo.Run("rm", "001")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Cancelled") {
		t.Errorf("expected cancellation message, got: %s", res.Stdout)
	}

	// File should remain
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected file to remain after declining")
	}
}

func TestRm_InteractiveConfirmEmpty(t *testing.T) {
	repo := rmRepo(t)

	// Simulate user pressing Enter (empty input = default No)
	oldStdin := rmStdinReader
	rmStdinReader = strings.NewReader("\n")
	defer func() { rmStdinReader = oldStdin }()

	res := repo.Run("rm", "001")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Cancelled") {
		t.Errorf("expected cancellation message, got: %s", res.Stdout)
	}

	// File should remain
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected file to remain after empty input")
	}
}

func TestRm_BlockedByDependency(t *testing.T) {
	// Task 001 exists; task 003 depends on 001
	depTask := `---
id: "003"
title: "Depends on setup"
status: pending
priority: medium
effort: small
dependencies: ["001"]
tags: []
created: 2026-02-08
---

# Depends on setup
`
	repo := newTaskRepo(t, map[string]string{
		"001-setup.md": rmTaskPending,
		"003-dep.md":   depTask,
	})

	// Without --force, should fail
	res := repo.Run("rm", "001")
	if res.Err == nil {
		t.Fatal("expected error when task is referenced")
	}
	if !strings.Contains(res.Err.Error(), "referenced by other tasks") {
		t.Errorf("expected reference error, got: %v", res.Err)
	}
	if !strings.Contains(res.Err.Error(), "003") {
		t.Errorf("expected referencing task ID in error, got: %v", res.Err)
	}

	// File should still exist
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected file to remain when blocked by reference")
	}
}

func TestRm_BlockedByParent(t *testing.T) {
	childTask := `---
id: "004"
title: "Child task"
status: pending
priority: medium
effort: small
dependencies: []
parent: "001"
tags: []
created: 2026-02-08
---

# Child task
`
	repo := newTaskRepo(t, map[string]string{
		"001-setup.md": rmTaskPending,
		"004-child.md": childTask,
	})

	res := repo.Run("rm", "001")
	if res.Err == nil {
		t.Fatal("expected error when task has child referencing it as parent")
	}
	if !strings.Contains(res.Err.Error(), "referenced by other tasks") {
		t.Errorf("expected reference error, got: %v", res.Err)
	}
	if !strings.Contains(res.Err.Error(), "004") {
		t.Errorf("expected child task ID in error, got: %v", res.Err)
	}
}

func TestRm_ForceDeletesReferencedTask(t *testing.T) {
	depTask := `---
id: "003"
title: "Depends on setup"
status: pending
priority: medium
effort: small
dependencies: ["001"]
tags: []
created: 2026-02-08
---

# Depends on setup
`
	repo := newTaskRepo(t, map[string]string{
		"001-setup.md": rmTaskPending,
		"003-dep.md":   depTask,
	})

	res := repo.Run("rm", "001", "--force")
	if res.Err != nil {
		t.Fatalf("expected --force to override reference check, got: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "Deleted 1 task") {
		t.Errorf("expected delete confirmation, got: %s", res.Stdout)
	}

	if _, err := os.Stat(repo.Path("001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected file to be deleted with --force")
	}
}

func TestRm_DeletesWorklog(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{"001-setup.md": rmTaskPending})
	repo.Write(".worklogs/001.md", "## 2026-02-08T10:00:00Z\n\nStarted work.\n")

	res := repo.Run("rm", "001", "--force")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Deleted worklog") {
		t.Errorf("expected worklog deletion message, got: %s", res.Stdout)
	}

	// Worklog file should be gone
	if _, err := os.Stat(repo.Path(".worklogs/001.md")); !os.IsNotExist(err) {
		t.Error("expected worklog file to be deleted")
	}

	// Empty .worklogs dir should be removed
	if _, err := os.Stat(repo.Path(".worklogs")); !os.IsNotExist(err) {
		t.Error("expected empty .worklogs directory to be removed")
	}
}

func TestRm_WorklogDirKeptIfNotEmpty(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-setup.md": rmTaskPending,
		"002-old.md":   rmTaskCompleted,
	})
	repo.Write(".worklogs/001.md", "## 2026-02-08T10:00:00Z\n\nWork.\n")
	repo.Write(".worklogs/002.md", "## 2026-02-08T10:00:00Z\n\nWork.\n")

	res := repo.Run("rm", "001", "--force")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	// 001 worklog gone, but .worklogs dir should remain (002.md still there)
	if _, err := os.Stat(repo.Path(".worklogs/001.md")); !os.IsNotExist(err) {
		t.Error("expected 001 worklog to be deleted")
	}
	if _, err := os.Stat(repo.Path(".worklogs")); err != nil {
		t.Error("expected .worklogs directory to remain (still has 002.md)")
	}
}

func TestRm_ShowsTaskDetails(t *testing.T) {
	repo := rmRepo(t)

	res := repo.Run("rm", "001", "--dry-run")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "001") {
		t.Errorf("expected task ID in output, got: %s", res.Stdout)
	}
	if !strings.Contains(res.Stdout, "Setup project") {
		t.Errorf("expected task title in output, got: %s", res.Stdout)
	}
}
