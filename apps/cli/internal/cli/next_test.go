package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/next"
)

// The next command's fixtures are command-specific scoring/ordering shapes that
// do not map onto the canonical testdata sets (dependency-chain, phases, etc.),
// so they are seeded inline with newTaskRepo. Each helper below returns the
// task files as a map for the harness rather than hand-rolling os.WriteFile.

// nextScoringFiles returns 10 task files designed to exercise the next command's
// scoring, filtering, and actionability logic.
//
// Task graph:
//
//	001 (completed, high, small)         - root, completed
//	002 (completed, medium, medium)      - depends on 001, completed
//	003 (pending, critical, small, cli)  - depends on 001 (completed) - actionable
//	004 (pending, high, large, cli)      - depends on 002 (completed) - actionable
//	005 (pending, low, small)            - no deps - actionable
//	006 (pending, high, medium)          - depends on 007 (pending) → blocked
//	007 (pending, medium, small)         - no deps - actionable
//	008 (in-progress, high, small, cli)  - depends on 001 (completed) - actionable
//	009 (pending, medium, large)         - depends on 006 (pending) → blocked
//	010 (pending, low, medium)           - depends on 003 → blocked (003 pending)
func nextScoringFiles() map[string]string {
	return map[string]string{
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
}

// nextStdout runs `next <args...>` against repo, fails on error, and returns stdout.
func nextStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"next"}, args...)...)
	if res.Err != nil {
		t.Fatalf("next %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

// recIDs extracts the recommendation IDs from JSON next output.
func recIDs(t *testing.T, output string) []string {
	t.Helper()
	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}
	ids := make([]string, len(recs))
	for i, r := range recs {
		ids[i] = r.ID
	}
	return ids
}

// setPhaseOrder configures the phase ordering in viper (as .taskmd.yaml would).
func setPhaseOrder(phases []string) {
	items := make([]any, len(phases))
	for i, p := range phases {
		items[i] = map[string]any{"id": p}
	}
	viper.Set("phases", items)
}

// runNextWithPhases runs `next` with a phase order configured in viper — the
// same effect a repo-local .taskmd.yaml would have under the real command. The
// order is seeded through RunWith so it survives the harness's hermetic reset.
func runNextWithPhases(t *testing.T, repo *taskRepo, phases []string, args ...string) cliResult {
	t.Helper()
	return repo.RunWith(func() { setPhaseOrder(phases) }, append([]string{"next"}, args...)...)
}

func TestNext_BasicRanking(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "20")

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
	repo := newTaskRepo(t, map[string]string{
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
	})

	output := nextStdout(t, repo, "--format", "json", "--limit", "20")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "2")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recs) != 2 {
		t.Errorf("Expected 2 recommendations with --limit 2, got %d", len(recs))
	}
}

func TestNext_LimitExceedsAvailable(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "100")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--filter", "tag=cli")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--filter", "priority=high")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--filter", "tag=cli", "--filter", "priority=high")

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
	repo := newTaskRepo(t, nextScoringFiles())

	res := repo.Run("next", "--format", "json", "--filter", "invalid")
	if res.Err == nil {
		t.Fatal("Expected error for invalid filter format")
	}

	if !strings.Contains(res.Err.Error(), "invalid filter format") {
		t.Errorf("Expected 'invalid filter format' error, got: %v", res.Err)
	}
}

func TestNext_UnsupportedFormat(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	res := repo.Run("next", "--format", "csv")
	if res.Err == nil {
		t.Fatal("Expected error for unsupported format")
	}

	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestNext_JSONFormat(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "2")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "yaml", "--limit", "2")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "table", "--limit", "3")

	if !strings.Contains(output, "Recommended tasks:") {
		t.Error("Expected table header 'Recommended tasks:'")
	}
	if !strings.Contains(output, "#") && !strings.Contains(output, "ID") {
		t.Error("Expected table column headers")
	}
	if !strings.Contains(output, "Effort") {
		t.Error("Expected 'Effort' column header in table output")
	}
}

func TestNext_Explain_TableFormat(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "table", "--explain", "--limit", "3")

	if !strings.Contains(output, "Recommended tasks:") {
		t.Error("Expected label 'Recommended tasks:'")
	}
	// Breakdown blocks list itemized components and a total, and omit the
	// compact table column headers.
	for _, want := range []string{"priority:", "total"} {
		if !strings.Contains(output, want) {
			t.Errorf("Expected --explain output to contain %q\nOutput:\n%s", want, output)
		}
	}
	if strings.Contains(output, "Effort") {
		t.Errorf("Did not expect compact table header 'Effort' in --explain output\nOutput:\n%s", output)
	}
}

func TestNext_Explain_ScaledBonusShowsMultiplier(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "table", "--explain", "--limit", "10")

	// The fixture has critical-path / downstream tasks, so at least one scaled
	// bonus (rendered as "base × mult)") must appear.
	if !strings.Contains(output, "×") {
		t.Errorf("Expected a scaled-bonus multiplier ('×') in --explain output\nOutput:\n%s", output)
	}
}

func TestNext_Explain_JSONBreakdownSumsToScore(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}
	if len(recs) == 0 {
		t.Fatal("Expected at least one recommendation")
	}

	for _, rec := range recs {
		if len(rec.ScoreBreakdown) == 0 {
			t.Errorf("rec %s: score_breakdown is empty", rec.ID)
			continue
		}
		sum := 0
		for _, c := range rec.ScoreBreakdown {
			sum += c.Points
		}
		if sum != rec.Score {
			t.Errorf("rec %s: score_breakdown sum = %d, want score %d (%+v)", rec.ID, sum, rec.Score, rec.ScoreBreakdown)
		}
	}
}

func TestNext_TableWithoutExplain_HasNoBreakdown(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "table", "--limit", "3")

	// Default table shows the compact columns, not the itemized breakdown.
	if !strings.Contains(output, "Effort") {
		t.Errorf("Expected compact table header 'Effort'\nOutput:\n%s", output)
	}
	if strings.Contains(output, "total") || strings.Contains(output, "priority:") {
		t.Errorf("Did not expect breakdown lines in default table output\nOutput:\n%s", output)
	}
}

func TestNext_NoActionableTasks(t *testing.T) {
	// Only completed tasks
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Done task"
status: completed
priority: high
dependencies: []
created: 2026-02-01
---`,
	})

	output := nextStdout(t, repo, "--format", "table")

	if !strings.Contains(output, "No actionable tasks found") {
		t.Errorf("Expected 'No actionable tasks found' message, got: %s", output)
	}
}

func TestNext_InProgressTasksIncluded(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

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
	// Two identical-scoring tasks with different IDs
	repo := newTaskRepo(t, map[string]string{
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
	})

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

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
			got := next.IsActionable(tt.task, taskMap, next.BuildChildrenMap(nil))
			if got != tt.expected {
				t.Errorf("isActionable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestScoreTask(t *testing.T) {
	criticalPath := map[string]bool{"cp1": true}
	// Use high-priority downstream so bonuses are at full weight (multiplier = 1.0)
	downstreamInfo := map[string]next.DownstreamInfo{
		"cp1": {Count: 3, MaxPriority: model.PriorityHigh},
		"ds1": {Count: 1, MaxPriority: model.PriorityHigh},
		"ds6": {Count: 6, MaxPriority: model.PriorityHigh},
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
			score, reasons := next.ScoreTask(tt.task, criticalPath, downstreamInfo, effort.Default())
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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--filter", "tag=api")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--quick-wins", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--quick-wins", "--filter", "tag=cli", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--quick-wins", "--limit", "1")

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
	// Only medium/large effort tasks
	repo := newTaskRepo(t, map[string]string{
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
	})

	output := nextStdout(t, repo, "--format", "table", "--quick-wins")

	if !strings.Contains(output, "No quick wins available") {
		t.Errorf("Expected 'No quick wins available' message, got: %s", output)
	}
}

func TestNext_QuickWins_TableFormat(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "table", "--quick-wins", "--limit", "3")

	if !strings.Contains(output, "Recommended quick wins:") {
		t.Error("Expected table header 'Recommended quick wins:'")
	}
}

func TestNext_QuickWins_YAMLFormat(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "yaml", "--quick-wins", "--limit", "2")

	// Verify it's valid YAML with effort field
	if !strings.Contains(output, "effort: small") {
		t.Error("Expected YAML output to contain 'effort: small'")
	}
}

func TestNext_Critical_HappyPath(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--critical", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--critical", "--filter", "tag=cli", "--limit", "10")

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
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--critical", "--limit", "1")

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
	// A scenario where --critical + tag filter yields no results.
	// Critical path: 001 → 002 (both pending, tagged "api")
	// Non-critical: 003 (pending, no deps, shorter path, tagged "docs")
	//
	// Filtering by tag=docs + --critical should find no tasks because
	// 003 is the only docs-tagged task and it's not on the critical path.
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "API foundation"
status: pending
priority: high
effort: large
dependencies: []
tags: ["api"]
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "API endpoints"
status: pending
priority: high
effort: large
dependencies: ["001"]
tags: ["api"]
created: 2026-02-02
---`,
		"003.md": `---
id: "003"
title: "Write docs"
status: pending
priority: low
effort: small
dependencies: []
tags: ["docs"]
created: 2026-02-03
---`,
	})

	output := nextStdout(t, repo, "--format", "table", "--critical", "--filter", "tag=docs")

	// 003 is actionable and docs-tagged but NOT on critical path (shorter chain)
	// So --critical + tag=docs should show no results
	if !strings.Contains(output, "No critical path tasks available") {
		t.Errorf("Expected 'No critical path tasks available' message, got: %s", output)
	}
}

func TestNext_Critical_CompletedDepsIgnored(t *testing.T) {
	// Completed dependency chains should NOT inflate critical path depth.
	// 001 (completed) → 003 (completed) → 004 (completed): all done, irrelevant
	// 002 (pending, depends on completed 001): only remaining task
	//
	// 002 should BE the critical path since it's the only remaining work.
	repo := newTaskRepo(t, map[string]string{
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
title: "Remaining task"
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
	})

	output := nextStdout(t, repo, "--format", "json", "--critical", "--limit", "10")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// 002 is the only pending task — it IS the critical path
	if len(recs) != 1 {
		t.Fatalf("Expected 1 critical path task, got %d", len(recs))
	}
	if recs[0].ID != "002" {
		t.Errorf("Expected task 002 on critical path, got %s", recs[0].ID)
	}
	if !recs[0].OnCriticalPath {
		t.Error("Expected task 002 to be marked on_critical_path")
	}
}

func TestNext_Critical_TableFormat(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "table", "--critical", "--limit", "3")

	if !strings.Contains(output, "Recommended critical path tasks:") {
		t.Error("Expected table header 'Recommended critical path tasks:'")
	}
}

func TestNext_QuickWins_Ranking(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--quick-wins", "--limit", "10")

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

func TestNext_ArchivedDependencySatisfied(t *testing.T) {
	// Active task 002 depends on archived completed task 001.
	repo := newTaskRepo(t, map[string]string{
		"002.md": `---
id: "002"
title: "Feature that depends on archived"
status: pending
priority: high
effort: small
dependencies: ["001"]
created: 2026-02-01
---`,
		"archive/001.md": `---
id: "001"
title: "Completed archived task"
status: completed
priority: high
effort: medium
created: 2026-01-01
---`,
	})

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Task 002 should be actionable because its dependency (001) is archived and completed
	if len(recs) != 1 {
		t.Fatalf("Expected 1 recommendation (002 unblocked by archived dep), got %d", len(recs))
	}
	if recs[0].ID != "002" {
		t.Errorf("Expected task 002, got %s", recs[0].ID)
	}
}

// scopeFiles returns task files with touches fields for scope tests.
func scopeFiles() map[string]string {
	return map[string]string{
		"001.md": `---
id: "001"
title: "Web dashboard"
status: pending
priority: high
effort: small
touches: ["web", "api"]
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "CLI refactor"
status: pending
priority: medium
effort: small
touches: ["cli"]
created: 2026-02-02
---`,
		"003.md": `---
id: "003"
title: "Web styling"
status: pending
priority: low
effort: large
touches: ["web"]
created: 2026-02-03
---`,
		"004.md": `---
id: "004"
title: "No scope task"
status: pending
priority: high
effort: small
created: 2026-02-04
---`,
	}
}

func TestNext_Scope_FiltersCorrectly(t *testing.T) {
	repo := newTaskRepo(t, scopeFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--scope", "web")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Only tasks 001 and 003 have touches containing "web"
	if len(recs) != 2 {
		t.Errorf("Expected 2 tasks with scope 'web', got %d", len(recs))
		for _, r := range recs {
			t.Logf("  %s: %s", r.ID, r.Title)
		}
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}
	if !ids["001"] || !ids["003"] {
		t.Errorf("Expected tasks 001 and 003 for scope 'web', got %v", ids)
	}
}

func TestNext_Scope_NoMatches(t *testing.T) {
	repo := newTaskRepo(t, scopeFiles())

	output := nextStdout(t, repo, "--format", "table", "--scope", "nonexistent")

	if !strings.Contains(output, `No actionable tasks found for scope "nonexistent"`) {
		t.Errorf("Expected scope-specific no-results message, got: %s", output)
	}
}

func TestNext_Scope_CombinedWithFilter(t *testing.T) {
	repo := newTaskRepo(t, scopeFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--scope", "web", "--filter", "priority=high")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Only task 001 matches: scope=web AND priority=high
	if len(recs) != 1 {
		t.Fatalf("Expected 1 task with scope=web + priority=high, got %d", len(recs))
	}
	if recs[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", recs[0].ID)
	}
}

func TestNext_Scope_CombinedWithQuickWins(t *testing.T) {
	repo := newTaskRepo(t, scopeFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--scope", "web", "--quick-wins")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Only task 001 matches: scope=web AND effort=small
	// Task 003 has scope=web but effort=large
	if len(recs) != 1 {
		t.Fatalf("Expected 1 task with scope=web + quick-wins, got %d", len(recs))
	}
	if recs[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", recs[0].ID)
	}
}

func TestNext_Scope_TableFormat(t *testing.T) {
	repo := newTaskRepo(t, scopeFiles())

	output := nextStdout(t, repo, "--format", "table", "--scope", "web")

	if !strings.Contains(output, "Recommended tasks (scope: web):") {
		t.Errorf("Expected scope label in table output, got: %s", output)
	}
}

func TestNext_Scope_WithoutScopeUnchanged(t *testing.T) {
	repo := newTaskRepo(t, scopeFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Without --scope, all 4 tasks should be actionable
	if len(recs) != 4 {
		t.Errorf("Without scope, expected all 4 actionable tasks, got %d", len(recs))
	}
}

// scopeDepFiles returns tasks where a scoped task depends on an unscoped task.
func scopeDepFiles() map[string]string {
	return map[string]string{
		"001.md": `---
id: "001"
title: "Setup database"
status: pending
priority: high
effort: small
created: 2026-02-01
---`,
		"002.md": `---
id: "002"
title: "Web dashboard"
status: pending
priority: medium
effort: small
touches: ["web"]
dependencies: ["001"]
created: 2026-02-02
---`,
		"003.md": `---
id: "003"
title: "Unrelated CLI task"
status: pending
priority: low
effort: small
touches: ["cli"]
created: 2026-02-03
---`,
		"004.md": `---
id: "004"
title: "Web styling"
status: pending
priority: low
effort: small
touches: ["web"]
created: 2026-02-04
---`,
	}
}

func TestNext_Scope_ExpandsDependencies(t *testing.T) {
	repo := newTaskRepo(t, scopeDepFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--scope", "web")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}

	// Task 001 (no touches) should be included because it blocks task 002 (touches web).
	// Task 004 directly touches web and is actionable.
	// Task 003 (touches cli) is unrelated and should be excluded.
	if !ids["001"] {
		t.Errorf("Expected task 001 (dependency of web task) to be included, got %v", ids)
	}
	if !ids["004"] {
		t.Errorf("Expected task 004 (directly touches web) to be included, got %v", ids)
	}
	if ids["003"] {
		t.Errorf("Task 003 (cli scope) should not appear in web scope results")
	}
}

func TestNext_Scope_ExactSkipsExpansion(t *testing.T) {
	repo := newTaskRepo(t, scopeDepFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--scope", "web", "--exact")

	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}

	// With --scope-exact, only tasks that directly touch "web" should appear.
	// Task 001 doesn't touch web — excluded even though it blocks a web task.
	// Task 002 touches web but is blocked (dep 001 pending) — excluded.
	// Task 004 touches web and is actionable — included.
	if ids["001"] {
		t.Errorf("With --scope-exact, task 001 (no touches) should not appear")
	}
	if len(recs) != 1 || recs[0].ID != "004" {
		t.Errorf("Expected only task 004, got %v", ids)
	}
}

// phaseFiles returns tasks tagged with distinct phases for --phase filter tests.
func phaseFiles() map[string]string {
	return map[string]string{
		"001.md": `---
id: "001"
title: "V0.2 task"
status: pending
priority: medium
phase: v0.2
---`,
		"002.md": `---
id: "002"
title: "V0.3 task"
status: pending
priority: medium
phase: v0.3
---`,
		"003.md": `---
id: "003"
title: "No phase task"
status: pending
priority: medium
---`,
	}
}

func TestNext_PhaseFilter(t *testing.T) {
	repo := newTaskRepo(t, phaseFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--phase", "v0.2")

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if len(recs) != 1 {
		t.Fatalf("Expected 1 recommendation for phase v0.2, got %d", len(recs))
	}
	if recs[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", recs[0].ID)
	}
}

func TestNext_PhaseFilterNoMatch(t *testing.T) {
	repo := newTaskRepo(t, phaseFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--phase", "v9.9")

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	if len(recs) != 0 {
		t.Errorf("Expected 0 recommendations for non-existent phase, got %d", len(recs))
	}
}

func TestLoadPhaseOrder_UsesIDWhenPresent(t *testing.T) {
	viper.Set("phases", []any{
		map[string]any{"id": "core-cli", "name": "Core CLI"},
		map[string]any{"id": "web-ui", "name": "Web UI"},
	})
	defer viper.Set("phases", nil)

	got := loadPhaseOrder()
	want := []string{"core-cli", "web-ui"}
	if len(got) != len(want) {
		t.Fatalf("loadPhaseOrder() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("loadPhaseOrder()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLoadPhaseOrder_SkipsPhasesWithoutID(t *testing.T) {
	viper.Set("phases", []any{
		map[string]any{"name": "Phase One"},
		map[string]any{"name": "Phase Two"},
	})
	defer viper.Set("phases", nil)

	got := loadPhaseOrder()
	if len(got) != 0 {
		t.Fatalf("loadPhaseOrder() = %v, want empty (phases without id should be skipped)", got)
	}
}

func TestLoadPhaseOrder_MixedIDAndNoID(t *testing.T) {
	viper.Set("phases", []any{
		map[string]any{"id": "core-cli", "name": "Core CLI"},
		map[string]any{"name": "Legacy Phase"},
	})
	defer viper.Set("phases", nil)

	got := loadPhaseOrder()
	want := []string{"core-cli"}
	if len(got) != len(want) {
		t.Fatalf("loadPhaseOrder() = %v, want %v", got, want)
	}
	if got[0] != want[0] {
		t.Errorf("loadPhaseOrder()[0] = %q, want %q", got[0], want[0])
	}
}

func TestLoadPhaseOrder_NilPhases(t *testing.T) {
	viper.Set("phases", nil)

	got := loadPhaseOrder()
	if got != nil {
		t.Errorf("loadPhaseOrder() = %v, want nil", got)
	}
}

// strictPhaseFiles returns tasks with different phases and priorities to test
// --strict-phases behavior.
//
// Task layout:
//
//	001: phase=v0.2, priority=low      - actionable
//	002: phase=v0.3, priority=critical - actionable (would normally outrank 001)
//	003: phase=v0.2, priority=medium   - actionable
//	004: no phase,   priority=high     - actionable
func strictPhaseFiles() map[string]string {
	return map[string]string{
		"001.md": `---
id: "001"
title: "Low priority v0.2 task"
status: pending
priority: low
phase: v0.2
---`,
		"002.md": `---
id: "002"
title: "Critical v0.3 task"
status: pending
priority: critical
phase: v0.3
---`,
		"003.md": `---
id: "003"
title: "Medium v0.2 task"
status: pending
priority: medium
phase: v0.2
---`,
		"004.md": `---
id: "004"
title: "High no-phase task"
status: pending
priority: high
---`,
	}
}

func TestNext_StrictPhasesOff_DefaultBehavior(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	res := runNextWithPhases(t, repo, []string{"v0.2", "v0.3"}, "--format", "json", "--limit", "10")
	if res.Err != nil {
		t.Fatalf("runNext failed: %v", res.Err)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, res.Stdout)
	}

	if len(recs) == 0 {
		t.Fatal("Expected recommendations, got none")
	}

	// Without strict phases, the critical-priority v0.3 task (002) should rank first
	if recs[0].ID != "002" {
		t.Errorf("Without --strict-phases, expected critical task 002 first, got %s", recs[0].ID)
	}
}

func TestNext_StrictPhasesOn_EarlierPhaseFirst(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	res := runNextWithPhases(t, repo, []string{"v0.2", "v0.3"}, "--format", "json", "--limit", "10", "--strict-phases")
	if res.Err != nil {
		t.Fatalf("runNext failed: %v", res.Err)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, res.Stdout)
	}

	if len(recs) != 4 {
		t.Fatalf("Expected 4 recommendations, got %d", len(recs))
	}

	// v0.2 tasks (001, 003) must come before v0.3 task (002)
	// Within v0.2, 003 (medium) should rank above 001 (low)
	if recs[0].ID != "003" {
		t.Errorf("Expected v0.2 medium task 003 first, got %s", recs[0].ID)
	}
	if recs[1].ID != "001" {
		t.Errorf("Expected v0.2 low task 001 second, got %s", recs[1].ID)
	}
	if recs[2].ID != "002" {
		t.Errorf("Expected v0.3 critical task 002 third, got %s", recs[2].ID)
	}
}

func TestNext_StrictPhases_SamePhaseUsesScore(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	res := runNextWithPhases(t, repo, []string{"v0.2", "v0.3"}, "--format", "json", "--limit", "10", "--strict-phases")
	if res.Err != nil {
		t.Fatalf("runNext failed: %v", res.Err)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, res.Stdout)
	}

	// Find the two v0.2 tasks (001=low, 003=medium)
	var v02Tasks []next.Recommendation
	for _, r := range recs {
		if r.ID == "001" || r.ID == "003" {
			v02Tasks = append(v02Tasks, r)
		}
	}
	if len(v02Tasks) != 2 {
		t.Fatalf("Expected 2 v0.2 tasks, got %d", len(v02Tasks))
	}
	// Medium priority (003) should score higher than low (001)
	if v02Tasks[0].ID != "003" {
		t.Errorf("Within v0.2, expected medium-priority 003 before low-priority 001, got %s first", v02Tasks[0].ID)
	}
}

func TestNext_StrictPhases_NoPhaseSortedLast(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	res := runNextWithPhases(t, repo, []string{"v0.2", "v0.3"}, "--format", "json", "--limit", "10", "--strict-phases")
	if res.Err != nil {
		t.Fatalf("runNext failed: %v", res.Err)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, res.Stdout)
	}

	if len(recs) != 4 {
		t.Fatalf("Expected 4 recommendations, got %d", len(recs))
	}

	// Task 004 (no phase) should be last
	if recs[3].ID != "004" {
		t.Errorf("Expected no-phase task 004 last, got %s", recs[3].ID)
	}
}

func TestNext_StrictPhases_WithPhaseFilter(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	res := runNextWithPhases(t, repo, []string{"v0.2", "v0.3"}, "--format", "json", "--limit", "10", "--strict-phases", "--phase", "v0.2")
	if res.Err != nil {
		t.Fatalf("runNext failed: %v", res.Err)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, res.Stdout)
	}

	// --phase v0.2 filters to only v0.2 tasks; --strict-phases still applies ordering
	if len(recs) != 2 {
		t.Fatalf("Expected 2 recommendations for phase v0.2, got %d", len(recs))
	}
	// Both are v0.2, so normal scoring: 003 (medium) before 001 (low)
	if recs[0].ID != "003" {
		t.Errorf("Expected 003 first within filtered v0.2, got %s", recs[0].ID)
	}
}

// strictPriorityFiles returns tasks where a bonus-laden lower-priority task
// normally outranks a bonus-less higher-priority one, to test --strict-priority
// behavior.
//
// Task layout:
//
//	crit: priority=critical               - actionable, no bonuses (score 40)
//	med:  priority=medium, effort=small   - actionable, on critical path with a
//	                                         high-priority downstream (score > 40)
//	dep:  priority=high, depends on med    - blocked (boosts med's score only)
func strictPriorityFiles() map[string]string {
	return map[string]string{
		"crit.md": `---
id: "crit"
title: "Critical task"
status: pending
priority: critical
---`,
		"med.md": `---
id: "med"
title: "Medium quick win"
status: pending
priority: medium
effort: small
---`,
		"dep.md": `---
id: "dep"
title: "High downstream task"
status: pending
priority: high
dependencies: ["med"]
---`,
	}
}

func TestNext_StrictPriorityOff_ScoreDominates(t *testing.T) {
	repo := newTaskRepo(t, strictPriorityFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10")

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Without --strict-priority, the bonus-laden medium task outscores critical.
	if recs[0].ID != "med" {
		t.Errorf("Without --strict-priority, expected bonus-laden 'med' first, got %s", recs[0].ID)
	}
}

func TestNext_StrictPriorityOn_HigherPriorityFirst(t *testing.T) {
	repo := newTaskRepo(t, strictPriorityFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--strict-priority")

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Only crit and med are actionable (dep is blocked).
	if len(recs) != 2 {
		t.Fatalf("Expected 2 recommendations, got %d", len(recs))
	}
	if recs[0].ID != "crit" {
		t.Errorf("Expected critical task first under --strict-priority, got %s", recs[0].ID)
	}
	if recs[1].ID != "med" {
		t.Errorf("Expected medium task second under --strict-priority, got %s", recs[1].ID)
	}
}

func TestNext_StrictPriority_ScoreBreaksTieWithinTier(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"plain.md": "---\nid: \"plain\"\ntitle: \"Plain medium\"\nstatus: pending\npriority: medium\n---",
		"quick.md": "---\nid: \"quick\"\ntitle: \"Quick medium\"\nstatus: pending\npriority: medium\neffort: small\n---",
	})

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--strict-priority")

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}

	// Same priority tier → score breaks the tie: the small-effort task ranks first.
	if recs[0].ID != "quick" {
		t.Errorf("Within the medium tier, expected higher-scoring 'quick' first, got %s", recs[0].ID)
	}
}

func TestNext_StrictPhasesAndPriority_PhasePrimary(t *testing.T) {
	repo := newTaskRepo(t, strictPhaseFiles())

	res := runNextWithPhases(t, repo, []string{"v0.2", "v0.3"}, "--format", "json", "--limit", "10", "--strict-phases", "--strict-priority")
	if res.Err != nil {
		t.Fatalf("runNext failed: %v", res.Err)
	}

	var recs []next.Recommendation
	if err := json.Unmarshal([]byte(res.Stdout), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, res.Stdout)
	}

	if len(recs) != 4 {
		t.Fatalf("Expected 4 recommendations, got %d", len(recs))
	}

	// Phase primary: v0.2 tasks (001 low, 003 medium) before v0.3 critical (002),
	// then the no-phase task (004) last.
	// Priority secondary within v0.2: 003 (medium) before 001 (low).
	gotIDs := make([]string, len(recs))
	for i, r := range recs {
		gotIDs[i] = r.ID
	}
	expected := []string{"003", "001", "002", "004"}
	for i, id := range expected {
		if recs[i].ID != id {
			t.Errorf("rank %d: expected %s, got %s (full order: %v)", i, id, recs[i].ID, gotIDs)
		}
	}
}

func TestNext_Columns_DefaultMatchesLegacy(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--limit", "3")

	// Default columns should produce headers: #, ID, Title, Priority, Effort, File, Reason
	for _, col := range []string{"#", "ID", "Title", "Priority", "Effort", "File", "Reason"} {
		if !strings.Contains(output, col) {
			t.Errorf("Default output missing expected column header %q", col)
		}
	}
}

func TestNext_Columns_CustomSelection(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--columns", "id,title,reason", "--limit", "3")

	lower := strings.ToLower(output)
	// Selected columns should be present
	for _, col := range []string{"id", "title", "reason"} {
		if !strings.Contains(lower, col) {
			t.Errorf("Custom output missing expected column %q", col)
		}
	}

	// Non-selected columns should not appear as headers
	// Check that "priority" and "effort" don't appear as column headers
	// (they could appear in data, but we check the header line)
	lines := strings.Split(output, "\n")
	if len(lines) > 2 {
		headerLine := strings.ToLower(lines[2]) // after label and blank line
		if strings.Contains(headerLine, "priority") {
			t.Errorf("Custom output should not contain 'priority' column header")
		}
		if strings.Contains(headerLine, "effort") {
			t.Errorf("Custom output should not contain 'effort' column header")
		}
	}
}

func TestNext_Columns_ScoreColumn(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--columns", "rank,id,score", "--limit", "3")

	if !strings.Contains(strings.ToLower(output), "score") {
		t.Error("Output should contain 'score' column")
	}
}

func TestNext_Columns_InvalidColumn(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	res := repo.Run("next", "--columns", "id,title,bogus", "--limit", "3")
	if res.Err == nil {
		t.Fatal("Expected error for invalid column name, got nil")
	}
	if !strings.Contains(res.Err.Error(), "bogus") {
		t.Errorf("Error should mention invalid column name 'bogus', got: %v", res.Err)
	}
}

func TestNext_Columns_JSONUnaffected(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--columns", "id,title", "--limit", "3")

	// JSON output should contain all fields regardless of --columns
	var recs []Recommendation
	if err := json.Unmarshal([]byte(output), &recs); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(recs) == 0 {
		t.Fatal("Expected recommendations in JSON output")
	}

	// JSON should still have fields like Score and Priority
	if recs[0].Score == 0 && recs[0].Priority == "" {
		t.Error("JSON output should contain all fields regardless of --columns flag")
	}
}

func TestNext_Columns_CaseInsensitive(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--columns", "ID, Title, Reason", "--limit", "3")

	lower := strings.ToLower(output)
	for _, col := range []string{"id", "title", "reason"} {
		if !strings.Contains(lower, col) {
			t.Errorf("Case-insensitive columns missing %q in output", col)
		}
	}
}

func TestParseNextColumns(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "default columns",
			input: "rank,id,title,priority,effort,file,reason",
			want:  []string{"rank", "id", "title", "priority", "effort", "file", "reason"},
		},
		{
			name:  "case insensitive with whitespace",
			input: " ID , Title , Score ",
			want:  []string{"id", "title", "score"},
		},
		{
			name:    "invalid column",
			input:   "id,invalid",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:  "single column",
			input: "id",
			want:  []string{"id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNextColumns(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNextColumns(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("parseNextColumns(%q) = %v, want %v", tt.input, got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("parseNextColumns(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

// parentFiles returns a small task set with a parent/child hierarchy for
// exercising --root against a parent task.
//
//	P (pending, high)           - parent, blocked by incomplete children
//	C1 (pending, medium, →P)    - actionable subtask
//	C2 (pending, high, →P)      - actionable subtask
//	U (pending, high)           - unrelated, not reachable from P
func parentFiles() map[string]string {
	return map[string]string{
		"P.md": `---
id: "P"
title: "Parent task"
status: pending
priority: high
---`,
		"C1.md": `---
id: "C1"
title: "Child one"
status: pending
priority: medium
parent: "P"
---`,
		"C2.md": `---
id: "C2"
title: "Child two"
status: pending
priority: high
parent: "P"
---`,
		"U.md": `---
id: "U"
title: "Unrelated task"
status: pending
priority: high
---`,
	}
}

func TestNext_RootLeafReturnsUpstreamOnly(t *testing.T) {
	// 009 depends on 006 depends on 007. Only 007 is actionable in that chain.
	// --root 009 should return just 007, excluding other actionable tasks.
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--root", "009")

	ids := recIDs(t, output)
	if len(ids) != 1 || ids[0] != "007" {
		t.Fatalf("Expected only upstream task 007, got %v", ids)
	}
}

func TestNext_RootActionableRootItself(t *testing.T) {
	// 004 depends on 002 (completed), so 004 itself is actionable.
	// --root 004 should return 004.
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--root", "004")

	ids := recIDs(t, output)
	if len(ids) != 1 || ids[0] != "004" {
		t.Fatalf("Expected root task 004 itself, got %v", ids)
	}
}

func TestNext_RootParentReturnsSubtasks(t *testing.T) {
	repo := newTaskRepo(t, parentFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--root", "P")

	ids := recIDs(t, output)
	got := map[string]bool{}
	for _, id := range ids {
		got[id] = true
	}
	if len(ids) != 2 || !got["C1"] || !got["C2"] {
		t.Fatalf("Expected actionable subtasks C1 and C2, got %v", ids)
	}
}

func TestNext_RootUnknownIDErrors(t *testing.T) {
	repo := newTaskRepo(t, nextScoringFiles())

	res := repo.Run("next", "--format", "json", "--root", "999")
	if res.Err == nil {
		t.Fatal("Expected error for unknown root ID, got nil")
	}
	if !strings.Contains(res.Err.Error(), "root task 999 not found") {
		t.Errorf("Expected 'root task 999 not found', got %q", res.Err.Error())
	}
}

func TestNext_RootCombinedWithFilter(t *testing.T) {
	// --root 009 reaches actionable task 007 (tag api). A matching filter
	// keeps it; a non-matching filter narrows the result to empty.
	repo := newTaskRepo(t, nextScoringFiles())

	output := nextStdout(t, repo, "--format", "json", "--limit", "10", "--root", "009", "--filter", "tag=api")
	if ids := recIDs(t, output); len(ids) != 1 || ids[0] != "007" {
		t.Fatalf("Expected 007 with matching filter, got %v", ids)
	}

	output = nextStdout(t, repo, "--format", "json", "--limit", "10", "--root", "009", "--filter", "tag=cli")
	if ids := recIDs(t, output); len(ids) != 0 {
		t.Fatalf("Expected no results with non-matching filter, got %v", ids)
	}
}
