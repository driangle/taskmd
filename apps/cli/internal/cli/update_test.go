package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/taskfile"
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

func createMultilineTagTestFile(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	content := `---
id: "010"
title: "Multiline tags task"
status: pending
priority: high
effort: small
dependencies: []
tags:
  - backend
  - api
created: 2026-02-08
---

# Multiline tags task

Task with multiline YAML tags.
`
	path := filepath.Join(tmpDir, "010-multiline.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return tmpDir
}

func resetUpdateFlags() {
	updateTaskID = ""
	updateStatus = ""
	updatePriority = ""
	updateEffort = ""
	updateDone = false
	updateAddTags = nil
	updateRemoveTags = nil
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
	statuses := []string{"pending", "in-progress", "completed", "blocked", "cancelled"}
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

func TestUpdate_CancelledStatus(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateStatus = "cancelled"

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error when setting status to cancelled: %v", err)
	}

	if !strings.Contains(output, "status: in-progress -> cancelled") {
		t.Errorf("Expected status change to cancelled in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), "status: cancelled") {
		t.Error("Expected file to contain status: cancelled")
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
		name      string
		lines     []string
		wantOpen  int
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
			open, closeIdx := taskfile.FindFrontmatterBounds(tt.lines)
			if open != tt.wantOpen || closeIdx != tt.wantClose {
				t.Errorf("findFrontmatterBounds() = (%d, %d), want (%d, %d)",
					open, closeIdx, tt.wantOpen, tt.wantClose)
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

func TestUpdate_AddSingleTag(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateAddTags = []string{"new-tag"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [infra] -> [infra, new-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra", "new-tag"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestUpdate_AddMultipleTags(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateAddTags = []string{"tag-a", "tag-b"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [infra] -> [infra, tag-a, tag-b]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra", "tag-a", "tag-b"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestUpdate_RemoveTag(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateRemoveTags = []string{"security"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [backend, security] -> [backend]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), `tags: ["backend"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestUpdate_AddAndRemoveTag(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateAddTags = []string{"new-feature"}
	updateRemoveTags = []string{"security"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [backend, security] -> [backend, new-feature]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "002-auth.md"))
	if !strings.Contains(string(content), `tags: ["backend", "new-feature"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", string(content))
	}
}

func TestUpdate_AddDuplicateTag(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateAddTags = []string{"infra"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tags should remain unchanged since "infra" already exists.
	if !strings.Contains(output, "tags: [infra] -> [infra]") {
		t.Errorf("Expected no-op tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra"]`) {
		t.Errorf("Expected tags to remain unchanged, got:\n%s", string(content))
	}
}

func TestUpdate_RemoveNonexistentTag(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateRemoveTags = []string{"nonexistent"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Tags should remain unchanged since "nonexistent" isn't present.
	if !strings.Contains(output, "tags: [infra] -> [infra]") {
		t.Errorf("Expected no-op tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	if !strings.Contains(string(content), `tags: ["infra"]`) {
		t.Errorf("Expected tags to remain unchanged, got:\n%s", string(content))
	}
}

func TestUpdate_TagOnlyUpdate(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateAddTags = []string{"new-tag"}

	// Should NOT produce "nothing to update" error.
	_, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("tag-only update should succeed, got error: %v", err)
	}
}

func TestUpdate_TagsWithOtherFlags(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateStatus = "completed"
	updateAddTags = []string{"done-tag"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "status: pending -> completed") {
		t.Error("Expected status change in output")
	}
	if !strings.Contains(output, "tags: [infra] -> [infra, done-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)
	if !strings.Contains(fileStr, "status: completed") {
		t.Error("Expected file to contain updated status")
	}
	if !strings.Contains(fileStr, `tags: ["infra", "done-tag"]`) {
		t.Errorf("Expected file to contain updated tags, got:\n%s", fileStr)
	}
}

func TestUpdate_TagsPreservedFormat(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "001"
	updateAddTags = []string{"extra"}

	_, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-setup.md"))
	fileStr := string(content)

	// Inline format should stay inline.
	if !strings.Contains(fileStr, `tags: ["infra", "extra"]`) {
		t.Errorf("Expected inline tag format to be preserved, got:\n%s", fileStr)
	}

	// Other fields should be preserved.
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be preserved")
	}
	if !strings.Contains(fileStr, "# Setup project") {
		t.Error("Expected body to be preserved")
	}
}

func TestUpdate_MultilineTagFormat(t *testing.T) {
	tmpDir := createMultilineTagTestFile(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "010"
	updateAddTags = []string{"new-tag"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "tags: [backend, api] -> [backend, api, new-tag]") {
		t.Errorf("Expected tag change in output, got: %s", output)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "010-multiline.md"))
	fileStr := string(content)

	// Multiline format should stay multiline.
	if !strings.Contains(fileStr, "tags:\n  - backend\n  - api\n  - new-tag") {
		t.Errorf("Expected multiline tag format to be preserved, got:\n%s", fileStr)
	}

	// Other fields should be preserved.
	if !strings.Contains(fileStr, "status: pending") {
		t.Error("Expected status to be preserved")
	}
	if !strings.Contains(fileStr, "# Multiline tags task") {
		t.Error("Expected body to be preserved")
	}
}

func TestUpdate_MultilineTagRemove(t *testing.T) {
	tmpDir := createMultilineTagTestFile(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "010"
	updateRemoveTags = []string{"api"}

	_, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "010-multiline.md"))
	fileStr := string(content)

	if !strings.Contains(fileStr, "tags:\n  - backend\ncreated:") {
		t.Errorf("Expected multiline format with 'api' removed, got:\n%s", fileStr)
	}
}

func TestUpdate_TagConfirmationOutput(t *testing.T) {
	tmpDir := createUpdateTestFiles(t)
	resetUpdateFlags()
	dir = tmpDir
	updateTaskID = "002"
	updateAddTags = []string{"feature"}
	updateRemoveTags = []string{"security"}

	output, err := captureUpdateOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Updated task 002") {
		t.Error("Expected confirmation message with task ID")
	}
	if !strings.Contains(output, "tags: [backend, security] -> [backend, feature]") {
		t.Errorf("Expected formatted tag change, got: %s", output)
	}
}

func TestComputeNewTags(t *testing.T) {
	tests := []struct {
		name       string
		current    []string
		addTags    []string
		removeTags []string
		want       []string
	}{
		{
			name:    "add to empty",
			current: nil,
			addTags: []string{"a", "b"},
			want:    []string{"a", "b"},
		},
		{
			name:       "remove from list",
			current:    []string{"a", "b", "c"},
			removeTags: []string{"b"},
			want:       []string{"a", "c"},
		},
		{
			name:       "add and remove",
			current:    []string{"a", "b"},
			addTags:    []string{"c"},
			removeTags: []string{"a"},
			want:       []string{"b", "c"},
		},
		{
			name:    "add duplicate is no-op",
			current: []string{"a", "b"},
			addTags: []string{"a"},
			want:    []string{"a", "b"},
		},
		{
			name:       "remove nonexistent is no-op",
			current:    []string{"a"},
			removeTags: []string{"z"},
			want:       []string{"a"},
		},
		{
			name:       "remove all tags",
			current:    []string{"a"},
			removeTags: []string{"a"},
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := taskfile.ComputeNewTags(tt.current, tt.addTags, tt.removeTags)
			if len(got) != len(tt.want) {
				t.Fatalf("computeNewTags() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("computeNewTags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
