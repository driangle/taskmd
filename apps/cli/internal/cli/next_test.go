package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/next"
)

// createNextTestTaskFiles creates a set of 10 task files designed to exercise
// the next command's scoring, filtering, and actionability logic.
//
// Task graph:
//
//	001 (completed, high, small)         - root, completed
//	002 (completed, medium, medium)      - depends on 001, completed
//	003 (pending, critical, small, cli)  - depends on 001 (completed) → actionable
//	004 (pending, high, large, cli)      - depends on 002 (completed) → actionable
//	005 (pending, low, small)            - no deps → actionable
//	006 (pending, high, medium)          - depends on 007 (pending) → blocked
//	007 (pending, medium, small)         - no deps → actionable
//	008 (in-progress, high, small, cli)  - depends on 001 (completed) → actionable
//	009 (pending, medium, large)         - depends on 006 (pending) → blocked
//	010 (pending, low, medium)           - depends on 003 → blocked (003 pending)
func createNextTestTaskFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	tasks := map[string]string{
		"001.md": `---
id: "001"
title: "Setup infrastructure"
status: completed
priority: high
effort: small
dependencies: []
tags: ["infra"]
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "Design API schema"
status: completed
priority: medium
effort: medium
dependencies: ["001"]
tags: ["api"]
created: 2026-02-02
---`,
		"003.md": `---
id: "003"
title: "Build CLI parser"
status: pending
priority: critical
effort: small
dependencies: ["001"]
tags: ["cli"]
created: 2026-02-03
---`,
		"004.md": `---
id: "004"
title: "Implement API endpoints"
status: pending
priority: high
effort: large
dependencies: ["002"]
tags: ["cli", "api"]
created: 2026-02-04
---`,
		"005.md": `---
id: "005"
title: "Write README"
status: pending
priority: low
effort: small
dependencies: []
tags: ["docs"]
created: 2026-02-05
---`,
		"006.md": `---
id: "006"
title: "Add authentication"
status: pending
priority: high
effort: medium
dependencies: ["007"]
tags: ["api"]
created: 2026-02-06
---`,
		"007.md": `---
id: "007"
title: "Create user model"
status: pending
priority: medium
effort: small
dependencies: []
tags: ["api"]
created: 2026-02-07
---`,
		"008.md": `---
id: "008"
title: "CLI help text"
status: in-progress
priority: high
effort: small
dependencies: ["001"]
tags: ["cli"]
created: 2026-02-08
---`,
		"009.md": `---
id: "009"
title: "Add OAuth support"
status: pending
priority: medium
effort: large
dependencies: ["006"]
tags: ["api"]
created: 2026-02-09
---`,
		"010.md": `---
id: "010"
title: "CLI integration tests"
status: pending
priority: low
effort: medium
dependencies: ["003"]
tags: ["cli", "test"]
created: 2026-02-10
---`,
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	return tmpDir
}

// captureNextOutput runs runNext and captures stdout
func captureNextOutput(t *testing.T, args []string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runNext(nextCmd, args)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

// resetNextFlags resets all next command flags to defaults
func resetNextFlags() {
	nextFormat = "table"
	nextLimit = 5
	nextFilters = []string{}
	nextQuickWins = false
	nextCritical = false
}

func TestNext_BasicRanking(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Actionable tasks: 003, 004, 005, 007, 008
	// Blocked: 006 (dep 007 pending), 009 (dep 006 pending), 010 (dep 003 pending)
	// Completed: 001, 002
	if len(recs) != 5 {
		t.Errorf("Expected 5 actionable tasks, got %d", len(recs))
		for _, r := range recs {
			t.Logf("  %s: %s (score=%d)", r.ID, r.Title, r.Score)
		}
	}

	// Verify sorted by score descending
	for i := 1; i < len(recs); i++ {
		if recs[i].Score > recs[i-1].Score {
			t.Errorf("Recommendations not sorted by score: [%d]=%d > [%d]=%d",
				i, recs[i].Score, i-1, recs[i-1].Score)
		}
	}

	// Verify rank field
	for i, rec := range recs {
		if rec.Rank != i+1 {
			t.Errorf("Expected rank %d, got %d for %s", i+1, rec.Rank, rec.ID)
		}
	}
}

func TestNext_BlockedTasksExcluded(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 20

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	blockedIDs := map[string]bool{"006": true, "009": true, "010": true}
	completedIDs := map[string]bool{"001": true, "002": true}

	for _, rec := range recs {
		if blockedIDs[rec.ID] {
			t.Errorf("Blocked task %s should not appear in recommendations", rec.ID)
		}
		if completedIDs[rec.ID] {
			t.Errorf("Completed task %s should not appear in recommendations", rec.ID)
		}
	}
}

func TestNext_CancelledTasksExcluded(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test tasks including cancelled ones
	tasks := map[string]string{
		"001-active.md": `---
id: "001"
title: "Active pending task"
status: pending
priority: high
effort: small
dependencies: []
tags: ["active"]
created: 2026-02-12
---`,
		"002-cancelled.md": `---
id: "002"
title: "Cancelled task"
status: cancelled
priority: critical
effort: small
dependencies: []
tags: ["old"]
created: 2026-02-10
---`,
		"003-completed.md": `---
id: "003"
title: "Completed task"
status: completed
priority: high
effort: medium
dependencies: []
tags: ["done"]
created: 2026-02-11
---`,
		"004-another-active.md": `---
id: "004"
title: "Another active task"
status: in-progress
priority: medium
effort: small
dependencies: []
tags: ["active"]
created: 2026-02-12
---`,
	}

	for filename, content := range tasks {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 20

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify no cancelled or completed tasks in recommendations
	for _, rec := range recs {
		if rec.ID == "002" {
			t.Errorf("Cancelled task 002 should NOT appear in recommendations")
		}
		if rec.ID == "003" {
			t.Errorf("Completed task 003 should NOT appear in recommendations")
		}
	}

	// Verify only actionable tasks (pending/in-progress) are included
	expectedIDs := map[string]bool{"001": true, "004": true}
	if len(recs) != len(expectedIDs) {
		t.Errorf("Expected %d actionable tasks, got %d", len(expectedIDs), len(recs))
	}

	for _, rec := range recs {
		if !expectedIDs[rec.ID] {
			t.Errorf("Unexpected task %s in recommendations", rec.ID)
		}
	}
}

func TestNext_LimitFlag(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 2

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recs) != 2 {
		t.Errorf("Expected 2 recommendations with --limit 2, got %d", len(recs))
	}
}

func TestNext_LimitExceedsAvailable(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 100

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should return all 5 actionable tasks, not error
	if len(recs) != 5 {
		t.Errorf("Expected 5 recommendations (all actionable), got %d", len(recs))
	}
}

func TestNext_FilterByTag(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10
	nextFilters = []string{"tag=cli"}

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// CLI-tagged actionable tasks: 003, 004, 008
	// 010 is CLI-tagged but blocked (dep 003 pending)
	if len(recs) != 3 {
		t.Errorf("Expected 3 CLI-tagged actionable tasks, got %d", len(recs))
		for _, r := range recs {
			t.Logf("  %s: %s", r.ID, r.Title)
		}
	}

	for _, rec := range recs {
		if rec.ID != "003" && rec.ID != "004" && rec.ID != "008" {
			t.Errorf("Unexpected task %s in CLI filter results", rec.ID)
		}
	}
}

func TestNext_FilterByPriority(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10
	nextFilters = []string{"priority=high"}

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// High priority actionable: 004, 008
	// 006 is high but blocked
	if len(recs) != 2 {
		t.Errorf("Expected 2 high-priority actionable tasks, got %d", len(recs))
		for _, r := range recs {
			t.Logf("  %s: %s (priority=%s)", r.ID, r.Title, r.Priority)
		}
	}

	for _, rec := range recs {
		if rec.Priority != "high" {
			t.Errorf("Expected priority=high, got %s for task %s", rec.Priority, rec.ID)
		}
	}
}

func TestNext_MultipleFilters(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10
	nextFilters = []string{"tag=cli", "priority=high"}

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// CLI + high priority actionable: 004, 008
	if len(recs) != 2 {
		t.Errorf("Expected 2 tasks matching tag=cli AND priority=high, got %d", len(recs))
	}
}

func TestNext_InvalidFilterFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextFilters = []string{"invalid"}

	_, err := captureNextOutput(t, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for invalid filter format")
	}

	if !strings.Contains(err.Error(), "invalid filter format") {
		t.Errorf("Expected 'invalid filter format' error, got: %v", err)
	}
}

func TestNext_UnsupportedFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "csv"

	_, err := captureNextOutput(t, []string{tmpDir})
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestNext_JSONFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 2

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	if len(recs) != 2 {
		t.Fatalf("Expected 2 recommendations, got %d", len(recs))
	}

	// Verify all fields are present
	rec := recs[0]
	if rec.Rank != 1 {
		t.Errorf("Expected rank=1, got %d", rec.Rank)
	}
	if rec.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if rec.Title == "" {
		t.Error("Expected non-empty Title")
	}
	if rec.Score <= 0 {
		t.Errorf("Expected positive score, got %d", rec.Score)
	}
	if len(rec.Reasons) == 0 {
		t.Error("Expected at least one reason")
	}
}

func TestNext_YAMLFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "yaml"
	nextLimit = 2

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	// Basic YAML structure check
	if !strings.Contains(output, "rank:") {
		t.Error("Expected YAML output to contain 'rank:'")
	}
	if !strings.Contains(output, "id:") {
		t.Error("Expected YAML output to contain 'id:'")
	}
	if !strings.Contains(output, "reasons:") {
		t.Error("Expected YAML output to contain 'reasons:'")
	}
}

func TestNext_TableFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "table"
	nextLimit = 3

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	if !strings.Contains(output, "Recommended tasks:") {
		t.Error("Expected table header 'Recommended tasks:'")
	}
	if !strings.Contains(output, "#") && !strings.Contains(output, "ID") {
		t.Error("Expected table column headers")
	}
}

func TestNext_NoActionableTasks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only completed tasks
	tasks := map[string]string{
		"001.md": `---
id: "001"
title: "Done task"
status: completed
priority: high
dependencies: []
created: 2026-02-01
---`,
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	resetNextFlags()
	nextFormat = "table"

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	if !strings.Contains(output, "No actionable tasks found") {
		t.Errorf("Expected 'No actionable tasks found' message, got: %s", output)
	}
}

func TestNext_InProgressTasksIncluded(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	found := false
	for _, rec := range recs {
		if rec.ID == "008" {
			found = true
			if rec.Status != "in-progress" {
				t.Errorf("Expected task 008 status=in-progress, got %s", rec.Status)
			}
		}
	}

	if !found {
		t.Error("Expected in-progress task 008 to be in recommendations")
	}
}

func TestNext_ReasonStrings(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	recMap := make(map[string]Recommendation)
	for _, rec := range recs {
		recMap[rec.ID] = rec
	}

	// Task 003: critical priority, small effort
	if rec, ok := recMap["003"]; ok {
		reasons := strings.Join(rec.Reasons, " ")
		if !strings.Contains(reasons, "critical priority") {
			t.Errorf("Expected 'critical priority' reason for task 003, got: %v", rec.Reasons)
		}
		if !strings.Contains(reasons, "quick win") {
			t.Errorf("Expected 'quick win' reason for task 003, got: %v", rec.Reasons)
		}
	} else {
		t.Error("Task 003 not found in recommendations")
	}

	// Task 007: medium priority, small effort, has downstream (006 depends on it)
	if rec, ok := recMap["007"]; ok {
		reasons := strings.Join(rec.Reasons, " ")
		if !strings.Contains(reasons, "unblocks") {
			t.Errorf("Expected 'unblocks' reason for task 007, got: %v", rec.Reasons)
		}
		if !strings.Contains(reasons, "quick win") {
			t.Errorf("Expected 'quick win' reason for task 007, got: %v", rec.Reasons)
		}
	} else {
		t.Error("Task 007 not found in recommendations")
	}

	// Task 008: high priority, small effort
	if rec, ok := recMap["008"]; ok {
		reasons := strings.Join(rec.Reasons, " ")
		if !strings.Contains(reasons, "high priority") {
			t.Errorf("Expected 'high priority' reason for task 008, got: %v", rec.Reasons)
		}
	} else {
		t.Error("Task 008 not found in recommendations")
	}
}

func TestNext_ScoringOrder(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	recMap := make(map[string]Recommendation)
	for _, rec := range recs {
		recMap[rec.ID] = rec
	}

	// Task 003 (critical+small) should score higher than 005 (low+small)
	if recMap["003"].Score <= recMap["005"].Score {
		t.Errorf("Expected task 003 (critical) to score higher than 005 (low): %d <= %d",
			recMap["003"].Score, recMap["005"].Score)
	}

	// Task 008 (high+small) should score higher than 005 (low+small)
	if recMap["008"].Score <= recMap["005"].Score {
		t.Errorf("Expected task 008 (high+small) to score higher than 005 (low+small): %d <= %d",
			recMap["008"].Score, recMap["005"].Score)
	}
}

func TestNext_TiedScoresBreakByID(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two identical-scoring tasks with different IDs
	tasks := map[string]string{
		"bbb.md": `---
id: "BBB"
title: "Task BBB"
status: pending
priority: medium
effort: medium
dependencies: []
created: 2026-02-01
---`,
		"aaa.md": `---
id: "AAA"
title: "Task AAA"
status: pending
priority: medium
effort: medium
dependencies: []
created: 2026-02-01
---`,
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recs) != 2 {
		t.Fatalf("Expected 2 recommendations, got %d", len(recs))
	}

	// Same score → alphabetical by ID
	if recs[0].ID != "AAA" || recs[1].ID != "BBB" {
		t.Errorf("Expected tied scores to sort by ID asc: got %s, %s", recs[0].ID, recs[1].ID)
	}
}

// Unit tests for helper functions

func TestHasUnmetDependencies(t *testing.T) {
	taskMap := map[string]*model.Task{
		"001": {ID: "001", Status: model.StatusCompleted},
		"002": {ID: "002", Status: model.StatusPending},
	}

	tests := []struct {
		name     string
		task     *model.Task
		expected bool
	}{
		{
			name:     "no dependencies",
			task:     &model.Task{ID: "100", Dependencies: nil},
			expected: false,
		},
		{
			name:     "all deps completed",
			task:     &model.Task{ID: "100", Dependencies: []string{"001"}},
			expected: false,
		},
		{
			name:     "dep pending",
			task:     &model.Task{ID: "100", Dependencies: []string{"002"}},
			expected: true,
		},
		{
			name:     "dep missing",
			task:     &model.Task{ID: "100", Dependencies: []string{"999"}},
			expected: true,
		},
		{
			name:     "mixed deps",
			task:     &model.Task{ID: "100", Dependencies: []string{"001", "002"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := next.HasUnmetDependencies(tt.task, taskMap)
			if got != tt.expected {
				t.Errorf("hasUnmetDependencies() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsActionable(t *testing.T) {
	taskMap := map[string]*model.Task{
		"001": {ID: "001", Status: model.StatusCompleted},
		"002": {ID: "002", Status: model.StatusPending},
	}

	tests := []struct {
		name     string
		task     *model.Task
		expected bool
	}{
		{
			name:     "pending no deps",
			task:     &model.Task{ID: "100", Status: model.StatusPending},
			expected: true,
		},
		{
			name:     "pending all deps completed",
			task:     &model.Task{ID: "100", Status: model.StatusPending, Dependencies: []string{"001"}},
			expected: true,
		},
		{
			name:     "pending unmet dep",
			task:     &model.Task{ID: "100", Status: model.StatusPending, Dependencies: []string{"002"}},
			expected: false,
		},
		{
			name:     "in-progress no deps",
			task:     &model.Task{ID: "100", Status: model.StatusInProgress},
			expected: true,
		},
		{
			name:     "completed",
			task:     &model.Task{ID: "100", Status: model.StatusCompleted},
			expected: false,
		},
		{
			name:     "blocked status",
			task:     &model.Task{ID: "100", Status: model.StatusBlocked},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := next.IsActionable(tt.task, taskMap)
			if got != tt.expected {
				t.Errorf("isActionable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestScoreTask(t *testing.T) {
	criticalPath := map[string]bool{"cp1": true}
	downstreamCounts := map[string]int{
		"cp1": 3,
		"ds1": 1,
		"ds6": 6,
	}

	tests := []struct {
		name          string
		task          *model.Task
		expectedScore int
		expectReason  string
	}{
		{
			name:          "critical priority",
			task:          &model.Task{ID: "t1", Priority: model.PriorityCritical},
			expectedScore: next.ScorePriorityCritical,
			expectReason:  "critical priority",
		},
		{
			name:          "high priority",
			task:          &model.Task{ID: "t2", Priority: model.PriorityHigh},
			expectedScore: next.ScorePriorityHigh,
			expectReason:  "high priority",
		},
		{
			name:          "medium priority no special reason",
			task:          &model.Task{ID: "t3", Priority: model.PriorityMedium},
			expectedScore: next.ScorePriorityMedium,
		},
		{
			name:          "low/unset priority",
			task:          &model.Task{ID: "t4"},
			expectedScore: next.ScorePriorityLow,
		},
		{
			name:          "small effort bonus",
			task:          &model.Task{ID: "t5", Effort: model.EffortSmall},
			expectedScore: next.ScorePriorityLow + next.ScoreEffortSmall,
			expectReason:  "quick win",
		},
		{
			name:          "critical path bonus",
			task:          &model.Task{ID: "cp1", Priority: model.PriorityMedium, Effort: model.EffortMedium},
			expectedScore: next.ScorePriorityMedium + next.ScoreCriticalPath + min(3*next.ScorePerDownstream, next.ScoreDownstreamMax) + next.ScoreEffortMedium,
			expectReason:  "on critical path",
		},
		{
			name:          "downstream 1 task",
			task:          &model.Task{ID: "ds1", Priority: model.PriorityMedium},
			expectedScore: next.ScorePriorityMedium + 1*next.ScorePerDownstream,
			expectReason:  "unblocks 1 task",
		},
		{
			name:          "downstream capped at max",
			task:          &model.Task{ID: "ds6", Priority: model.PriorityMedium},
			expectedScore: next.ScorePriorityMedium + next.ScoreDownstreamMax,
			expectReason:  "unblocks 6 tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, reasons := next.ScoreTask(tt.task, criticalPath, downstreamCounts)
			if score != tt.expectedScore {
				t.Errorf("scoreTask() score = %d, want %d", score, tt.expectedScore)
			}
			if tt.expectReason != "" {
				found := false
				for _, r := range reasons {
					if strings.Contains(r, tt.expectReason) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected reason containing %q, got %v", tt.expectReason, reasons)
				}
			}
		})
	}
}

func TestNext_DownstreamCountUsesFullGraph(t *testing.T) {
	// Verify that downstream counts reflect the full graph, not just filtered results
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextLimit = 10
	nextFilters = []string{"tag=api"}

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Task 007 should have downstream count reflecting ALL tasks that depend on it
	// (006, 009) — computed from the full graph, even though we filtered by tag=api
	for _, rec := range recs {
		if rec.ID == "007" {
			if rec.DownstreamCount < 2 {
				t.Errorf("Expected task 007 downstream_count >= 2 (full graph), got %d", rec.DownstreamCount)
			}
			return
		}
	}

	t.Error("Expected task 007 in api-filtered results")
}

func TestNext_QuickWins_HappyPath(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextQuickWins = true
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Actionable small tasks: 003, 005, 007, 008
	// NOT included: 004 (large), blocked tasks
	if len(recs) != 4 {
		t.Errorf("Expected 4 quick wins, got %d", len(recs))
		for _, r := range recs {
			t.Logf("  %s: %s (effort=%s)", r.ID, r.Title, r.Effort)
		}
	}

	// Verify all returned tasks have effort: small
	for _, rec := range recs {
		if rec.Effort != "small" {
			t.Errorf("Quick wins should only include small effort tasks, got %s for task %s", rec.Effort, rec.ID)
		}
	}

	// Verify expected tasks
	expectedIDs := map[string]bool{"003": true, "005": true, "007": true, "008": true}
	for _, rec := range recs {
		if !expectedIDs[rec.ID] {
			t.Errorf("Unexpected task %s in quick wins", rec.ID)
		}
	}
}

func TestNext_QuickWins_WithFilter(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextQuickWins = true
	nextFilters = []string{"tag=cli"}
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// CLI-tagged, actionable, small effort: 003, 008
	if len(recs) != 2 {
		t.Errorf("Expected 2 CLI quick wins, got %d", len(recs))
		for _, r := range recs {
			t.Logf("  %s: %s (effort=%s)", r.ID, r.Title, r.Effort)
		}
	}

	for _, rec := range recs {
		if rec.Effort != "small" {
			t.Errorf("Expected small effort, got %s for task %s", rec.Effort, rec.ID)
		}
		if rec.ID != "003" && rec.ID != "008" {
			t.Errorf("Unexpected task %s in filtered quick wins", rec.ID)
		}
	}
}

func TestNext_QuickWins_WithLimit(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextQuickWins = true
	nextLimit = 1

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recs) != 1 {
		t.Errorf("Expected 1 quick win with --limit 1, got %d", len(recs))
	}

	if len(recs) > 0 && recs[0].Effort != "small" {
		t.Errorf("Expected small effort, got %s", recs[0].Effort)
	}
}

func TestNext_QuickWins_NoQuickWinsAvailable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only medium/large effort tasks
	tasks := map[string]string{
		"001.md": `---
id: "001"
title: "Large task"
status: pending
priority: high
effort: large
dependencies: []
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "Medium task"
status: pending
priority: medium
effort: medium
dependencies: []
created: 2026-02-02
---`,
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	resetNextFlags()
	nextFormat = "table"
	nextQuickWins = true

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	if !strings.Contains(output, "No quick wins available") {
		t.Errorf("Expected 'No quick wins available' message, got: %s", output)
	}
}

func TestNext_QuickWins_TableFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "table"
	nextQuickWins = true
	nextLimit = 3

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	if !strings.Contains(output, "Recommended quick wins:") {
		t.Error("Expected table header 'Recommended quick wins:'")
	}
}

func TestNext_QuickWins_YAMLFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "yaml"
	nextQuickWins = true
	nextLimit = 2

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	// Verify it's valid YAML with effort field
	if !strings.Contains(output, "effort: small") {
		t.Error("Expected YAML output to contain 'effort: small'")
	}
}

func TestNext_Critical_HappyPath(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextCritical = true
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify all returned tasks are on critical path
	for _, rec := range recs {
		if !rec.OnCriticalPath {
			t.Errorf("Critical filter should only include critical path tasks, got task %s", rec.ID)
		}
	}

	// All tasks should have on_critical_path: true
	if len(recs) > 0 {
		allCritical := true
		for _, rec := range recs {
			if !rec.OnCriticalPath {
				allCritical = false
				break
			}
		}
		if !allCritical {
			t.Error("All recommendations should be on critical path")
		}
	}
}

func TestNext_Critical_WithFilter(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextCritical = true
	nextFilters = []string{"tag=cli"}
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify all tasks are CLI-tagged AND on critical path
	for _, rec := range recs {
		if !rec.OnCriticalPath {
			t.Errorf("Expected task %s to be on critical path", rec.ID)
		}
	}
}

func TestNext_Critical_WithLimit(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextCritical = true
	nextLimit = 1

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recs) > 1 {
		t.Errorf("Expected at most 1 recommendation with --limit 1, got %d", len(recs))
	}

	if len(recs) > 0 && !recs[0].OnCriticalPath {
		t.Error("Expected critical path task")
	}
}

func TestNext_Critical_NoCriticalTasksAvailable(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a scenario where actionable tasks are NOT on critical path
	// Critical path: 001 -> 003 (longer chain)
	// Non-critical: 002 (shorter parallel path)
	tasks := map[string]string{
		"001.md": `---
id: "001"
title: "Root task"
status: completed
priority: high
effort: large
dependencies: []
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "Short path task"
status: pending
priority: low
effort: small
dependencies: ["001"]
created: 2026-02-02
---`,
		"003.md": `---
id: "003"
title: "Long path intermediate"
status: completed
priority: high
effort: large
dependencies: ["001"]
created: 2026-02-03
---`,
		"004.md": `---
id: "004"
title: "Long path final"
status: completed
priority: high
effort: large
dependencies: ["003"]
created: 2026-02-04
---`,
	}

	for filename, content := range tasks {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	resetNextFlags()
	nextFormat = "table"
	nextCritical = true

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	// Task 002 is actionable but not on critical path (001->003->004 is longer)
	// So filtering by --critical should show no results
	if !strings.Contains(output, "No critical path tasks available") {
		t.Errorf("Expected 'No critical path tasks available' message, got: %s", output)
	}
}

func TestNext_Critical_TableFormat(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "table"
	nextCritical = true
	nextLimit = 3

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	if !strings.Contains(output, "Recommended critical path tasks:") {
		t.Error("Expected table header 'Recommended critical path tasks:'")
	}
}

func TestNext_QuickWins_Ranking(t *testing.T) {
	tmpDir := createNextTestTaskFiles(t)

	resetNextFlags()
	nextFormat = "json"
	nextQuickWins = true
	nextLimit = 10

	output, err := captureNextOutput(t, []string{tmpDir})
	if err != nil {
		t.Fatalf("runNext failed: %v", err)
	}

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify still ranked by score
	for i := 1; i < len(recs); i++ {
		if recs[i].Score > recs[i-1].Score {
			t.Errorf("Quick wins not sorted by score: [%d]=%d > [%d]=%d",
				i, recs[i].Score, i-1, recs[i-1].Score)
		}
	}

	// Task 003 (critical+small) should rank higher than 005 (low+small)
	recMap := make(map[string]Recommendation)
	for _, rec := range recs {
		recMap[rec.ID] = rec
	}

	if rec003, ok := recMap["003"]; ok {
		if rec005, ok := recMap["005"]; ok {
			if rec003.Rank > rec005.Rank {
				t.Errorf("Expected task 003 (critical) to rank higher than 005 (low): rank %d > %d",
					rec003.Rank, rec005.Rank)
			}
		}
	}
}
