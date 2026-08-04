package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

// Canonical, recurring task shapes live under testdata/ and are loaded with
// newTaskRepoFromFixture (see testdata/README.md):
//   - "dependency-chain"  — 001-setup -> 002-auth -> 003-ui
//   - "parent-children"   — 010 parent with children 011/012
//   - "phases"            — p01 (phase: beta) and p02 (no phase)
//   - "subdir-projects"   — cli/ and backend/ tasks with an ambiguous filename
// Only genuinely one-off fixtures (e.g. markdownFixtures below) stay inline.

// getStdout runs `get <args...>` against repo, fails on error, and returns stdout.
func getStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"get"}, args...)...)
	if res.Err != nil {
		t.Fatalf("get %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestGet_ExactMatchByID(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	output := getStdout(t, repo, "001")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
	if !strings.Contains(output, "Title: Setup project") {
		t.Error("Expected output to contain task title")
	}
}

func TestGet_ExactMatchByTitle(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	output := getStdout(t, repo, "Setup project")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
}

func TestGet_ExactMatchByTitle_CaseInsensitive(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	output := getStdout(t, repo, "setup PROJECT")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected case-insensitive title match to find task 001")
	}
}

func TestGet_IDPrecedenceOverTitle(t *testing.T) {
	// A task whose title matches another task's ID: ID match must win.
	repo := newTaskRepo(t, map[string]string{
		"task1.md": `---
id: "abc"
title: "First task"
status: pending
priority: low
dependencies: []
tags: []
created: 2026-02-08
---

# First task
`,
		"task2.md": `---
id: "xyz"
title: "abc"
status: pending
priority: high
dependencies: []
tags: []
created: 2026-02-08
---

# abc task
`,
	})

	output := getStdout(t, repo, "abc")

	// Should match by ID (task1), not by title (task2)
	if !strings.Contains(output, "Title: First task") {
		t.Error("Expected ID match to take precedence over title match")
	}
}

func TestGet_TaskNotFound_ExactMode(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	res := repo.Run("get", "nonexistent", "--exact")
	if res.Err == nil {
		t.Fatal("Expected error for non-matching query in exact mode")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestGet_TaskNotFound_NoMatches(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	// very high threshold so nothing matches
	res := repo.Run("get", "zzzzzzzzzzzzzzz", "--threshold", "0.99")
	if res.Err == nil {
		t.Fatal("Expected error for garbage query")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

// TestGet_Formats groups the per-format output variants (text/json/yaml plus the
// unsupported-format error) into one table. Each format asserts differently
// (substring checks vs. JSON unmarshal), so rows carry a check closure rather
// than a shared expected value. These subtests do NOT use t.Parallel(): repo.Run
// swaps the process-global os.Stdout and mutates package-global flag vars, so
// command-driven tests must stay serial until that state is per-invocation (see
// the harness note in harness_test.go).
func TestGet_Formats(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	tests := []struct {
		name    string
		args    []string
		wantErr string // non-empty means the run should fail with this substring
		check   func(t *testing.T, output string)
	}{
		{name: "text format", args: []string{"002"}, check: assertGetText},
		{name: "json format", args: []string{"002", "--format", "json"}, check: assertGetJSON},
		{name: "yaml format", args: []string{"001", "--format", "yaml"}, check: assertGetYAML},
		{name: "unsupported format", args: []string{"001", "--format", "csv"}, wantErr: "unsupported format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := repo.Run(append([]string{"get"}, tt.args...)...)
			if tt.wantErr != "" {
				if res.Err == nil {
					t.Fatalf("Expected error for %s", tt.name)
				}
				if !strings.Contains(res.Err.Error(), tt.wantErr) {
					t.Errorf("Expected %q error, got: %v", tt.wantErr, res.Err)
				}
				return
			}
			if res.Err != nil {
				t.Fatalf("unexpected error: %v", res.Err)
			}
			tt.check(t, res.Stdout)
		})
	}
}

func assertGetText(t *testing.T, output string) {
	t.Helper()
	expected := []string{
		"Task: 002",
		"Title: Implement authentication",
		"Status: in-progress",
		"Priority: critical",
		"Effort: large",
		"Tags: backend, security",
		"Created: 2026-02-08",
		"File:",
		"Description:",
		"Add JWT-based auth with refresh tokens.",
		"Dependencies:",
		"Depends on: 001 (Setup project)",
	}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q", exp)
		}
	}
}

func assertGetJSON(t *testing.T, output string) {
	t.Helper()
	var result getOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}
	if result.ID != "002" {
		t.Errorf("Expected ID '002', got %q", result.ID)
	}
	if result.Title != "Implement authentication" {
		t.Errorf("Expected title 'Implement authentication', got %q", result.Title)
	}
	if result.Status != "in-progress" {
		t.Errorf("Expected status 'in-progress', got %q", result.Status)
	}
	if result.Content == "" {
		t.Error("Expected non-empty content in JSON output")
	}
	if len(result.Dependencies.DependsOn) != 1 || result.Dependencies.DependsOn[0].ID != "001" {
		t.Error("Expected depends_on to contain task 001")
	}
}

func assertGetYAML(t *testing.T, output string) {
	t.Helper()
	expected := []string{"id: \"001\"", "title: Setup project", "status: completed"}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected YAML output to contain %q", exp)
		}
	}
}

func TestGet_FuzzyMatch_Substring(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	// "auth" is a substring of "Implement authentication" — should fuzzy match.
	// Simulate selecting option 1.
	getStdinReader = strings.NewReader("1\n")
	defer func() { getStdinReader = os.Stdin }()

	output := getStdout(t, repo, "auth")

	if !strings.Contains(output, "Task: 002") {
		t.Error("Expected fuzzy substring match to find task 002")
	}
}

func TestGet_FuzzyMatch_Selection(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	// "ui" should fuzzy match "Build UI components"
	getStdinReader = strings.NewReader("1\n")
	defer func() { getStdinReader = os.Stdin }()

	output := getStdout(t, repo, "ui")

	if !strings.Contains(output, "Task: 003") {
		t.Error("Expected fuzzy match selection to return task 003")
	}
}

func TestGet_FuzzyMatch_Cancel(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	getStdinReader = strings.NewReader("0\n")
	defer func() { getStdinReader = os.Stdin }()

	res := repo.Run("get", "auth")
	if res.Err == nil {
		t.Fatal("Expected error when user cancels selection")
	}
	if !strings.Contains(res.Err.Error(), "cancelled") {
		t.Errorf("Expected 'cancelled' error, got: %v", res.Err)
	}
}

func TestGet_Threshold(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	// very high threshold
	res := repo.Run("get", "aut", "--threshold", "0.95")
	if res.Err == nil {
		t.Fatal("Expected error when threshold filters out matches")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestGet_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("get", "anything")
	if res.Err == nil {
		t.Fatal("Expected error for empty directory")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestGet_Dependencies(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "dependency-chain")

	// Task 003 depends on 002, and 002 blocks 003
	output := getStdout(t, repo, "003")

	if !strings.Contains(output, "Depends on: 002 (Implement authentication)") {
		t.Error("Expected depends-on info for task 003")
	}

	// Check that task 002 shows it blocks 003
	output = getStdout(t, repo, "002")
	if !strings.Contains(output, "Blocks: 003 (Build UI components)") {
		t.Error("Expected blocks info for task 002")
	}
}

// --- Unit tests for helper functions ---

func TestFindExactMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Setup project"},
		{ID: "002", Title: "Auth service"},
	}

	// Match by ID
	if task := findExactMatch("001", tasks); task == nil || task.ID != "001" {
		t.Error("Expected to find task 001 by ID")
	}

	// Match by title (case-insensitive)
	if task := findExactMatch("auth service", tasks); task == nil || task.ID != "002" {
		t.Error("Expected to find task 002 by title")
	}

	// No match
	if task := findExactMatch("nonexistent", tasks); task != nil {
		t.Error("Expected nil for non-matching query")
	}
}

func TestCalculateSimilarity(t *testing.T) {
	tests := []struct {
		query    string
		target   string
		minScore float64
		maxScore float64
	}{
		{"auth", "Implement authentication", 0.7, 1.0}, // substring
		{"setup", "Setup project", 0.7, 1.0},           // substring
		{"setup project", "Setup project", 1.0, 1.0},   // exact
		{"zzzzz", "Setup project", 0.0, 0.3},           // no relation
		{"seutp", "Setup project", 0.3, 0.8},           // typo
	}

	for _, tt := range tests {
		score := calculateSimilarity(tt.query, tt.target)
		if score < tt.minScore || score > tt.maxScore {
			t.Errorf("calculateSimilarity(%q, %q) = %.2f, expected [%.2f, %.2f]",
				tt.query, tt.target, score, tt.minScore, tt.maxScore)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
	}

	for _, tt := range tests {
		result := levenshtein(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("levenshtein(%q, %q) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// --- File path matching tests ---

func TestGet_FilePathMatch_FullRelativePath(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "subdir-projects")

	output := getStdout(t, repo, "cli/042-task.md")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected full relative path to match task cli-042")
	}
}

func TestGet_FilePathMatch_FilenameWithExtension(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "subdir-projects")

	output := getStdout(t, repo, "042-task.md")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected filename with extension to match task cli-042")
	}
}

func TestGet_FilePathMatch_FilenameWithoutExtension(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "subdir-projects")

	output := getStdout(t, repo, "042-task")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected filename without extension to match task cli-042")
	}
}

func TestGet_FilePathMatch_AmbiguousFilename(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "subdir-projects")

	// "055-api.md" exists in both cli/ and backend/ — should be ambiguous
	res := repo.Run("get", "055-api.md")
	if res.Err == nil {
		t.Fatal("Expected error for ambiguous filename")
	}
	if !strings.Contains(res.Err.Error(), "ambiguous") {
		t.Errorf("Expected 'ambiguous' error, got: %v", res.Err)
	}
}

func TestGet_FilePathMatch_ExactPathResolvesAmbiguity(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "subdir-projects")

	// Full relative path should resolve ambiguity
	output := getStdout(t, repo, "backend/055-api.md")

	if !strings.Contains(output, "Task: backend-055") {
		t.Error("Expected exact path to resolve ambiguity and match backend-055")
	}
}

func TestGet_FilePathMatch_IDStillTakesPriority(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "subdir-projects")

	// "cli-042" is a task ID — should match by ID, not filepath
	output := getStdout(t, repo, "cli-042")

	if !strings.Contains(output, "Task: cli-042") {
		t.Error("Expected ID match to still work")
	}
	if !strings.Contains(output, "Title: CLI task") {
		t.Error("Expected ID match to return CLI task")
	}
}

func TestFindFilePathMatch(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task A", FilePath: "cli/001-setup.md"},
		{ID: "002", Title: "Task B", FilePath: "backend/002-api.md"},
		{ID: "003", Title: "Task C", FilePath: "cli/003-shared.md"},
		{ID: "004", Title: "Task D", FilePath: "backend/003-shared.md"},
	}

	// Exact full path match
	task, err := findFilePathMatch("cli/001-setup.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "001" {
		t.Error("Expected exact path match to find task 001")
	}

	// Filename with extension (unique)
	task, err = findFilePathMatch("002-api.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "002" {
		t.Error("Expected filename match to find task 002")
	}

	// Filename without extension (unique)
	task, err = findFilePathMatch("001-setup", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "001" {
		t.Error("Expected filename without extension to find task 001")
	}

	// Ambiguous filename
	task, err = findFilePathMatch("003-shared.md", tasks)
	if err == nil {
		t.Fatal("Expected error for ambiguous filename")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("Expected 'ambiguous' error, got: %v", err)
	}
	if task != nil {
		t.Error("Expected nil task for ambiguous match")
	}

	// No match
	task, err = findFilePathMatch("nonexistent.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task != nil {
		t.Error("Expected nil for no match")
	}

	// Exact path takes priority over ambiguous filename
	task, err = findFilePathMatch("cli/003-shared.md", tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task == nil || task.ID != "003" {
		t.Error("Expected exact path to resolve ambiguity")
	}
}

func TestFuzzyMatchTasks(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Setup project"},
		{ID: "002", Title: "Implement authentication"},
		{ID: "003", Title: "Build UI components"},
	}

	// "auth" should match task 002 via substring
	matches := fuzzyMatchTasks("auth", tasks, 0.6)
	if len(matches) == 0 {
		t.Fatal("Expected at least one fuzzy match for 'auth'")
	}
	if matches[0].Task.ID != "002" {
		t.Errorf("Expected top match to be task 002, got %s", matches[0].Task.ID)
	}

	// Very high threshold should filter everything
	matches = fuzzyMatchTasks("auth", tasks, 0.99)
	if len(matches) != 0 {
		t.Errorf("Expected no matches with threshold 0.99, got %d", len(matches))
	}
}

func TestGet_ParentDisplay(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "parent-children")

	output := getStdout(t, repo, "011")

	if !strings.Contains(output, "Parent:") {
		t.Error("Expected output to contain 'Parent:'")
	}
	if !strings.Contains(output, "010") {
		t.Error("Expected output to contain parent ID '010'")
	}
}

func TestGet_ChildrenDisplay(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "parent-children")

	output := getStdout(t, repo, "010")

	if !strings.Contains(output, "Children:") {
		t.Error("Expected output to contain 'Children:'")
	}
	if !strings.Contains(output, "011") {
		t.Error("Expected output to contain child ID '011'")
	}
	if !strings.Contains(output, "pending") {
		t.Error("Expected output to contain child status 'pending'")
	}
	if !strings.Contains(output, "012") {
		t.Error("Expected output to contain child ID '012'")
	}
	if !strings.Contains(output, "completed") {
		t.Error("Expected output to contain child status 'completed'")
	}
}

func TestGet_ParentJSON(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "parent-children")

	output := getStdout(t, repo, "011", "--format", "json")

	var result getOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if result.Parent == nil {
		t.Fatal("Expected parent in JSON output")
	}
	if result.Parent.ID != "010" {
		t.Errorf("Expected parent ID '010', got %q", result.Parent.ID)
	}
}

func TestGet_ChildrenJSON(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "parent-children")

	output := getStdout(t, repo, "010", "--format", "json")

	var result getOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(result.Children))
	}

	childByID := make(map[string]depEntry)
	for _, c := range result.Children {
		childByID[c.ID] = c
	}

	if c, ok := childByID["011"]; !ok {
		t.Error("Expected child with ID '011'")
	} else if c.Status != "pending" {
		t.Errorf("Expected child 011 status 'pending', got %q", c.Status)
	}

	if c, ok := childByID["012"]; !ok {
		t.Error("Expected child with ID '012'")
	} else if c.Status != "completed" {
		t.Errorf("Expected child 012 status 'completed', got %q", c.Status)
	}
}

func TestGet_LeafTaskNoChildren(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "parent-children")

	output := getStdout(t, repo, "011")

	if strings.Contains(output, "Children:") {
		t.Error("Leaf task should not have 'Children:' section")
	}
}

// markdownFixtures returns a task whose body exercises markdown rendering. Kept
// inline (not promoted to testdata/) on purpose: the body's exact markdown
// constructs are the subject of these tests, so co-locating them with the
// assertions keeps the test self-explanatory.
func markdownFixtures() map[string]string {
	return map[string]string{
		"md-001-test.md": `---
id: "md-001"
title: "Markdown test task"
status: pending
priority: medium
dependencies: []
tags: []
created: 2026-02-08
---

# Heading

This has **bold** and ` + "`code`" + ` text.

- [ ] Unchecked
- [x] Checked
`,
	}
}

func TestGet_RawMarkdown(t *testing.T) {
	repo := newTaskRepo(t, markdownFixtures())

	output := getStdout(t, repo, "md-001", "--raw-markdown")

	// Raw mode: markdown delimiters should be preserved
	if !strings.Contains(output, "# Heading") {
		t.Error("Expected raw '# Heading' preserved with --raw-markdown")
	}
	if !strings.Contains(output, "**bold**") {
		t.Error("Expected raw '**bold**' preserved with --raw-markdown")
	}
	if !strings.Contains(output, "- [ ] Unchecked") {
		t.Error("Expected raw '- [ ]' preserved with --raw-markdown")
	}
	if !strings.Contains(output, "- [x] Checked") {
		t.Error("Expected raw '- [x]' preserved with --raw-markdown")
	}
}

func TestGet_FormattedMarkdown(t *testing.T) {
	repo := newTaskRepo(t, markdownFixtures())

	// Force color so the renderer emits ANSI (default noColor=false after reset).
	forceColor = true
	defer func() { forceColor = false }()

	output := getStdout(t, repo, "md-001")

	// Formatted mode: markdown delimiters should be stripped
	if strings.Contains(output, "# Heading") {
		t.Error("Expected '# Heading' to be formatted, not raw")
	}
	if strings.Contains(output, "**bold**") {
		t.Error("Expected '**bold**' to be formatted, not raw")
	}
	if !strings.Contains(output, "Heading") {
		t.Error("Expected heading text preserved after formatting")
	}
	if !strings.Contains(output, "bold") {
		t.Error("Expected bold text preserved after formatting")
	}
	// Should contain ANSI codes since we forced color
	if !strings.Contains(output, "\x1b[") {
		t.Error("Expected ANSI codes in formatted output")
	}
}

func TestGet_PhaseText(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "phases")

	output := getStdout(t, repo, "p01")

	if !strings.Contains(output, "Phase: beta") {
		t.Error("Expected output to contain 'Phase: beta'")
	}
}

func TestGet_PhaseOmittedWhenEmpty_Text(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "phases")

	output := getStdout(t, repo, "p02")

	if strings.Contains(output, "Phase:") {
		t.Error("Expected phase to be omitted when empty")
	}
}

func TestGet_PhaseJSON(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "phases")

	output := getStdout(t, repo, "p01", "--format", "json")

	var result getOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if result.Phase != "beta" {
		t.Errorf("Expected phase 'beta', got %q", result.Phase)
	}
}

func TestGet_PhaseOmittedWhenEmpty_JSON(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "phases")

	output := getStdout(t, repo, "p02", "--format", "json")

	if strings.Contains(output, `"phase"`) {
		t.Error("Expected phase to be omitted from JSON when empty")
	}
}

func TestGet_PhaseYAML(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "phases")

	output := getStdout(t, repo, "p01", "--format", "yaml")

	if !strings.Contains(output, "phase: beta") {
		t.Errorf("Expected YAML output to contain 'phase: beta', got:\n%s", output)
	}
}

func TestGet_PhaseOmittedWhenEmpty_YAML(t *testing.T) {
	repo := newTaskRepoFromFixture(t, "phases")

	output := getStdout(t, repo, "p02", "--format", "yaml")

	if strings.Contains(output, "phase:") {
		t.Error("Expected phase to be omitted from YAML when empty")
	}
}
