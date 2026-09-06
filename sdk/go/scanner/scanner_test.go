package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_Scan(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test task files
	testFiles := map[string]string{
		"task1.md": `---
id: "001"
title: "Task 1"
status: pending
priority: high
---
# Task 1`,
		"subdir/task2.md": `---
id: "002"
title: "Task 2"
status: completed
priority: low
---
# Task 2`,
		"subdir/nested/task3.md": `---
id: "003"
title: "Task 3"
status: in-progress
priority: medium
---
# Task 3`,
		"README.md": `# Not a task file
This is just a regular markdown file without frontmatter.`,
		"invalid.md": `---
missing: required fields
---
# Invalid task`,
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", fullPath, err)
		}
	}

	// Create scanner and scan
	scanner := NewScanner(tmpDir, false, nil)
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 3 valid tasks
	if len(result.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(result.Tasks))
	}

	// Verify task IDs
	foundIDs := make(map[string]bool)
	for _, task := range result.Tasks {
		foundIDs[task.ID] = true
	}

	expectedIDs := []string{"001", "002", "003"}
	for _, id := range expectedIDs {
		if !foundIDs[id] {
			t.Errorf("Expected to find task with ID %s", id)
		}
	}
}

func TestScanner_SkipsHiddenDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a task in a hidden directory
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}

	hiddenTask := filepath.Join(hiddenDir, "task.md")
	content := `---
id: "hidden"
title: "Hidden Task"
status: pending
---
# Hidden`
	if err := os.WriteFile(hiddenTask, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write hidden task: %v", err)
	}

	scanner := NewScanner(tmpDir, false, nil)
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should not find any tasks (hidden directory is skipped)
	if len(result.Tasks) != 0 {
		t.Errorf("Expected 0 tasks (hidden dir should be skipped), got %d", len(result.Tasks))
	}
}

func createTaskFile(t *testing.T, dir, id string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	content := "---\nid: \"" + id + "\"\ntitle: \"Task " + id + "\"\nstatus: pending\n---\n# Task " + id
	if err := os.WriteFile(filepath.Join(dir, id+".md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
}

func TestScanner_IgnoreConfiguredDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "active"), "001")
	createTaskFile(t, filepath.Join(tmpDir, "drafts"), "002")

	s := NewScanner(tmpDir, false, []string{"drafts"})
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(result.Tasks))
	}
	if result.Tasks[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", result.Tasks[0].ID)
	}
}

func TestScanner_IgnoreMultipleDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "active"), "001")
	createTaskFile(t, filepath.Join(tmpDir, "drafts"), "002")
	createTaskFile(t, filepath.Join(tmpDir, "templates"), "003")

	s := NewScanner(tmpDir, false, []string{"drafts", "templates"})
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(result.Tasks))
	}
	if result.Tasks[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", result.Tasks[0].ID)
	}
}

func TestScanner_IgnoreNestedDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "project", "active"), "001")
	createTaskFile(t, filepath.Join(tmpDir, "project", "drafts"), "002")

	s := NewScanner(tmpDir, false, []string{"drafts"})
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(result.Tasks))
	}
	if result.Tasks[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", result.Tasks[0].ID)
	}
}

func TestScanner_IgnoreWithHiddenDirsStillWorks(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "active"), "001")
	createTaskFile(t, filepath.Join(tmpDir, ".hidden"), "002")
	createTaskFile(t, filepath.Join(tmpDir, "drafts"), "003")

	s := NewScanner(tmpDir, false, []string{"drafts"})
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(result.Tasks))
	}
	if result.Tasks[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", result.Tasks[0].ID)
	}
}

func TestScanner_NoIgnoreConfig(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "active"), "001")
	createTaskFile(t, filepath.Join(tmpDir, "custom"), "002")

	// nil ignore list - only hardcoded defaults apply
	s := NewScanner(tmpDir, false, nil)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(result.Tasks))
	}
}

func TestScanner_DefaultSkipDirsStillApply(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "active"), "001")
	createTaskFile(t, filepath.Join(tmpDir, "node_modules"), "002")
	createTaskFile(t, filepath.Join(tmpDir, "vendor"), "003")

	// Even with custom ignore dirs, defaults still apply
	s := NewScanner(tmpDir, false, []string{"custom"})
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(result.Tasks))
	}
	if result.Tasks[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", result.Tasks[0].ID)
	}
}

func TestDeriveGroupFromPath(t *testing.T) {
	tests := []struct {
		name     string
		rootDir  string
		filePath string
		expected string
	}{
		{
			name:     "file in subdirectory",
			rootDir:  "/projects/tasks",
			filePath: "/projects/tasks/cli/task.md",
			expected: "cli",
		},
		{
			name:     "file in nested subdirectory",
			rootDir:  "/projects/tasks",
			filePath: "/projects/tasks/frontend/components/task.md",
			expected: "components",
		},
		{
			name:     "file in root directory",
			rootDir:  "/projects/tasks",
			filePath: "/projects/tasks/task.md",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deriveGroupFromPath(tt.rootDir, tt.filePath)
			if result != tt.expected {
				t.Errorf("deriveGroupFromPath() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func createArchivedTaskFile(t *testing.T, dir, id, status string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	content := "---\nid: \"" + id + "\"\ntitle: \"Task " + id + "\"\nstatus: " + status + "\n---\n# Task " + id
	if err := os.WriteFile(filepath.Join(dir, id+".md"), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}
}

func TestScanner_ScanArchive_FindsArchivedTasks(t *testing.T) {
	tmpDir := t.TempDir()

	// Active task
	createTaskFile(t, filepath.Join(tmpDir, "cli"), "001")
	// Archived tasks
	createArchivedTaskFile(t, filepath.Join(tmpDir, "archive"), "050", "completed")
	createArchivedTaskFile(t, filepath.Join(tmpDir, "archive"), "051", "completed")

	s := NewScanner(tmpDir, false, nil)
	archived, err := s.ScanArchive()
	if err != nil {
		t.Fatalf("ScanArchive failed: %v", err)
	}

	if len(archived) != 2 {
		t.Fatalf("Expected 2 archived tasks, got %d", len(archived))
	}

	ids := make(map[string]bool)
	for _, task := range archived {
		ids[task.ID] = true
	}
	if !ids["050"] || !ids["051"] {
		t.Errorf("Expected archived tasks 050 and 051, got %v", ids)
	}
}

func TestScanner_ScanArchive_NestedArchive(t *testing.T) {
	tmpDir := t.TempDir()

	// Archive nested in a subdirectory
	createArchivedTaskFile(t, filepath.Join(tmpDir, "cli", "archive"), "060", "completed")
	createTaskFile(t, filepath.Join(tmpDir, "cli"), "001")

	s := NewScanner(tmpDir, false, nil)
	archived, err := s.ScanArchive()
	if err != nil {
		t.Fatalf("ScanArchive failed: %v", err)
	}

	if len(archived) != 1 {
		t.Fatalf("Expected 1 archived task, got %d", len(archived))
	}
	if archived[0].ID != "060" {
		t.Errorf("Expected task 060, got %s", archived[0].ID)
	}
}

func TestScanner_ScanArchive_NoArchiveDir(t *testing.T) {
	tmpDir := t.TempDir()

	createTaskFile(t, filepath.Join(tmpDir, "cli"), "001")

	s := NewScanner(tmpDir, false, nil)
	archived, err := s.ScanArchive()
	if err != nil {
		t.Fatalf("ScanArchive failed: %v", err)
	}

	if len(archived) != 0 {
		t.Errorf("Expected 0 archived tasks, got %d", len(archived))
	}
}

func TestScanner_ScanArchive_EmptyArchive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty archive directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "archive"), 0755); err != nil {
		t.Fatalf("Failed to create archive dir: %v", err)
	}

	s := NewScanner(tmpDir, false, nil)
	archived, err := s.ScanArchive()
	if err != nil {
		t.Fatalf("ScanArchive failed: %v", err)
	}

	if len(archived) != 0 {
		t.Errorf("Expected 0 archived tasks from empty archive, got %d", len(archived))
	}
}

func TestScanner_SkippedSegment(t *testing.T) {
	s := NewScanner("/tmp/tasks", false, []string{"content"})

	tests := []struct {
		name    string
		relPath string
		want    string
	}{
		{"scannable", "cli", ""},
		{"nested scannable", "cli/backend", ""},
		{"configured ignore", "content", "content"},
		{"configured ignore nested", "a/content", "content"},
		{"configured ignore as parent", "content/a", "content"},
		{"default skip dir", "node_modules", "node_modules"},
		{"hidden dir", ".drafts", ".drafts"},
		{"hidden nested", "a/.drafts/b", ".drafts"},
		{"empty path", "", ""},
		{"dot path", ".", ""},
		{"parent traversal is not a hidden dir", "../cli", ""},
		{"substring is not a match", "contents", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.SkippedSegment(tt.relPath); got != tt.want {
				t.Errorf("SkippedSegment(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}
