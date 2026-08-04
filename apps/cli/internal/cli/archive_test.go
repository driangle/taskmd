package cli

import (
	"os"
	"strings"
	"testing"
)

const taskCompleted = `---
id: "001"
title: "Setup project"
status: completed
priority: high
effort: small
dependencies: []
tags: ["infra"]
created: 2026-02-08
---

# Setup project
`

const taskCancelled = `---
id: "002"
title: "Old feature"
status: cancelled
priority: low
effort: medium
dependencies: []
tags: ["backend"]
created: 2026-02-08
---

# Old feature
`

const taskPending = `---
id: "003"
title: "New feature"
status: pending
priority: high
effort: large
dependencies: []
tags: ["frontend"]
created: 2026-02-08
---

# New feature
`

const taskCompletedBackend = `---
id: "004"
title: "API endpoint"
status: completed
priority: medium
effort: small
dependencies: []
tags: ["backend"]
created: 2026-02-08
---

# API endpoint
`

// archiveFiles is the canonical archive fixture: four tasks with mixed statuses
// and tags so the status/tag/completed/cancelled filters can be exercised. Kept
// inline (not promoted to testdata/) because this archive-specific status mix is
// the subject of these tests.
func archiveFiles() map[string]string {
	return map[string]string{
		"001-setup.md": taskCompleted,
		"002-old.md":   taskCancelled,
		"003-new.md":   taskPending,
		"004-api.md":   taskCompletedBackend,
	}
}

// archiveStdout runs `archive <args...>` against repo, fails on error, and
// returns stdout.
func archiveStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"archive"}, args...)...)
	if res.Err != nil {
		t.Fatalf("archive %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestArchive_ByID(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--id", "001", "--yes")

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", output)
	}

	// Source should be gone
	if _, err := os.Stat(repo.Path("001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected source file to be removed")
	}

	// Should exist in archive
	archived := repo.Path("archive/001-setup.md")
	if _, err := os.Stat(archived); err != nil {
		t.Errorf("expected archived file at %s: %v", archived, err)
	}
}

func TestArchive_ByMultipleIDs(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--id", "001", "--id", "002", "--yes")

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived, got: %s", output)
	}

	if _, err := os.Stat(repo.Path("archive/001-setup.md")); err != nil {
		t.Error("expected 001 in archive")
	}
	if _, err := os.Stat(repo.Path("archive/002-old.md")); err != nil {
		t.Error("expected 002 in archive")
	}
}

func TestArchive_AllCompleted(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--all-completed", "--yes")

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 completed tasks archived, got: %s", output)
	}

	// Completed tasks should be gone
	if _, err := os.Stat(repo.Path("001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected 001 to be moved")
	}
	if _, err := os.Stat(repo.Path("004-api.md")); !os.IsNotExist(err) {
		t.Error("expected 004 to be moved")
	}

	// Non-completed tasks should remain
	if _, err := os.Stat(repo.Path("002-old.md")); err != nil {
		t.Error("expected 002 (cancelled) to remain")
	}
	if _, err := os.Stat(repo.Path("003-new.md")); err != nil {
		t.Error("expected 003 (pending) to remain")
	}
}

func TestArchive_AllCancelled(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--all-cancelled", "--yes")

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected 1 cancelled task archived, got: %s", output)
	}

	if _, err := os.Stat(repo.Path("archive/002-old.md")); err != nil {
		t.Error("expected 002 in archive")
	}
}

func TestArchive_ByStatus(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--status", "completed", "--yes")

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived by status, got: %s", output)
	}
}

func TestArchive_ByTag(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--tag", "backend", "--yes")

	// Tasks 002 and 004 have "backend" tag
	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived by tag, got: %s", output)
	}
}

func TestArchive_CombinedFilters(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--all-completed", "--tag", "backend", "--yes")

	// Only task 004 is both completed AND has "backend" tag
	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected 1 task with combined filter, got: %s", output)
	}

	if _, err := os.Stat(repo.Path("archive/004-api.md")); err != nil {
		t.Error("expected 004 in archive")
	}
	// 001 is completed but no "backend" tag — should remain
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected 001 to remain (no backend tag)")
	}
}

func TestArchive_DryRun(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--all-completed", "--dry-run")

	if !strings.Contains(output, "Dry run") {
		t.Errorf("expected dry run message, got: %s", output)
	}
	if !strings.Contains(output, "Archive 2 task(s)") {
		t.Errorf("expected preview of 2 tasks, got: %s", output)
	}

	// Files should NOT be moved
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected source file to remain after dry run")
	}
	if _, err := os.Stat(repo.Path("004-api.md")); err != nil {
		t.Error("expected source file to remain after dry run")
	}

	// No archive directory should exist
	if _, err := os.Stat(repo.Path("archive")); !os.IsNotExist(err) {
		t.Error("expected no archive directory after dry run")
	}
}

func TestArchive_Delete(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--id", "001", "--delete", "--force")

	if !strings.Contains(output, "Deleted 1 task(s)") {
		t.Errorf("expected delete confirmation, got: %s", output)
	}

	// File should be gone
	if _, err := os.Stat(repo.Path("001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}

	// Should NOT be in archive
	if _, err := os.Stat(repo.Path("archive/001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected no archive copy when using --delete")
	}
}

func TestArchive_NoMatchingTasks(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	res := repo.Run("archive", "--id", "nonexistent", "--yes")
	if res.Err == nil {
		t.Fatal("expected error for no matching tasks")
	}
	if !strings.Contains(res.Err.Error(), "no tasks match") {
		t.Errorf("expected 'no tasks match' error, got: %v", res.Err)
	}
}

func TestArchive_NoFiltersProvided(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	res := repo.Run("archive")
	if res.Err == nil {
		t.Fatal("expected error when no filters provided")
	}
	if !strings.Contains(res.Err.Error(), "specify tasks to archive") {
		t.Errorf("expected 'specify tasks' error, got: %v", res.Err)
	}
}

func TestArchive_PreservesDirectoryStructure(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-setup.md":   taskCompleted,
		"cli/004-api.md": taskCompletedBackend,
	})

	output := archiveStdout(t, repo, "--all-completed", "--yes")

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived, got: %s", output)
	}

	// Root-level task archived at root of archive
	if _, err := os.Stat(repo.Path("archive/001-setup.md")); err != nil {
		t.Error("expected 001 at archive root")
	}

	// Subdir task preserves subdir structure
	if _, err := os.Stat(repo.Path("archive/cli/004-api.md")); err != nil {
		t.Error("expected 004 at archive/cli/")
	}
}

func TestArchive_ConflictingDestination(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	// Pre-create archive with conflicting file
	repo.Write("archive/001-setup.md", "existing")

	res := repo.Run("archive", "--id", "001", "--yes")
	if res.Err == nil {
		t.Fatal("expected error for conflicting destination")
	}
	if !strings.Contains(res.Err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", res.Err)
	}
}

func TestArchive_RequiresConfirmation_Declined(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())
	archiveStdin = strings.NewReader("n\n")
	defer func() { archiveStdin = os.Stdin }()

	res := repo.Run("archive", "--all-completed")
	if res.Err == nil {
		t.Fatal("expected error when user declines")
	}
	if !strings.Contains(res.Err.Error(), "archive cancelled") {
		t.Errorf("expected 'archive cancelled' error, got: %v", res.Err)
	}

	// Files should not have moved
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected file to remain when user declines")
	}
}

func TestArchive_RequiresConfirmation_Accepted(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())
	archiveStdin = strings.NewReader("y\n")
	defer func() { archiveStdin = os.Stdin }()

	res := repo.Run("archive", "--all-completed")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Archived 2 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", res.Stdout)
	}
}

func TestArchive_RequiresConfirmation_EmptyInput(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())
	archiveStdin = strings.NewReader("\n")
	defer func() { archiveStdin = os.Stdin }()

	res := repo.Run("archive", "--all-completed")
	if res.Err == nil {
		t.Fatal("expected error when user presses enter without input")
	}
	if !strings.Contains(res.Err.Error(), "archive cancelled") {
		t.Errorf("expected 'archive cancelled' error, got: %v", res.Err)
	}
}

func TestArchive_DeleteRequiresForceOrConfirmation(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())
	archiveStdin = strings.NewReader("n\n")
	defer func() { archiveStdin = os.Stdin }()

	res := repo.Run("archive", "--id", "001", "--delete")
	if res.Err == nil {
		t.Fatal("expected error when delete declined")
	}
	if !strings.Contains(res.Err.Error(), "delete cancelled") {
		t.Errorf("expected 'delete cancelled' error, got: %v", res.Err)
	}

	// File should not have been deleted
	if _, err := os.Stat(repo.Path("001-setup.md")); err != nil {
		t.Error("expected file to remain when user declines")
	}
}

func TestArchive_DeleteWithYesFlag(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "--id", "001", "--delete", "--yes")

	if !strings.Contains(output, "Deleted 1 task(s)") {
		t.Errorf("expected delete confirmation, got: %s", output)
	}
}

func TestArchive_MutuallyExclusive_AllCompletedAndStatus(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("archive", "--all-completed", "--status", "completed")
	if res.Err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(res.Err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", res.Err)
	}
}

func TestArchive_MutuallyExclusive_AllCancelledAndStatus(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("archive", "--all-cancelled", "--status", "cancelled")
	if res.Err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(res.Err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", res.Err)
	}
}

func TestArchive_MutuallyExclusive_AllCompletedAndAllCancelled(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("archive", "--all-completed", "--all-cancelled")
	if res.Err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(res.Err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", res.Err)
	}
}

func TestArchive_InvalidStatus(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("archive", "--status", "bogus")
	if res.Err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(res.Err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %v", res.Err)
	}
}

func TestArchive_ScannerSkipsArchiveDir(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	// Archive task 001
	if res := repo.Run("archive", "--id", "001", "--yes"); res.Err != nil {
		t.Fatalf("archive failed: %v", res.Err)
	}

	// Now try to archive "001" again — it should not be found since
	// the scanner skips the archive directory
	res := repo.Run("archive", "--id", "001", "--yes")
	if res.Err == nil {
		t.Fatal("expected error: task should not be found after archiving")
	}
	if !strings.Contains(res.Err.Error(), "no tasks match") {
		t.Errorf("expected 'no tasks match' error, got: %v", res.Err)
	}
}

func TestArchive_IDWithStatusFilter(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	// 003 is pending; --status completed should exclude it.
	res := repo.Run("archive", "--id", "003", "--status", "completed", "--yes")
	if res.Err == nil {
		t.Fatal("expected error: 003 is pending, not completed")
	}
	if !strings.Contains(res.Err.Error(), "no tasks match") {
		t.Errorf("expected 'no tasks match' error, got: %v", res.Err)
	}
}

func TestArchive_PositionalArg(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "001", "--yes")

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", output)
	}

	if _, err := os.Stat(repo.Path("archive/001-setup.md")); err != nil {
		t.Error("expected 001 in archive")
	}
}

func TestArchive_PositionalArgWithInteractiveConfirm(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())
	archiveStdin = strings.NewReader("y\n")
	defer func() { archiveStdin = os.Stdin }()

	res := repo.Run("archive", "001")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Archived 1 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", res.Stdout)
	}
}

func TestArchive_PositionalArgWithYesFlag(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	output := archiveStdout(t, repo, "001", "--yes")

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", output)
	}
}

func TestArchive_PositionalArgBackwardCompatible(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())

	// --id flag still works
	output := archiveStdout(t, repo, "--id", "001", "--yes")

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", output)
	}
}

func TestArchive_InteractiveConfirm_CaseInsensitive(t *testing.T) {
	repo := newTaskRepo(t, archiveFiles())
	archiveStdin = strings.NewReader("Y\n")
	defer func() { archiveStdin = os.Stdin }()

	res := repo.Run("archive", "--all-completed")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "Archived 2 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", res.Stdout)
	}
}
