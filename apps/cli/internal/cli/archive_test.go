package cli

import (
	"bytes"
	"os"
	"path/filepath"
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

func createArchiveTestFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	files := map[string]string{
		"001-setup.md": taskCompleted,
		"002-old.md":   taskCancelled,
		"003-new.md":   taskPending,
		"004-api.md":   taskCompletedBackend,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	return tmpDir
}

func createArchiveTestFilesWithSubdir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "cli")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "001-setup.md"), []byte(taskCompleted), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "004-api.md"), []byte(taskCompletedBackend), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	return tmpDir
}

func resetArchiveFlags() {
	archiveIDs = nil
	archiveStatus = ""
	archiveAllCompleted = false
	archiveAllCancelled = false
	archiveTag = ""
	archiveDryRun = false
	archiveYes = false
	archiveDelete = false
	archiveForce = false
	taskDir = "."
}

func captureArchiveOutput(t *testing.T) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runArchive(archiveCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestArchive_ByID(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected archive confirmation, got: %s", output)
	}

	// Source should be gone
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected source file to be removed")
	}

	// Should exist in archive
	archived := filepath.Join(tmpDir, "archive", "001-setup.md")
	if _, err := os.Stat(archived); err != nil {
		t.Errorf("expected archived file at %s: %v", archived, err)
	}
}

func TestArchive_ByMultipleIDs(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"001", "002"}
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived, got: %s", output)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "001-setup.md")); err != nil {
		t.Error("expected 001 in archive")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "002-old.md")); err != nil {
		t.Error("expected 002 in archive")
	}
}

func TestArchive_AllCompleted(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveAllCompleted = true
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 completed tasks archived, got: %s", output)
	}

	// Completed tasks should be gone
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected 001 to be moved")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "004-api.md")); !os.IsNotExist(err) {
		t.Error("expected 004 to be moved")
	}

	// Non-completed tasks should remain
	if _, err := os.Stat(filepath.Join(tmpDir, "002-old.md")); err != nil {
		t.Error("expected 002 (cancelled) to remain")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "003-new.md")); err != nil {
		t.Error("expected 003 (pending) to remain")
	}
}

func TestArchive_AllCancelled(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveAllCancelled = true
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected 1 cancelled task archived, got: %s", output)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "002-old.md")); err != nil {
		t.Error("expected 002 in archive")
	}
}

func TestArchive_ByStatus(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveStatus = "completed"
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived by status, got: %s", output)
	}
}

func TestArchive_ByTag(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveTag = "backend"
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tasks 002 and 004 have "backend" tag
	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived by tag, got: %s", output)
	}
}

func TestArchive_CombinedFilters(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveAllCompleted = true
	archiveTag = "backend"
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only task 004 is both completed AND has "backend" tag
	if !strings.Contains(output, "Archived 1 task(s)") {
		t.Errorf("expected 1 task with combined filter, got: %s", output)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "004-api.md")); err != nil {
		t.Error("expected 004 in archive")
	}
	// 001 is completed but no "backend" tag — should remain
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); err != nil {
		t.Error("expected 001 to remain (no backend tag)")
	}
}

func TestArchive_DryRun(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveAllCompleted = true
	archiveDryRun = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Dry run") {
		t.Errorf("expected dry run message, got: %s", output)
	}
	if !strings.Contains(output, "Archive 2 task(s)") {
		t.Errorf("expected preview of 2 tasks, got: %s", output)
	}

	// Files should NOT be moved
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); err != nil {
		t.Error("expected source file to remain after dry run")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "004-api.md")); err != nil {
		t.Error("expected source file to remain after dry run")
	}

	// No archive directory should exist
	if _, err := os.Stat(filepath.Join(tmpDir, "archive")); !os.IsNotExist(err) {
		t.Error("expected no archive directory after dry run")
	}
}

func TestArchive_Delete(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveDelete = true
	archiveForce = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Deleted 1 task(s)") {
		t.Errorf("expected delete confirmation, got: %s", output)
	}

	// File should be gone
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}

	// Should NOT be in archive
	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "001-setup.md")); !os.IsNotExist(err) {
		t.Error("expected no archive copy when using --delete")
	}
}

func TestArchive_NoMatchingTasks(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"nonexistent"}
	archiveYes = true

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error for no matching tasks")
	}
	if !strings.Contains(err.Error(), "no tasks match") {
		t.Errorf("expected 'no tasks match' error, got: %v", err)
	}
}

func TestArchive_NoFiltersProvided(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error when no filters provided")
	}
	if !strings.Contains(err.Error(), "specify tasks to archive") {
		t.Errorf("expected 'specify tasks' error, got: %v", err)
	}
}

func TestArchive_PreservesDirectoryStructure(t *testing.T) {
	tmpDir := createArchiveTestFilesWithSubdir(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveAllCompleted = true
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Archived 2 task(s)") {
		t.Errorf("expected 2 tasks archived, got: %s", output)
	}

	// Root-level task archived at root of archive
	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "001-setup.md")); err != nil {
		t.Error("expected 001 at archive root")
	}

	// Subdir task preserves subdir structure
	if _, err := os.Stat(filepath.Join(tmpDir, "archive", "cli", "004-api.md")); err != nil {
		t.Error("expected 004 at archive/cli/")
	}
}

func TestArchive_ConflictingDestination(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()

	// Pre-create archive with conflicting file
	archivePath := filepath.Join(tmpDir, "archive")
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(archivePath, "001-setup.md"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveYes = true

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error for conflicting destination")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestArchive_RequiresConfirmation(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveAllCompleted = true
	// Neither --yes nor --dry-run

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error when no confirmation flag")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected confirmation hint in error, got: %v", err)
	}

	// Files should not have moved
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); err != nil {
		t.Error("expected file to remain without confirmation")
	}
}

func TestArchive_DeleteRequiresForce(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveDelete = true
	// No --force or --yes

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error when delete without force/yes")
	}

	// File should not have been deleted
	if _, err := os.Stat(filepath.Join(tmpDir, "001-setup.md")); err != nil {
		t.Error("expected file to remain without force flag")
	}
}

func TestArchive_DeleteWithYesFlag(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveDelete = true
	archiveYes = true

	output, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Deleted 1 task(s)") {
		t.Errorf("expected delete confirmation, got: %s", output)
	}
}

func TestArchive_MutuallyExclusive_AllCompletedAndStatus(t *testing.T) {
	resetArchiveFlags()
	archiveAllCompleted = true
	archiveStatus = "completed"

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}

func TestArchive_MutuallyExclusive_AllCancelledAndStatus(t *testing.T) {
	resetArchiveFlags()
	archiveAllCancelled = true
	archiveStatus = "cancelled"

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}

func TestArchive_MutuallyExclusive_AllCompletedAndAllCancelled(t *testing.T) {
	resetArchiveFlags()
	archiveAllCompleted = true
	archiveAllCancelled = true

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}

func TestArchive_InvalidStatus(t *testing.T) {
	resetArchiveFlags()
	archiveStatus = "bogus"

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %v", err)
	}
}

func TestArchive_ScannerSkipsArchiveDir(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()

	// Archive task 001
	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveYes = true

	_, err := captureArchiveOutput(t)
	if err != nil {
		t.Fatalf("archive failed: %v", err)
	}

	// Now try to archive "001" again — it should not be found since
	// the scanner skips the archive directory
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"001"}
	archiveYes = true

	_, err = captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error: task should not be found after archiving")
	}
	if !strings.Contains(err.Error(), "no tasks match") {
		t.Errorf("expected 'no tasks match' error, got: %v", err)
	}
}

func TestArchive_IDWithStatusFilter(t *testing.T) {
	tmpDir := createArchiveTestFiles(t)
	resetArchiveFlags()
	taskDir = tmpDir
	archiveIDs = []string{"003"} // pending task
	archiveStatus = "completed"  // only match completed
	archiveYes = true

	_, err := captureArchiveOutput(t)
	if err == nil {
		t.Fatal("expected error: 003 is pending, not completed")
	}
	if !strings.Contains(err.Error(), "no tasks match") {
		t.Errorf("expected 'no tasks match' error, got: %v", err)
	}
}
