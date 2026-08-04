package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// setPhaseConfig seeds phase config into viper as a repo .taskmd.yaml would.
// Passing nil clears it. The harness runs commands' RunE directly and skips
// cobra's config discovery, so phase config must be injected (via RunWith) to
// survive the hermetic reset.
func setPhaseConfig(phases []map[string]any) {
	if phases == nil {
		viper.Set("phases", nil)
		return
	}
	// Convert to []any so viper.Get returns a type parsePhasesConfig can assert.
	items := make([]any, len(phases))
	for i, p := range phases {
		items[i] = p
	}
	viper.Set("phases", items)
}

// runPhasesWith runs `phases <args...>` against repo with phase config seeded
// through RunWith so it survives the harness reset.
func runPhasesWith(t *testing.T, repo *taskRepo, phases []map[string]any, args ...string) cliResult {
	t.Helper()
	return repo.RunWith(func() { setPhaseConfig(phases) }, append([]string{"phases"}, args...)...)
}

// phasesTestFiles is the canonical phases-command fixture: three mvp tasks (one
// each pending/completed/in-progress), one v2 task, and one un-phased task.
// Kept inline because this specific status/phase mix is the subject of these
// tests.
func phasesTestFiles() map[string]string {
	return map[string]string{
		"001.md": `---
id: "001"
title: "MVP task A"
status: pending
priority: high
phase: mvp
---`,
		"002.md": `---
id: "002"
title: "MVP task B"
status: completed
priority: medium
phase: mvp
---`,
		"003.md": `---
id: "003"
title: "MVP task C"
status: in-progress
priority: low
phase: mvp
---`,
		"004.md": `---
id: "004"
title: "V2 task"
status: pending
priority: medium
phase: v2
---`,
		"005.md": `---
id: "005"
title: "No phase task"
status: pending
priority: low
---`,
	}
}

func TestPhases_TableOutput(t *testing.T) {
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP", "due": "2026-06-01"},
		{"id": "v2", "name": "Version 2"},
	})
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}
	stdout := res.Stdout

	for _, expected := range []string{"ID", "Name", "Tasks", "Done", "Progress", "Due"} {
		if !strings.Contains(stdout, expected) {
			t.Errorf("table output missing header %q:\n%s", expected, stdout)
		}
	}
	if !strings.Contains(stdout, "mvp") {
		t.Errorf("table output missing mvp phase:\n%s", stdout)
	}
	if !strings.Contains(stdout, "MVP") {
		t.Errorf("table output missing MVP name:\n%s", stdout)
	}
	if !strings.Contains(stdout, "33%") {
		t.Errorf("table output missing 33%% progress for mvp (1/3 done):\n%s", stdout)
	}
	if !strings.Contains(stdout, "2026-06-01") {
		t.Errorf("table output missing due date:\n%s", stdout)
	}
}

func TestPhases_JSONOutput(t *testing.T) {
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP", "due": "2026-06-01"},
		{"id": "v2", "name": "Version 2"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}
	stdout := res.Stdout

	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, stdout)
	}

	// mvp, v2, plus the synthetic unassigned row for task 005 (no phase).
	if len(summaries) != 3 {
		t.Fatalf("expected 3 summaries (mvp, v2, unassigned), got %d", len(summaries))
	}

	mvp := summaries[0]
	if mvp.ID != "mvp" {
		t.Errorf("first phase ID = %q, want mvp", mvp.ID)
	}
	if mvp.Tasks != 3 {
		t.Errorf("mvp tasks = %d, want 3", mvp.Tasks)
	}
	if mvp.Done != 1 {
		t.Errorf("mvp done = %d, want 1", mvp.Done)
	}
	if mvp.Progress != "33%" {
		t.Errorf("mvp progress = %q, want 33%%", mvp.Progress)
	}
	if mvp.Due != "2026-06-01" {
		t.Errorf("mvp due = %q, want 2026-06-01", mvp.Due)
	}
	if mvp.ByStatus["pending"] != 1 {
		t.Errorf("mvp by_status[pending] = %d, want 1", mvp.ByStatus["pending"])
	}
	if mvp.ByStatus["completed"] != 1 {
		t.Errorf("mvp by_status[completed] = %d, want 1", mvp.ByStatus["completed"])
	}
	if mvp.ByStatus["in-progress"] != 1 {
		t.Errorf("mvp by_status[in-progress] = %d, want 1", mvp.ByStatus["in-progress"])
	}

	v2 := summaries[1]
	if v2.Tasks != 1 {
		t.Errorf("v2 tasks = %d, want 1", v2.Tasks)
	}
	if v2.Done != 0 {
		t.Errorf("v2 done = %d, want 0", v2.Done)
	}
	if v2.Progress != "0%" {
		t.Errorf("v2 progress = %q, want 0%%", v2.Progress)
	}

	unassigned := summaries[2]
	if unassigned.ID != unassignedPhaseID {
		t.Errorf("third summary ID = %q, want %q", unassigned.ID, unassignedPhaseID)
	}
	if unassigned.Tasks != 1 {
		t.Errorf("unassigned tasks = %d, want 1 (task 005)", unassigned.Tasks)
	}
}

func TestPhases_YAMLOutput(t *testing.T) {
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	}, "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}
	stdout := res.Stdout

	if !strings.Contains(stdout, "id: mvp") {
		t.Errorf("YAML output missing 'id: mvp':\n%s", stdout)
	}
	if !strings.Contains(stdout, "tasks: 3") {
		t.Errorf("YAML output missing 'tasks: 3':\n%s", stdout)
	}
}

func TestPhases_NoPhasesConfigured(t *testing.T) {
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, nil)
	if res.Err != nil {
		t.Fatalf("runPhases should not error when no phases configured: %v", res.Err)
	}

	if !strings.Contains(res.Stderr, "No phases configured") {
		t.Errorf("expected helpful message about no phases, got stderr:\n%s", res.Stderr)
	}
}

func TestPhases_OrphanedPhaseValues(t *testing.T) {
	repo := newTaskRepo(t, map[string]string{
		"001.md": `---
id: "001"
title: "Orphaned phase task"
status: pending
phase: unknown-phase
---`,
	})

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	})
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	if !strings.Contains(res.Stderr, "undefined phase") {
		t.Errorf("expected warning about undefined phase, got stderr:\n%s", res.Stderr)
	}
	if !strings.Contains(res.Stderr, "unknown-phase") {
		t.Errorf("expected orphaned phase name in warning, got stderr:\n%s", res.Stderr)
	}
}

func TestPhases_OrphanedPhaseWarningsSortedDeterministically(t *testing.T) {
	// Tasks referencing multiple undefined phases.
	files := map[string]string{}
	for i, phase := range []string{"zebra", "alpha", "middle"} {
		id := "00" + string(rune('1'+i))
		files[id+".md"] = "---\nid: \"" + id + "\"\ntitle: \"Task\"\nstatus: pending\nphase: " + phase + "\n---"
	}
	repo := newTaskRepo(t, files)

	phases := []map[string]any{{"id": "mvp", "name": "MVP"}}

	// Run multiple times and verify consistent ordering.
	var firstStderr string
	for i := 0; i < 5; i++ {
		res := runPhasesWith(t, repo, phases)
		if res.Err != nil {
			t.Fatalf("runPhases failed: %v", res.Err)
		}
		stderr := res.Stderr
		if i == 0 {
			firstStderr = stderr
			// Verify alphabetical order: alpha, middle, zebra.
			alphaIdx := strings.Index(stderr, "alpha")
			middleIdx := strings.Index(stderr, "middle")
			zebraIdx := strings.Index(stderr, "zebra")
			if alphaIdx == -1 || middleIdx == -1 || zebraIdx == -1 {
				t.Fatalf("expected all three undefined phases in warnings, got:\n%s", stderr)
			}
			if !(alphaIdx < middleIdx && middleIdx < zebraIdx) {
				t.Errorf("expected warnings in alphabetical order (alpha < middle < zebra), got:\n%s", stderr)
			}
		} else if stderr != firstStderr {
			t.Errorf("warning order changed between runs.\nFirst:\n%s\nRun %d:\n%s", firstStderr, i+1, stderr)
		}
	}
}

func TestPhases_CancelledTasksExcludedFromProgress(t *testing.T) {
	// 10 tasks total: 3 completed, 2 cancelled, 5 pending
	// Expected: 8 active tasks (10 - 2 cancelled), 3 done → 38% progress
	tasks := []struct {
		id     string
		status string
	}{
		{"001", "completed"},
		{"002", "completed"},
		{"003", "completed"},
		{"004", "cancelled"},
		{"005", "cancelled"},
		{"006", "pending"},
		{"007", "pending"},
		{"008", "pending"},
		{"009", "pending"},
		{"010", "pending"},
	}
	files := map[string]string{}
	for _, task := range tasks {
		files[task.id+".md"] = "---\nid: \"" + task.id + "\"\ntitle: \"Task " + task.id + "\"\nstatus: " + task.status + "\nphase: mvp\n---"
	}
	repo := newTaskRepo(t, files)

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if len(summaries) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(summaries))
	}

	mvp := summaries[0]

	// Cancelled tasks excluded from Tasks count
	if mvp.Tasks != 8 {
		t.Errorf("tasks = %d, want 8 (10 total minus 2 cancelled)", mvp.Tasks)
	}
	if mvp.Done != 3 {
		t.Errorf("done = %d, want 3", mvp.Done)
	}
	if mvp.Progress != "38%" {
		t.Errorf("progress = %q, want 38%% (3/8)", mvp.Progress)
	}

	// Cancelled tasks still visible in ByStatus
	if mvp.ByStatus["cancelled"] != 2 {
		t.Errorf("by_status[cancelled] = %d, want 2", mvp.ByStatus["cancelled"])
	}
	if mvp.ByStatus["completed"] != 3 {
		t.Errorf("by_status[completed] = %d, want 3", mvp.ByStatus["completed"])
	}
	if mvp.ByStatus["pending"] != 5 {
		t.Errorf("by_status[pending] = %d, want 5", mvp.ByStatus["pending"])
	}
}

func TestPhases_InvalidFormat(t *testing.T) {
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	}, "--format", "invalid")
	if res.Err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	if !strings.Contains(res.Err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", res.Err)
	}
}

func TestPhases_EmptyPhaseHasZeroProgress(t *testing.T) {
	repo := newTaskRepo(t, nil)

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "future", "name": "Future Work"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if len(summaries) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(summaries))
	}
	if summaries[0].Tasks != 0 {
		t.Errorf("expected 0 tasks, got %d", summaries[0].Tasks)
	}
	if summaries[0].Progress != "0%" {
		t.Errorf("expected 0%% progress, got %q", summaries[0].Progress)
	}
}

func TestPhases_MissingIDWarnsAndExcludes(t *testing.T) {
	// Create tasks: 1 completed + 2 pending.
	tasks := []struct {
		id     string
		status string
	}{
		{"001", "completed"},
		{"002", "pending"},
		{"003", "pending"},
	}
	files := map[string]string{}
	for _, task := range tasks {
		files[task.id+".md"] = "---\nid: \"" + task.id + "\"\ntitle: \"Task " + task.id + "\"\nstatus: " + task.status + "\n---"
	}
	repo := newTaskRepo(t, files)

	// Configure phases WITHOUT id fields — only names.
	res := runPhasesWith(t, repo, []map[string]any{
		{"name": "Core CLI"},
		{"name": "Web Dashboard"},
		{"name": "Documentation"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}
	stderr := res.Stderr

	// Phases without id should produce warnings.
	for _, name := range []string{"Core CLI", "Web Dashboard", "Documentation"} {
		if !strings.Contains(stderr, name) {
			t.Errorf("expected warning mentioning phase %q, got stderr:\n%s", name, stderr)
		}
	}
	if !strings.Contains(stderr, "missing \"id\"") {
		t.Errorf("expected warning about missing id, got stderr:\n%s", stderr)
	}

	// Since all phases lack an id, none should be valid → "No phases configured".
	if !strings.Contains(stderr, "No phases configured") {
		t.Errorf("expected 'No phases configured' when all phases lack id, got stderr:\n%s", stderr)
	}
}

func findSummary(summaries []PhaseSummary, id string) (PhaseSummary, bool) {
	for _, s := range summaries {
		if s.ID == id {
			return s, true
		}
	}
	return PhaseSummary{}, false
}

func TestPhases_UnassignedRowShownInTable(t *testing.T) {
	// phasesTestFiles includes task 005 with no phase.
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
		{"id": "v2", "name": "Version 2"},
	})
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}
	stdout := res.Stdout

	if !strings.Contains(stdout, "unassigned") {
		t.Errorf("table output missing unassigned row:\n%s", stdout)
	}
	if !strings.Contains(stdout, "(unassigned)") {
		t.Errorf("table output missing (unassigned) display name:\n%s", stdout)
	}
}

func TestPhases_UnassignedRowOmittedWhenAllAssigned(t *testing.T) {
	// Every task is assigned to a phase.
	files := map[string]string{}
	for _, task := range []struct{ id, status string }{
		{"001", "completed"},
		{"002", "pending"},
	} {
		files[task.id+".md"] = "---\nid: \"" + task.id + "\"\ntitle: \"Task\"\nstatus: " + task.status + "\nphase: mvp\n---"
	}
	repo := newTaskRepo(t, files)

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if _, ok := findSummary(summaries, unassignedPhaseID); ok {
		t.Errorf("expected no unassigned row when all tasks assigned, got:\n%s", res.Stdout)
	}
}

func TestPhases_UnassignedProgressAndCounts(t *testing.T) {
	// Unassigned (no phase): 4 active tasks (1 completed, 3 pending), 1 cancelled.
	// Expected: Tasks=4, Done=1, Progress=25%, cancelled excluded from count.
	tasks := []struct {
		id, status, phase string
	}{
		{"001", "completed", "mvp"}, // assigned, keeps mvp non-empty
		{"010", "completed", ""},
		{"011", "pending", ""},
		{"012", "pending", ""},
		{"013", "pending", ""},
		{"014", "cancelled", ""},
	}
	files := map[string]string{}
	for _, task := range tasks {
		content := "---\nid: \"" + task.id + "\"\ntitle: \"Task\"\nstatus: " + task.status + "\n"
		if task.phase != "" {
			content += "phase: " + task.phase + "\n"
		}
		content += "---"
		files[task.id+".md"] = content
	}
	repo := newTaskRepo(t, files)

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	unassigned, ok := findSummary(summaries, unassignedPhaseID)
	if !ok {
		t.Fatalf("expected unassigned summary, got:\n%s", res.Stdout)
	}
	if unassigned.Name != "(unassigned)" {
		t.Errorf("unassigned name = %q, want (unassigned)", unassigned.Name)
	}
	if unassigned.Tasks != 4 {
		t.Errorf("unassigned tasks = %d, want 4 (cancelled excluded)", unassigned.Tasks)
	}
	if unassigned.Done != 1 {
		t.Errorf("unassigned done = %d, want 1", unassigned.Done)
	}
	if unassigned.Progress != "25%" {
		t.Errorf("unassigned progress = %q, want 25%%", unassigned.Progress)
	}
	if unassigned.ByStatus["cancelled"] != 1 {
		t.Errorf("unassigned by_status[cancelled] = %d, want 1", unassigned.ByStatus["cancelled"])
	}

	// The unassigned row must come after the configured phases.
	if summaries[len(summaries)-1].ID != unassignedPhaseID {
		t.Errorf("unassigned row should be last, got id %q", summaries[len(summaries)-1].ID)
	}
}

func TestPhases_UnassignedOmittedWhenOnlyCancelled(t *testing.T) {
	// One assigned task keeps the phase list non-empty; unassigned tasks are all cancelled.
	files := map[string]string{}
	for _, task := range []struct{ id, status, phase string }{
		{"001", "pending", "mvp"},
		{"010", "cancelled", ""},
		{"011", "cancelled", ""},
	} {
		content := "---\nid: \"" + task.id + "\"\ntitle: \"Task\"\nstatus: " + task.status + "\n"
		if task.phase != "" {
			content += "phase: " + task.phase + "\n"
		}
		content += "---"
		files[task.id+".md"] = content
	}
	repo := newTaskRepo(t, files)

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if _, ok := findSummary(summaries, unassignedPhaseID); ok {
		t.Errorf("expected no unassigned row when only cancelled unassigned tasks exist, got:\n%s", res.Stdout)
	}
}

func TestPhases_UnassignedInYAMLOutput(t *testing.T) {
	repo := newTaskRepo(t, phasesTestFiles())

	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
		{"id": "v2", "name": "Version 2"},
	}, "--format", "yaml")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	if !strings.Contains(res.Stdout, "id: unassigned") {
		t.Errorf("YAML output missing 'id: unassigned':\n%s", res.Stdout)
	}
}

func TestPhases_MixedMissingAndValidIDs(t *testing.T) {
	// Create tasks assigned to the valid phases.
	files := map[string]string{}
	for _, task := range []struct {
		id, status, phase string
	}{
		{"001", "completed", "mvp"},
		{"002", "pending", "mvp"},
		{"003", "pending", "v2"},
	} {
		files[task.id+".md"] = "---\nid: \"" + task.id + "\"\ntitle: \"Task\"\nstatus: " + task.status + "\nphase: " + task.phase + "\n---"
	}
	repo := newTaskRepo(t, files)

	// One valid phase with id, two without.
	res := runPhasesWith(t, repo, []map[string]any{
		{"id": "mvp", "name": "MVP"},
		{"name": "No ID Phase"},
		{"id": "v2", "name": "Version 2"},
	}, "--format", "json")
	if res.Err != nil {
		t.Fatalf("runPhases failed: %v", res.Err)
	}

	// Warning about the phase without id.
	if !strings.Contains(res.Stderr, "No ID Phase") {
		t.Errorf("expected warning about phase without id, got stderr:\n%s", res.Stderr)
	}

	// Only valid phases should appear in output.
	var summaries []PhaseSummary
	if err := json.Unmarshal([]byte(res.Stdout), &summaries); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, res.Stdout)
	}

	if len(summaries) != 2 {
		t.Fatalf("expected 2 phases (only those with id), got %d", len(summaries))
	}
	if summaries[0].ID != "mvp" {
		t.Errorf("first phase ID = %q, want mvp", summaries[0].ID)
	}
	if summaries[0].Progress != "50%" {
		t.Errorf("mvp progress = %q, want 50%%", summaries[0].Progress)
	}
	if summaries[1].ID != "v2" {
		t.Errorf("second phase ID = %q, want v2", summaries[1].ID)
	}
}
