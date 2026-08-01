package feed

import (
	"os"
	"strings"
	"testing"
	"time"
)

const sampleGitLogOutput = `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
Alice
2026-02-28 10:30:00 +0000
chore: update task 042 status

M	tasks/cli/042-add-auth.md

bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
Bob
2026-02-27 14:00:00 +0000
feat: add new task 043

A	tasks/cli/043-new-feature.md
R100	tasks/old/010-rename-me.md	tasks/cli/010-renamed.md
`

func TestParseGitLogOutput(t *testing.T) {
	entries := ParseGitLogOutput(sampleGitLogOutput)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	e := entries[0]
	if e.Hash != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Errorf("unexpected hash: %s", e.Hash)
	}
	if e.Author != "Alice" {
		t.Errorf("unexpected author: %s", e.Author)
	}
	if e.Message != "chore: update task 042 status" {
		t.Errorf("unexpected message: %s", e.Message)
	}
	if len(e.Files) != 1 || e.Files[0].Status != "modified" {
		t.Errorf("unexpected files: %+v", e.Files)
	}
	if e.Files[0].TaskID != "042" {
		t.Errorf("expected task ID 042, got %s", e.Files[0].TaskID)
	}
}

func TestParseGitLogOutput_Empty(t *testing.T) {
	if entries := ParseGitLogOutput(""); len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
	if entries := ParseGitLogOutput("  \n\n "); len(entries) != 0 {
		t.Errorf("expected 0 entries for whitespace, got %d", len(entries))
	}
}

func TestExtractTaskIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"tasks/cli/042-add-auth.md", "042"},
		{"tasks/043-new-feature.md", "043"},
		{"tasks/cli/01kjmg6sc-implement-feed.md", "01kjmg6sc"},
		{"README.md", ""},
	}
	for _, tt := range tests {
		if got := ExtractTaskIDFromPath(tt.path); got != tt.expected {
			t.Errorf("ExtractTaskIDFromPath(%q) = %q, want %q", tt.path, got, tt.expected)
		}
	}
}

func TestBuildGitLogArgs(t *testing.T) {
	args := BuildGitLogArgs("tasks", 10, "7d", "cli")
	hasLimit, hasSince, hasScope := false, false, false
	for _, arg := range args {
		switch {
		case arg == "-10":
			hasLimit = true
		case arg == "--since=7.days.ago":
			hasSince = true
		}
		if len(arg) > 0 && arg[len(arg)-1] == 'd' {
			// skip
		}
		if len(arg) >= 3 && arg[len(arg)-3:] == "cli" || (len(arg) > 3 && contains(arg, "cli")) {
			hasScope = true
		}
	}
	if !hasLimit {
		t.Error("expected -10 in args")
	}
	if !hasSince {
		t.Error("expected --since=7.days.ago in args")
	}
	if !hasScope {
		t.Error("expected cli scope in args")
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestNormalizeSince(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"2d", "2.days.ago"},
		{"1w", "1.weeks.ago"},
		{"3m", "3.months.ago"},
		{"1y", "1.years.ago"},
		{"2026-02-28", "2026-02-28"},
		{"", ""},
		{"d", "d"},
	}
	for _, tt := range tests {
		if got := NormalizeSince(tt.input); got != tt.expected {
			t.Errorf("NormalizeSince(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMergeEntries(t *testing.T) {
	a := []FeedEntry{
		{Source: "git", Timestamp: mustParseTime("2026-02-28T10:00:00Z"), Message: "git-1"},
		{Source: "git", Timestamp: mustParseTime("2026-02-26T10:00:00Z"), Message: "git-2"},
	}
	b := []FeedEntry{
		{Source: "worklog", Timestamp: mustParseTime("2026-02-27T12:00:00Z"), Message: "wl-1"},
		{Source: "worklog", Timestamp: mustParseTime("2026-02-25T08:00:00Z"), Message: "wl-2"},
	}

	merged := MergeEntries(a, b)
	if len(merged) != 4 {
		t.Fatalf("expected 4, got %d", len(merged))
	}

	expected := []string{"git-1", "wl-1", "git-2", "wl-2"}
	for i, e := range merged {
		if e.Message != expected[i] {
			t.Errorf("merged[%d] = %q, want %q", i, e.Message, expected[i])
		}
	}
}

func TestMergeEntries_EmptySlices(t *testing.T) {
	entries := []FeedEntry{{Source: "git", Message: "only"}}
	if got := MergeEntries(entries, nil); len(got) != 1 {
		t.Errorf("merge with nil b: expected 1, got %d", len(got))
	}
	if got := MergeEntries(nil, entries); len(got) != 1 {
		t.Errorf("merge with nil a: expected 1, got %d", len(got))
	}
}

func TestParseSinceTime(t *testing.T) {
	ts := ParseSinceTime("2026-02-15")
	if ts.IsZero() || ts.Year() != 2026 || ts.Month() != 2 || ts.Day() != 15 {
		t.Errorf("unexpected date: %v", ts)
	}

	ts = ParseSinceTime("7d")
	if ts.IsZero() {
		t.Error("expected non-zero time for 7d")
	}

	ts = ParseSinceTime("garbage")
	if !ts.IsZero() {
		t.Errorf("expected zero time for invalid input, got %v", ts)
	}
}

func TestScanWorklogEntries(t *testing.T) {
	tmpDir := t.TempDir()
	wlDir := tmpDir + "/cli/.worklogs"
	if err := os.MkdirAll(wlDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := "## 2026-02-15T10:00:00Z\n\nStarted implementation.\n\n## 2026-02-15T14:30:00Z\n\nCompleted login.\n"
	if err := os.WriteFile(wlDir+"/015.md", []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	entries := ScanWorklogEntries(tmpDir, "cli", "", false)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Message != "Completed login." {
		t.Errorf("expected newest first, got: %s", entries[0].Message)
	}
	if entries[0].Source != "worklog" {
		t.Errorf("expected source worklog, got: %s", entries[0].Source)
	}
	if entries[0].TaskID != "015" {
		t.Errorf("expected taskID 015, got: %s", entries[0].TaskID)
	}
}

func TestEnrichEntriesWithDiffAnalysis(t *testing.T) {
	entries := []FeedEntry{
		{
			Hash: "abc123",
			Files: []FileChange{
				{Path: "tasks/042-auth.md", Status: "modified"},
			},
		},
	}

	mockShow := func(hash, path string) (string, error) {
		if hash == "abc123^" {
			return "---\nstatus: pending\n---\n# Task", nil
		}
		return "---\nstatus: completed\n---\n# Task", nil
	}

	EnrichEntriesWithDiffAnalysis(entries, mockShow)

	fc := entries[0].Files[0]
	if fc.TaskStatus != "completed" {
		t.Errorf("expected taskStatus completed, got %q", fc.TaskStatus)
	}
	if len(fc.FieldChanges) != 1 || fc.FieldChanges[0].Field != "status" {
		t.Errorf("expected status field change, got %+v", fc.FieldChanges)
	}
}

func TestQuery(t *testing.T) {
	entries, err := Query(Options{
		TasksDir: t.TempDir(),
		Limit:    5,
		Source:   "worklog",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty dir yields no entries
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestQuery_RequiresGitLogFn(t *testing.T) {
	_, err := Query(Options{
		TasksDir: t.TempDir(),
		Source:   "git",
	})
	if err == nil {
		t.Fatal("expected error when GitLogFn is nil")
	}
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestBuildGitLogArgsForFile(t *testing.T) {
	args := BuildGitLogArgsForFile(10, "7d", "tasks/cli/042-add-auth.md")

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--follow") {
		t.Errorf("expected --follow in args, got: %v", args)
	}
	if !strings.Contains(joined, "-10") {
		t.Errorf("expected -10 limit in args, got: %v", args)
	}
	if !strings.Contains(joined, "--since=7.days.ago") {
		t.Errorf("expected normalized --since in args, got: %v", args)
	}
	// The task file must be the final pathspec, after the "--" separator.
	if args[len(args)-1] != "tasks/cli/042-add-auth.md" {
		t.Errorf("expected task file as last arg, got: %v", args)
	}
	if args[len(args)-2] != "--" {
		t.Errorf("expected -- before task file, got: %v", args)
	}
}

func TestBuildGitLogArgsForFile_NoSince(t *testing.T) {
	args := BuildGitLogArgsForFile(20, "", "tasks/cli/042-add-auth.md")
	if strings.Contains(strings.Join(args, " "), "--since") {
		t.Errorf("did not expect --since when empty, got: %v", args)
	}
}

func TestParseFileChangeLine_Copy(t *testing.T) {
	// A copy (git --follow surfaces task creation from a template as C<score>)
	// should be treated as a creation of the destination file.
	fc := parseFileChangeLine("C051\t.taskmd/templates/feature.md\ttasks/cli/042-add-auth.md")
	if fc == nil {
		t.Fatal("expected a file change for a copy line")
	}
	if fc.Status != "created" {
		t.Errorf("expected status created, got %q", fc.Status)
	}
	if fc.Path != "tasks/cli/042-add-auth.md" {
		t.Errorf("expected destination path, got %q", fc.Path)
	}
	if fc.TaskID != "042" {
		t.Errorf("expected taskID 042, got %q", fc.TaskID)
	}
}

func TestFilterEntriesByTaskID(t *testing.T) {
	entries := []FeedEntry{
		{Files: []FileChange{
			{Path: "tasks/cli/042-add-auth.md", TaskID: "042", Status: "modified"},
			{Path: "apps/cli/internal/cli/templates/task_feature.md", TaskID: "", Status: "created"},
		}},
		{Files: []FileChange{
			{Path: "tasks/cli/099-other.md", TaskID: "099", Status: "modified"},
		}},
	}

	got := filterEntriesByTaskID(entries, "042")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry after filtering, got %d", len(got))
	}
	if len(got[0].Files) != 1 || got[0].Files[0].TaskID != "042" {
		t.Errorf("expected only the 042 file change, got %+v", got[0].Files)
	}
}

func TestFilterWorklogsByTask(t *testing.T) {
	entries := []FeedEntry{
		{Source: "worklog", TaskID: "042", Message: "keep"},
		{Source: "worklog", TaskID: "099", Message: "drop"},
	}
	got := filterWorklogsByTask(entries, "042")
	if len(got) != 1 || got[0].Message != "keep" {
		t.Errorf("expected only the 042 worklog entry, got %+v", got)
	}
	// Empty task id is a no-op.
	if got := filterWorklogsByTask(entries, ""); len(got) != 2 {
		t.Errorf("expected passthrough for empty task id, got %d", len(got))
	}
}

func TestQuery_TaskFileScopesToTask(t *testing.T) {
	// git log --follow output that includes an unrelated look-alike file the
	// follow traced through; only the target task's changes should survive.
	gitOut := `aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
Alice
2026-02-28 10:30:00 +0000
chore: complete task 042

M	tasks/cli/042-add-auth.md

bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
Bob
2026-02-20 09:00:00 +0000
chore: add templates

A	apps/cli/internal/cli/templates/task_feature.md
`
	gitLog := func(_ string, _ []string) (string, error) { return gitOut, nil }

	entries, err := Query(Options{
		TasksDir: "tasks",
		Source:   "git",
		TaskFile: "tasks/cli/042-add-auth.md",
		GitLogFn: gitLog,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry scoped to task 042, got %d", len(entries))
	}
	if entries[0].Files[0].TaskID != "042" {
		t.Errorf("expected only task 042 changes, got %+v", entries[0].Files)
	}
}
