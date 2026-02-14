package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveScanDir_NoArgs(t *testing.T) {
	taskDir = "."
	got := ResolveScanDir([]string{})
	if got != "." {
		t.Errorf("ResolveScanDir([]) = %q, want %q", got, ".")
	}
}

func TestResolveScanDir_WithPositionalArg(t *testing.T) {
	taskDir = "."
	got := ResolveScanDir([]string{"/some/path"})
	if got != "/some/path" {
		t.Errorf("ResolveScanDir([\"/some/path\"]) = %q, want %q", got, "/some/path")
	}
}

func TestResolveScanDir_DirFlagUsed(t *testing.T) {
	taskDir = "/custom/dir"
	got := ResolveScanDir([]string{})
	if got != "/custom/dir" {
		t.Errorf("ResolveScanDir([]) with dir=%q = %q, want %q", "/custom/dir", got, "/custom/dir")
	}
	taskDir = "." // reset
}

func TestResolveScanDir_PositionalArgOverridesFlag(t *testing.T) {
	taskDir = "/flag/dir"
	got := ResolveScanDir([]string{"/positional/dir"})
	if got != "/positional/dir" {
		t.Errorf("ResolveScanDir([\"/positional/dir\"]) with dir=%q = %q, want %q", "/flag/dir", got, "/positional/dir")
	}
	taskDir = "." // reset
}

func TestDirFlag_ListIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal task file
	taskContent := `---
id: "001"
title: "Test Task"
status: pending
priority: high
---
# Test Task
`
	err := os.WriteFile(filepath.Join(tmpDir, "001.md"), []byte(taskContent), 0644)
	if err != nil {
		t.Fatalf("failed to write test task file: %v", err)
	}

	// Save and reset flags
	oldDir := taskDir
	oldListFormat := listFormat
	oldFilters := listFilters
	oldSort := listSort
	oldColumns := listColumns
	defer func() {
		taskDir = oldDir
		listFormat = oldListFormat
		listFilters = oldFilters
		listSort = oldSort
		listColumns = oldColumns
	}()

	// Use --task-dir flag (no positional arg)
	taskDir = tmpDir
	listFormat = "json"
	listFilters = []string{}
	listSort = ""
	listColumns = "id,title,status,priority"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runList(listCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("runList with --dir flag failed: %v", err)
	}

	// Verify output contains the task
	var tasks []map[string]any
	if err := json.Unmarshal([]byte(output), &tasks); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, output)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0]["id"] != "001" {
		t.Errorf("expected task id=001, got %v", tasks[0]["id"])
	}
}
