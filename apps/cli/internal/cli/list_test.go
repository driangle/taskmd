package cli

import (
	"testing"
	"time"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestMatchesFilter(t *testing.T) {
	task := &model.Task{
		ID:           "001",
		Title:        "Test Task",
		Status:       model.StatusPending,
		Priority:     model.PriorityHigh,
		Effort:       model.EffortSmall,
		Dependencies: []string{"002"},
		Tags:         []string{"cli", "test"},
		Group:        "testing",
	}

	tests := []struct {
		name     string
		field    string
		value    string
		expected bool
	}{
		{"status match", "status", "pending", true},
		{"status no match", "status", "completed", false},
		{"priority match", "priority", "high", true},
		{"priority no match", "priority", "low", false},
		{"id match", "id", "001", true},
		{"id no match", "id", "002", false},
		{"title contains", "title", "test", true},
		{"title not contains", "title", "xyz", false},
		{"group match", "group", "testing", true},
		{"blocked true", "blocked", "true", true},
		{"blocked false", "blocked", "false", false},
		{"tag exists", "tag", "cli", true},
		{"tag not exists", "tag", "missing", false},
		{"unknown field", "unknown", "value", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(task, tt.field, tt.value)
			if result != tt.expected {
				t.Errorf("matchesFilter(%s, %s) = %v, want %v", tt.field, tt.value, result, tt.expected)
			}
		})
	}
}

func TestApplyFilters(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Status: model.StatusPending, Priority: model.PriorityHigh},
		{ID: "002", Status: model.StatusCompleted, Priority: model.PriorityLow},
		{ID: "003", Status: model.StatusPending, Priority: model.PriorityMedium},
		{ID: "004", Status: model.StatusPending, Priority: model.PriorityHigh},
	}

	tests := []struct {
		name        string
		filters     []string
		expectedLen int
		wantErr     bool
	}{
		{"single filter by status", []string{"status=pending"}, 3, false},
		{"single filter by priority", []string{"priority=high"}, 2, false},
		{"multiple filters AND", []string{"status=pending", "priority=high"}, 2, false},
		{"no matches", []string{"status=blocked"}, 0, false},
		{"multiple filters no matches", []string{"status=pending", "priority=low"}, 0, false},
		{"invalid filter format", []string{"invalid"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := applyFilters(tasks, tt.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) != tt.expectedLen {
				t.Errorf("applyFilters() got %d tasks, want %d", len(result), tt.expectedLen)
			}
		})
	}
}

func TestMatchesAllFilters(t *testing.T) {
	task := &model.Task{
		ID:       "001",
		Status:   model.StatusPending,
		Priority: model.PriorityHigh,
		Effort:   model.EffortSmall,
	}

	tests := []struct {
		name     string
		filters  []filterCriteria
		expected bool
	}{
		{
			name:     "single matching filter",
			filters:  []filterCriteria{{field: "status", value: "pending"}},
			expected: true,
		},
		{
			name:     "multiple matching filters",
			filters:  []filterCriteria{{field: "status", value: "pending"}, {field: "priority", value: "high"}},
			expected: true,
		},
		{
			name:     "one non-matching filter",
			filters:  []filterCriteria{{field: "status", value: "pending"}, {field: "priority", value: "low"}},
			expected: false,
		},
		{
			name:     "all non-matching filters",
			filters:  []filterCriteria{{field: "status", value: "completed"}, {field: "priority", value: "low"}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesAllFilters(task, tt.filters)
			if result != tt.expected {
				t.Errorf("matchesAllFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

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
