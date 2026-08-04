package cli

import (
	"os"
	"strings"
	"testing"
)

// edgeCaseFiles returns a single well-formed task used by the edge-case tests
// that exercise invalid flags and the set/graph commands against a valid repo.
func edgeCaseFiles() map[string]string {
	return map[string]string{
		"001-task.md": `---
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
`,
	}
}

// --- Empty directory ---

func TestEdge_EmptyDirectory_List(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("list")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "No tasks found") {
		t.Errorf("expected 'No tasks found' message, got: %q", res.Stdout)
	}
}

func TestEdge_EmptyDirectory_Board(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := boardStdout(t, repo)
	if strings.TrimSpace(output) != "" {
		t.Errorf("expected empty output for board on empty dir, got: %q", output)
	}
}

// --- Malformed frontmatter ---

func TestEdge_MalformedFrontmatter(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-broken.md": `---
id: "001"
title: "Broken YAML
status: pending
---

body
`,
	})

	// Should not crash; scanner may skip malformed files
	_ = repo.Run("list")
}

// --- Empty file ---

func TestEdge_EmptyFile(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{"empty.md": ""})

	res := repo.Run("list")
	if res.Err != nil {
		t.Fatalf("unexpected error on empty file: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "No tasks found") {
		t.Errorf("expected 'No tasks found' for empty file, got: %q", res.Stdout)
	}
}

// --- File with frontmatter but no body ---

func TestEdge_FrontmatterNoBody(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-nobody.md": `---
id: "001"
title: "No body task"
status: pending
priority: high
effort: small
dependencies: []
tags: []
created: 2026-02-08
---
`,
	})

	res := repo.Run("list")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !strings.Contains(res.Stdout, "001") {
		t.Errorf("expected task 001 in output, got: %q", res.Stdout)
	}
}

// --- Non-existent directory ---

func TestEdge_NonExistentDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("list", "/tmp/nonexistent-taskmd-dir-xyz-"+t.Name())
	// Scanner should either error or return empty results
	if res.Err != nil {
		if !strings.Contains(res.Err.Error(), "scan failed") {
			t.Errorf("expected 'scan failed' error, got: %v", res.Err)
		}
	}
	// If no error, that's also acceptable (scanner returns empty)
}

// --- Invalid sort field ---

func TestEdge_InvalidSortField(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("list", "--sort", "nonexistent")
	if res.Err == nil {
		t.Fatal("expected error for invalid sort field")
	}
	if !strings.Contains(res.Err.Error(), "invalid sort field") {
		t.Errorf("expected 'invalid sort field' in error, got: %v", res.Err)
	}
}

// --- Invalid format value ---

func TestEdge_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("list", "--format", "xml")
	if res.Err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' in error, got: %v", res.Err)
	}
}

// --- Invalid filter syntax ---

func TestEdge_InvalidFilterSyntax(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("list", "--filter", "status-pending") // missing =
	if res.Err == nil {
		t.Fatal("expected error for invalid filter syntax")
	}
	if !strings.Contains(res.Err.Error(), "invalid filter format") {
		t.Errorf("expected 'invalid filter format' error, got: %v", res.Err)
	}
}

// --- Set command: invalid enum values with suggestions ---

func TestEdge_Set_InvalidStatusSuggestion(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("set", "001", "--status", "pnding")
	if res.Err == nil {
		t.Fatal("expected error for invalid status")
	}
	errMsg := res.Err.Error()
	if !strings.Contains(errMsg, "invalid status") {
		t.Errorf("expected 'invalid status' error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `did you mean "pending"`) {
		t.Errorf("expected suggestion for 'pending', got: %s", errMsg)
	}
}

func TestEdge_Set_InvalidPrioritySuggestion(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("set", "001", "--priority", "hgh")
	if res.Err == nil {
		t.Fatal("expected error for invalid priority")
	}
	errMsg := res.Err.Error()
	if !strings.Contains(errMsg, "invalid priority") {
		t.Errorf("expected 'invalid priority' error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, `did you mean "high"`) {
		t.Errorf("expected suggestion for 'high', got: %s", errMsg)
	}
}

// --- Set command: dry-run ---

func TestEdge_Set_DryRunNoWrite(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("set", "001", "--status", "completed", "--dry-run")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	output := res.Stdout

	if !strings.Contains(output, "Dry run") {
		t.Error("expected 'Dry run' message in output")
	}

	// Verify file was NOT changed
	content, _ := os.ReadFile(repo.Path("001-task.md"))
	if strings.Contains(string(content), "status: completed") {
		t.Error("expected file NOT to be modified during dry run")
	}
}

// --- Graph: conflicting flags ---

func TestEdge_Graph_UpstreamAndDownstream(t *testing.T) {
	repo := newTaskRepo(t, edgeCaseFiles())

	res := repo.Run("graph", "--upstream", "--downstream")
	if res.Err == nil {
		t.Fatal("expected error when using both --upstream and --downstream")
	}
	if !strings.Contains(res.Err.Error(), "cannot use both") {
		t.Errorf("expected 'cannot use both' error, got: %v", res.Err)
	}
}
