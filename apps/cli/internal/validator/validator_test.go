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

func TestValidate_CancelledStatus(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*model.Task
		wantErrs int
	}{
		{
			name: "valid cancelled status",
			tasks: []*model.Task{
				{
					ID:     "001",
					Title:  "Cancelled Task",
					Status: model.StatusCancelled,
				},
			},
			wantErrs: 0,
		},
		{
			name: "all valid statuses including cancelled",
			tasks: []*model.Task{
				{ID: "001", Title: "Pending", Status: model.StatusPending},
				{ID: "002", Title: "In Progress", Status: model.StatusInProgress},
				{ID: "003", Title: "Completed", Status: model.StatusCompleted},
				{ID: "004", Title: "Blocked", Status: model.StatusBlocked},
				{ID: "005", Title: "Cancelled", Status: model.StatusCancelled},
			},
			wantErrs: 0,
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

func TestValidate_MissingParent(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task 1", Parent: "999"},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	if result.Errors != 1 {
		t.Errorf("Expected 1 error for missing parent, got %d", result.Errors)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.TaskID == "001" && issue.Level == LevelError {
			found = true
		}
	}
	if !found {
		t.Error("Expected missing parent error for task 001")
	}
}

func TestValidate_ParentSelfReference(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task 1", Parent: "001"},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	if result.Warnings != 1 {
		t.Errorf("Expected 1 warning for self-referencing parent, got %d", result.Warnings)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.TaskID == "001" && issue.Level == LevelWarning {
			found = true
		}
	}
	if !found {
		t.Error("Expected self-reference warning for task 001")
	}
}

func TestValidate_ParentCycle(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Task 1", Parent: "002"},
		{ID: "002", Title: "Task 2", Parent: "001"},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	foundCycleError := false
	for _, issue := range result.Issues {
		if issue.Level == LevelError && issue.Message != "" {
			foundCycleError = true
		}
	}
	if !foundCycleError {
		t.Error("Expected parent cycle error")
	}
}

func TestValidate_ValidParent(t *testing.T) {
	tasks := []*model.Task{
		{ID: "001", Title: "Parent Task"},
		{ID: "002", Title: "Child Task", Parent: "001"},
	}

	v := NewValidator(false)
	result := v.Validate(tasks)

	if result.Errors != 0 {
		t.Errorf("Expected no errors for valid parent, got %d", result.Errors)
		for _, issue := range result.Issues {
			t.Logf("  Issue: [%s] %s: %s", issue.Level, issue.TaskID, issue.Message)
		}
	}
	if result.Warnings != 0 {
		t.Errorf("Expected no warnings for valid parent, got %d", result.Warnings)
	}
}

// --- Config validation tests ---

func TestValidateConfig_ValidScopes(t *testing.T) {
	v := NewValidator(false)
	config := &ConfigData{
		Scopes: map[string]ScopeConfig{
			"cli/graph": {Paths: []string{"apps/cli/internal/graph/"}},
			"web/board": {Paths: []string{"apps/web/src/components/board/"}},
		},
		TopKeys:    []string{"scopes", "dir", "web"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	if result.Errors != 0 {
		t.Errorf("Expected no errors, got %d", result.Errors)
		for _, issue := range result.Issues {
			t.Logf("  Issue: [%s] %s", issue.Level, issue.Message)
		}
	}
	if result.Warnings != 0 {
		t.Errorf("Expected no warnings, got %d", result.Warnings)
	}
}

func TestValidateConfig_MissingPaths(t *testing.T) {
	v := NewValidator(false)
	config := &ConfigData{
		Scopes: map[string]ScopeConfig{
			"cli/graph": {Paths: nil}, // missing paths field
		},
		TopKeys:    []string{"scopes"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	if result.Errors != 1 {
		t.Errorf("Expected 1 error for missing paths, got %d", result.Errors)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Level == LevelError && issue.Message == "scope 'cli/graph' is missing required field: paths" {
			found = true
		}
	}
	if !found {
		t.Error("Expected missing paths error for scope cli/graph")
	}
}

func TestValidateConfig_EmptyPaths(t *testing.T) {
	v := NewValidator(false)
	config := &ConfigData{
		Scopes: map[string]ScopeConfig{
			"cli/graph": {Paths: []string{}}, // empty array
		},
		TopKeys:    []string{"scopes"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	if result.Errors != 1 {
		t.Errorf("Expected 1 error for empty paths, got %d", result.Errors)
	}

	found := false
	for _, issue := range result.Issues {
		if issue.Level == LevelError && issue.Message == "scope 'cli/graph' has empty paths array" {
			found = true
		}
	}
	if !found {
		t.Error("Expected empty paths error for scope cli/graph")
	}
}

func TestValidateConfig_UnknownKeys(t *testing.T) {
	v := NewValidator(false)
	config := &ConfigData{
		Scopes:     map[string]ScopeConfig{},
		TopKeys:    []string{"scopes", "dir", "banana", "foobar"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	if result.Warnings != 2 {
		t.Errorf("Expected 2 warnings for unknown keys, got %d", result.Warnings)
		for _, issue := range result.Issues {
			t.Logf("  Issue: [%s] %s", issue.Level, issue.Message)
		}
	}
	if result.Errors != 0 {
		t.Errorf("Expected no errors, got %d", result.Errors)
	}
}

func TestValidateConfig_NilConfig(t *testing.T) {
	v := NewValidator(false)
	result := v.ValidateConfig(nil)

	if result.Errors != 0 || result.Warnings != 0 {
		t.Errorf("Expected no issues for nil config, got %d errors, %d warnings", result.Errors, result.Warnings)
	}
}

func TestValidateConfig_NoScopes(t *testing.T) {
	v := NewValidator(false)
	config := &ConfigData{
		Scopes:     nil,
		TopKeys:    []string{"dir", "web"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	if result.Errors != 0 || result.Warnings != 0 {
		t.Errorf("Expected no issues when no scopes section, got %d errors, %d warnings", result.Errors, result.Warnings)
	}
}

func TestValidateTouches_ValidReferences(t *testing.T) {
	v := NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Task 1", Touches: []string{"cli/graph", "web/board"}},
		{ID: "002", Title: "Task 2", Touches: []string{"cli/graph"}},
	}
	scopes := map[string]bool{"cli/graph": true, "web/board": true}

	result := v.ValidateTouchesAgainstScopes(tasks, scopes)

	if result.Warnings != 0 {
		t.Errorf("Expected no warnings, got %d", result.Warnings)
		for _, issue := range result.Issues {
			t.Logf("  Issue: [%s] %s", issue.Level, issue.Message)
		}
	}
}

func TestValidateTouches_UndefinedScope(t *testing.T) {
	v := NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Task 1", Touches: []string{"cli/graph", "unknown/scope"}},
		{ID: "002", Title: "Task 2", Touches: []string{"unknown/scope"}}, // same unknown scope, should not duplicate
		{ID: "003", Title: "Task 3", Touches: []string{"another/missing"}},
	}
	scopes := map[string]bool{"cli/graph": true}

	result := v.ValidateTouchesAgainstScopes(tasks, scopes)

	if result.Warnings != 2 {
		t.Errorf("Expected 2 deduplicated warnings, got %d", result.Warnings)
		for _, issue := range result.Issues {
			t.Logf("  Issue: [%s] task=%s %s", issue.Level, issue.TaskID, issue.Message)
		}
	}

	// Verify the first occurrence is reported (task 001 for unknown/scope)
	found := false
	for _, issue := range result.Issues {
		if issue.TaskID == "001" && issue.Message == "touches references undefined scope: 'unknown/scope'" {
			found = true
		}
	}
	if !found {
		t.Error("Expected warning for task 001 referencing unknown/scope")
	}
}

func TestValidateTouches_NoScopesConfigured(t *testing.T) {
	v := NewValidator(false)
	tasks := []*model.Task{
		{ID: "001", Title: "Task 1", Touches: []string{"anything"}},
	}

	// nil scopes
	result := v.ValidateTouchesAgainstScopes(tasks, nil)
	if result.Warnings != 0 {
		t.Errorf("Expected no warnings for nil scopes, got %d", result.Warnings)
	}

	// empty scopes
	result = v.ValidateTouchesAgainstScopes(tasks, map[string]bool{})
	if result.Warnings != 0 {
		t.Errorf("Expected no warnings for empty scopes, got %d", result.Warnings)
	}
}

func TestValidateConfig_ScopesWithDescription(t *testing.T) {
	v := NewValidator(false)
	config := &ConfigData{
		Scopes: map[string]ScopeConfig{
			"cli/graph": {
				Description: "Graph visualization",
				Paths:       []string{"apps/cli/internal/graph/"},
			},
		},
		TopKeys:    []string{"scopes"},
		ConfigPath: ".taskmd.yaml",
	}

	result := v.ValidateConfig(config)

	if result.Errors != 0 {
		t.Errorf("Expected no errors, got %d", result.Errors)
		for _, issue := range result.Issues {
			t.Logf("  Issue: [%s] %s", issue.Level, issue.Message)
		}
	}
}

func TestValidateConfig_ScopeDescriptionInErrorMessage(t *testing.T) {
	v := NewValidator(false)

	t.Run("missing paths with description", func(t *testing.T) {
		config := &ConfigData{
			Scopes: map[string]ScopeConfig{
				"cli/graph": {Description: "Graph visualization", Paths: nil},
			},
			TopKeys:    []string{"scopes"},
			ConfigPath: ".taskmd.yaml",
		}

		result := v.ValidateConfig(config)

		if result.Errors != 1 {
			t.Fatalf("Expected 1 error, got %d", result.Errors)
		}

		want := "scope 'cli/graph' (Graph visualization) is missing required field: paths"
		if result.Issues[0].Message != want {
			t.Errorf("got message %q, want %q", result.Issues[0].Message, want)
		}
	})

	t.Run("empty paths with description", func(t *testing.T) {
		config := &ConfigData{
			Scopes: map[string]ScopeConfig{
				"cli/graph": {Description: "Graph visualization", Paths: []string{}},
			},
			TopKeys:    []string{"scopes"},
			ConfigPath: ".taskmd.yaml",
		}

		result := v.ValidateConfig(config)

		if result.Errors != 1 {
			t.Fatalf("Expected 1 error, got %d", result.Errors)
		}

		want := "scope 'cli/graph' (Graph visualization) has empty paths array"
		if result.Issues[0].Message != want {
			t.Errorf("got message %q, want %q", result.Issues[0].Message, want)
		}
	})

	t.Run("missing paths without description unchanged", func(t *testing.T) {
		config := &ConfigData{
			Scopes: map[string]ScopeConfig{
				"cli/graph": {Paths: nil},
			},
			TopKeys:    []string{"scopes"},
			ConfigPath: ".taskmd.yaml",
		}

		result := v.ValidateConfig(config)

		if result.Errors != 1 {
			t.Fatalf("Expected 1 error, got %d", result.Errors)
		}

		want := "scope 'cli/graph' is missing required field: paths"
		if result.Issues[0].Message != want {
			t.Errorf("got message %q, want %q", result.Issues[0].Message, want)
		}
	})
}
