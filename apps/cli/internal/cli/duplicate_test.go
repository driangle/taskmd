package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

func TestFindDuplicatesByID(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", FilePath: "a/001.md"},
		{ID: "002", FilePath: "a/002.md"},
		{ID: "001", FilePath: "b/001.md"},
		{ID: "003", FilePath: "a/003.md"},
	}

	t.Run("returns all matches for duplicate ID", func(t *testing.T) {
		matches := findDuplicatesByID("001", tasks)
		if len(matches) != 2 {
			t.Fatalf("expected 2 matches, got %d", len(matches))
		}
		if matches[0].FilePath != "a/001.md" || matches[1].FilePath != "b/001.md" {
			t.Errorf("unexpected paths: %v, %v", matches[0].FilePath, matches[1].FilePath)
		}
	})

	t.Run("returns single match for unique ID", func(t *testing.T) {
		matches := findDuplicatesByID("002", tasks)
		if len(matches) != 1 {
			t.Fatalf("expected 1 match, got %d", len(matches))
		}
	})

	t.Run("returns empty for unknown ID", func(t *testing.T) {
		matches := findDuplicatesByID("999", tasks)
		if len(matches) != 0 {
			t.Fatalf("expected 0 matches, got %d", len(matches))
		}
	})
}

func TestFindAllDuplicateIDs(t *testing.T) {
	t.Run("finds duplicates", func(t *testing.T) {
		tasks := []*model.Task{
			{ID: "001", FilePath: "a/001.md"},
			{ID: "002", FilePath: "a/002.md"},
			{ID: "001", FilePath: "b/001.md"},
			{ID: "003", FilePath: "a/003.md"},
			{ID: "003", FilePath: "b/003.md"},
		}

		dupes := findAllDuplicateIDs(tasks)
		if len(dupes) != 2 {
			t.Fatalf("expected 2 duplicate IDs, got %d", len(dupes))
		}
		if len(dupes["001"]) != 2 {
			t.Errorf("expected 2 paths for ID 001, got %d", len(dupes["001"]))
		}
		if len(dupes["003"]) != 2 {
			t.Errorf("expected 2 paths for ID 003, got %d", len(dupes["003"]))
		}
	})

	t.Run("returns empty when no duplicates", func(t *testing.T) {
		tasks := []*model.Task{
			{ID: "001", FilePath: "a/001.md"},
			{ID: "002", FilePath: "a/002.md"},
		}

		dupes := findAllDuplicateIDs(tasks)
		if len(dupes) != 0 {
			t.Fatalf("expected no duplicates, got %d", len(dupes))
		}
	})
}

func TestWarnDuplicateIDs(t *testing.T) {
	t.Run("prints warning for duplicates", func(t *testing.T) {
		tasks := []*model.Task{
			{ID: "042", FilePath: "tasks/cli/042-foo.md"},
			{ID: "042", FilePath: "tasks/web/042-bar.md"},
			{ID: "001", FilePath: "tasks/001.md"},
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		dupes := warnDuplicateIDs(tasks)

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if len(dupes) != 1 {
			t.Fatalf("expected 1 duplicate ID, got %d", len(dupes))
		}

		if !strings.Contains(output, `ID "042"`) {
			t.Errorf("expected warning to mention ID 042, got: %s", output)
		}
		if !strings.Contains(output, "tasks/cli/042-foo.md") {
			t.Errorf("expected warning to mention file path, got: %s", output)
		}
		if !strings.Contains(output, "taskmd deduplicate") {
			t.Errorf("expected warning to mention deduplicate command, got: %s", output)
		}
	})

	t.Run("no output when no duplicates", func(t *testing.T) {
		tasks := []*model.Task{
			{ID: "001", FilePath: "a/001.md"},
			{ID: "002", FilePath: "a/002.md"},
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		dupes := warnDuplicateIDs(tasks)

		w.Close()
		os.Stderr = oldStderr

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if len(dupes) != 0 {
			t.Fatalf("expected no duplicates, got %d", len(dupes))
		}
		if output != "" {
			t.Errorf("expected no output, got: %s", output)
		}
	})
}

func TestFormatDuplicatePaths(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", FilePath: "a/001.md"},
		{ID: "001", FilePath: "b/001.md"},
	}

	result := formatDuplicatePaths(tasks)
	if !strings.Contains(result, "  - a/001.md") {
		t.Errorf("expected bulleted path, got: %s", result)
	}
	if !strings.Contains(result, "  - b/001.md") {
		t.Errorf("expected bulleted path, got: %s", result)
	}
}

func TestFormatDuplicatePathsWithTitles(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", FilePath: "a/001.md"},
		{ID: "001", Title: "Task B", FilePath: "b/001.md"},
	}

	result := formatDuplicatePathsWithTitles(tasks)
	if !strings.Contains(result, "  - a/001.md (Task A)") {
		t.Errorf("expected bulleted path with title, got: %s", result)
	}
	if !strings.Contains(result, "  - b/001.md (Task B)") {
		t.Errorf("expected bulleted path with title, got: %s", result)
	}
}

// Helper to create task files with duplicate IDs for integration tests.
func createDuplicateTestFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	groupA := filepath.Join(tmpDir, "groupA")
	groupB := filepath.Join(tmpDir, "groupB")
	if err := os.MkdirAll(groupA, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(groupB, 0755); err != nil {
		t.Fatal(err)
	}

	taskA := `---
id: "042"
title: "Task A"
status: pending
priority: high
effort: small
dependencies: []
tags: []
created: 2026-02-08
---

# Task A
`
	taskB := `---
id: "042"
title: "Task B"
status: pending
priority: medium
effort: medium
dependencies: []
tags: []
created: 2026-02-08
---

# Task B
`
	uniqueTask := `---
id: "001"
title: "Unique Task"
status: pending
priority: low
effort: small
dependencies: []
tags: []
created: 2026-02-08
---

# Unique Task
`

	for name, content := range map[string]string{
		filepath.Join(groupA, "042-task-a.md"): taskA,
		filepath.Join(groupB, "042-task-b.md"): taskB,
		filepath.Join(tmpDir, "001-unique.md"): uniqueTask,
	} {
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	return tmpDir
}

func TestResolveTask_DuplicateIDError(t *testing.T) {
	tasks := []*model.Task{
		{ID: "042", Title: "Task A", FilePath: "groupA/042-task-a.md"},
		{ID: "042", Title: "Task B", FilePath: "groupB/042-task-b.md"},
		{ID: "001", Title: "Unique Task", FilePath: "001-unique.md"},
	}

	_, err := resolveTask("042", tasks, true, 0.6)
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate task ID") {
		t.Errorf("expected 'duplicate task ID' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "042") {
		t.Errorf("expected error to mention ID 042, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Task A") {
		t.Errorf("expected error to mention title 'Task A', got: %v", err)
	}
	if !strings.Contains(err.Error(), "Task B") {
		t.Errorf("expected error to mention title 'Task B', got: %v", err)
	}
	if !strings.Contains(err.Error(), "groupA/042-task-a.md") {
		t.Errorf("expected error to mention file path, got: %v", err)
	}
	if !strings.Contains(err.Error(), "groupB/042-task-b.md") {
		t.Errorf("expected error to mention file path, got: %v", err)
	}
	if !strings.Contains(err.Error(), "taskmd deduplicate") {
		t.Errorf("expected error to mention deduplicate command, got: %v", err)
	}
}

func TestResolveTask_DuplicateIDError_ByTitle(t *testing.T) {
	tasks := []*model.Task{
		{ID: "042", Title: "Task A", FilePath: "groupA/042-task-a.md"},
		{ID: "042", Title: "Task B", FilePath: "groupB/042-task-b.md"},
		{ID: "001", Title: "Unique Task", FilePath: "001-unique.md"},
	}

	// Matching by title should still detect duplicates on the resolved task's ID
	_, err := resolveTask("Task A", tasks, true, 0.6)
	if err == nil {
		t.Fatal("expected error for duplicate ID when matched by title, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate task ID") {
		t.Errorf("expected 'duplicate task ID' error, got: %v", err)
	}
}

func TestResolveTask_UniqueID(t *testing.T) {
	tasks := []*model.Task{
		{ID: "042", Title: "Task A", FilePath: "groupA/042-task-a.md"},
		{ID: "042", Title: "Task B", FilePath: "groupB/042-task-b.md"},
		{ID: "001", Title: "Unique Task", FilePath: "001-unique.md"},
	}

	task, err := resolveTask("001", tasks, true, 0.6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil {
		t.Fatal("expected task, got nil")
	}
	if task.ID != "001" {
		t.Errorf("expected ID 001, got %s", task.ID)
	}
}

func TestRunGet_DuplicateIDError(t *testing.T) {
	repo := &taskRepo{t: t, Dir: createDuplicateTestFiles(t)}

	err := repo.Run("get", "042").Err
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate task ID") {
		t.Errorf("expected 'duplicate task ID' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "042") {
		t.Errorf("expected error to mention ID 042, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Task A") {
		t.Errorf("expected error to mention title 'Task A', got: %v", err)
	}
	if !strings.Contains(err.Error(), "Task B") {
		t.Errorf("expected error to mention title 'Task B', got: %v", err)
	}
}

func TestRunGet_DuplicateID_UniqueTaskWorks(t *testing.T) {
	repo := &taskRepo{t: t, Dir: createDuplicateTestFiles(t)}

	output := getStdout(t, repo, "001")
	if !strings.Contains(output, "Unique Task") {
		t.Error("expected unique task to still work")
	}
}

func TestRunStatus_DuplicateIDError(t *testing.T) {
	tmpDir := createDuplicateTestFiles(t)
	statusFormat = "text"
	statusExact = false
	statusThreshold = 0.6
	statusMinimal = false
	statusStatusline = false
	statusScope = ""
	taskDir = tmpDir

	err := runStatusSingle("042")
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate task ID") {
		t.Errorf("expected 'duplicate task ID' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "042") {
		t.Errorf("expected error to mention ID 042, got: %v", err)
	}
}

func TestRunSet_DuplicateIDError(t *testing.T) {
	tmpDir := createDuplicateTestFiles(t)

	taskDir = tmpDir
	setTaskID = "042"
	setStatus = "completed"
	setPriority = ""
	setEffort = ""
	setType = ""
	setOwner = ""
	setParent = ""
	setDone = false
	setDryRun = false
	setVerify = false
	setAddTags = nil
	setRemoveTags = nil
	setAddPRs = nil
	setRemovePRs = nil
	setAddTouches = nil
	setRemoveTouches = nil
	setDependsOn = ""

	err := runSet(setCmd, []string{"042"})
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "refusing to modify") {
		t.Errorf("expected 'refusing to modify' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "042") {
		t.Errorf("expected error to mention ID 042, got: %v", err)
	}
}

func TestRunRm_DuplicateIDError(t *testing.T) {
	tmpDir := createDuplicateTestFiles(t)

	taskDir = tmpDir
	rmForce = true
	rmDryRun = false

	err := runRm(rmCmd, []string{"042"})
	if err == nil {
		t.Fatal("expected error for duplicate ID, got nil")
	}
	if !strings.Contains(err.Error(), "refusing to delete") {
		t.Errorf("expected 'refusing to delete' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "042") {
		t.Errorf("expected error to mention ID 042, got: %v", err)
	}
}
