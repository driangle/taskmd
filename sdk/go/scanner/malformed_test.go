package scanner

import (
	"path/filepath"
	"strings"
	"testing"
)

// scanFixture scans a fixture set under testdata/ and returns the result.
func scanFixture(t *testing.T, set string) *ScanResult {
	t.Helper()
	result, err := NewScanner(filepath.Join("testdata", set), false, nil).Scan()
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	return result
}

// testdata/malformed-task reproduces GitHub issue #19: a good task next to a
// file whose frontmatter has task keys but a YAML syntax error (mixed
// flow/block list style). The broken file must be reported, not skipped.
func TestScanner_Scan_ReportsMalformedTaskFrontmatter(t *testing.T) {
	result := scanFixture(t, "malformed-task")

	if len(result.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(result.Tasks))
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 scan error, got %d", len(result.Errors))
	}
	if !strings.HasSuffix(result.Errors[0].FilePath, "003-lorem.md") {
		t.Errorf("expected error for 003-lorem.md, got %s", result.Errors[0].FilePath)
	}
	if !strings.Contains(result.Errors[0].Error.Error(), "YAML") {
		t.Errorf("expected a YAML parse error, got: %v", result.Errors[0].Error)
	}
}

// testdata/foreign-broken holds a docs page with malformed frontmatter but no
// task-signature keys — it must stay silently skipped.
func TestScanner_Scan_SkipsForeignMarkdownWithBrokenFrontmatter(t *testing.T) {
	result := scanFixture(t, "foreign-broken")

	if len(result.Errors) != 0 {
		t.Errorf("expected no scan errors for foreign markdown, got %d: %v",
			len(result.Errors), result.Errors)
	}
	if len(result.Tasks) != 0 {
		t.Errorf("expected no tasks, got %d", len(result.Tasks))
	}
}

// testdata/foreign-valid holds a page with valid YAML frontmatter but no
// id/title — not a task, and no error (unchanged behavior).
func TestScanner_Scan_SkipsValidFrontmatterNonTasks(t *testing.T) {
	result := scanFixture(t, "foreign-valid")

	if len(result.Errors) != 0 {
		t.Errorf("expected no scan errors, got %d: %v", len(result.Errors), result.Errors)
	}
	if len(result.Tasks) != 0 {
		t.Errorf("expected no tasks, got %d", len(result.Tasks))
	}
}

// testdata/unclosed holds a task file that never closes its frontmatter
// block — task-like, so it must be reported.
func TestScanner_Scan_ReportsUnclosedTaskFrontmatter(t *testing.T) {
	result := scanFixture(t, "unclosed")

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 scan error for unclosed frontmatter, got %d", len(result.Errors))
	}
	if !strings.HasSuffix(result.Errors[0].FilePath, "004-unclosed.md") {
		t.Errorf("expected error for 004-unclosed.md, got %s", result.Errors[0].FilePath)
	}
}
