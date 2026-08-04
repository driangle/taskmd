package cli

import (
	"encoding/json"
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
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Test Task"
status: pending
priority: high
---
# Test Task
`,
	})

	// Use --task-dir flag (no positional arg) to verify the flag is honored.
	res := repo.Run("list", "--task-dir", repo.Dir, "--format", "json", "--columns", "id,title,status,priority")
	if res.Err != nil {
		t.Fatalf("runList with --dir flag failed: %v", res.Err)
	}

	// Verify output contains the task
	var tasks []map[string]any
	if err := json.Unmarshal([]byte(res.Stdout), &tasks); err != nil {
		t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, res.Stdout)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0]["id"] != "001" {
		t.Errorf("expected task id=001, got %v", tasks[0]["id"])
	}
}
