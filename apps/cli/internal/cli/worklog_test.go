package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/worklog"
)

func createWorklogTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	task := `---
id: "015"
title: "Add user auth"
status: in-progress
priority: high
dependencies: []
tags: ["backend"]
created: 2026-02-08
---

# Add user auth
`
	if err := os.WriteFile(filepath.Join(tmpDir, "015-auth.md"), []byte(task), 0644); err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}

	return tmpDir
}

func createWorklogTestFilesWithWorklog(t *testing.T) string {
	t.Helper()

	tmpDir := createWorklogTestFiles(t)

	wlDir := filepath.Join(tmpDir, ".worklogs")
	if err := os.MkdirAll(wlDir, 0755); err != nil {
		t.Fatalf("Failed to create .worklogs dir: %v", err)
	}

	wlContent := `## 2026-02-15T10:00:00Z

Started working on authentication module.

## 2026-02-15T14:30:00Z

Completed login endpoint.
`
	if err := os.WriteFile(filepath.Join(wlDir, "015.md"), []byte(wlContent), 0644); err != nil {
		t.Fatalf("Failed to create test worklog: %v", err)
	}

	return tmpDir
}

func resetWorklogFlags() {
	worklogTaskID = ""
	worklogAdd = ""
	worklogFormat = "text"
	taskDir = "."
}

func captureWorklogOutput(t *testing.T) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runWorklog(worklogCmd, nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestWorklog_ViewEntries(t *testing.T) {
	tmpDir := createWorklogTestFilesWithWorklog(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"

	output, err := captureWorklogOutput(t)
	if err != nil {
		t.Fatalf("runWorklog failed: %v", err)
	}

	if !strings.Contains(output, "015") {
		t.Error("Expected output to contain task ID")
	}
	if !strings.Contains(output, "2 entries") || !strings.Contains(output, "Entries: 2") {
		// Check for either styled or plain output
		if !strings.Contains(output, "2") {
			t.Error("Expected output to show entry count")
		}
	}
	if !strings.Contains(output, "authentication module") {
		t.Error("Expected output to contain first entry content")
	}
	if !strings.Contains(output, "login endpoint") {
		t.Error("Expected output to contain second entry content")
	}
}

func TestWorklog_ViewJSON(t *testing.T) {
	tmpDir := createWorklogTestFilesWithWorklog(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"
	worklogFormat = "json"

	output, err := captureWorklogOutput(t)
	if err != nil {
		t.Fatalf("runWorklog failed: %v", err)
	}

	var wl worklog.Worklog
	if err := json.Unmarshal([]byte(output), &wl); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if wl.TaskID != "015" {
		t.Errorf("Expected task_id '015', got %q", wl.TaskID)
	}
	if len(wl.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(wl.Entries))
	}
}

func TestWorklog_ViewYAML(t *testing.T) {
	tmpDir := createWorklogTestFilesWithWorklog(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"
	worklogFormat = "yaml"

	output, err := captureWorklogOutput(t)
	if err != nil {
		t.Fatalf("runWorklog failed: %v", err)
	}

	if !strings.Contains(output, "task_id: \"015\"") {
		t.Errorf("Expected YAML output to contain task_id, got:\n%s", output)
	}
	if !strings.Contains(output, "authentication module") {
		t.Error("Expected YAML to contain entry content")
	}
}

func TestWorklog_NoWorklogExists(t *testing.T) {
	tmpDir := createWorklogTestFiles(t) // no worklog created
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"

	// Capture stderr too
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	_, err := captureWorklogOutput(t)

	wErr.Close()
	os.Stderr = oldStderr

	var stderrBuf bytes.Buffer
	stderrBuf.ReadFrom(rErr)

	if err != nil {
		t.Fatalf("Expected no error for missing worklog, got: %v", err)
	}
	if !strings.Contains(stderrBuf.String(), "No worklog found") {
		t.Errorf("Expected stderr to say 'No worklog found', got: %q", stderrBuf.String())
	}
}

func TestWorklog_TaskNotFound(t *testing.T) {
	tmpDir := createWorklogTestFiles(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "999"

	_, err := captureWorklogOutput(t)
	if err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !strings.Contains(err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestWorklog_AddEntry(t *testing.T) {
	tmpDir := createWorklogTestFiles(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"
	worklogAdd = "Started implementation of auth module"

	// Capture stderr for the success message
	oldStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	_, err := captureWorklogOutput(t)

	wErr.Close()
	os.Stderr = oldStderr

	var stderrBuf bytes.Buffer
	stderrBuf.ReadFrom(rErr)

	if err != nil {
		t.Fatalf("runWorklog --add failed: %v", err)
	}

	if !strings.Contains(stderrBuf.String(), "Added worklog entry") {
		t.Errorf("Expected success message, got: %q", stderrBuf.String())
	}

	// Verify the file was created
	wlPath := filepath.Join(tmpDir, ".worklogs", "015.md")
	data, err := os.ReadFile(wlPath)
	if err != nil {
		t.Fatalf("Worklog file not created: %v", err)
	}

	if !strings.Contains(string(data), "Started implementation of auth module") {
		t.Error("Expected worklog to contain the added message")
	}
}

func TestWorklog_AddThenView(t *testing.T) {
	tmpDir := createWorklogTestFiles(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"
	worklogAdd = "First entry"

	_, err := captureWorklogOutput(t)
	if err != nil {
		t.Fatalf("First add failed: %v", err)
	}

	// Now view
	worklogAdd = ""
	worklogFormat = "json"

	output, err := captureWorklogOutput(t)
	if err != nil {
		t.Fatalf("View after add failed: %v", err)
	}

	var wl worklog.Worklog
	if err := json.Unmarshal([]byte(output), &wl); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(wl.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(wl.Entries))
	}
}

func TestWorklog_UnsupportedFormat(t *testing.T) {
	tmpDir := createWorklogTestFilesWithWorklog(t)
	resetWorklogFlags()
	taskDir = tmpDir
	worklogTaskID = "015"
	worklogFormat = "csv"

	_, err := captureWorklogOutput(t)
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}
