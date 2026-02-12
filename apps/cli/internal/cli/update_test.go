package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createUpdateTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	tasks := map[string]string{
		"001-setup.md": `---
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

Initial project setup with build tooling.
`,
		"002-auth.md": `---
id: "002"
title: "Implement authentication"
status: in-progress
priority: critical
effort: large
dependencies: ["001"]
tags: ["backend", "security"]
created: 2026-02-08
---

# Implement authentication

Add JWT-based auth with refresh tokens.
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: blocked
priority: medium
effort: medium
dependencies: ["002"]
tags: ["frontend"]
created: 2026-02-08
---

# Build UI components

Create reusable component library.
`,
	}

	for filename, content := range tasks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return tmpDir
}

func resetUpdateFlags() {
	updateTaskID = ""
	updateStatus = ""
	updatePriority = ""
	updateEffort = ""
	updateDone = false
	dir = "."
}

func captureUpdateOutput(t *testing.T) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runUpdate(updateCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestUpdate_Status(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateStatus = "completed"

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation message")
	}
	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected status change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain updated status")
	}
}

func TestUpdate_Priority(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updatePriority = "low"

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "priority: high -> low") {
		t.Errorf("Expected priority change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "priority: low") {
		t.Error("Expected file to contain updated priority")
	}
}

func TestUpdate_Effort(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateEffort = "small"

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "effort: large -> small") {
		t.Errorf("Expected effort change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), "effort: small") {
		t.Error("Expected file to contain updated effort")
	}
}

func TestUpdate_DoneFlag(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateDone = true

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> completed") {
		t.Errorf("Expected --done to set status to completed, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), "status: completed") {
		t.Error("Expected file to contain completed status")
	}
}

func TestUpdate_MultipleFields(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "003"
	updateStatus = "in-progress"
	updatePriority = "critical"
	updateEffort = "large"

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: blocked -> in-progress") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "priority: medium -> critical") {
		t.Error("Expected priority change in output")
	}
	if !strings.Contains(output, "effort: medium -> large") {
		t.Error("Expected effort change in output")
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "003-ui.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: in-progress") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("Expected file to contain updated priority")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("Expected file to contain updated effort")
	}
}

func TestUpdate_AllValidStatuses(t *testing.T) {
	statuses := []string{"pending", "in-progress", "completed", "blocked"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			tmpDir := createUpdateTestFiles(t)
			resetUpdateFlags()
			dir = tmpDir
			updateTaskID = "001"
			updateStatus = status

			_, err := captureUpdateOutput(t)
			if err != nil {
				t.Fatalf("unexpected error for status %q: %v", status, err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
			if !strings.Contains(string(content), "status: "+status) {
				t.Errorf("Expected file to contain status: %s", status)
			}
		})
	}
}

func TestUpdate_AllValidPriorities(t *testing.T) {
	priorities := []string{"low", "medium", "high", "critical"}
	for _, priority := range priorities {
		t.Run(priority, func(t *testing.T) {
			tmpDir := createUpdateTestFiles(t)
			resetUpdateFlags()
			dir = tmpDir
			updateTaskID = "001"
			updatePriority = priority

			_, err := captureUpdateOutput(t)
			if err != nil {
				t.Fatalf("unexpected error for priority %q: %v", priority, err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
			if !strings.Contains(string(content), "priority: "+priority) {
				t.Errorf("Expected file to contain priority: %s", priority)
			}
		})
	}
}

func TestUpdate_AllValidEfforts(t *testing.T) {
	efforts := []string{"small", "medium", "large"}
	for _, effort := range efforts {
		t.Run(effort, func(t *testing.T) {
			tmpDir := createUpdateTestFiles(t)
			resetUpdateFlags()
			dir = tmpDir
			updateTaskID = "002"
			updateEffort = effort

			_, err := captureUpdateOutput(t)
			if err != nil {
				t.Fatalf("unexpected error for effort %q: %v", effort, err)
			}

			content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
			if !strings.Contains(string(content), "effort: "+effort) {
				t.Errorf("Expected file to contain effort: %s", effort)
			}
		})
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateStatus = "invalid"

	_, err := captureUpdateOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("Expected 'invalid status' error, got: %v", err)
	}
}

func TestUpdate_InvalidPriority(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updatePriority = "urgent"

	_, err := captureUpdateOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid priority")
	}
	if !strings.Contains(err.Error(), "invalid priority") {
		t.Errorf("Expected 'invalid priority' error, got: %v", err)
	}
}

func TestUpdate_InvalidEffort(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateEffort = "huge"

	_, err := captureUpdateOutput(t)
	if err == nil {
		t.Fatal("Expected error for invalid effort")
	}
	if !strings.Contains(err.Error(), "invalid effort") {
		t.Errorf("Expected 'invalid effort' error, got: %v", err)
	}
}

func TestUpdate_TaskNotFound(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "nonexistent"
	updateStatus = "completed"

	_, err := captureUpdateOutput(t)
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestUpdate_NoFlagsProvided(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"

	_, err := captureUpdateOutput(t)
	if err == nil {
		t.Fatal("Expected error when no update flags provided")
	}
	if !strings.Contains(err.Error(), "nothing to update") {
		t.Errorf("Expected 'nothing to update' error, got: %v", err)
	}
}

func TestUpdate_DoneWithStatusMutuallyExclusive(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateDone = true
	updateStatus = "blocked"

	// Mark the --status flag as changed to simulate CLI usage
	updateCmd.Flags().Set("status", "blocked")
	defer func() {
		// Reset the changed state by creating a fresh flag set lookup
		updateCmd.Flags().Set("status", "")
	}()

	_, err := captureUpdateOutput(t)
	if err == nil {
		t.Fatal("Expected error when --done and --status are both set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("Expected 'mutually exclusive' error, got: %v", err)
	}
}

func TestUpdate_BodyPreserved(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateStatus = "completed"

	_, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	fileStr := string(content)

	if !strings.Contains(fileStr, "# Implement authentication") {
		t.Error("Expected body heading to be preserved")
	}
	if !strings.Contains(fileStr, "Add JWT-based auth with refresh tokens.") {
		t.Error("Expected body content to be preserved")
	}
}

func TestUpdate_OtherFieldsPreserved(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateStatus = "completed"

	_, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	fileStr := string(content)

	// Verify non-updated fields are preserved
	if !strings.Contains(fileStr, "priority: critical") {
		t.Error("Expected priority to be preserved")
	}
	if !strings.Contains(fileStr, "effort: large") {
		t.Error("Expected effort to be preserved")
	}
	if !strings.Contains(fileStr, `dependencies: ["001"]`) {
		t.Error("Expected dependencies to be preserved")
	}
	if !strings.Contains(fileStr, `tags: ["backend", "security"]`) {
		t.Error("Expected tags to be preserved")
	}
	if !strings.Contains(fileStr, "created: 2026-02-08") {
		t.Error("Expected created date to be preserved")
	}
}

func TestUpdate_FrontmatterBounds(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		wantOpen int
		wantClose int
	}{
		{
			name:      "standard frontmatter",
			lines:     []string{"---", "id: foo", "---", "body"},
			wantOpen:  0,
			wantClose: 2,
		},
		{
			name:      "no frontmatter",
			lines:     []string{"# Just a heading", "body"},
			wantOpen:  -1,
			wantClose: -1,
		},
		{
			name:      "unclosed frontmatter",
			lines:     []string{"---", "id: foo"},
			wantOpen:  -1,
			wantClose: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			open, close := findFrontmatterBounds(tt.lines)
			if open != tt.wantOpen || close != tt.wantClose {
				t.Errorf("findFrontmatterBounds() = (%d, %d), want (%d, %d)",
					open, close, tt.wantOpen, tt.wantClose)
			}
		})
	}
}

func TestUpdate_MatchByTitle(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "Setup project"
	updateStatus = "completed"

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 001") {
		t.Error("Expected confirmation for task found by title match")
	}
}
