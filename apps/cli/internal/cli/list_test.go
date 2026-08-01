package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/driangle/taskmd/sdk/go/model"
)

func TestSortTasks(t *testing.T) {
	now := time.Now()
	tasks := []*model.Task{
		{ID: "003", Title: "C", Status: model.StatusPending, Priority: model.PriorityLow, Effort: model.EffortLarge, Created: model.NewFlexibleTime(now.Add(2 * time.Hour))},
		{ID: "001", Title: "A", Status: model.StatusCompleted, Priority: model.PriorityHigh, Effort: model.EffortSmall, Created: model.NewFlexibleTime(now)},
		{ID: "002", Title: "B", Status: model.StatusInProgress, Priority: model.PriorityCritical, Effort: model.EffortMedium, Created: model.NewFlexibleTime(now.Add(1 * time.Hour))},
	}

	tests := []struct {
		name      string
		sortField string
		firstID   string
		wantErr   bool
	}{
		{"sort by id", "id", "001", false},
		{"sort by title", "title", "001", false},
		{"sort by priority", "priority", "002", false},
		{"sort by effort", "effort", "001", false},
		{"sort by created", "created", "001", false},
		{"invalid sort field", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksCopy := make([]*model.Task, len(tasks))
			copy(tasksCopy, tasks)

			err := sortTasks(tasksCopy, tt.sortField)
			if (err != nil) != tt.wantErr {
				t.Errorf("sortTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tasksCopy[0].ID != tt.firstID {
				t.Errorf("sortTasks() first task ID = %s, want %s", tasksCopy[0].ID, tt.firstID)
			}
		})
	}
}

func TestGetColumnValue(t *testing.T) {
	created := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	task := &model.Task{
		ID:           "001",
		Title:        "Test Task",
		Status:       model.StatusPending,
		Priority:     model.PriorityHigh,
		Effort:       model.EffortSmall,
		Group:        "testing",
		Owner:        "alice",
		Created:      model.NewFlexibleTime(created),
		Dependencies: []string{"002", "003"},
		Tags:         []string{"cli", "test"},
	}

	tests := []struct {
		name     string
		column   string
		expected string
	}{
		{"id column", "id", "001"},
		{"title column", "title", "Test Task"},
		{"status column", "status", "pending"},
		{"priority column", "priority", "high"},
		{"effort column", "effort", "small"},
		{"group column", "group", "testing"},
		{"owner column", "owner", "alice"},
		{"created column", "created", "2026-02-08"},
		{"deps column", "deps", "002,003"},
		{"tags column", "tags", "cli,test"},
		{"unknown column", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getColumnValue(task, tt.column)
			if result != tt.expected {
				t.Errorf("getColumnValue(%s) = %s, want %s", tt.column, result, tt.expected)
			}
		})
	}
}

// resetListFlags resets list command flags to defaults before each test.
func resetListFlags() {
	listFilters = []string{}
	listSort = ""
	listColumns = "id,title,status,priority,file"
	listLimit = 0
	listScope = ""
	listPhase = ""
	listStatus = ""
	listPriority = ""
	noColor = true
}

// captureListTableOutput runs outputTable and captures stdout.
func captureListTableOutput(t *testing.T, tasks []*model.Task, columns string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTable(tasks, columns)
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("outputTable failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestListCommand_TableColorEnabled(t *testing.T) {
	resetListFlags()
	noColor = false
	forceColor = true
	defer func() { forceColor = false }()
	os.Unsetenv("NO_COLOR")

	tasks := []*model.Task{
		{ID: "001", Title: "Test Task", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "test.md"},
	}

	output := captureListTableOutput(t, tasks, "id,title,status,priority")

	// With colors enabled, output should contain ANSI escape codes
	if !strings.Contains(output, "\x1b[") {
		t.Error("Expected colored table output to contain ANSI escape codes")
	}

	// Task data should still be present
	if !strings.Contains(output, "001") {
		t.Error("Expected task ID in output")
	}
	if !strings.Contains(output, "pending") {
		t.Error("Expected status in output")
	}
}

func TestListCommand_TableNoColorFlag(t *testing.T) {
	resetListFlags()
	// noColor is already true from resetListFlags
	os.Unsetenv("NO_COLOR")

	tasks := []*model.Task{
		{ID: "001", Title: "Test Task", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "test.md"},
	}

	output := captureListTableOutput(t, tasks, "id,title,status,priority")

	// With no-color, output should NOT contain ANSI escape codes
	if strings.Contains(output, "\x1b[") {
		t.Error("Expected no ANSI codes in no-color table output")
	}

	// Task data should still be present
	if !strings.Contains(output, "001") {
		t.Error("Expected task ID in output")
	}
	if !strings.Contains(output, "pending") {
		t.Error("Expected status in output")
	}
}

func TestListCommand_TableNoColorEnvVar(t *testing.T) {
	resetListFlags()
	noColor = false // enable via flag, but env var should override
	forceColor = true
	defer func() { forceColor = false }()

	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	tasks := []*model.Task{
		{ID: "001", Title: "Test Task", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "test.md"},
	}

	output := captureListTableOutput(t, tasks, "id,title,status,priority")

	// NO_COLOR env var should disable colors
	if strings.Contains(output, "\x1b[") {
		t.Error("Expected no ANSI codes when NO_COLOR env var is set")
	}
}

func TestListCommand_TableColorColumns(t *testing.T) {
	resetListFlags()
	noColor = false
	forceColor = true
	defer func() { forceColor = false }()
	os.Unsetenv("NO_COLOR")

	tasks := []*model.Task{
		{ID: "001", Title: "Test Task", Status: model.StatusCompleted, Priority: model.PriorityCritical, FilePath: "test.md"},
		{ID: "002", Title: "Another", Status: model.StatusInProgress, Priority: model.PriorityLow, FilePath: "test2.md"},
	}

	output := captureListTableOutput(t, tasks, "id,title,status,priority")

	// Verify colored output contains task data
	if !strings.Contains(output, "001") {
		t.Error("Expected task 001 in output")
	}
	if !strings.Contains(output, "002") {
		t.Error("Expected task 002 in output")
	}
	if !strings.Contains(output, "Test Task") {
		t.Error("Expected title in output")
	}
}

func TestGetColumnValue_Parent(t *testing.T) {
	task := &model.Task{
		ID:     "001",
		Title:  "Test Task",
		Parent: "010",
	}

	result := getColumnValue(task, "parent")
	if result != "010" {
		t.Errorf("getColumnValue(parent) = %s, want 010", result)
	}

	// Empty parent
	task2 := &model.Task{ID: "002", Title: "No Parent"}
	result2 := getColumnValue(task2, "parent")
	if result2 != "" {
		t.Errorf("getColumnValue(parent) for task without parent = %q, want empty", result2)
	}
}

func TestListCommand_ParentColumn(t *testing.T) {
	resetListFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Parent", Status: model.StatusPending},
		{ID: "002", Title: "Child", Status: model.StatusPending, Parent: "001"},
	}

	output := captureListTableOutput(t, tasks, "id,title,parent")

	if !strings.Contains(output, "parent") {
		t.Error("Expected 'parent' column header in output")
	}
	if !strings.Contains(output, "001") {
		t.Error("Expected parent value '001' in output")
	}
}

func TestListCommand_TypeColumn(t *testing.T) {
	resetListFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Feature", Status: model.StatusPending, Type: model.TypeFeature},
		{ID: "002", Title: "Bug", Status: model.StatusPending, Type: model.TypeBug},
		{ID: "003", Title: "No type", Status: model.StatusPending},
	}

	output := captureListTableOutput(t, tasks, "id,title,type")

	if !strings.Contains(output, "type") {
		t.Error("Expected 'type' column header in output")
	}
	if !strings.Contains(output, "feature") {
		t.Error("Expected 'feature' in output")
	}
	if !strings.Contains(output, "bug") {
		t.Error("Expected 'bug' in output")
	}
}

func TestListCommand_SeparatorAlignment(t *testing.T) {
	resetListFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "A very long task title here", Status: model.StatusInProgress, Priority: model.PriorityHigh, FilePath: "tasks/cli/001.md"},
		{ID: "002", Title: "Short", Status: model.StatusPending, Priority: model.PriorityCritical, FilePath: "tasks/002.md"},
	}

	output := captureListTableOutput(t, tasks, "id,title,status,priority,file")
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (header, separator, data), got %d", len(lines))
	}

	// The separator is the second line. With tabwriter, columns are tab-aligned,
	// so we check that each separator segment (dashes) is at least as wide as
	// its corresponding header.
	headerLine := lines[0]
	separatorLine := lines[1]

	// tabwriter replaces tabs with spaces, so split on 2+ spaces
	headerCols := splitTableColumns(headerLine)
	sepCols := splitTableColumns(separatorLine)

	if len(headerCols) != len(sepCols) {
		t.Fatalf("Header has %d columns, separator has %d columns", len(headerCols), len(sepCols))
	}

	for i, header := range headerCols {
		sep := sepCols[i]
		if len(sep) < len(header) {
			t.Errorf("Separator column %d (%q, len=%d) is shorter than header (%q, len=%d)",
				i, sep, len(sep), header, len(header))
		}
		// Separator should be all dashes
		if strings.Trim(sep, "-") != "" {
			t.Errorf("Separator column %d should be all dashes, got %q", i, sep)
		}
	}

	// Also verify that "title" separator matches the longest title, not just the header
	titleIdx := -1
	for i, h := range headerCols {
		if h == "title" {
			titleIdx = i
			break
		}
	}
	if titleIdx >= 0 {
		titleSep := sepCols[titleIdx]
		longestTitle := "A very long task title here"
		if len(titleSep) < len(longestTitle) {
			t.Errorf("Title separator (%d dashes) should be at least as wide as longest title (%d chars)",
				len(titleSep), len(longestTitle))
		}
	}
}

func TestListCommand_ColorAlignmentMatchesPlain(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task One", Status: model.StatusInProgress, Priority: model.PriorityHigh, FilePath: "a.md"},
		{ID: "002", Title: "Task Two", Status: model.StatusPending, Priority: model.PriorityCritical, FilePath: "b.md"},
	}

	// Capture with no color
	resetListFlags()
	noColor = true
	plainOutput := captureListTableOutput(t, tasks, "id,title,status,priority,file")

	// Capture with color
	resetListFlags()
	noColor = false
	forceColor = true
	defer func() { forceColor = false }()
	os.Unsetenv("NO_COLOR")
	colorOutput := captureListTableOutput(t, tasks, "id,title,status,priority,file")

	plainLines := strings.Split(strings.TrimRight(plainOutput, "\n"), "\n")
	colorLines := strings.Split(strings.TrimRight(colorOutput, "\n"), "\n")

	if len(plainLines) != len(colorLines) {
		t.Fatalf("Line count mismatch: plain=%d, color=%d", len(plainLines), len(colorLines))
	}

	// Strip ANSI codes and compare visible widths per line
	for i := range plainLines {
		plainLen := len(strings.TrimRight(plainLines[i], " "))
		strippedColor := StripANSI(colorLines[i])
		colorLen := len(strings.TrimRight(strippedColor, " "))
		if plainLen != colorLen {
			t.Errorf("Line %d visible width mismatch: plain=%d, color(stripped)=%d\n  plain: %q\n  color: %q",
				i, plainLen, colorLen, plainLines[i], strippedColor)
		}
	}
}

// splitTableColumns splits a table-formatted line into columns.
// Columns are separated by 2+ spaces.
func splitTableColumns(line string) []string {
	parts := strings.Fields(line)
	return parts
}

func TestListCommand_LimitFlag(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "A", Status: model.StatusPending, Priority: model.PriorityLow, FilePath: "a.md"},
		{ID: "002", Title: "B", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "b.md"},
		{ID: "003", Title: "C", Status: model.StatusPending, Priority: model.PriorityCritical, FilePath: "c.md"},
		{ID: "004", Title: "D", Status: model.StatusPending, Priority: model.PriorityMedium, FilePath: "d.md"},
		{ID: "005", Title: "E", Status: model.StatusPending, Priority: model.PriorityLow, FilePath: "e.md"},
	}

	tests := []struct {
		name          string
		limit         int
		sort          string
		expectedCount int
		firstID       string
	}{
		{"no limit", 0, "", 5, "001"},
		{"limit less than total", 3, "", 3, "001"},
		{"limit equal to total", 5, "", 5, "001"},
		{"limit greater than total", 10, "", 5, "001"},
		{"limit 1", 1, "", 1, "001"},
		{"limit with sort by priority", 2, "priority", 2, "003"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetListFlags()
			listLimit = tt.limit

			tasksCopy := make([]*model.Task, len(tasks))
			copy(tasksCopy, tasks)

			if tt.sort != "" {
				if err := sortTasks(tasksCopy, tt.sort); err != nil {
					t.Fatalf("sortTasks() error: %v", err)
				}
			}

			if listLimit > 0 && listLimit < len(tasksCopy) {
				tasksCopy = tasksCopy[:listLimit]
			}

			if len(tasksCopy) != tt.expectedCount {
				t.Errorf("got %d tasks, want %d", len(tasksCopy), tt.expectedCount)
			}
			if tasksCopy[0].ID != tt.firstID {
				t.Errorf("first task ID = %s, want %s", tasksCopy[0].ID, tt.firstID)
			}
		})
	}
}

func TestListCommand_LimitTableOutput(t *testing.T) {
	resetListFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "First", Status: model.StatusPending, Priority: model.PriorityHigh, FilePath: "a.md"},
		{ID: "002", Title: "Second", Status: model.StatusPending, Priority: model.PriorityLow, FilePath: "b.md"},
		{ID: "003", Title: "Third", Status: model.StatusPending, Priority: model.PriorityMedium, FilePath: "c.md"},
	}

	// Apply limit before outputting
	limited := tasks[:2]

	output := captureListTableOutput(t, limited, "id,title,status")

	if !strings.Contains(output, "001") {
		t.Error("Expected task 001 in output")
	}
	if !strings.Contains(output, "002") {
		t.Error("Expected task 002 in output")
	}
	if strings.Contains(output, "003") {
		t.Error("Task 003 should not appear in limited output")
	}
}

func TestListCommand_LimitJSONOutput(t *testing.T) {
	resetListFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "First", Status: model.StatusPending},
		{ID: "002", Title: "Second", Status: model.StatusPending},
		{ID: "003", Title: "Third", Status: model.StatusPending},
	}

	limited := tasks[:2]

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(limited)
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("outputJSON failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "001") {
		t.Error("Expected task 001 in JSON output")
	}
	if !strings.Contains(output, "002") {
		t.Error("Expected task 002 in JSON output")
	}
	if strings.Contains(output, "003") {
		t.Error("Task 003 should not appear in limited JSON output")
	}
}

func TestFilterTasksByScope(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Web feature", Touches: []string{"web", "api"}},
		{ID: "002", Title: "CLI feature", Touches: []string{"cli"}},
		{ID: "003", Title: "Web styling", Touches: []string{"web"}},
		{ID: "004", Title: "No scope", Touches: nil},
	}

	tests := []struct {
		name    string
		scope   string
		wantIDs []string
	}{
		{"exact match", "cli", []string{"002"}},
		{"multiple matches", "web", []string{"001", "003"}},
		{"wildcard match", "web*", []string{"001", "003"}},
		{"wildcard all", "c*", []string{"002"}},
		{"no match", "nonexistent", nil},
		{"api match", "api", []string{"001"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTasksByScope(tasks, tt.scope)
			if len(result) != len(tt.wantIDs) {
				t.Fatalf("got %d tasks, want %d", len(result), len(tt.wantIDs))
			}
			for i, task := range result {
				if task.ID != tt.wantIDs[i] {
					t.Errorf("task[%d].ID = %s, want %s", i, task.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

// createScopedListTaskFiles creates test task files with touches fields for list tests.
func createScopedListTaskFiles(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	taskFiles := map[string]string{
		"001-web.md":     "---\nid: \"001\"\ntitle: \"Web feature\"\nstatus: pending\npriority: high\ntouches: [\"web\", \"api\"]\ncreated: 2026-01-01\n---\n# Web feature\n",
		"002-cli.md":     "---\nid: \"002\"\ntitle: \"CLI feature\"\nstatus: pending\npriority: medium\ntouches: [\"cli\"]\ncreated: 2026-01-01\n---\n# CLI feature\n",
		"003-web.md":     "---\nid: \"003\"\ntitle: \"Web styling\"\nstatus: pending\npriority: low\ntouches: [\"web\"]\ncreated: 2026-01-01\n---\n# Web styling\n",
		"004-noscope.md": "---\nid: \"004\"\ntitle: \"No scope task\"\nstatus: pending\npriority: medium\ncreated: 2026-01-01\n---\n# No scope task\n",
	}

	for name, content := range taskFiles {
		if err := os.WriteFile(tmpDir+"/"+name, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return tmpDir
}

// captureRunList runs runList and captures stdout.
func captureRunList(t *testing.T, args []string) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runList(listCmd, args)
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runList failed: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestListCommand_ScopeFilter(t *testing.T) {
	tmpDir := createScopedListTaskFiles(t)

	t.Run("scope filters to matching tasks", func(t *testing.T) {
		resetListFlags()
		listFormat = "json"
		listScope = "web"

		output := captureRunList(t, []string{tmpDir})

		if !strings.Contains(output, "001") {
			t.Error("Expected task 001 (touches web)")
		}
		if !strings.Contains(output, "003") {
			t.Error("Expected task 003 (touches web)")
		}
		if strings.Contains(output, `"004"`) {
			t.Error("Task 004 should be excluded (no touches)")
		}
	})

	t.Run("scope with wildcard", func(t *testing.T) {
		resetListFlags()
		listFormat = "json"
		listScope = "w*"

		output := captureRunList(t, []string{tmpDir})

		if !strings.Contains(output, "001") || !strings.Contains(output, "003") {
			t.Error("Expected tasks 001 and 003 to match w* wildcard")
		}
	})

	t.Run("scope with no matches", func(t *testing.T) {
		resetListFlags()
		listFormat = "table"
		listScope = "nonexistent"

		output := captureRunList(t, []string{tmpDir})

		if !strings.Contains(output, "No tasks found") {
			t.Error("Expected 'No tasks found' message for non-matching scope")
		}
	})

	t.Run("scope combined with filter", func(t *testing.T) {
		resetListFlags()
		listFormat = "json"
		listScope = "web"
		listFilters = []string{"priority=high"}

		output := captureRunList(t, []string{tmpDir})

		if !strings.Contains(output, "001") {
			t.Error("Expected task 001 (web + high priority)")
		}
		if strings.Contains(output, `"003"`) {
			t.Error("Task 003 should be excluded (web but low priority)")
		}
	})
}

func TestOutputJSON_NilTasks_ReturnsEmptyArray(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON(nil)
	if err != nil {
		w.Close()
		os.Stdout = oldStdout
		t.Fatalf("outputJSON failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if output != "[]" {
		t.Errorf("expected empty JSON array '[]', got %q", output)
	}
}

func TestGetColumnValue_Phase(t *testing.T) {
	task := &model.Task{
		ID:    "001",
		Title: "Test Task",
		Phase: "v0.2",
	}

	result := getColumnValue(task, "phase")
	if result != "v0.2" {
		t.Errorf("getColumnValue(phase) = %s, want v0.2", result)
	}

	// Empty phase
	task2 := &model.Task{ID: "002", Title: "No Phase"}
	result2 := getColumnValue(task2, "phase")
	if result2 != "" {
		t.Errorf("getColumnValue(phase) for task without phase = %q, want empty", result2)
	}
}

func TestListCommand_PhaseColumn(t *testing.T) {
	resetListFlags()

	tasks := []*model.Task{
		{ID: "001", Title: "Feature A", Status: model.StatusPending, Phase: "v0.2"},
		{ID: "002", Title: "Feature B", Status: model.StatusPending, Phase: "v0.3"},
		{ID: "003", Title: "Feature C", Status: model.StatusPending},
	}

	output := captureListTableOutput(t, tasks, "id,title,phase")

	if !strings.Contains(output, "phase") {
		t.Error("Expected 'phase' column header in output")
	}
	if !strings.Contains(output, "v0.2") {
		t.Error("Expected phase value 'v0.2' in output")
	}
	if !strings.Contains(output, "v0.3") {
		t.Error("Expected phase value 'v0.3' in output")
	}
}

func TestApplyShortcutFilters_Phase(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Feature A", Status: model.StatusPending, Phase: "v0.2"},
		{ID: "002", Title: "Feature B", Status: model.StatusCompleted, Phase: "v0.3"},
		{ID: "003", Title: "Feature C", Status: model.StatusPending, Phase: "v0.2"},
		{ID: "004", Title: "Feature D", Status: model.StatusPending},
	}

	tests := []struct {
		name    string
		s       FilterShortcuts
		wantIDs []string
	}{
		{"exact match", FilterShortcuts{Phase: "v0.2"}, []string{"001", "003"}},
		{"single match", FilterShortcuts{Phase: "v0.3"}, []string{"002"}},
		{"no match", FilterShortcuts{Phase: "v1.0"}, nil},
		{"phase none", FilterShortcuts{Phase: "none"}, []string{"004"}},
		{"phase any", FilterShortcuts{Phase: "any"}, []string{"001", "002", "003"}},
		{"phase none + status pending", FilterShortcuts{Phase: "none", Status: "pending"}, []string{"004"}},
		{"phase any + status pending", FilterShortcuts{Phase: "any", Status: "pending"}, []string{"001", "003"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applyShortcutFilters(tasks, tt.s)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != len(tt.wantIDs) {
				t.Fatalf("got %d tasks, want %d", len(result), len(tt.wantIDs))
			}
			for i, task := range result {
				if task.ID != tt.wantIDs[i] {
					t.Errorf("task[%d].ID = %s, want %s", i, task.ID, tt.wantIDs[i])
				}
			}
		})
	}
}
