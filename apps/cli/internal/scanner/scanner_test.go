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
	scanner := NewScanner(tmpDir, false)
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

	scanner := NewScanner(tmpDir, false)
	result, err := scanner.Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should not find any tasks (hidden directory is skipped)
	if len(result.Tasks) != 0 {
		t.Errorf("Expected 0 tasks (hidden dir should be skipped), got %d", len(result.Tasks))
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
