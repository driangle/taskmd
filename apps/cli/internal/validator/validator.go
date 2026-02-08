package validator

import (
	"fmt"
	"strings"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
)

// ValidationLevel represents the severity of a validation issue
type ValidationLevel string

const (
	LevelError   ValidationLevel = "error"
	LevelWarning ValidationLevel = "warning"
)

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Level    ValidationLevel `json:"level"`
	TaskID   string          `json:"task_id,omitempty"`
	FilePath string          `json:"file_path,omitempty"`
	Message  string          `json:"message"`
}

// ValidationResult contains all validation issues found
type ValidationResult struct {
	Issues    []ValidationIssue `json:"issues"`
	Errors    int               `json:"errors"`
	Warnings  int               `json:"warnings"`
	TaskCount int               `json:"task_count"`
}

// IsValid returns true if there are no errors
func (vr *ValidationResult) IsValid() bool {
	return vr.Errors == 0
}

// HasWarnings returns true if there are warnings
func (vr *ValidationResult) HasWarnings() bool {
	return vr.Warnings > 0
}

// AddIssue adds a validation issue and updates counters
func (vr *ValidationResult) AddIssue(level ValidationLevel, taskID, filePath, message string) {
	vr.Issues = append(vr.Issues, ValidationIssue{
		Level:    level,
		TaskID:   taskID,
		FilePath: filePath,
		Message:  message,
	})

	if level == LevelError {
		vr.Errors++
	} else if level == LevelWarning {
		vr.Warnings++
	}
}

// Validator validates task collections
type Validator struct {
	strict bool
}

// NewValidator creates a new validator
func NewValidator(strict bool) *Validator {
	return &Validator{strict: strict}
}

// Validate performs all validation checks on a set of tasks
func (v *Validator) Validate(tasks []*model.Task) *ValidationResult {
	result := &ValidationResult{
		Issues:    make([]ValidationIssue, 0),
		TaskCount: len(tasks),
	}

	// Build task ID map for lookups
	taskMap := make(map[string]*model.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Run validation checks
	v.checkRequiredFields(tasks, result)
	v.checkInvalidFieldValues(tasks, result)
	v.checkDuplicateIDs(tasks, result)
	v.checkMissingDependencies(tasks, taskMap, result)
	v.checkCircularDependencies(tasks, taskMap, result)

	// Strict mode additional checks
	if v.strict {
		v.checkStrictWarnings(tasks, result)
	}

	return result
}

// checkRequiredFields validates that tasks have required fields
func (v *Validator) checkRequiredFields(tasks []*model.Task, result *ValidationResult) {
	for _, task := range tasks {
		if task.ID == "" {
			result.AddIssue(LevelError, "", task.FilePath, "task is missing required field: id")
		}
		if task.Title == "" {
			result.AddIssue(LevelError, task.ID, task.FilePath, "task is missing required field: title")
		}
	}
}

// checkInvalidFieldValues validates enum field values
func (v *Validator) checkInvalidFieldValues(tasks []*model.Task, result *ValidationResult) {
	validStatuses := map[model.Status]bool{
		model.StatusPending:    true,
		model.StatusInProgress: true,
		model.StatusCompleted:  true,
		model.StatusBlocked:    true,
		"":                     true, // Empty is allowed (will default)
	}

	validPriorities := map[model.Priority]bool{
		model.PriorityLow:      true,
		model.PriorityMedium:   true,
		model.PriorityHigh:     true,
		model.PriorityCritical: true,
		"":                     true, // Empty is allowed (will default)
	}

	validEfforts := map[model.Effort]bool{
		model.EffortSmall:  true,
		model.EffortMedium: true,
		model.EffortLarge:  true,
		"":                 true, // Empty is allowed (will default)
	}

	for _, task := range tasks {
		if !validStatuses[task.Status] {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("invalid status: '%s' (valid values: pending, in-progress, completed, blocked)", task.Status))
		}

		if !validPriorities[task.Priority] {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("invalid priority: '%s' (valid values: low, medium, high, critical)", task.Priority))
		}

		if !validEfforts[task.Effort] {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("invalid effort: '%s' (valid values: small, medium, large)", task.Effort))
		}
	}
}

// checkDuplicateIDs checks for duplicate task IDs
func (v *Validator) checkDuplicateIDs(tasks []*model.Task, result *ValidationResult) {
	seen := make(map[string][]string) // ID -> file paths

	for _, task := range tasks {
		if task.ID == "" {
			continue // Skip, already reported in checkRequiredFields
		}
		seen[task.ID] = append(seen[task.ID], task.FilePath)
	}

	for id, paths := range seen {
		if len(paths) > 1 {
			result.AddIssue(LevelError, id, strings.Join(paths, ", "),
				fmt.Sprintf("duplicate task ID '%s' found in %d files", id, len(paths)))
		}
	}
}

// checkMissingDependencies checks for references to non-existent tasks
func (v *Validator) checkMissingDependencies(tasks []*model.Task, taskMap map[string]*model.Task, result *ValidationResult) {
	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			if _, exists := taskMap[depID]; !exists {
				result.AddIssue(LevelError, task.ID, task.FilePath,
					fmt.Sprintf("dependency references non-existent task: '%s'", depID))
			}
		}
	}
}

// checkCircularDependencies detects cycles in the dependency graph
//
//nolint:gocognit // TODO: refactor to reduce complexity
func (v *Validator) checkCircularDependencies(tasks []*model.Task, taskMap map[string]*model.Task, result *ValidationResult) {
	// Build adjacency list
	graph := make(map[string][]string)
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}

	// Track visit states: 0 = unvisited, 1 = visiting, 2 = visited
	visitState := make(map[string]int)
	path := []string{}

	var hasCycle func(string) bool
	hasCycle = func(taskID string) bool {
		if visitState[taskID] == 1 {
			// Found a cycle - build cycle path
			cycleStart := -1
			for i, id := range path {
				if id == taskID {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cyclePath := append(path[cycleStart:], taskID)
				result.AddIssue(LevelError, taskID, taskMap[taskID].FilePath,
					fmt.Sprintf("circular dependency detected: %s", strings.Join(cyclePath, " -> ")))
			}
			return true
		}

		if visitState[taskID] == 2 {
			return false // Already fully processed
		}

		visitState[taskID] = 1
		path = append(path, taskID)

		for _, depID := range graph[taskID] {
			if _, exists := taskMap[depID]; !exists {
				continue // Skip missing dependencies (already reported)
			}
			if hasCycle(depID) {
				visitState[taskID] = 2
				path = path[:len(path)-1]
				return true
			}
		}

		visitState[taskID] = 2
		path = path[:len(path)-1]
		return false
	}

	// Check each task for cycles
	for taskID := range taskMap {
		if visitState[taskID] == 0 {
			hasCycle(taskID)
		}
	}
}

// checkStrictWarnings performs additional checks in strict mode
func (v *Validator) checkStrictWarnings(tasks []*model.Task, result *ValidationResult) {
	for _, task := range tasks {
		// Warn about tasks with no status
		if task.Status == "" {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task has no status specified (will default to pending)")
		}

		// Warn about tasks with no priority
		if task.Priority == "" {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task has no priority specified (will default to medium)")
		}

		// Warn about tasks with no effort
		if task.Effort == "" {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task has no effort specified (will default to medium)")
		}

		// Warn about tasks with no group
		if task.Group == "" {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task has no group specified")
		}

		// Warn about tasks with no tags
		if len(task.Tags) == 0 {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task has no tags")
		}

		// Warn about empty body
		if strings.TrimSpace(task.Body) == "" {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task has no description/body content")
		}
	}
}
