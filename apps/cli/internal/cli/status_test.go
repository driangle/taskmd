package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// newStatusRepo seeds the dependency-chain fixture and overlays the two fields
// the status tests assert on: owner: alice on task 002 and parent: "001" on task
// 003. The rest of the shape is identical to dependency-chain.
func newStatusRepo(t *testing.T) *taskRepo {
	t.Helper()
	repo := newTaskRepoFromFixture(t, "dependency-chain")
	repo.Write("002-auth.md", `---
id: "002"
title: "Implement authentication"
status: in-progress
priority: critical
effort: large
dependencies: ["001"]
tags: ["backend", "security"]
owner: "alice"
created: 2026-02-08
---

# Implement authentication

Add JWT-based auth with refresh tokens.
`)
	repo.Write("003-ui.md", `---
id: "003"
title: "Build UI components"
status: pending
priority: medium
effort: medium
dependencies: ["002"]
tags: ["frontend"]
parent: "001"
created: 2026-02-08
---

# Build UI components

Create reusable component library.
`)
	return repo
}

// newStatusChildrenRepo builds a parent/child/grandchild tree: 010 (in-progress)
// with children 011 (pending) and 012 (completed), and grandchild 013 under 011.
// Kept inline because the specific in-progress parent and the grandchild depth are
// the subject of these tests and differ from the shared parent-children fixture.
func newStatusChildrenRepo(t *testing.T) *taskRepo {
	t.Helper()
	return newTaskRepo(t, map[string]string{
		"010-parent.md": `---
id: "010"
title: "Parent task"
status: in-progress
tags: []
dependencies: []
---

# Parent task
`,
		"011-child-a.md": `---
id: "011"
title: "Child A"
status: pending
parent: "010"
tags: []
dependencies: []
---

# Child A
`,
		"012-child-b.md": `---
id: "012"
title: "Child B"
status: completed
parent: "010"
tags: []
dependencies: []
---

# Child B
`,
		"013-grandchild.md": `---
id: "013"
title: "Grandchild"
status: pending
parent: "011"
tags: []
dependencies: []
---

# Grandchild
`,
	})
}

// newStatusMultipleInProgressRepo builds two in-progress tasks plus one pending.
func newStatusMultipleInProgressRepo(t *testing.T) *taskRepo {
	t.Helper()
	return newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "First task"
status: in-progress
tags: []
dependencies: []
---
`,
		"002.md": `---
id: "002"
title: "Second task"
status: in-progress
tags: []
dependencies: []
---
`,
		"003.md": `---
id: "003"
title: "Pending task"
status: pending
tags: []
dependencies: []
---
`,
	})
}

// statusStdout runs `status <args...>` against repo, fails on error, and returns
// stdout.
func statusStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"status"}, args...)...)
	if res.Err != nil {
		t.Fatalf("status %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestStatus_ExactMatchByID(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "001")

	if !strings.Contains(output, "Task: 001") {
		t.Error("Expected output to contain 'Task: 001'")
	}
	if !strings.Contains(output, "Title: Setup project") {
		t.Error("Expected output to contain task title")
	}
}

func TestStatus_TextFormat(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "002")

	expected := []string{
		"Task: 002",
		"Title: Implement authentication",
		"Status: in-progress",
		"Priority: critical",
		"Effort: large",
		"Tags: backend, security",
		"Owner: alice",
		"Created: 2026-02-08",
		"Dependencies: 001",
		"File:",
	}

	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q", exp)
		}
	}

	// Verify no body content is present
	if strings.Contains(output, "Description:") {
		t.Error("Status output should not contain Description section")
	}
	if strings.Contains(output, "Add JWT-based auth") {
		t.Error("Status output should not contain body content")
	}
}

func TestStatus_TextFormat_ParentField(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "003")

	if !strings.Contains(output, "Parent: 001") {
		t.Error("Expected output to contain 'Parent: 001'")
	}
}

func TestStatus_JSONFormat(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "002", "--format", "json")

	var result statusOutput
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
	if result.Priority != "critical" {
		t.Errorf("Expected priority 'critical', got %q", result.Priority)
	}
	if result.Effort != "large" {
		t.Errorf("Expected effort 'large', got %q", result.Effort)
	}
	if result.Owner != "alice" {
		t.Errorf("Expected owner 'alice', got %q", result.Owner)
	}
	if len(result.Dependencies) != 1 || result.Dependencies[0] != "001" {
		t.Errorf("Expected dependencies [001], got %v", result.Dependencies)
	}

	// Verify no content/body field in JSON
	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse raw JSON: %v", err)
	}
	if _, ok := raw["content"]; ok {
		t.Error("JSON output should not contain 'content' key")
	}
	if _, ok := raw["body"]; ok {
		t.Error("JSON output should not contain 'body' key")
	}
}

func TestStatus_YAMLFormat(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "001", "--format", "yaml")

	expected := []string{"id: \"001\"", "title: Setup project", "status: completed"}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected YAML output to contain %q", exp)
		}
	}

	// Verify no content field
	if strings.Contains(output, "content:") {
		t.Error("YAML output should not contain 'content' field")
	}
}

func TestStatus_UnsupportedFormat(t *testing.T) {
	repo := newStatusRepo(t)

	res := repo.Run("status", "001", "--format", "csv")
	if res.Err == nil {
		t.Fatal("Expected error for unsupported format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestStatus_TaskNotFound_ExactMode(t *testing.T) {
	repo := newStatusRepo(t)

	res := repo.Run("status", "nonexistent", "--exact")
	if res.Err == nil {
		t.Fatal("Expected error for non-matching query in exact mode")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestStatus_FuzzyMatch(t *testing.T) {
	repo := newStatusRepo(t)

	// "auth" is a substring of "Implement authentication"
	statusStdinReader = strings.NewReader("1\n")
	defer func() { statusStdinReader = os.Stdin }()

	output := statusStdout(t, repo, "auth")

	if !strings.Contains(output, "Task: 002") {
		t.Error("Expected fuzzy match to find task 002")
	}
}

func TestStatus_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := repo.Run("status", "anything")
	if res.Err == nil {
		t.Fatal("Expected error for empty directory")
	}
	if !strings.Contains(res.Err.Error(), "task not found") {
		t.Errorf("Expected 'task not found' error, got: %v", res.Err)
	}
}

func TestStatus_ChildrenTree(t *testing.T) {
	repo := newStatusChildrenRepo(t)

	output := statusStdout(t, repo, "010")

	if !strings.Contains(output, "Children:") {
		t.Error("Expected output to contain 'Children:' section")
	}
	if !strings.Contains(output, "011") {
		t.Error("Expected output to contain child ID '011'")
	}
	if !strings.Contains(output, "Child A") {
		t.Error("Expected output to contain child title 'Child A'")
	}
	if !strings.Contains(output, "012") {
		t.Error("Expected output to contain child ID '012'")
	}
	if !strings.Contains(output, "Child B") {
		t.Error("Expected output to contain child title 'Child B'")
	}
	if !strings.Contains(output, "013") {
		t.Error("Expected output to contain grandchild ID '013'")
	}
	if !strings.Contains(output, "Grandchild") {
		t.Error("Expected output to contain grandchild title 'Grandchild'")
	}
}

func TestStatus_ChildrenTree_JSON(t *testing.T) {
	repo := newStatusChildrenRepo(t)

	output := statusStdout(t, repo, "010", "--format", "json")

	var result statusOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if len(result.Children) == 0 {
		t.Fatal("Expected children in JSON output")
	}

	// Find child with grandchild
	var foundGrandchild bool
	for _, child := range result.Children {
		if child.ID == "011" {
			if len(child.Children) == 0 {
				t.Error("Expected child 011 to have grandchild")
			} else if child.Children[0].ID == "013" {
				foundGrandchild = true
			}
		}
	}
	if !foundGrandchild {
		t.Error("Expected to find grandchild 013 under child 011")
	}
}

func TestStatus_ChildrenTree_Circular(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"020-a.md": `---
id: "020"
title: "Task A"
status: pending
parent: "021"
tags: []
dependencies: []
---
`,
		"021-b.md": `---
id: "021"
title: "Task B"
status: pending
parent: "020"
tags: []
dependencies: []
---
`,
	})

	// Should not hang or panic
	output := statusStdout(t, repo, "020")

	if !strings.Contains(output, "Task: 020") {
		t.Error("Expected output to contain 'Task: 020'")
	}
}

func TestStatus_MinimalFlag(t *testing.T) {
	repo := newStatusChildrenRepo(t)

	output := statusStdout(t, repo, "010", "--minimal")

	if strings.Contains(output, "Children:") {
		t.Error("--minimal should suppress children section")
	}
}

func TestStatus_MinimalFlag_JSON(t *testing.T) {
	repo := newStatusChildrenRepo(t)

	output := statusStdout(t, repo, "010", "--minimal", "--format", "json")

	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if _, ok := raw["children"]; ok {
		t.Error("--minimal JSON output should not contain 'children' key")
	}
}

func TestStatus_NoChildren(t *testing.T) {
	repo := newStatusChildrenRepo(t)

	// Task 012 has no children
	output := statusStdout(t, repo, "012")

	if strings.Contains(output, "Children:") {
		t.Error("Task with no children should not show 'Children:' section")
	}
}

func TestStatus_NoBodyInOutput(t *testing.T) {
	repo := newStatusRepo(t)

	// Text format
	output := statusStdout(t, repo, "001")
	if strings.Contains(output, "Initial project setup") {
		t.Error("Text output should not contain task body")
	}

	// JSON format
	output = statusStdout(t, repo, "001", "--format", "json")

	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}
	if _, ok := raw["content"]; ok {
		t.Error("JSON should not have 'content' field")
	}
}

func TestStatus_NoArgs_ZeroInProgress(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Done task"
status: completed
tags: []
dependencies: []
---
`,
	})

	res := repo.Run("status")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Stdout != "" {
		t.Errorf("expected empty stdout, got: %q", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "No tasks currently in progress") {
		t.Errorf("expected informational message on stderr, got: %q", res.Stderr)
	}
}

func TestStatus_NoArgs_ZeroInProgress_JSON(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Done task"
status: completed
tags: []
dependencies: []
---
`,
	})

	res := repo.Run("status", "--format", "json")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if strings.TrimSpace(res.Stdout) != "[]" {
		t.Errorf("expected empty JSON array, got: %q", res.Stdout)
	}
}

func TestStatus_NoArgs_ZeroInProgress_YAML(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Done task"
status: completed
tags: []
dependencies: []
---
`,
	})

	res := repo.Run("status", "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if strings.TrimSpace(res.Stdout) != "[]" {
		t.Errorf("expected empty YAML array, got: %q", res.Stdout)
	}
}

func TestStatus_NoArgs_OneInProgress(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo)

	if !strings.Contains(output, "Task: 002") {
		t.Error("Expected output to contain 'Task: 002'")
	}
	if !strings.Contains(output, "Implement authentication") {
		t.Error("Expected output to contain task title")
	}
}

func TestStatus_NoArgs_MultipleInProgress(t *testing.T) {
	repo := newStatusMultipleInProgressRepo(t)

	output := statusStdout(t, repo)

	if !strings.Contains(output, "First task") {
		t.Error("Expected output to contain 'First task'")
	}
	if !strings.Contains(output, "Second task") {
		t.Error("Expected output to contain 'Second task'")
	}
	if strings.Contains(output, "Pending task") {
		t.Error("Should not contain pending task")
	}
}

func TestStatus_Statusline_OneTask(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "--statusline")

	expected := "#002 Implement authentication\n"
	if output != expected {
		t.Errorf("expected %q, got: %q", expected, output)
	}
}

func TestStatus_Statusline_MultipleTasks(t *testing.T) {
	repo := newStatusMultipleInProgressRepo(t)

	output := statusStdout(t, repo, "--statusline")

	if !strings.Contains(output, "(+1 more)") {
		t.Errorf("expected '(+1 more)' in output, got: %q", output)
	}
}

func TestStatus_Statusline_NoTasks(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Done"
status: completed
tags: []
dependencies: []
---
`,
	})

	output := statusStdout(t, repo, "--statusline")
	if output != "" {
		t.Errorf("expected empty output, got: %q", output)
	}
}

func TestStatus_Statusline_LongTitle(t *testing.T) {
	longTitle := "This is a very long task title that exceeds thirty characters"
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "` + longTitle + `"
status: in-progress
tags: []
dependencies: []
---
`,
	})

	output := statusStdout(t, repo, "--statusline")

	expected := "#001 " + longTitle + "\n"
	if output != expected {
		t.Errorf("expected %q, got: %q", expected, output)
	}
}

func TestStatus_NoArgs_Scope(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"cli/001.md": `---
id: "001"
title: "CLI task"
status: in-progress
tags: []
dependencies: []
---
`,
		"web/002.md": `---
id: "002"
title: "Web task"
status: in-progress
tags: []
dependencies: []
---
`,
	})

	output := statusStdout(t, repo, "--scope", "cli")

	if !strings.Contains(output, "CLI task") {
		t.Error("Expected output to contain 'CLI task'")
	}
	if strings.Contains(output, "Web task") {
		t.Error("Should not contain 'Web task' when scope is 'cli'")
	}
}

func TestStatus_NoArgs_JSON(t *testing.T) {
	repo := newStatusMultipleInProgressRepo(t)

	output := statusStdout(t, repo, "--format", "json")

	var results []statusOutput
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("Failed to parse JSON array: %v\nOutput: %s", err, output)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestStatus_Blocked(t *testing.T) {
	repo := newStatusRepo(t)

	// Task 003 depends on 002 which is in-progress (not completed)
	output := statusStdout(t, repo, "003")

	if !strings.Contains(output, "Blocked: Yes (blocked by: 002)") {
		t.Errorf("Expected blocked indicator with 002, got:\n%s", output)
	}
}

func TestStatus_Unblocked(t *testing.T) {
	repo := newStatusRepo(t)

	// Task 002 depends on 001 which is completed
	output := statusStdout(t, repo, "002")

	if !strings.Contains(output, "Blocked: No") {
		t.Errorf("Expected 'Blocked: No', got:\n%s", output)
	}
}

func TestStatus_NoDependencies_NoBlockedIndicator(t *testing.T) {
	repo := newStatusRepo(t)

	// Task 001 has no dependencies
	output := statusStdout(t, repo, "001")

	if strings.Contains(output, "Blocked:") {
		t.Errorf("Expected no blocked indicator for task without dependencies, got:\n%s", output)
	}
}

func TestStatus_Blocked_JSON(t *testing.T) {
	repo := newStatusRepo(t)

	// Task 003 is blocked (depends on 002 which is in-progress)
	output := statusStdout(t, repo, "003", "--format", "json")

	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	blocked, ok := raw["blocked"]
	if !ok {
		t.Fatal("Expected 'blocked' field in JSON output")
	}
	if blocked != true {
		t.Errorf("Expected blocked=true, got %v", blocked)
	}

	blockedBy, ok := raw["blocked_by"]
	if !ok {
		t.Fatal("Expected 'blocked_by' field in JSON output")
	}
	arr, ok := blockedBy.([]any)
	if !ok || len(arr) != 1 || arr[0] != "002" {
		t.Errorf("Expected blocked_by=[002], got %v", blockedBy)
	}
}

func TestStatus_Unblocked_JSON(t *testing.T) {
	repo := newStatusRepo(t)

	// Task 002 depends on 001 (completed) → unblocked
	output := statusStdout(t, repo, "002", "--format", "json")

	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	blocked, ok := raw["blocked"]
	if !ok {
		t.Fatal("Expected 'blocked' field in JSON output")
	}
	if blocked != false {
		t.Errorf("Expected blocked=false, got %v", blocked)
	}

	if _, ok := raw["blocked_by"]; ok {
		t.Error("Expected no 'blocked_by' field when unblocked")
	}
}

func TestStatus_NoDependencies_JSON(t *testing.T) {
	repo := newStatusRepo(t)

	// Task 001 has no dependencies
	output := statusStdout(t, repo, "001", "--format", "json")

	var raw map[string]any
	if err := json.Unmarshal([]byte(output), &raw); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if _, ok := raw["blocked"]; ok {
		t.Error("Expected no 'blocked' field for task without dependencies")
	}
	if _, ok := raw["blocked_by"]; ok {
		t.Error("Expected no 'blocked_by' field for task without dependencies")
	}
}

func TestStatus_NoArgs_YAML(t *testing.T) {
	repo := newStatusRepo(t)

	output := statusStdout(t, repo, "--format", "yaml")

	if !strings.Contains(output, "id: \"002\"") {
		t.Errorf("Expected YAML output to contain in-progress task, got: %s", output)
	}
}
