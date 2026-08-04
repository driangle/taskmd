package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

// snapshotFiles is the snapshot-specific task set: a 001 -> 002 -> 003 chain
// (with effort/group fields the derived-field tests depend on) plus an
// independent task 004. Kept inline because these exact dependency and derived
// shapes are the subject of the command-level snapshot tests.
func snapshotFiles() map[string]string {
	return map[string]string{
		"001-root.md": `---
id: "001"
title: "Root Task"
status: completed
priority: high
effort: small
dependencies: []
tags: ["core"]
created: 2026-02-08
---

Root task.
`,
		"002-depends-on-001.md": `---
id: "002"
title: "Middle Task"
status: pending
priority: medium
effort: medium
dependencies: ["001"]
tags: ["core"]
created: 2026-02-08
---

Depends on 001.
`,
		"003-depends-on-002.md": `---
id: "003"
title: "Leaf Task"
status: pending
priority: low
effort: large
dependencies: ["002"]
tags: ["extra"]
group: "backend"
created: 2026-02-08
---

Depends on 002, forming a chain 001 -> 002 -> 003.
`,
		"004-no-deps.md": `---
id: "004"
title: "Independent Task"
status: pending
priority: high
effort: small
dependencies: []
tags: []
created: 2026-02-08
---

No dependencies.
`,
	}
}

// snapshotStdout runs `snapshot <args...>` against repo, fails on error, and
// returns stdout.
func snapshotStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"snapshot"}, args...)...)
	if res.Err != nil {
		t.Fatalf("snapshot %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

// --- Analysis function tests (pure, no I/O) ---

func TestBuildTaskMap(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "A"},
		{ID: "002", Title: "B"},
	}

	taskMap := buildTaskMap(tasks)

	if len(taskMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(taskMap))
	}
	if taskMap["001"].Title != "A" {
		t.Errorf("taskMap[001].Title = %q, want %q", taskMap["001"].Title, "A")
	}
	if taskMap["002"].Title != "B" {
		t.Errorf("taskMap[002].Title = %q, want %q", taskMap["002"].Title, "B")
	}
}

func TestIsTaskBlocked(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Done", Status: model.StatusCompleted},
		{ID: "002", Title: "Pending", Status: model.StatusPending, Dependencies: []string{"001"}},
		{ID: "003", Title: "Blocked", Status: model.StatusPending, Dependencies: []string{"001", "999"}},
		{ID: "004", Title: "No deps", Status: model.StatusPending},
	}
	taskMap := buildTaskMap(tasks)

	tests := []struct {
		name    string
		taskID  string
		blocked bool
	}{
		{"all deps completed but task pending", "002", true},
		{"missing dep", "003", true},
		{"no deps", "004", false},
		{"completed task with deps", "001", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := taskMap[tt.taskID]
			got := isTaskBlocked(task, taskMap)
			if got != tt.blocked {
				t.Errorf("isTaskBlocked(%s) = %v, want %v", tt.taskID, got, tt.blocked)
			}
		})
	}
}

func TestGroupSnapshots(t *testing.T) {
	snapshots := []TaskSnapshot{
		{ID: "001", Status: "completed", Priority: "high", Effort: "small", Group: "backend"},
		{ID: "002", Status: "pending", Priority: "medium", Effort: "medium"},
		{ID: "003", Status: "pending", Priority: "low", Effort: "large", Group: "frontend"},
	}

	tests := []struct {
		name    string
		groupBy string
		wantKey string
	}{
		{"by status", "status", "pending"},
		{"by priority", "priority", "high"},
		{"by effort", "effort", "small"},
		{"by group", "group", "backend"},
		{"unknown field", "unknown", "ungrouped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := groupSnapshots(snapshots, tt.groupBy)
			if _, ok := groups[tt.wantKey]; !ok {
				t.Errorf("expected key %q in groups, got keys: %v", tt.wantKey, groupKeys(groups))
			}
		})
	}

	// Empty field → "none"
	t.Run("empty field becomes none", func(t *testing.T) {
		groups := groupSnapshots(snapshots, "group")
		if _, ok := groups["none"]; !ok {
			t.Errorf("expected 'none' key for empty group field, got keys: %v", groupKeys(groups))
		}
	})
}

func groupKeys(m map[string][]TaskSnapshot) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func TestCalculateDepthMap(t *testing.T) {
	// Chain: 001 <- 002 <- 003
	tasks := []*model.Task{
		{ID: "001", Dependencies: []string{}},
		{ID: "002", Dependencies: []string{"001"}},
		{ID: "003", Dependencies: []string{"002"}},
	}
	taskMap := buildTaskMap(tasks)
	depthMap := calculateDepthMap(tasks, taskMap)

	if depthMap["001"] != 1 {
		t.Errorf("depth[001] = %d, want 1", depthMap["001"])
	}
	if depthMap["002"] != 2 {
		t.Errorf("depth[002] = %d, want 2", depthMap["002"])
	}
	if depthMap["003"] != 3 {
		t.Errorf("depth[003] = %d, want 3", depthMap["003"])
	}
}

func TestCalculateDepthMap_NoDeps(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Dependencies: []string{}},
		{ID: "002", Dependencies: []string{}},
	}
	taskMap := buildTaskMap(tasks)
	depthMap := calculateDepthMap(tasks, taskMap)

	if depthMap["001"] != 1 {
		t.Errorf("depth[001] = %d, want 1", depthMap["001"])
	}
	if depthMap["002"] != 1 {
		t.Errorf("depth[002] = %d, want 1", depthMap["002"])
	}
}

func TestCalculateTopologicalOrder(t *testing.T) {
	// Chain: 001 <- 002 <- 003
	tasks := []*model.Task{
		{ID: "001", Dependencies: []string{}},
		{ID: "002", Dependencies: []string{"001"}},
		{ID: "003", Dependencies: []string{"002"}},
	}
	taskMap := buildTaskMap(tasks)
	order := calculateTopologicalOrder(tasks, taskMap)

	// Dependencies must have lower order than dependents
	if order["001"] >= order["002"] {
		t.Errorf("order[001]=%d should be < order[002]=%d", order["001"], order["002"])
	}
	if order["002"] >= order["003"] {
		t.Errorf("order[002]=%d should be < order[003]=%d", order["002"], order["003"])
	}

	// All tasks should get unique orders
	seen := make(map[int]string)
	for id, o := range order {
		if prev, exists := seen[o]; exists {
			t.Errorf("duplicate order %d for tasks %s and %s", o, prev, id)
		}
		seen[o] = id
	}
}

func TestCalculateCriticalPathTasks(t *testing.T) {
	// Chain: 001 <- 002 <- 003 (depth 3, longest)
	// Branch: 001 <- 004 (depth 2, shorter)
	tasks := []*model.Task{
		{ID: "001", Dependencies: []string{}},
		{ID: "002", Dependencies: []string{"001"}},
		{ID: "003", Dependencies: []string{"002"}},
		{ID: "004", Dependencies: []string{"001"}},
	}
	taskMap := buildTaskMap(tasks)
	critical := calculateCriticalPathTasks(tasks, taskMap)

	// 001, 002, 003 should be on critical path
	if !critical["001"] {
		t.Error("expected 001 on critical path")
	}
	if !critical["002"] {
		t.Error("expected 002 on critical path")
	}
	if !critical["003"] {
		t.Error("expected 003 on critical path")
	}
	// 004 is on a shorter branch
	if critical["004"] {
		t.Error("expected 004 NOT on critical path")
	}
}

func TestTaskToSnapshot_CoreOnly(t *testing.T) {
	task := &model.Task{
		ID:       "001",
		Title:    "Test",
		Status:   model.StatusPending,
		Priority: model.PriorityHigh,
		Effort:   model.EffortSmall,
		Group:    "backend",
		FilePath: "tasks/001.md",
	}
	taskMap := map[string]*model.Task{"001": task}

	snapshot := taskToSnapshot(task, true, false, nil, nil, nil, taskMap)

	if snapshot.ID != "001" {
		t.Errorf("ID = %q, want %q", snapshot.ID, "001")
	}
	if snapshot.Title != "Test" {
		t.Errorf("Title = %q, want %q", snapshot.Title, "Test")
	}
	// Core-only: status, priority, effort, group, filepath should be omitted
	if snapshot.Status != "" {
		t.Errorf("Status = %q, want empty (core-only mode)", snapshot.Status)
	}
	if snapshot.Priority != "" {
		t.Errorf("Priority = %q, want empty (core-only mode)", snapshot.Priority)
	}
	if snapshot.FilePath != "" {
		t.Errorf("FilePath = %q, want empty (core-only mode)", snapshot.FilePath)
	}
}

func TestTaskToSnapshot_WithDerived(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Status: model.StatusCompleted, Dependencies: []string{}},
		{ID: "002", Status: model.StatusPending, Dependencies: []string{"001"}},
	}
	taskMap := buildTaskMap(tasks)
	depthMap := calculateDepthMap(tasks, taskMap)
	topoOrder := calculateTopologicalOrder(tasks, taskMap)
	criticalPath := calculateCriticalPathTasks(tasks, taskMap)

	snapshot := taskToSnapshot(tasks[1], false, true, depthMap, topoOrder, criticalPath, taskMap)

	if snapshot.IsBlocked == nil {
		t.Fatal("expected IsBlocked to be set")
	}
	if snapshot.DependencyDepth == nil {
		t.Fatal("expected DependencyDepth to be set")
	}
	if snapshot.TopologicalOrder == nil {
		t.Fatal("expected TopologicalOrder to be set")
	}
}

// --- Command-level tests ---

func TestRunSnapshot_JSONOutput(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo)

	var result SnapshotOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	if len(result.Tasks) != 4 {
		t.Errorf("expected 4 tasks, got %d", len(result.Tasks))
	}
}

func TestRunSnapshot_YAMLOutput(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo, "--format", "yaml")

	if !strings.Contains(output, "id:") || !strings.Contains(output, "title:") {
		t.Errorf("YAML output missing expected fields:\n%s", output)
	}
}

func TestRunSnapshot_MarkdownOutput(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo, "--format", "md")

	if !strings.Contains(output, "###") {
		t.Errorf("markdown output missing headings:\n%s", output)
	}
}

func TestRunSnapshot_CoreFlag(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo, "--core")

	var result SnapshotOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	for _, snap := range result.Tasks {
		if snap.Status != "" {
			t.Errorf("core-only: task %s has status %q, want empty", snap.ID, snap.Status)
		}
		if snap.Priority != "" {
			t.Errorf("core-only: task %s has priority %q, want empty", snap.ID, snap.Priority)
		}
	}
}

func TestRunSnapshot_DerivedFlag(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo, "--derived")

	var result SnapshotOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// At least one task should have derived fields
	foundDerived := false
	for _, snap := range result.Tasks {
		if snap.IsBlocked != nil || snap.DependencyDepth != nil {
			foundDerived = true
			break
		}
	}
	if !foundDerived {
		t.Error("derived flag set but no derived fields found in output")
	}
}

func TestRunSnapshot_GroupBy(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo, "--group-by", "status")

	var result SnapshotOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(result.Groups) == 0 {
		t.Error("expected grouped output, got empty groups")
	}
}

func TestRunSnapshot_FileOutput(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	outFile := repo.Path("snapshot.json")

	snapshotStdout(t, repo, "--out", outFile)

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var result SnapshotOutput
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("output file is not valid JSON: %v", err)
	}
	if len(result.Tasks) != 4 {
		t.Errorf("expected 4 tasks in file, got %d", len(result.Tasks))
	}
}

func TestRunSnapshot_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	res := repo.Run("snapshot", "--format", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestRunSnapshot_MarkdownGroupedByStatus(t *testing.T) {
	repo := newTaskRepo(t, snapshotFiles())

	output := snapshotStdout(t, repo, "--format", "md", "--group-by", "status")

	// Should have group headings (capitalized status values)
	if !strings.Contains(output, "## Completed") {
		t.Errorf("expected '## Completed' heading in grouped markdown, got:\n%s", output)
	}
	if !strings.Contains(output, "## Pending") {
		t.Errorf("expected '## Pending' heading in grouped markdown, got:\n%s", output)
	}

	// Should still contain task entries
	if !strings.Contains(output, "### [001]") {
		t.Errorf("expected task 001 entry in output, got:\n%s", output)
	}
}

func TestOutputSnapshotMarkdown_Grouped(t *testing.T) {
	snapshots := []TaskSnapshot{
		{ID: "001", Title: "Done Task", Status: "completed", Priority: "high"},
		{ID: "002", Title: "Pending Task", Status: "pending", Priority: "medium"},
		{ID: "003", Title: "Another Pending", Status: "pending", Priority: "low"},
	}

	outPath := filepath.Join(t.TempDir(), "out.md")
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	err = outputSnapshotMarkdown(snapshots, f, "status")
	f.Close()
	if err != nil {
		t.Fatalf("outputSnapshotMarkdown failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	output := string(data)

	if !strings.Contains(output, "## Completed") {
		t.Errorf("expected '## Completed' heading, got:\n%s", output)
	}
	if !strings.Contains(output, "## Pending") {
		t.Errorf("expected '## Pending' heading, got:\n%s", output)
	}
	if !strings.Contains(output, "### [001] Done Task") {
		t.Errorf("expected task 001 entry, got:\n%s", output)
	}
	if !strings.Contains(output, "### [002] Pending Task") {
		t.Errorf("expected task 002 entry, got:\n%s", output)
	}
}

func TestOutputSnapshotMarkdown_Ungrouped(t *testing.T) {
	snapshots := []TaskSnapshot{
		{ID: "001", Title: "Task A", Status: "pending"},
		{ID: "002", Title: "Task B", Status: "completed"},
	}

	outPath := filepath.Join(t.TempDir(), "out.md")
	f, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	err = outputSnapshotMarkdown(snapshots, f, "")
	f.Close()
	if err != nil {
		t.Fatalf("outputSnapshotMarkdown failed: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	output := string(data)

	// Should NOT have level-2 group headings (but will have ### task headings)
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "## ") {
			t.Errorf("expected no group headings in ungrouped output, found: %s", line)
		}
	}
	// Should have task entries
	if !strings.Contains(output, "### [001] Task A") {
		t.Errorf("expected task 001 entry, got:\n%s", output)
	}
}
