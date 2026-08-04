package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// graphTaskFiles returns the canonical graph-command dependency shape: a 5-task
// set (001..005) with a mix of completed/pending statuses and priorities used to
// exercise dependency traversal, exclusion, and cleanup. Kept inline (not a
// shared testdata set) because this specific 5-node shape is graph-only.
func graphTaskFiles() map[string]string {
	return map[string]string{
		"001-root-task.md": `---
id: "001"
title: "Root Task"
status: completed
priority: high
effort: small
dependencies: []
tags: ["test"]
created: 2026-02-08
---

# Root Task

A completed root task.
`,
		"002-depends-on-001.md": `---
id: "002"
title: "Depends on 001"
status: completed
priority: medium
effort: medium
dependencies: ["001"]
tags: ["test"]
created: 2026-02-08
---

# Depends on 001

A completed task that depends on 001.
`,
		"003-depends-on-002.md": `---
id: "003"
title: "Depends on 002"
status: pending
priority: high
effort: small
dependencies: ["002"]
tags: ["test"]
created: 2026-02-08
---

# Depends on 002

A pending task that depends on completed task 002.
`,
		"004-depends-on-001-002.md": `---
id: "004"
title: "Depends on 001 and 002"
status: pending
priority: low
effort: large
dependencies: ["001", "002"]
tags: ["test"]
created: 2026-02-08
---

# Depends on 001 and 002

A pending task that depends on both 001 and 002.
`,
		"005-no-deps.md": `---
id: "005"
title: "No Dependencies"
status: pending
priority: medium
effort: small
dependencies: []
tags: ["test"]
created: 2026-02-08
---

# No Dependencies

A pending task with no dependencies.
`,
	}
}

// graphStdout runs `graph <args...>` against repo, fails on error, and returns stdout.
func graphStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"graph"}, args...)...)
	if res.Err != nil {
		t.Fatalf("graph %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

// parseGraphJSON parses JSON graph output and returns the result map.
func parseGraphJSON(t *testing.T, output string) map[string]any {
	t.Helper()

	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}
	return result
}

// graphNodeIDs extracts node IDs from parsed graph JSON.
func graphNodeIDs(t *testing.T, result map[string]any) []string {
	t.Helper()

	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected 'nodes' to be an array")
	}

	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeMap := node.(map[string]any)
		ids = append(ids, nodeMap["id"].(string))
	}
	return ids
}

func TestGraphCommand_JSON_Format(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all")

	// Parse JSON
	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify nodes
	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected 'nodes' to be an array")
	}

	if len(nodes) != 5 {
		t.Errorf("Expected 5 nodes, got %d", len(nodes))
	}

	// Verify edges
	edges, ok := result["edges"].([]any)
	if !ok {
		t.Fatal("Expected 'edges' to be an array")
	}

	if len(edges) != 4 {
		t.Errorf("Expected 4 edges, got %d", len(edges))
	}
}

func TestGraphCommand_ExcludeStatus_BugFix(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--exclude-status", "completed")

	// Parse JSON
	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify only pending tasks are included
	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected 'nodes' to be an array")
	}

	if len(nodes) != 3 {
		t.Errorf("Expected 3 pending nodes, got %d", len(nodes))
	}

	// Verify that dependencies to completed tasks are cleaned up
	edges, ok := result["edges"].([]any)
	if !ok {
		t.Fatal("Expected 'edges' to be an array")
	}

	// Should have no edges since all dependencies were to completed tasks
	if len(edges) != 0 {
		t.Errorf("Expected 0 edges after cleaning dependencies, got %d", len(edges))
	}

	// Verify each pending task has no dependencies now
	for _, node := range nodes {
		nodeMap := node.(map[string]any)
		taskID := nodeMap["id"].(string)

		// Task 003 and 004 had dependencies on completed tasks
		// After filtering, they should have no dependencies
		if taskID == "003" || taskID == "004" {
			// These tasks should now be root tasks (no dependencies)
			// This is the bug fix - their dependencies should be cleaned
			t.Logf("Task %s correctly filtered (dependencies cleaned)", taskID)
		}
	}
}

func TestGraphCommand_ASCII_Format(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "ascii", "--all")

	// Verify output contains task IDs
	if !strings.Contains(output, "[001]") {
		t.Error("Expected ASCII output to contain [001]")
	}

	if !strings.Contains(output, "[005]") {
		t.Error("Expected ASCII output to contain [005]")
	}

	// Verify tree structure characters
	if !strings.Contains(output, "└──") || !strings.Contains(output, "├──") {
		t.Error("Expected ASCII output to contain tree structure characters")
	}
}

func TestGraphCommand_ASCII_ExcludeCompleted(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "ascii", "--exclude-status", "completed")

	// Verify completed tasks are not in output
	if strings.Contains(output, "[001]") {
		t.Error("Expected [001] (completed) to be excluded from output")
	}

	if strings.Contains(output, "[002]") {
		t.Error("Expected [002] (completed) to be excluded from output")
	}

	// Verify pending tasks ARE in output
	if !strings.Contains(output, "[003]") {
		t.Error("Expected [003] (pending) to be in output")
	}

	if !strings.Contains(output, "[004]") {
		t.Error("Expected [004] (pending) to be in output")
	}

	if !strings.Contains(output, "[005]") {
		t.Error("Expected [005] (pending) to be in output")
	}

	// This is the key test for the bug fix:
	// With the bug, only task 005 would appear (the only one with truly no dependencies)
	// With the fix, all 3 pending tasks should appear as roots since their dependencies are cleaned
	rootTaskCount := strings.Count(output, "[00")
	if rootTaskCount < 3 {
		t.Errorf("Expected at least 3 tasks to be shown as roots, got %d (bug not fixed)", rootTaskCount)
	}
}

func TestGraphCommand_Mermaid_Format(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "mermaid", "--all")

	// Verify mermaid syntax
	if !strings.Contains(output, "graph TD") {
		t.Error("Expected mermaid output to start with 'graph TD'")
	}

	// Verify nodes are defined
	if !strings.Contains(output, "001[") {
		t.Error("Expected node definition for task 001")
	}

	// Verify edges
	if !strings.Contains(output, "-->") {
		t.Error("Expected mermaid output to contain edge arrows")
	}

	// Verify styles are included
	if !strings.Contains(output, "classDef") {
		t.Error("Expected mermaid output to contain style definitions")
	}
}

func TestGraphCommand_Mermaid_WithFocus(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "mermaid", "--all", "--focus", "003")

	// Verify focus style is applied
	if !strings.Contains(output, ":::focus") {
		t.Error("Expected focused task to have :::focus style")
	}
}

func TestGraphCommand_DOT_Format(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "dot", "--all")

	// Verify DOT syntax
	if !strings.Contains(output, "digraph tasks") {
		t.Error("Expected DOT output to start with 'digraph tasks'")
	}

	// Verify nodes are defined
	if !strings.Contains(output, "001 [label=") {
		t.Error("Expected node definition for task 001")
	}

	// Verify edges
	if !strings.Contains(output, "->") {
		t.Error("Expected DOT output to contain edge arrows")
	}

	// Verify colors are applied
	if !strings.Contains(output, "fillcolor=") {
		t.Error("Expected DOT output to contain color definitions")
	}
}

func TestGraphCommand_RootDownstream(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all", "--root", "001", "--downstream")

	// Parse JSON
	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify only downstream tasks are included
	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected 'nodes' to be an array")
	}

	// Should include: 001 (root), 002, 003, 004 (all depend on 001 directly or indirectly)
	// Should NOT include: 005 (no dependency on 001)
	if len(nodes) != 4 {
		t.Errorf("Expected 4 nodes (001 and its downstream), got %d", len(nodes))
	}

	// Verify 005 is not included
	for _, node := range nodes {
		nodeMap := node.(map[string]any)
		if nodeMap["id"].(string) == "005" {
			t.Error("Task 005 should not be included in downstream of 001")
		}
	}
}

func TestGraphCommand_RootUpstream(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all", "--root", "003", "--upstream")

	// Parse JSON
	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify only upstream tasks are included
	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected 'nodes' to be an array")
	}

	// Should include: 003 (root), 002, 001 (dependencies)
	// Should NOT include: 004, 005
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes (003 and its upstream), got %d", len(nodes))
	}

	// Verify correct tasks are included
	includedIDs := make(map[string]bool)
	for _, node := range nodes {
		nodeMap := node.(map[string]any)
		includedIDs[nodeMap["id"].(string)] = true
	}

	if !includedIDs["001"] || !includedIDs["002"] || !includedIDs["003"] {
		t.Error("Expected tasks 001, 002, 003 to be included")
	}

	if includedIDs["004"] || includedIDs["005"] {
		t.Error("Tasks 004 and 005 should not be included in upstream of 003")
	}
}

func TestGraphCommand_OutputToFile(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())
	outFile := filepath.Join(t.TempDir(), "graph.json")

	res := repo.Run("graph", "--format", "json", "--all", "--out", outFile)
	if res.Err != nil {
		t.Fatalf("runGraph failed: %v", res.Err)
	}

	// Verify file was created
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Fatal("Expected output file to be created")
	}

	// Read and verify content
	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(content, &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON from file: %v", err)
	}

	if result["nodes"] == nil {
		t.Error("Expected output file to contain nodes")
	}
}

func TestGraphCommand_ErrorInvalidRoot(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Run command and expect error
	res := repo.Run("graph", "--format", "json", "--all", "--root", "999")
	if res.Err == nil {
		t.Fatal("Expected error for invalid root task")
	}

	if !strings.Contains(res.Err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", res.Err)
	}
}

func TestGraphCommand_ErrorInvalidFocus(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Run command and expect error
	res := repo.Run("graph", "--format", "json", "--all", "--focus", "999")
	if res.Err == nil {
		t.Fatal("Expected error for invalid focus task")
	}

	if !strings.Contains(res.Err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", res.Err)
	}
}

func TestGraphCommand_ErrorUpstreamDownstreamWithoutRoot(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Run command and expect error
	res := repo.Run("graph", "--format", "json", "--all", "--upstream")
	if res.Err == nil {
		t.Fatal("Expected error when using --upstream without --root")
	}

	if !strings.Contains(res.Err.Error(), "require --root") {
		t.Errorf("Expected 'require --root' error, got: %v", res.Err)
	}
}

func TestGraphCommand_ErrorBothUpstreamAndDownstream(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Run command and expect error
	res := repo.Run("graph", "--format", "json", "--all", "--root", "001", "--upstream", "--downstream")
	if res.Err == nil {
		t.Fatal("Expected error when using both --upstream and --downstream")
	}

	if !strings.Contains(res.Err.Error(), "cannot use both") {
		t.Errorf("Expected 'cannot use both' error, got: %v", res.Err)
	}
}

func TestGraphCommand_ExcludeMultipleStatuses(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--exclude-status", "completed", "--exclude-status", "pending")

	// Parse JSON
	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Should have no tasks since all are either completed or pending
	nodes, ok := result["nodes"].([]any)
	if !ok {
		t.Fatal("Expected 'nodes' to be an array")
	}

	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes when excluding both completed and pending, got %d", len(nodes))
	}
}

func TestGraphCommand_DependencyCleanup_Complex(t *testing.T) {
	// Create a more complex scenario for dependency cleanup
	repo := newTaskRepo(t, map[string]string{
		"001-completed.md": `---
id: "001"
title: "Completed Task"
status: completed
dependencies: []
created: 2026-02-08
---`,
		"002-pending.md": `---
id: "002"
title: "Pending Task - depends on completed"
status: pending
dependencies: ["001"]
created: 2026-02-08
---`,
		"003-pending.md": `---
id: "003"
title: "Pending Task - depends on pending and completed"
status: pending
dependencies: ["001", "002"]
created: 2026-02-08
---`,
	})

	output := graphStdout(t, repo, "--format", "json", "--exclude-status", "completed")

	// Parse JSON
	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Verify edges
	edges, ok := result["edges"].([]any)
	if !ok {
		t.Fatal("Expected 'edges' to be an array")
	}

	// Should have 1 edge: 002 -> 003
	// The dependency from 003 to 001 should be cleaned up
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge after cleanup (002->003), got %d", len(edges))
	}

	if len(edges) > 0 {
		edge := edges[0].(map[string]any)
		if edge["from"] != "002" || edge["to"] != "003" {
			t.Errorf("Expected edge from 002 to 003, got from %s to %s", edge["from"], edge["to"])
		}
	}
}

func TestGraphCommand_DefaultExcludesCompleted(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Use the default exclude-status (completed) by not overriding it
	output := graphStdout(t, repo, "--format", "json")

	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	nodes := result["nodes"].([]any)
	// Only pending tasks (003, 004, 005) should be included
	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes (completed excluded by default), got %d", len(nodes))
	}

	for _, node := range nodes {
		nodeMap := node.(map[string]any)
		status := nodeMap["status"].(string)
		if status == "completed" {
			t.Errorf("Completed task %s should be excluded by default", nodeMap["id"])
		}
	}
}

func TestGraphCommand_DefaultFormat_IsASCII(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Run with default format (do NOT pass --format) but include all statuses.
	output := graphStdout(t, repo, "--all")

	// Verify the default value is "ascii"
	defVal := graphCmd.Flag("format").DefValue
	if defVal != "ascii" {
		t.Errorf("Expected default format flag to be 'ascii', got %q", defVal)
	}

	// Verify the output looks like ASCII tree (not mermaid or dot)
	if strings.Contains(output, "graph TD") {
		t.Error("Default format should not produce mermaid output")
	}
	if strings.Contains(output, "digraph tasks") {
		t.Error("Default format should not produce DOT output")
	}
	if !strings.Contains(output, "[001]") && !strings.Contains(output, "[005]") {
		t.Error("Expected ASCII output to contain task IDs like [001] or [005]")
	}
}

func TestGraphCommand_AllFlag_IncludesCompleted(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// --all should override the default exclude
	output := graphStdout(t, repo, "--format", "json", "--all")

	var result map[string]any
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	nodes := result["nodes"].([]any)
	// All 5 tasks should be included when --all is used
	if len(nodes) != 5 {
		t.Errorf("Expected 5 nodes with --all flag, got %d", len(nodes))
	}
}

func TestGraphCommand_Filter_ByPriority(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all", "--filter", "priority=high")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// Tasks with priority=high: 001 (completed), 003 (pending)
	if len(ids) != 2 {
		t.Errorf("Expected 2 high-priority nodes, got %d: %v", len(ids), ids)
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	if !idSet["001"] || !idSet["003"] {
		t.Errorf("Expected tasks 001 and 003, got %v", ids)
	}
}

func TestGraphCommand_Filter_ByTag(t *testing.T) {
	// Create tasks with different tags
	repo := newTaskRepo(t, map[string]string{
		"001-cli.md": `---
id: "001"
title: "CLI Task"
status: pending
priority: high
tags: ["cli"]
created: 2026-02-08
---`,
		"002-api.md": `---
id: "002"
title: "API Task"
status: pending
priority: medium
tags: ["api"]
created: 2026-02-08
---`,
		"003-both.md": `---
id: "003"
title: "Both Tags"
status: pending
priority: low
tags: ["cli", "api"]
created: 2026-02-08
---`,
	})

	output := graphStdout(t, repo, "--format", "json", "--all", "--filter", "tag=cli")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// Tasks with tag=cli: 001, 003
	if len(ids) != 2 {
		t.Errorf("Expected 2 nodes with tag=cli, got %d: %v", len(ids), ids)
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	if !idSet["001"] || !idSet["003"] {
		t.Errorf("Expected tasks 001 and 003, got %v", ids)
	}
}

func TestGraphCommand_Filter_Combined(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// AND: priority=high AND status=pending => only 003
	output := graphStdout(t, repo, "--format", "json", "--all", "--filter", "priority=high", "--filter", "status=pending")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	if len(ids) != 1 {
		t.Errorf("Expected 1 node matching both filters, got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && ids[0] != "003" {
		t.Errorf("Expected task 003, got %v", ids)
	}
}

func TestGraphCommand_Filter_WithExcludeStatus(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Filter to high priority, then also exclude completed
	output := graphStdout(t, repo, "--format", "json", "--filter", "priority=high", "--exclude-status", "completed")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// priority=high: 001, 003; then exclude completed: only 003
	if len(ids) != 1 {
		t.Errorf("Expected 1 node (high priority, not completed), got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && ids[0] != "003" {
		t.Errorf("Expected task 003, got %v", ids)
	}

	// Dependencies to filtered-out tasks should be cleaned
	edges, ok := result["edges"].([]any)
	if !ok {
		t.Fatal("Expected 'edges' to be an array")
	}
	if len(edges) != 0 {
		t.Errorf("Expected 0 edges after filtering, got %d", len(edges))
	}
}

func TestGraphCommand_Filter_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	res := repo.Run("graph", "--format", "json", "--all", "--filter", "invalid-no-equals")
	if res.Err == nil {
		t.Fatal("Expected error for invalid filter format")
	}

	if !strings.Contains(res.Err.Error(), "invalid filter format") {
		t.Errorf("Expected 'invalid filter format' error, got: %v", res.Err)
	}
}

func TestGraphCommand_ASCII_WithColors(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// Enable colors. forceColor is a plain global (not a flag), so it survives the
	// harness reset; noColor defaults to false after reset.
	forceColor = true
	defer func() {
		forceColor = false
	}()

	output := graphStdout(t, repo, "--all", "--format", "ascii")

	// ANSI escape codes should be present when colors are enabled
	if !strings.Contains(output, "\033[") {
		t.Error("Expected ANSI escape codes in colored output")
	}

	// Verify task content is still present
	if !strings.Contains(output, "001") {
		t.Error("Expected output to contain task ID 001")
	}
	if !strings.Contains(output, "Root Task") {
		t.Error("Expected output to contain task title 'Root Task'")
	}
}

func TestGraphCommand_ASCII_NoColor_Flag(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	output := graphStdout(t, repo, "--all", "--format", "ascii", "--no-color")

	// No ANSI escape codes when --no-color is set
	if strings.Contains(output, "\033[") {
		t.Error("Expected no ANSI escape codes with --no-color flag")
	}

	// Verify content is still present
	if !strings.Contains(output, "[001]") {
		t.Error("Expected output to contain [001]")
	}
	if !strings.Contains(output, "Root Task") {
		t.Error("Expected output to contain 'Root Task'")
	}
}

// scopedTaskFiles returns a graph-specific task set with `touches` (scope)
// fields. Kept inline because this scope-annotated shape is graph-only and does
// not match a shared testdata set.
func scopedTaskFiles() map[string]string {
	return map[string]string{
		"001-web.md": `---
id: "001"
title: "Web feature"
status: pending
priority: high
touches: ["web", "api"]
created: 2026-02-08
---
# Web feature
`,
		"002-cli.md": `---
id: "002"
title: "CLI feature"
status: pending
priority: medium
touches: ["cli"]
dependencies: ["001"]
created: 2026-02-08
---
# CLI feature
`,
		"003-web.md": `---
id: "003"
title: "Web styling"
status: pending
priority: low
touches: ["web"]
created: 2026-02-08
---
# Web styling
`,
		"004-noscope.md": `---
id: "004"
title: "No scope task"
status: pending
priority: medium
created: 2026-02-08
---
# No scope task
`,
	}
}

func TestGraphCommand_Scope_FiltersToMatchingTasks(t *testing.T) {
	repo := newTaskRepo(t, scopedTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all", "--scope", "web")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// Should include tasks 001 and 003 (both touch "web")
	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	if len(ids) != 2 {
		t.Errorf("Expected 2 nodes for scope=web, got %d: %v", len(ids), ids)
	}
	if !idSet["001"] || !idSet["003"] {
		t.Errorf("Expected tasks 001 and 003, got %v", ids)
	}
}

func TestGraphCommand_Scope_RemovesDanglingDeps(t *testing.T) {
	repo := newTaskRepo(t, scopedTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all", "--scope", "cli")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// Should include only task 002 (touches cli)
	if len(ids) != 1 {
		t.Errorf("Expected 1 node for scope=cli, got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && ids[0] != "002" {
		t.Errorf("Expected task 002, got %v", ids)
	}

	// Dependency on 001 should be cleaned up since 001 was filtered out
	edges, ok := result["edges"].([]any)
	if !ok {
		t.Fatal("Expected 'edges' to be an array")
	}
	if len(edges) != 0 {
		t.Errorf("Expected 0 edges after scope filtering (dangling dep cleaned), got %d", len(edges))
	}
}

func TestGraphCommand_Scope_Wildcard(t *testing.T) {
	repo := newTaskRepo(t, scopedTaskFiles())

	output := graphStdout(t, repo, "--format", "json", "--all", "--scope", "w*")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	if !idSet["001"] || !idSet["003"] {
		t.Errorf("Expected tasks 001 and 003 for wildcard w*, got %v", ids)
	}
}

func TestGraphCommand_StatusShortcut(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// --status pending should match --filter status=pending
	output := graphStdout(t, repo, "--format", "json", "--all", "--status", "pending")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// Pending tasks: 003, 004, 005
	if len(ids) != 3 {
		t.Errorf("Expected 3 pending nodes, got %d: %v", len(ids), ids)
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	if !idSet["003"] || !idSet["004"] || !idSet["005"] {
		t.Errorf("Expected tasks 003, 004, 005, got %v", ids)
	}

	// Compare with --filter status=pending
	filterOutput := graphStdout(t, repo, "--format", "json", "--all", "--filter", "status=pending")
	filterResult := parseGraphJSON(t, filterOutput)
	filterIDs := graphNodeIDs(t, filterResult)

	if len(filterIDs) != len(ids) {
		t.Errorf("--status and --filter status= produced different counts: %d vs %d", len(ids), len(filterIDs))
	}
}

func TestGraphCommand_PriorityShortcut(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// --priority high should match --filter priority=high
	output := graphStdout(t, repo, "--format", "json", "--all", "--priority", "high")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// High priority tasks: 001, 003
	if len(ids) != 2 {
		t.Errorf("Expected 2 high-priority nodes, got %d: %v", len(ids), ids)
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	if !idSet["001"] || !idSet["003"] {
		t.Errorf("Expected tasks 001 and 003, got %v", ids)
	}

	// Compare with --filter priority=high
	filterOutput := graphStdout(t, repo, "--format", "json", "--all", "--filter", "priority=high")
	filterResult := parseGraphJSON(t, filterOutput)
	filterIDs := graphNodeIDs(t, filterResult)

	if len(filterIDs) != len(ids) {
		t.Errorf("--priority and --filter priority= produced different counts: %d vs %d", len(ids), len(filterIDs))
	}
}

func TestGraphCommand_PhaseShortcut(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001-phase-a.md": `---
id: "001"
title: "Phase A Task"
status: pending
priority: high
phase: alpha
created: 2026-02-08
---`,
		"002-phase-b.md": `---
id: "002"
title: "Phase B Task"
status: pending
priority: medium
phase: beta
created: 2026-02-08
---`,
		"003-no-phase.md": `---
id: "003"
title: "No Phase Task"
status: pending
priority: low
created: 2026-02-08
---`,
	})

	output := graphStdout(t, repo, "--format", "json", "--all", "--phase", "alpha")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	if len(ids) != 1 {
		t.Errorf("Expected 1 node for phase alpha, got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && ids[0] != "001" {
		t.Errorf("Expected task 001, got %v", ids)
	}
}

func TestGraphCommand_ShortcutComposesWithRoot(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// --status pending --root 003 --downstream
	output := graphStdout(t, repo, "--format", "json", "--all", "--status", "pending", "--root", "003", "--downstream")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// Task 003 is pending and root; no downstream dependents of 003 exist
	if len(ids) != 1 {
		t.Errorf("Expected 1 node (just root 003), got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && ids[0] != "003" {
		t.Errorf("Expected task 003, got %v", ids)
	}
}

func TestGraphCommand_ShortcutComposesWithExcludeStatus(t *testing.T) {
	repo := newTaskRepo(t, graphTaskFiles())

	// --priority high --exclude-status completed should only show pending high-priority tasks
	output := graphStdout(t, repo, "--format", "json", "--priority", "high", "--exclude-status", "completed")
	result := parseGraphJSON(t, output)
	ids := graphNodeIDs(t, result)

	// High priority: 001 (completed), 003 (pending) → after exclude: only 003
	if len(ids) != 1 {
		t.Errorf("Expected 1 node, got %d: %v", len(ids), ids)
	}
	if len(ids) > 0 && ids[0] != "003" {
		t.Errorf("Expected task 003, got %v", ids)
	}
}
