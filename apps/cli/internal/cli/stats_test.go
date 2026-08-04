package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/driangle/taskmd/sdk/go/metrics"
	"github.com/driangle/taskmd/sdk/go/model"
)

// statsFiles is the stats-specific fixture: three tasks with varying
// status/priority/effort. Kept inline because the exact status/priority/effort
// mix (2 pending, 1 completed) is the subject of these breakdown assertions.
func statsFiles() map[string]string {
	return map[string]string{
		"001-pending-high.md": `---
id: "001"
title: "High Priority Pending"
status: pending
priority: high
effort: small
dependencies: []
tags: ["api"]
created: 2026-02-08
---

Pending high-priority task.
`,
		"002-completed-medium.md": `---
id: "002"
title: "Completed Medium"
status: completed
priority: medium
effort: medium
dependencies: []
tags: ["api"]
created: 2026-02-08
---

Completed medium-priority task.
`,
		"003-pending-low.md": `---
id: "003"
title: "Pending Low with Dep"
status: pending
priority: low
effort: large
dependencies: ["002"]
tags: ["frontend"]
created: 2026-02-08
---

Pending task with dependency.
`,
	}
}

// statsPhaseFiles is the phase-grouping fixture: tasks spread across v0.2/v0.3
// plus one un-phased task. Kept inline because the per-phase counts are what
// the group-by-phase tests assert on.
func statsPhaseFiles() map[string]string {
	return map[string]string{
		"001.md": `---
id: "001"
title: "V0.2 task A"
status: pending
priority: high
phase: v0.2
---`,
		"002.md": `---
id: "002"
title: "V0.2 task B"
status: completed
priority: medium
phase: v0.2
---`,
		"003.md": `---
id: "003"
title: "V0.3 task"
status: pending
priority: low
phase: v0.3
---`,
		"004.md": `---
id: "004"
title: "No phase"
status: pending
priority: medium
---`,
	}
}

// statsStdout runs `stats <args...>` against repo, fails on error, returns stdout.
func statsStdout(t *testing.T, repo *taskRepo, args ...string) string {
	t.Helper()
	res := repo.Run(append([]string{"stats"}, args...)...)
	if res.Err != nil {
		t.Fatalf("stats %v failed: %v", args, res.Err)
	}
	return res.Stdout
}

func TestRunStats_JSONOutput(t *testing.T) {
	repo := newTaskRepo(t, statsFiles())

	output := statsStdout(t, repo, "--format", "json")

	var m metrics.Metrics
	if err := json.Unmarshal([]byte(output), &m); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	if m.TotalTasks != 3 {
		t.Errorf("total_tasks = %d, want 3", m.TotalTasks)
	}
	if m.TasksByStatus[model.StatusPending] != 2 {
		t.Errorf("tasks_by_status[pending] = %d, want 2", m.TasksByStatus[model.StatusPending])
	}
	if m.TasksByStatus[model.StatusCompleted] != 1 {
		t.Errorf("tasks_by_status[completed] = %d, want 1", m.TasksByStatus[model.StatusCompleted])
	}
}

func TestRunStats_YAMLOutput(t *testing.T) {
	repo := newTaskRepo(t, statsFiles())

	output := statsStdout(t, repo, "--format", "yaml")

	if !strings.Contains(output, "totaltasks:") {
		t.Errorf("YAML output missing 'totaltasks:':\n%s", output)
	}
}

func TestRunStats_TableOutput(t *testing.T) {
	repo := newTaskRepo(t, statsFiles())

	output := statsStdout(t, repo, "--format", "table")

	for _, expected := range []string{"TASK STATISTICS", "BY STATUS:", "BY PRIORITY:", "BY EFFORT:"} {
		if !strings.Contains(output, expected) {
			t.Errorf("table output missing %q:\n%s", expected, output)
		}
	}
}

func TestRunStats_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, statsFiles())

	res := repo.Run("stats", "--format", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestRunStats_EmptyDir(t *testing.T) {
	repo := newTaskRepo(t, nil)

	output := statsStdout(t, repo, "--format", "json")

	var m metrics.Metrics
	if err := json.Unmarshal([]byte(output), &m); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	if m.TotalTasks != 0 {
		t.Errorf("total_tasks = %d, want 0", m.TotalTasks)
	}
}

func TestOutputStatsTable_EmptyBreakdowns(t *testing.T) {
	m := &metrics.Metrics{
		TotalTasks:      0,
		TasksByStatus:   map[model.Status]int{},
		TasksByPriority: map[model.Priority]int{},
		TasksByEffort:   map[model.Effort]int{},
		TasksByType:     map[model.TaskType]int{},
	}

	var err error
	output, _ := captureOutput(t, func() {
		err = outputStatsTable(m, "")
	})

	if err != nil {
		t.Fatalf("outputStatsTable failed: %v", err)
	}

	// Each empty breakdown should print "(none)"
	count := strings.Count(output, "(none)")
	if count < 3 {
		t.Errorf("expected at least 3 '(none)' strings (status, priority, effort), got %d\noutput:\n%s", count, output)
	}
}

func TestRunStats_GroupByPhase_Table(t *testing.T) {
	repo := newTaskRepo(t, statsPhaseFiles())

	output := statsStdout(t, repo, "--group-by", "phase")

	if !strings.Contains(output, "BY PHASE:") {
		t.Errorf("expected BY PHASE section in output:\n%s", output)
	}
	if !strings.Contains(output, "v0.2:") {
		t.Errorf("expected v0.2 phase in output:\n%s", output)
	}
	if !strings.Contains(output, "v0.3:") {
		t.Errorf("expected v0.3 phase in output:\n%s", output)
	}
}

func TestRunStats_GroupByPhase_JSON(t *testing.T) {
	repo := newTaskRepo(t, statsPhaseFiles())

	output := statsStdout(t, repo, "--format", "json", "--group-by", "phase")

	var m metrics.Metrics
	if err := json.Unmarshal([]byte(output), &m); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, output)
	}

	if m.TasksByPhase["v0.2"] != 2 {
		t.Errorf("tasks_by_phase[v0.2] = %d, want 2", m.TasksByPhase["v0.2"])
	}
	if m.TasksByPhase["v0.3"] != 1 {
		t.Errorf("tasks_by_phase[v0.3] = %d, want 1", m.TasksByPhase["v0.3"])
	}
}

func TestRunStats_InvalidGroupBy(t *testing.T) {
	repo := newTaskRepo(t, statsPhaseFiles())

	res := repo.Run("stats", "--group-by", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid group-by, got nil")
	}
	if !strings.Contains(res.Err.Error(), "unsupported group-by field") {
		t.Errorf("expected 'unsupported group-by field' error, got: %v", res.Err)
	}
}

func TestRunStats_PhaseShownWhenPresent(t *testing.T) {
	repo := newTaskRepo(t, statsPhaseFiles())

	// No --group-by flag, but tasks have phases — section should still appear
	output := statsStdout(t, repo)

	if !strings.Contains(output, "BY PHASE:") {
		t.Errorf("expected BY PHASE section when phases exist:\n%s", output)
	}
}
