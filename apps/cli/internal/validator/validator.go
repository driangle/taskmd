package validator

import (
	"fmt"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
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

// ConfigData holds parsed config file data for validation.
// Extracted in the CLI layer so the validator stays viper-free.
type ConfigData struct {
	Scopes     map[string]ScopeConfig
	TopKeys    []string
	ConfigPath string
}

// ScopeConfig holds the configuration for a single scope entry.
type ScopeConfig struct {
	Paths []string // nil means the paths field was absent
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
	v.checkMissingParent(tasks, taskMap, result)
	v.checkParentSelfReference(tasks, result)
	v.checkParentCycles(tasks, taskMap, result)

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
		model.StatusCancelled:  true,
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
				fmt.Sprintf("invalid status: '%s' (valid values: pending, in-progress, completed, blocked, cancelled)", task.Status))
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

// checkMissingParent checks that parent references an existing task
func (v *Validator) checkMissingParent(tasks []*model.Task, taskMap map[string]*model.Task, result *ValidationResult) {
	for _, task := range tasks {
		if task.Parent == "" {
			continue
		}
		if _, exists := taskMap[task.Parent]; !exists {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("parent references non-existent task: '%s'", task.Parent))
		}
	}
}

// checkParentSelfReference warns if a task lists itself as parent
func (v *Validator) checkParentSelfReference(tasks []*model.Task, result *ValidationResult) {
	for _, task := range tasks {
		if task.Parent != "" && task.Parent == task.ID {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				"task references itself as parent")
		}
	}
}

// checkParentCycles detects cycles in the parent chain (e.g. A→B→A)
func (v *Validator) checkParentCycles(tasks []*model.Task, taskMap map[string]*model.Task, result *ValidationResult) {
	for _, task := range tasks {
		if task.Parent == "" {
			continue
		}
		visited := map[string]bool{task.ID: true}
		current := task.Parent
		for current != "" {
			if visited[current] {
				result.AddIssue(LevelError, task.ID, task.FilePath,
					fmt.Sprintf("parent cycle detected: task '%s' creates a cycle via '%s'", task.ID, current))
				break
			}
			visited[current] = true
			parent, exists := taskMap[current]
			if !exists {
				break // missing parent already reported
			}
			current = parent.Parent
		}
	}
}

// ValidateConfig checks the .taskmd.yaml config file for issues.
// Returns an empty result if config is nil.
func (v *Validator) ValidateConfig(config *ConfigData) *ValidationResult {
	result := &ValidationResult{
		Issues: make([]ValidationIssue, 0),
	}
	if config == nil {
		return result
	}

	v.checkConfigScopes(config, result)
	v.checkUnknownConfigKeys(config, result)

	return result
}

// checkConfigScopes validates each scope entry has a non-empty paths array.
func (v *Validator) checkConfigScopes(config *ConfigData, result *ValidationResult) {
	for name, scope := range config.Scopes {
		if scope.Paths == nil {
			result.AddIssue(LevelError, "", config.ConfigPath,
				fmt.Sprintf("scope '%s' is missing required field: paths", name))
		} else if len(scope.Paths) == 0 {
			result.AddIssue(LevelError, "", config.ConfigPath,
				fmt.Sprintf("scope '%s' has empty paths array", name))
		}
	}
}

var knownConfigKeys = map[string]bool{
	"dir":      true,
	"task-dir": true,
	"web":      true,
	"scopes":   true,
	"sync":     true,
	"ignore":   true,
}

// checkUnknownConfigKeys warns about unrecognized top-level config keys.
func (v *Validator) checkUnknownConfigKeys(config *ConfigData, result *ValidationResult) {
	for _, key := range config.TopKeys {
		if !knownConfigKeys[key] {
			result.AddIssue(LevelWarning, "", config.ConfigPath,
				fmt.Sprintf("unknown config key: '%s'", key))
		}
	}
}

// ValidateTouchesAgainstScopes warns when tasks reference undefined scopes.
// Skips validation if knownScopes is nil or empty.
func (v *Validator) ValidateTouchesAgainstScopes(tasks []*model.Task, knownScopes map[string]bool) *ValidationResult {
	result := &ValidationResult{
		Issues: make([]ValidationIssue, 0),
	}
	if len(knownScopes) == 0 {
		return result
	}

	reported := make(map[string]bool)
	for _, task := range tasks {
		for _, scope := range task.Touches {
			if !knownScopes[scope] && !reported[scope] {
				reported[scope] = true
				result.AddIssue(LevelWarning, task.ID, task.FilePath,
					fmt.Sprintf("touches references undefined scope: '%s'", scope))
			}
		}
	}

	return result
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
