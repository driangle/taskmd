package validator

import (
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestValidate_RequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*model.Task
		wantErrs int
	}{
		{
			name: "valid task with all required fields",
			tasks: []*model.Task{
				{
					ID:    "001",
					Title: "Test Task",
				},
			},
			wantErrs: 0,
		},
		{
			name: "missing ID",
			tasks: []*model.Task{
				{
					Title: "Test Task",
				},
			},
			wantErrs: 1,
		},
		{
			name: "missing title",
			tasks: []*model.Task{
				{
					ID: "001",
				},
			},
			wantErrs: 1,
		},
		{
			name: "missing both ID and title",
			tasks: []*model.Task{
				{},
			},
			wantErrs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			result := v.Validate(tt.tasks)

			if result.Errors != tt.wantErrs {
				t.Errorf("Validate() errors = %d, want %d", result.Errors, tt.wantErrs)
			}
		})
	}
}

func TestValidate_InvalidFieldValues(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*model.Task
		wantErrs int
	}{
		{
			name: "valid enum values",
			tasks: []*model.Task{
				{
					ID:       "001",
					Title:    "Test",
					Status:   model.StatusPending,
					Priority: model.PriorityHigh,
					Effort:   model.EffortMedium,
				},
			},
			wantErrs: 0,
		},
		{
			name: "invalid status",
			tasks: []*model.Task{
				{
					ID:     "001",
					Title:  "Test",
					Status: "invalid-status",
				},
			},
			wantErrs: 1,
		},
		{
			name: "invalid priority",
			tasks: []*model.Task{
				{
					ID:       "001",
					Title:    "Test",
					Priority: "urgent",
				},
			},
			wantErrs: 1,
		},
		{
			name: "invalid effort",
			tasks: []*model.Task{
				{
					ID:     "001",
					Title:  "Test",
					Effort: "huge",
				},
			},
			wantErrs: 1,
		},
		{
			name: "multiple invalid values",
			tasks: []*model.Task{
				{
					ID:       "001",
					Title:    "Test",
					Status:   "bad",
					Priority: "wrong",
					Effort:   "nope",
				},
			},
			wantErrs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			result := v.Validate(tt.tasks)

			if result.Errors != tt.wantErrs {
				t.Errorf("Validate() errors = %d, want %d", result.Errors, tt.wantErrs)
				for _, issue := range result.Issues {
					t.Logf("  Issue: %s", issue.Message)
				}
			}
		})
	}
}

func TestValidate_DuplicateIDs(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:       "001",
			Title:    "Task 1",
			FilePath: "/path/to/task1.md",
		},
		{
			ID:       "001",
			Title:    "Task 2",
			FilePath: "/path/to/task2.md",
		},
		{
			ID:       "002",
			Title:    "Task 3",
			FilePath: "/path/to/task3.md",
		},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	if result.Errors != 1 {
		t.Errorf("Expected 1 error for duplicate IDs, got %d", result.Errors)
	}

	// Check that the error message mentions both files
	foundDuplicateError := false
	for _, issue := range result.Issues {
		if issue.TaskID == "001" && issue.Level == LevelError {
			foundDuplicateError = true
		}
	}

	if !foundDuplicateError {
		t.Error("Expected duplicate ID error for task 001")
	}
}

func TestValidate_MissingDependencies(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:           "001",
			Title:        "Task 1",
			Dependencies: []string{"002", "999"}, // 999 doesn't exist
		},
		{
			ID:    "002",
			Title: "Task 2",
		},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	if result.Errors != 1 {
		t.Errorf("Expected 1 error for missing dependency, got %d", result.Errors)
	}

	// Check that the error mentions the missing task ID
	foundMissingDep := false
	for _, issue := range result.Issues {
		if issue.TaskID == "001" && issue.Level == LevelError {
			foundMissingDep = true
		}
	}

	if !foundMissingDep {
		t.Error("Expected missing dependency error for task 001")
	}
}

func TestValidate_CircularDependencies(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*model.Task
		wantErrs bool
	}{
		{
			name: "simple cycle: A -> B -> A",
			tasks: []*model.Task{
				{
					ID:           "A",
					Title:        "Task A",
					Dependencies: []string{"B"},
				},
				{
					ID:           "B",
					Title:        "Task B",
					Dependencies: []string{"A"},
				},
			},
			wantErrs: true,
		},
		{
			name: "three-way cycle: A -> B -> C -> A",
			tasks: []*model.Task{
				{
					ID:           "A",
					Title:        "Task A",
					Dependencies: []string{"B"},
				},
				{
					ID:           "B",
					Title:        "Task B",
					Dependencies: []string{"C"},
				},
				{
					ID:           "C",
					Title:        "Task C",
					Dependencies: []string{"A"},
				},
			},
			wantErrs: true,
		},
		{
			name: "self-cycle: A -> A",
			tasks: []*model.Task{
				{
					ID:           "A",
					Title:        "Task A",
					Dependencies: []string{"A"},
				},
			},
			wantErrs: true,
		},
		{
			name: "no cycle: linear chain A -> B -> C",
			tasks: []*model.Task{
				{
					ID:           "A",
					Title:        "Task A",
					Dependencies: []string{"B"},
				},
				{
					ID:           "B",
					Title:        "Task B",
					Dependencies: []string{"C"},
				},
				{
					ID:    "C",
					Title: "Task C",
				},
			},
			wantErrs: false,
		},
		{
			name: "no cycle: diamond dependency A -> B,C -> D",
			tasks: []*model.Task{
				{
					ID:           "A",
					Title:        "Task A",
					Dependencies: []string{"B", "C"},
				},
				{
					ID:           "B",
					Title:        "Task B",
					Dependencies: []string{"D"},
				},
				{
					ID:           "C",
					Title:        "Task C",
					Dependencies: []string{"D"},
				},
				{
					ID:    "D",
					Title: "Task D",
				},
			},
			wantErrs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(false)
			result := v.Validate(tt.tasks)

			hasCircularError := false
			for _, issue := range result.Issues {
				if issue.Level == LevelError {
					hasCircularError = true
					break
				}
			}

			if hasCircularError != tt.wantErrs {
				t.Errorf("Validate() circular dependency error = %v, want %v", hasCircularError, tt.wantErrs)
				for _, issue := range result.Issues {
					t.Logf("  Issue: [%s] %s", issue.Level, issue.Message)
				}
			}
		})
	}
}

func TestValidate_StrictMode(t *testing.T) {
	task := &model.Task{
		ID:    "001",
		Title: "Test Task",
		// Missing optional fields: Status, Priority, Effort, Group, Tags, Body
	}

	// Non-strict mode should not produce warnings
	v := NewValidator(false)
	result := v.Validate([]*model.Task{task})

	if result.Warnings > 0 {
		t.Errorf("Non-strict mode should not produce warnings, got %d", result.Warnings)
	}

	// Strict mode should produce warnings for missing optional fields
	vStrict := NewValidator(true)
	resultStrict := vStrict.Validate([]*model.Task{task})

	// Should have warnings for: status, priority, effort, group, tags, body = 6 warnings
	if resultStrict.Warnings < 5 {
		t.Errorf("Strict mode should produce multiple warnings, got %d", resultStrict.Warnings)
		for _, issue := range resultStrict.Issues {
			t.Logf("  Issue: [%s] %s", issue.Level, issue.Message)
		}
	}
}

func TestValidationResult_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		result    *ValidationResult
		wantValid bool
	}{
		{
			name:      "no issues",
			result:    &ValidationResult{},
			wantValid: true,
		},
		{
			name: "only warnings",
			result: &ValidationResult{
				Warnings: 3,
			},
			wantValid: true,
		},
		{
			name: "has errors",
			result: &ValidationResult{
				Errors: 1,
			},
			wantValid: false,
		},
		{
			name: "errors and warnings",
			result: &ValidationResult{
				Errors:   1,
				Warnings: 2,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestValidate_ComplexScenario(t *testing.T) {
	// Create a complex scenario with multiple issues
	tasks := []*model.Task{
		{
			// Valid task
			ID:       "001",
			Title:    "Valid Task",
			Status:   model.StatusPending,
			Priority: model.PriorityHigh,
			Effort:   model.EffortSmall,
		},
		{
			// Missing title
			ID: "002",
		},
		{
			// Invalid status
			ID:     "003",
			Title:  "Task 3",
			Status: "wrong",
		},
		{
			// Duplicate ID with task 001
			ID:       "001",
			Title:    "Duplicate",
			FilePath: "/path/duplicate.md",
		},
		{
			// Missing dependency
			ID:           "004",
			Title:        "Task 4",
			Dependencies: []string{"999"},
		},
		{
			// Part of circular dependency
			ID:           "005",
			Title:        "Task 5",
			Dependencies: []string{"006"},
		},
		{
			// Part of circular dependency
			ID:           "006",
			Title:        "Task 6",
			Dependencies: []string{"005"},
		},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	// Should have multiple errors:
	// - Missing title (002)
	// - Invalid status (003)
	// - Duplicate ID (001)
	// - Missing dependency (004)
	// - Circular dependency (005/006)
	if result.Errors < 4 {
		t.Errorf("Expected at least 4 errors, got %d", result.Errors)
		for _, issue := range result.Issues {
			t.Logf("  [%s] %s: %s", issue.Level, issue.TaskID, issue.Message)
		}
	}

	if result.IsValid() {
		t.Error("Expected validation to fail with multiple errors")
	}
}
