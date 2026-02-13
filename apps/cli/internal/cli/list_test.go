package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestSortTasks(t *testing.T) {
	now := time.Now()
	tasks := []*model.Task{
		{ID: "003", Title: "C", Status: model.StatusPending, Priority: model.PriorityLow, Effort: model.EffortLarge, Created: now.Add(2 * time.Hour)},
		{ID: "001", Title: "A", Status: model.StatusCompleted, Priority: model.PriorityHigh, Effort: model.EffortSmall, Created: now},
		{ID: "002", Title: "B", Status: model.StatusInProgress, Priority: model.PriorityCritical, Effort: model.EffortMedium, Created: now.Add(1 * time.Hour)},
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
		Created:      created,
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
	listNoColor = true
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
	listNoColor = false
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
	// listNoColor is already true from resetListFlags
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
	listNoColor = false // enable via flag, but env var should override

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
	listNoColor = false
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
