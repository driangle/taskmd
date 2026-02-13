package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Helpers ---

func captureListOutput(t *testing.T, scanDir string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, []string{scanDir})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

// --- Empty directory ---

func TestEdge_EmptyDirectory_List(t *testing.T) {
	tmpDir := t.TempDir()
	resetListFlags()
	noColor = true

	output, err := captureListOutput(t, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "No tasks found") {
		t.Errorf("expected 'No tasks found' message, got: %q", output)
	}
}

func TestEdge_EmptyDirectory_Board(t *testing.T) {
	tmpDir := t.TempDir()
	resetBoardFlags()

	output := captureBoardOutput(t, tmpDir)
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for board on empty dir, got: %q", output)
	}
}

// --- Malformed frontmatter ---

func TestEdge_MalformedFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "001"
title: "Broken YAML
status: pending
---

body
`
	writeTestFile(t, tmpDir, "001-broken.md", content)

	resetListFlags()
	noColor = true
	// Should not crash; scanner may skip malformed files
	_, _ = captureListOutput(t, tmpDir)
}

// --- Empty file ---

func TestEdge_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, tmpDir, "empty.md", "")

	resetListFlags()
	noColor = true
	output, err := captureListOutput(t, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error on empty file: %v", err)
	}
	if !strings.Contains(output, "No tasks found") {
		t.Errorf("expected 'No tasks found' for empty file, got: %q", output)
	}
}

// --- File with frontmatter but no body ---

func TestEdge_FrontmatterNoBody(t *testing.T) {
	tmpDir := t.TempDir()
	content := `---
id: "001"
title: "No body task"
status: pending
priority: high
effort: small
dependencies: []
tags: []
created: 2026-02-08
---
`
	writeTestFile(t, tmpDir, "001-nobody.md", content)

	resetListFlags()
	noColor = true
	output, err := captureListOutput(t, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "001") {
		t.Errorf("expected task 001 in output, got: %q", output)
	}
}

// --- Non-existent directory ---

func TestEdge_NonExistentDirectory(t *testing.T) {
	resetListFlags()
	noColor = true
	_, err := captureListOutput(t, "/tmp/nonexistent-taskmd-dir-xyz-"+t.Name())
	// Scanner should either error or return empty results
	if err != nil {
		if !strings.Contains(err.Error(), "scan failed") {
			t.Errorf("expected 'scan failed' error, got: %v", err)
		}
	}
	// If no error, that's also acceptable (scanner returns empty)
}

// --- Invalid sort field ---

func TestEdge_InvalidSortField(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)
	resetListFlags()
	noColor = true
	listSort = "nonexistent"

	_, err := captureListOutput(t, tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid sort field")
	}
	if !strings.Contains(err.Error(), "invalid sort field") {
		t.Errorf("expected 'invalid sort field' in error, got: %v", err)
	}
}

// --- Invalid format value ---

func TestEdge_InvalidFormat(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)
	resetListFlags()
	noColor = true
	format = "xml"

	_, err := captureListOutput(t, tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' in error, got: %v", err)
	}

	// Reset format
	format = "table"
}

// --- Invalid filter syntax ---

func TestEdge_InvalidFilterSyntax(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)
	resetListFlags()
	noColor = true
	listFilters = []string{"status-pending"} // missing =

	_, err := captureListOutput(t, tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid filter syntax")
	}
	if !strings.Contains(err.Error(), "invalid filter format") {
		t.Errorf("expected 'invalid filter format' error, got: %v", err)
	}
}

// --- Set command: invalid enum values with suggestions ---

func TestEdge_Set_InvalidStatusSuggestion(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)
	resetSetFlags()
	dir = tmpDir
	setTaskID = "001"
	setStatus = "pnding"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `did you mean "pending"`) {
		t.Errorf("expected suggestion for 'pending', got: %s", errMsg)
	}
}

func TestEdge_Set_InvalidPrioritySuggestion(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)
	resetSetFlags()
	dir = tmpDir
	setTaskID = "001"
	setPriority = "hgh"

	_, err := captureSetOutput(t)
	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "invalid priority") {
		t.Errorf("expected 'invalid priority' error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `did you mean "high"`) {
		t.Errorf("expected suggestion for 'high', got: %s", errMsg)
	}
}

// --- Set command: dry-run ---

func TestEdge_Set_DryRunNoWrite(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)
	resetSetFlags()
	dir = tmpDir
	setTaskID = "001"
	setStatus = "completed"
	setDryRun = true
	defer func() { setDryRun = false }()

	output, err := captureSetOutput(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Dry run") {
		t.Error("expected 'Dry run' message in output")
	}

	// Verify file was NOT changed
	content, _ := os.ReadFile(filepath.Join(tmpDir, "001-task.md"))
	if strings.Contains(string(content), "status: completed") {
		t.Error("expected file NOT to be modified during dry run")
	}
}

// --- Graph: conflicting flags ---

func TestEdge_Graph_UpstreamAndDownstream(t *testing.T) {
	tmpDir := createEdgeCaseTestFiles(t)

	// Reset graph flags
	graphFormat = "ascii"
	graphRoot = ""
	graphFocus = ""
	graphUpstream = true
	graphDownstream = true
	graphExcludeStatus = []string{"completed"}
	graphAll = false
	graphOut = ""
	graphFilters = []string{}

	err := runGraph(graphCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("expected error when using both --upstream and --downstream")
	}
	if !strings.Contains(err.Error(), "cannot use both") {
		t.Errorf("expected 'cannot use both' error, got: %v", err)
	}
}

// --- helpers ---

func createEdgeCaseTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	content := `---
id: "001"
title: "Edge case task"
status: pending
priority: high
effort: small
dependencies: []
tags: ["test"]
created: 2026-02-08
---

# Edge case task

A basic task for edge case testing.
`
	writeTestFile(t, tmpDir, "001-task.md", content)
	return tmpDir
}

func writeTestFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filename, err)
	}
}
