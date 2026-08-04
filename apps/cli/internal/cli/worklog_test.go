package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/worklog"
)

// Worklog fixtures are command-specific (a single in-progress task plus an
// optional pre-seeded .worklogs entry), so they stay inline here.
const worklogTaskFile = `---
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

const worklogEntries = `## 2026-02-15T10:00:00Z

Started working on authentication module.

## 2026-02-15T14:30:00Z

Completed login endpoint.
`

func newWorklogRepo(t *testing.T) *taskRepo {
	t.Helper()
	return newTaskRepo(t, map[string]string{"015-auth.md": worklogTaskFile})
}

func newWorklogRepoWithLog(t *testing.T) *taskRepo {
	t.Helper()
	repo := newWorklogRepo(t)
	repo.Write(".worklogs/015.md", worklogEntries)
	return repo
}

// worklogStdout runs `worklog <args...>`, fails on error, returns stdout.
func worklogStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"worklog"}, args...)...)
	if res.Err != nil {
		t.Fatalf("worklog %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestWorklog_ViewEntries(t *testing.T) {
	repo := newWorklogRepoWithLog(t)

	output := worklogStdout(t, repo, "015")

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
	repo := newWorklogRepoWithLog(t)

	output := worklogStdout(t, repo, "015", "--format", "json")

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
	repo := newWorklogRepoWithLog(t)

	output := worklogStdout(t, repo, "015", "--format", "yaml")

	if !strings.Contains(output, "task_id: \"015\"") {
		t.Errorf("Expected YAML output to contain task_id, got:\n%s", output)
	}
	if !strings.Contains(output, "authentication module") {
		t.Error("Expected YAML to contain entry content")
	}
}

func TestWorklog_NoWorklogExists(t *testing.T) {
	repo := newWorklogRepo(t) // no worklog created

	res := repo.Run("worklog", "015")
	if res.Err != nil {
		t.Fatalf("Expected no error for missing worklog, got: %v", res.Err)
	}
	if !strings.Contains(res.Stderr, "No worklog found") {
		t.Errorf("Expected stderr to say 'No worklog found', got: %q", res.Stderr)
	}
}

func TestWorklog_TaskNotFound(t *testing.T) {
	repo := newWorklogRepo(t)

	res := repo.Run("worklog", "999")
	if res.Err == nil {
		t.Fatal("Expected error for non-existent task")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestWorklog_AddEntry(t *testing.T) {
	repo := newWorklogRepo(t)

	res := repo.Run("worklog", "015", "--add", "Started implementation of auth module")
	if res.Err != nil {
		t.Fatalf("runWorklog --add failed: %v", res.Err)
	}

	if !strings.Contains(res.Stderr, "Added worklog entry") {
		t.Errorf("Expected success message, got: %q", res.Stderr)
	}

	// Verify the file was created
	data, err := os.ReadFile(repo.Path(".worklogs/015.md"))
	if err != nil {
		t.Fatalf("Worklog file not created: %v", err)
	}

	if !strings.Contains(string(data), "Started implementation of auth module") {
		t.Error("Expected worklog to contain the added message")
	}
}

func TestWorklog_AddThenView(t *testing.T) {
	repo := newWorklogRepo(t)

	res := repo.Run("worklog", "015", "--add", "First entry")
	if res.Err != nil {
		t.Fatalf("First add failed: %v", res.Err)
	}

	// Now view
	output := worklogStdout(t, repo, "015", "--format", "json")

	var wl worklog.Worklog
	if err := json.Unmarshal([]byte(output), &wl); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(wl.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(wl.Entries))
	}
}

func TestWorklog_UnsupportedFormat(t *testing.T) {
	repo := newWorklogRepoWithLog(t)

	res := repo.Run("worklog", "015", "--format", "csv")
	if res.Err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", res.Err)
	}
}
