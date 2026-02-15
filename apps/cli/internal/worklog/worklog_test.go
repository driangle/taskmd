package worklog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseWorklog_MultipleEntries(t *testing.T) {
	content := `## 2026-02-15T10:00:00Z

Started working on authentication module.

## 2026-02-15T14:30:00Z

Completed login endpoint. Blocked on task 012 for user model.
`
	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, "015.md")
	if err := os.WriteFile(wlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	wl, err := ParseWorklog(wlFile)
	if err != nil {
		t.Fatalf("ParseWorklog failed: %v", err)
	}

	if wl.TaskID != "015" {
		t.Errorf("Expected TaskID '015', got %q", wl.TaskID)
	}
	if len(wl.Entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(wl.Entries))
	}

	if wl.Entries[0].Timestamp.Hour() != 10 {
		t.Errorf("Expected first entry at 10:00, got %v", wl.Entries[0].Timestamp)
	}
	if !strings.Contains(wl.Entries[0].Content, "authentication module") {
		t.Errorf("Expected first entry content to contain 'authentication module', got %q", wl.Entries[0].Content)
	}

	if wl.Entries[1].Timestamp.Hour() != 14 {
		t.Errorf("Expected second entry at 14:30, got %v", wl.Entries[1].Timestamp)
	}
	if !strings.Contains(wl.Entries[1].Content, "Blocked on task 012") {
		t.Errorf("Expected second entry to mention blocker, got %q", wl.Entries[1].Content)
	}
}

func TestParseWorklog_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, "001.md")
	if err := os.WriteFile(wlFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	wl, err := ParseWorklog(wlFile)
	if err != nil {
		t.Fatalf("ParseWorklog failed: %v", err)
	}

	if len(wl.Entries) != 0 {
		t.Errorf("Expected 0 entries for empty file, got %d", len(wl.Entries))
	}
}

func TestParseWorklog_FileNotFound(t *testing.T) {
	_, err := ParseWorklog("/nonexistent/path/worklog.md")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestParseWorklog_MalformedTimestamp(t *testing.T) {
	content := `## not-a-timestamp

Some content here.

## 2026-02-15T10:00:00Z

Valid entry.
`
	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, "002.md")
	if err := os.WriteFile(wlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	wl, err := ParseWorklog(wlFile)
	if err != nil {
		t.Fatalf("ParseWorklog failed: %v", err)
	}

	// Malformed entry should be skipped, only valid one parsed
	if len(wl.Entries) != 1 {
		t.Fatalf("Expected 1 entry (skipping malformed), got %d", len(wl.Entries))
	}
	if !strings.Contains(wl.Entries[0].Content, "Valid entry") {
		t.Errorf("Expected valid entry content, got %q", wl.Entries[0].Content)
	}
}

func TestParseWorklog_NoEntries(t *testing.T) {
	content := "Just some random text without any timestamp headings.\n"
	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, "003.md")
	if err := os.WriteFile(wlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	wl, err := ParseWorklog(wlFile)
	if err != nil {
		t.Fatalf("ParseWorklog failed: %v", err)
	}

	if len(wl.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(wl.Entries))
	}
}

func TestParseWorklog_TimezoneOffset(t *testing.T) {
	content := "## 2026-02-15T10:00:00+05:30\n\nEntry with timezone offset.\n"
	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, "004.md")
	if err := os.WriteFile(wlFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	wl, err := ParseWorklog(wlFile)
	if err != nil {
		t.Fatalf("ParseWorklog failed: %v", err)
	}

	if len(wl.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(wl.Entries))
	}
	if wl.Entries[0].Timestamp.Hour() != 10 {
		t.Errorf("Expected hour 10, got %d", wl.Entries[0].Timestamp.Hour())
	}
}

func TestWorklogPath(t *testing.T) {
	tests := []struct {
		taskFile string
		taskID   string
		expected string
	}{
		{
			taskFile: "tasks/cli/015-auth.md",
			taskID:   "015",
			expected: filepath.Join("tasks", "cli", ".worklogs", "015.md"),
		},
		{
			taskFile: "tasks/020-frontend.md",
			taskID:   "020",
			expected: filepath.Join("tasks", ".worklogs", "020.md"),
		},
		{
			taskFile: "/abs/tasks/cli/042-task.md",
			taskID:   "cli-042",
			expected: filepath.Join("/abs", "tasks", "cli", ".worklogs", "cli-042.md"),
		},
	}

	for _, tt := range tests {
		result := WorklogPath(tt.taskFile, tt.taskID)
		if result != tt.expected {
			t.Errorf("WorklogPath(%q, %q) = %q, want %q", tt.taskFile, tt.taskID, result, tt.expected)
		}
	}
}

func TestAppendEntry_CreatesFileAndDir(t *testing.T) {
	tmpDir := t.TempDir()
	wlFile := filepath.Join(tmpDir, ".worklogs", "015.md")

	err := AppendEntry(wlFile, "First entry")
	if err != nil {
		t.Fatalf("AppendEntry failed: %v", err)
	}

	// Verify file was created
	data, err := os.ReadFile(wlFile)
	if err != nil {
		t.Fatalf("Failed to read created worklog: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "## ") {
		t.Error("Expected timestamp header in worklog")
	}
	if !strings.Contains(content, "First entry") {
		t.Error("Expected message in worklog")
	}
	// First entry should not start with a newline
	if strings.HasPrefix(content, "\n") {
		t.Error("First entry should not start with a leading newline")
	}
}

func TestAppendEntry_AppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	wlDir := filepath.Join(tmpDir, ".worklogs")
	if err := os.MkdirAll(wlDir, 0755); err != nil {
		t.Fatal(err)
	}

	wlFile := filepath.Join(wlDir, "015.md")
	existing := "## 2026-02-15T10:00:00Z\n\nExisting entry.\n"
	if err := os.WriteFile(wlFile, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	err := AppendEntry(wlFile, "New entry")
	if err != nil {
		t.Fatalf("AppendEntry failed: %v", err)
	}

	data, err := os.ReadFile(wlFile)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "Existing entry") {
		t.Error("Expected existing content to be preserved")
	}
	if !strings.Contains(content, "New entry") {
		t.Error("Expected new entry to be appended")
	}

	// Parse and verify both entries
	wl, err := ParseWorklog(wlFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(wl.Entries) != 2 {
		t.Errorf("Expected 2 entries after append, got %d", len(wl.Entries))
	}
}

func TestDeriveTaskID(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"tasks/cli/.worklogs/015.md", "015"},
		{"tasks/.worklogs/abc-042.md", "abc-042"},
		{"/abs/path/.worklogs/001.md", "001"},
	}

	for _, tt := range tests {
		result := deriveTaskID(tt.path)
		if result != tt.expected {
			t.Errorf("deriveTaskID(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	// File doesn't exist
	if Exists(filepath.Join(tmpDir, "nonexistent.md")) {
		t.Error("Expected Exists to return false for missing file")
	}

	// Create file
	wlFile := filepath.Join(tmpDir, "015.md")
	if err := os.WriteFile(wlFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	if !Exists(wlFile) {
		t.Error("Expected Exists to return true for existing file")
	}
}
