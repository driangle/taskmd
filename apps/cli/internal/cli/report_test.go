package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newReportRepo seeds a report-specific repo covering all statuses, priorities,
// efforts, types, and a 3-level dependency chain for critical-path testing. Kept
// inline (not a shared fixture) because the type field and the exact 001→002→004→005
// chain are the subject of these tests.
func newReportRepo(t *testing.T) *taskRepo {
	t.Helper()
	return newTaskRepo(t, map[string]string{
		"001-setup.md": `---
id: "001"
title: "Setup project"
status: completed
priority: high
effort: small
type: chore
dependencies: []
tags: ["infra"]
created: 2026-02-08
---

# Setup project
`,
		"002-auth.md": `---
id: "002"
title: "Implement authentication"
status: in-progress
priority: critical
effort: large
type: feature
dependencies: ["001"]
tags: ["backend", "security"]
created: 2026-02-08
---

# Implement authentication
`,
		"003-ui.md": `---
id: "003"
title: "Build UI components"
status: pending
priority: medium
effort: medium
type: feature
dependencies: []
tags: ["frontend"]
created: 2026-02-08
---

# Build UI components
`,
		"004-api.md": `---
id: "004"
title: "Design API endpoints"
status: pending
priority: high
effort: medium
type: feature
dependencies: ["002"]
tags: ["backend"]
created: 2026-02-08
---

# Design API endpoints
`,
		"005-deploy.md": `---
id: "005"
title: "Setup deployment"
status: pending
priority: low
effort: large
type: chore
dependencies: ["004"]
tags: ["infra"]
created: 2026-02-08
---

# Setup deployment
`,
		"006-docs.md": `---
id: "006"
title: "Write documentation"
status: pending
priority: low
effort: small
type: docs
dependencies: []
tags: ["docs"]
created: 2026-02-08
---

# Write documentation
`,
	})
}

// reportStdout runs `report <args...>` against repo, fails on error, and returns
// stdout.
func reportStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"report"}, args...)...)
	if res.Err != nil {
		t.Fatalf("report %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

// --- Markdown tests ---

func TestReportCommand_MarkdownDefault(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	if !strings.Contains(output, "# Project Report") {
		t.Error("Expected '# Project Report' heading")
	}
	if !strings.Contains(output, "## Summary") {
		t.Error("Expected '## Summary' section")
	}
	if !strings.Contains(output, "## Tasks by Status") {
		t.Error("Expected '## Tasks by Status' section")
	}
	if !strings.Contains(output, "## Critical Path") {
		t.Error("Expected '## Critical Path' section")
	}
	if !strings.Contains(output, "## Blocked Tasks") {
		t.Error("Expected '## Blocked Tasks' section")
	}
}

func TestReportCommand_MarkdownSummaryMetrics(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	if !strings.Contains(output, "| Total Tasks | 6 |") {
		t.Error("Expected total tasks = 6 in summary table")
	}
	if !strings.Contains(output, "| Critical Path Length |") {
		t.Error("Expected critical path length in summary")
	}
	if !strings.Contains(output, "| Avg Dependencies |") {
		t.Error("Expected avg dependencies in summary")
	}
}

func TestReportCommand_MarkdownStatusBreakdown(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	if !strings.Contains(output, "### By Status") {
		t.Error("Expected '### By Status' breakdown")
	}
	if !strings.Contains(output, "pending: 4") {
		t.Error("Expected 4 pending tasks in breakdown")
	}
	if !strings.Contains(output, "completed: 1") {
		t.Error("Expected 1 completed task in breakdown")
	}
	if !strings.Contains(output, "in-progress: 1") {
		t.Error("Expected 1 in-progress task in breakdown")
	}
}

func TestReportCommand_MarkdownTypeBreakdown(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	if !strings.Contains(output, "### By Type") {
		t.Error("Expected '### By Type' breakdown")
	}
	if !strings.Contains(output, "feature: 3") {
		t.Error("Expected 3 feature tasks in type breakdown")
	}
	if !strings.Contains(output, "chore: 2") {
		t.Error("Expected 2 chore tasks in type breakdown")
	}
	if !strings.Contains(output, "docs: 1") {
		t.Error("Expected 1 docs task in type breakdown")
	}
}

func TestReportCommand_MarkdownTypeBreakdownHiddenWhenEmpty(t *testing.T) {
	// Create a task without a type field
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Untyped task"
status: pending
dependencies: []
created: 2026-02-08
---
# Untyped task
`,
	})

	output := reportStdout(t, repo)

	if strings.Contains(output, "### By Type") {
		t.Error("Expected no '### By Type' section when no tasks have a type set")
	}
}

func TestReportCommand_MarkdownGroups(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	// Default group-by is status
	if !strings.Contains(output, "### pending (4)") {
		t.Error("Expected '### pending (4)' group header")
	}
	if !strings.Contains(output, "### completed (1)") {
		t.Error("Expected '### completed (1)' group header")
	}
}

func TestReportCommand_MarkdownCriticalPath(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	// 001 -> 002 -> 004 -> 005 is the critical path (depth 4)
	if !strings.Contains(output, "[001]") {
		t.Error("Expected task 001 on critical path")
	}
	if !strings.Contains(output, "[005]") {
		t.Error("Expected task 005 on critical path")
	}
}

func TestReportCommand_MarkdownBlockedTasks(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	// Tasks 002, 004, 005 have deps and are not completed
	if !strings.Contains(output, "Waiting on:") {
		t.Error("Expected 'Waiting on:' info for blocked tasks")
	}
}

func TestReportCommand_MarkdownIncludeGraph(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--include-graph")

	if !strings.Contains(output, "## Dependency Graph") {
		t.Error("Expected '## Dependency Graph' section")
	}
	if !strings.Contains(output, "```mermaid") {
		t.Error("Expected mermaid code block")
	}
	if !strings.Contains(output, "graph TD") {
		t.Error("Expected 'graph TD' in mermaid output")
	}
}

func TestReportCommand_MarkdownNoGraph(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	if strings.Contains(output, "## Dependency Graph") {
		t.Error("Expected no graph section without --include-graph")
	}
}

// --- Group-by variations ---

func TestReportCommand_GroupByPriority(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--group-by", "priority")

	if !strings.Contains(output, "## Tasks by Priority") {
		t.Error("Expected '## Tasks by Priority' section")
	}
	if !strings.Contains(output, "### critical") {
		t.Error("Expected '### critical' group header")
	}
	if !strings.Contains(output, "### high") {
		t.Error("Expected '### high' group header")
	}
}

func TestReportCommand_GroupByEffort(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--group-by", "effort")

	if !strings.Contains(output, "## Tasks by Effort") {
		t.Error("Expected '## Tasks by Effort' section")
	}
	if !strings.Contains(output, "### small") {
		t.Error("Expected '### small' group header")
	}
	if !strings.Contains(output, "### medium") {
		t.Error("Expected '### medium' group header")
	}
	if !strings.Contains(output, "### large") {
		t.Error("Expected '### large' group header")
	}
}

func TestReportCommand_GroupByType(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--group-by", "type")

	if !strings.Contains(output, "## Tasks by Type") {
		t.Error("Expected '## Tasks by Type' section")
	}
	if !strings.Contains(output, "### feature") {
		t.Error("Expected '### feature' group header")
	}
	if !strings.Contains(output, "### chore") {
		t.Error("Expected '### chore' group header")
	}
}

func TestReportCommand_GroupByTag(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--group-by", "tag")

	if !strings.Contains(output, "## Tasks by Tag") {
		t.Error("Expected '## Tasks by Tag' section")
	}
	if !strings.Contains(output, "### backend") {
		t.Error("Expected '### backend' group header")
	}
	if !strings.Contains(output, "### infra") {
		t.Error("Expected '### infra' group header")
	}
}

// --- JSON tests ---

func TestReportCommand_JSONStructure(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if _, ok := result["summary"]; !ok {
		t.Error("Expected 'summary' key in JSON output")
	}
	if _, ok := result["groups"]; !ok {
		t.Error("Expected 'groups' key in JSON output")
	}
	if _, ok := result["critical_path"]; !ok {
		t.Error("Expected 'critical_path' key in JSON output")
	}
	if _, ok := result["blocked_tasks"]; !ok {
		t.Error("Expected 'blocked_tasks' key in JSON output")
	}
	if _, ok := result["group_by"]; !ok {
		t.Error("Expected 'group_by' key in JSON output")
	}
}

func TestReportCommand_JSONNoGraphByDefault(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, ok := result["graph"]; ok {
		t.Error("Expected no 'graph' key without --include-graph")
	}
}

func TestReportCommand_JSONIncludeGraph(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "json", "--include-graph")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	graphData, ok := result["graph"]
	if !ok {
		t.Fatal("Expected 'graph' key with --include-graph")
	}

	graphMap, ok := graphData.(map[string]any)
	if !ok {
		t.Fatal("Expected graph to be an object")
	}

	if _, ok := graphMap["nodes"]; !ok {
		t.Error("Expected 'nodes' in graph data")
	}
	if _, ok := graphMap["edges"]; !ok {
		t.Error("Expected 'edges' in graph data")
	}
}

func TestReportCommand_JSONGroupByPriority(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "json", "--group-by", "priority")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["group_by"] != "priority" {
		t.Errorf("Expected group_by = 'priority', got %v", result["group_by"])
	}
}

// --- HTML tests ---

func TestReportCommand_HTMLDoctype(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "html")

	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Expected <!DOCTYPE html> in HTML output")
	}
}

func TestReportCommand_HTMLStyle(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "html")

	if !strings.Contains(output, "<style>") {
		t.Error("Expected <style> tag in HTML output")
	}
}

func TestReportCommand_HTMLSections(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "html")

	sections := []string{
		"<h1>Project Report</h1>",
		"<h2>Summary</h2>",
		"<h2>Critical Path</h2>",
		"<h2>Blocked Tasks</h2>",
	}

	for _, s := range sections {
		if !strings.Contains(output, s) {
			t.Errorf("Expected HTML to contain %q", s)
		}
	}
}

func TestReportCommand_HTMLTypeBreakdown(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "html")

	if !strings.Contains(output, "<h3>By Type</h3>") {
		t.Error("Expected '<h3>By Type</h3>' in HTML output")
	}
	if !strings.Contains(output, "feature: 3") {
		t.Error("Expected 'feature: 3' in HTML type breakdown")
	}
}

func TestReportCommand_HTMLMermaid(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "html", "--include-graph")

	if !strings.Contains(output, "mermaid") {
		t.Error("Expected mermaid reference in HTML with --include-graph")
	}
	if !strings.Contains(output, "<h2>Dependency Graph</h2>") {
		t.Error("Expected Dependency Graph heading in HTML")
	}
}

func TestReportCommand_HTMLNoMermaidWithoutFlag(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo, "--format", "html")

	if strings.Contains(output, "Dependency Graph") {
		t.Error("Expected no Dependency Graph section without --include-graph")
	}
}

// --- Output file tests ---

func TestReportCommand_OutputToFile(t *testing.T) {
	repo := newReportRepo(t)
	outFile := filepath.Join(t.TempDir(), "report.md")

	if res := repo.Run("report", "--out", outFile); res.Err != nil {
		t.Fatalf("report failed: %v", res.Err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "# Project Report") {
		t.Error("Expected file to contain '# Project Report'")
	}
}

func TestReportCommand_OutputHTMLToFile(t *testing.T) {
	repo := newReportRepo(t)
	outFile := filepath.Join(t.TempDir(), "report.html")

	if res := repo.Run("report", "--format", "html", "--out", outFile); res.Err != nil {
		t.Fatalf("report failed: %v", res.Err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !strings.Contains(string(content), "<!DOCTYPE html>") {
		t.Error("Expected HTML file to contain DOCTYPE")
	}
}

// --- Edge cases ---

func TestReportCommand_EmptyDirectory(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := reportStdout(t, repo)

	if !strings.Contains(output, "# Project Report") {
		t.Error("Expected report heading even for empty directory")
	}
	if !strings.Contains(output, "| Total Tasks | 0 |") {
		t.Error("Expected total tasks = 0 for empty directory")
	}
}

func TestReportCommand_NoDependencies(t *testing.T) {
	// Create two independent tasks
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Task one"
status: pending
dependencies: []
created: 2026-02-08
---
# Task one
`,
		"002.md": `---
id: "002"
title: "Task two"
status: pending
dependencies: []
created: 2026-02-08
---
# Task two
`,
	})

	output := reportStdout(t, repo)

	// Independent tasks still appear on the critical path (all at depth 1)
	if !strings.Contains(output, "## Critical Path") {
		t.Error("Expected '## Critical Path' section")
	}
	if !strings.Contains(output, "No blocked tasks.") {
		t.Error("Expected 'No blocked tasks.' for independent tasks")
	}
}

func TestReportCommand_AllCompleted(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Done task"
status: completed
dependencies: []
created: 2026-02-08
---
# Done task
`,
	})

	output := reportStdout(t, repo)

	if !strings.Contains(output, "completed: 1") {
		t.Error("Expected completed count of 1")
	}
	if !strings.Contains(output, "No blocked tasks.") {
		t.Error("Expected no blocked tasks when all completed")
	}
}

func TestReportCommand_InvalidFormat(t *testing.T) {
	repo := newReportRepo(t)

	res := repo.Run("report", "--format", "xml")
	if res.Err == nil {
		t.Fatal("Expected error for invalid format")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestReportCommand_InvalidGroupBy(t *testing.T) {
	repo := newReportRepo(t)

	res := repo.Run("report", "--group-by", "nonexistent")
	if res.Err == nil {
		t.Fatal("Expected error for invalid group-by field")
	}
	if !strings.Contains(res.Err.Error(), "unsupported group-by field") {
		t.Errorf("Expected 'unsupported group-by field' error, got: %v", res.Err)
	}
}

func TestReportCommand_CriticalPathOrdering(t *testing.T) {
	repo := newReportRepo(t)

	output := reportStdout(t, repo)

	// Critical path: 001 -> 002 -> 004 -> 005
	// They should appear in depth order (root first)
	idx001 := strings.Index(output, "1. [001]")
	idx005 := strings.LastIndex(output, "[005]")

	if idx001 == -1 {
		t.Error("Expected task 001 in critical path")
	}
	if idx005 == -1 {
		t.Error("Expected task 005 in critical path")
	}
	if idx001 != -1 && idx005 != -1 && idx001 > idx005 {
		t.Error("Expected task 001 to appear before task 005 in critical path")
	}
}

// --- Color tests ---

func TestReportCommand_ColorEnabled(t *testing.T) {
	repo := newReportRepo(t)
	forceColor = true
	defer func() { forceColor = false }()

	os.Unsetenv("NO_COLOR")

	output := reportStdout(t, repo)

	// With colors enabled, output should contain ANSI escape codes
	if !strings.Contains(output, "\x1b[") {
		t.Error("Expected colored report output to contain ANSI escape codes")
	}

	// Content should still be present
	if !strings.Contains(output, "Project Report") {
		t.Error("Expected report heading in colored output")
	}
	if !strings.Contains(output, "001") {
		t.Error("Expected task IDs in colored output")
	}
}

func TestReportCommand_NoColorFlag(t *testing.T) {
	repo := newReportRepo(t)

	os.Unsetenv("NO_COLOR")

	output := reportStdout(t, repo, "--no-color")

	// With --no-color flag, output should NOT contain ANSI escape codes
	if strings.Contains(output, "\x1b[") {
		t.Error("Expected no ANSI codes in --no-color report output")
	}

	// Content should still be present
	if !strings.Contains(output, "Project Report") {
		t.Error("Expected report heading in no-color output")
	}
	if !strings.Contains(output, "001") {
		t.Error("Expected task IDs in no-color output")
	}
}

func TestReportCommand_NoColorEnvVar(t *testing.T) {
	repo := newReportRepo(t)

	t.Setenv("NO_COLOR", "1")

	output := reportStdout(t, repo)

	if strings.Contains(output, "\x1b[") {
		t.Error("Expected no ANSI codes when NO_COLOR env var is set")
	}
}

func TestReportCommand_FileOutputNoColor(t *testing.T) {
	repo := newReportRepo(t)
	outFile := filepath.Join(t.TempDir(), "report.md")

	forceColor = true
	defer func() { forceColor = false }()

	os.Unsetenv("NO_COLOR")

	if res := repo.Run("report", "--out", outFile); res.Err != nil {
		t.Fatalf("report failed: %v", res.Err)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// File output should NOT contain ANSI escape codes
	if strings.Contains(string(content), "\x1b[") {
		t.Error("Expected no ANSI codes in file output")
	}

	// Content should still be present
	if !strings.Contains(string(content), "Project Report") {
		t.Error("Expected report heading in file output")
	}
}

func TestReportCommand_ColoredSections(t *testing.T) {
	repo := newReportRepo(t)
	forceColor = true
	defer func() { forceColor = false }()

	os.Unsetenv("NO_COLOR")

	output := reportStdout(t, repo)

	// Status breakdown should contain colored status labels
	if !strings.Contains(output, "pending") {
		t.Error("Expected 'pending' in status breakdown")
	}
	if !strings.Contains(output, "completed") {
		t.Error("Expected 'completed' in status breakdown")
	}

	// Critical path should contain task IDs and statuses
	if !strings.Contains(output, "001") {
		t.Error("Expected task 001 in critical path")
	}

	// Section headings should be present (formatted by formatLabel)
	if !strings.Contains(output, "Summary") {
		t.Error("Expected Summary section heading")
	}
	if !strings.Contains(output, "Critical Path") {
		t.Error("Expected Critical Path section heading")
	}
	if !strings.Contains(output, "Blocked Tasks") {
		t.Error("Expected Blocked Tasks section heading")
	}
}
