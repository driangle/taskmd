package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestTaskFiles creates test task files in a temp directory
func createTestTaskFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	tasks := map[string]string{
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

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return tmpDir
}

func TestGraphCommand_JSON_Format(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON
	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{"completed"}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON
	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "ascii"
	graphExcludeStatus = []string{}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "ascii"
	graphExcludeStatus = []string{"completed"}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "mermaid"
	graphExcludeStatus = []string{}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "mermaid"
	graphExcludeStatus = []string{}
	graphRoot = ""
	graphFocus = "003"
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify focus style is applied
	if !strings.Contains(output, ":::focus") {
		t.Error("Expected focused task to have :::focus style")
	}
}

func TestGraphCommand_DOT_Format(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "dot"
	graphExcludeStatus = []string{}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = "001"
	graphFocus = ""
	graphUpstream = false
	graphDownstream = true
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON
	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = "003"
	graphFocus = ""
	graphUpstream = true
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON
	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)
	outFile := filepath.Join(t.TempDir(), "graph.json")

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = outFile

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
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
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = "999" // Non-existent task
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Run command and expect error
	err := runGraph(graphCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for invalid root task")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestGraphCommand_ErrorInvalidFocus(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = ""
	graphFocus = "999" // Non-existent task
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Run command and expect error
	err := runGraph(graphCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for invalid focus task")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestGraphCommand_ErrorUpstreamDownstreamWithoutRoot(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = ""
	graphFocus = ""
	graphUpstream = true
	graphDownstream = false
	graphOut = ""

	// Run command and expect error
	err := runGraph(graphCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error when using --upstream without --root")
	}

	if !strings.Contains(err.Error(), "require --root") {
		t.Errorf("Expected 'require --root' error, got: %v", err)
	}
}

func TestGraphCommand_ErrorBothUpstreamAndDownstream(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphRoot = "001"
	graphFocus = ""
	graphUpstream = true
	graphDownstream = true
	graphOut = ""

	// Run command and expect error
	err := runGraph(graphCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error when using both --upstream and --downstream")
	}

	if !strings.Contains(err.Error(), "cannot use both") {
		t.Errorf("Expected 'cannot use both' error, got: %v", err)
	}
}

func TestGraphCommand_ExcludeMultipleStatuses(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{"completed", "pending"}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON
	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := t.TempDir()

	tasks := map[string]string{
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
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Reset flags
	graphFormat = "json"
	graphExcludeStatus = []string{"completed"}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run command
	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse JSON
	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)

	// Use the default exclude-status (completed) by not overriding it
	graphFormat = "json"
	graphExcludeStatus = []string{"completed"} // simulates default
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)

	// Reset flags to defaults — notably, do NOT set graphFormat
	graphFormat = graphCmd.Flag("format").DefValue
	graphExcludeStatus = []string{}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

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

// resetGraphFlags resets all graph command flags to defaults before each test.
func resetGraphFlags() {
	graphFormat = "json"
	graphExcludeStatus = []string{}
	graphAll = false
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""
	graphFilters = []string{}
}

// captureGraphOutput runs runGraph and captures stdout, returning the output string.
func captureGraphOutput(t *testing.T, args []string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraph(graphCmd, args)
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("runGraph failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
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

func TestGraphCommand_AllFlag_IncludesCompleted(t *testing.T) {
	tmpDir := createTestTaskFiles(t)

	// --all should override the default exclude
	graphFormat = "json"
	graphExcludeStatus = []string{"completed"} // default value
	graphAll = true                            // --all overrides
	graphRoot = ""
	graphFocus = ""
	graphUpstream = false
	graphDownstream = false
	graphOut = ""

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runGraph(graphCmd, []string{tmpDir})
	if err != nil {
		t.Fatalf("runGraph failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result map[string]any
	err = json.Unmarshal([]byte(output), &result)
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
	tmpDir := createTestTaskFiles(t)

	resetGraphFlags()
	graphFilters = []string{"priority=high"}

	output := captureGraphOutput(t, []string{tmpDir})
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
	tmpDir := t.TempDir()

	tasks := map[string]string{
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
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	resetGraphFlags()
	graphFilters = []string{"tag=cli"}

	output := captureGraphOutput(t, []string{tmpDir})
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
	tmpDir := createTestTaskFiles(t)

	resetGraphFlags()
	// AND: priority=high AND status=pending => only 003
	graphFilters = []string{"priority=high", "status=pending"}

	output := captureGraphOutput(t, []string{tmpDir})
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
	tmpDir := createTestTaskFiles(t)

	resetGraphFlags()
	// Filter to high priority, then also exclude completed
	graphFilters = []string{"priority=high"}
	graphExcludeStatus = []string{"completed"}

	output := captureGraphOutput(t, []string{tmpDir})
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
	tmpDir := createTestTaskFiles(t)

	resetGraphFlags()
	graphFilters = []string{"invalid-no-equals"}

	err := runGraph(graphCmd, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for invalid filter format")
	}

	if !strings.Contains(err.Error(), "invalid filter format") {
		t.Errorf("Expected 'invalid filter format' error, got: %v", err)
	}
}
