package validator

import (
	"fmt"
	"strings"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
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
	Scopes map[string]ScopeConfig
	Phases []PhaseConfig
	// StructuralErrors holds problems detected while parsing the raw config
	// that can't be represented in the typed fields above — e.g. a section
	// with the wrong YAML container type (a list where a map is expected).
	// The CLI layer populates these; ValidateConfig surfaces them as errors.
	StructuralErrors []string
	TopKeys          []string
	ConfigPath       string
	Workflow         string
	ID               *IDConfig
	// Effort holds the configured effort vocabulary, lowest to highest.
	// Nil means the key was absent and the default vocabulary applies.
	Effort []string
}

// PhaseConfig holds the configuration for a single phase entry.
type PhaseConfig struct {
	ID          string
	Name        string
	Description string
	Due         model.FlexibleTime
}

// IDConfig holds the configuration for task ID generation.
type IDConfig struct {
	Strategy string // "sequential", "prefixed", or "random"
	Prefix   string // required when strategy is "prefixed"
	Length   int    // ID length (e.g. 6 for random IDs)
	Padding  int    // zero-padding width for sequential IDs
}

// ScopeConfig holds the configuration for a single scope entry.
type ScopeConfig struct {
	Description string   // optional human-readable description
	Paths       []string // nil means the paths field was absent
}

// Validator validates task collections
type Validator struct {
	strict      bool
	externalIDs map[string]bool // IDs known to exist but not subject to validation (e.g. archived tasks)
	efforts     effort.Scale    // zero value is the default effort vocabulary
}

// NewValidator creates a new validator
func NewValidator(strict bool) *Validator {
	return &Validator{strict: strict}
}

// SetEffortScale sets the project's effort vocabulary, used to validate the
// effort field. When unset, the default small, medium, large applies.
func (v *Validator) SetEffortScale(efforts effort.Scale) {
	v.efforts = efforts
}

// SetExternalIDs sets task IDs that are known to exist externally (e.g. in archive).
// These IDs satisfy dependency and parent checks but are not validated themselves.
func (v *Validator) SetExternalIDs(ids map[string]bool) {
	v.externalIDs = ids
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
		model.StatusInReview:   true,
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

	validTypes := map[model.TaskType]bool{
		model.TypeFeature:     true,
		model.TypeBug:         true,
		model.TypeImprovement: true,
		model.TypeChore:       true,
		model.TypeDocs:        true,
		"":                    true, // Empty is allowed (optional field)
	}

	for _, task := range tasks {
		if !validStatuses[task.Status] {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("invalid status: '%s' (valid values: pending, in-progress, completed, in-review, blocked, cancelled)", task.Status))
		}

		if !validPriorities[task.Priority] {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("invalid priority: '%s' (valid values: low, medium, high, critical)", task.Priority))
		}

		// Empty effort is allowed (will default).
		if task.Effort != "" && !v.efforts.Contains(string(task.Effort)) {
			result.AddIssue(LevelError, task.ID, task.FilePath,
				fmt.Sprintf("invalid effort: '%s' (valid values: %s)", task.Effort, v.efforts))
		}

		if !validTypes[task.Type] {
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				fmt.Sprintf("invalid type: '%s' (valid values: feature, bug, improvement, chore, docs)", task.Type))
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
				if v.externalIDs[depID] {
					continue
				}
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
			if v.externalIDs[task.Parent] {
				continue
			}
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

	v.checkStructuralErrors(config, result)
	v.checkConfigScopes(config, result)
	v.checkUnknownConfigKeys(config, result)
	v.checkWorkflowValue(config, result)
	v.checkIDConfig(config, result)
	v.checkPhaseConfig(config, result)
	v.checkEffortConfig(config, result)

	return result
}

// validIDStrategies lists the allowed values for the id.strategy field.
var validIDStrategies = map[string]bool{
	"sequential": true,
	"prefixed":   true,
	"random":     true,
	"ulid":       true,
}

// checkIDConfig validates the id config section.
func (v *Validator) checkIDConfig(config *ConfigData, result *ValidationResult) {
	if config.ID == nil {
		return
	}
	id := config.ID

	if id.Strategy != "" && !validIDStrategies[id.Strategy] {
		result.AddIssue(LevelError, "", config.ConfigPath,
			fmt.Sprintf("invalid id strategy: '%s' (valid values: sequential, prefixed, random, ulid)", id.Strategy))
	}

	if id.Strategy == "prefixed" && id.Prefix == "" {
		result.AddIssue(LevelError, "", config.ConfigPath,
			"id strategy 'prefixed' requires a non-empty prefix")
	}

	if id.Length < 0 {
		result.AddIssue(LevelError, "", config.ConfigPath,
			fmt.Sprintf("id length must not be negative, got %d", id.Length))
	}

	if id.Padding < 0 {
		result.AddIssue(LevelError, "", config.ConfigPath,
			fmt.Sprintf("id padding must not be negative, got %d", id.Padding))
	}
}

// checkEffortConfig validates the configured effort vocabulary. The rules live in
// effort.NewScale, so the config and the runtime vocabulary can never disagree.
func (v *Validator) checkEffortConfig(config *ConfigData, result *ValidationResult) {
	if config.Effort == nil {
		return
	}
	if _, err := effort.NewScale(config.Effort); err != nil {
		result.AddIssue(LevelError, "", config.ConfigPath,
			fmt.Sprintf("invalid effort config: %s", err))
	}
}

// checkPhaseConfig validates phase entries for structural issues and duplicates.
func (v *Validator) checkPhaseConfig(config *ConfigData, result *ValidationResult) {
	if len(config.Phases) == 0 {
		return
	}

	seenIDs := make(map[string]bool)
	seenNames := make(map[string]bool)

	for i, phase := range config.Phases {
		// Structural checks
		if phase.Name == "" {
			result.AddIssue(LevelError, "", config.ConfigPath,
				fmt.Sprintf("phase at index %d is missing required field: name", i))
		}

		if phase.ID == "" {
			label := phase.Name
			if label == "" {
				label = fmt.Sprintf("index %d", i)
			}
			result.AddIssue(LevelWarning, "", config.ConfigPath,
				fmt.Sprintf("phase '%s' is missing field: id (falling back to name)", label))
		}

		// Duplicate checks
		if phase.ID != "" {
			if seenIDs[phase.ID] {
				result.AddIssue(LevelWarning, "", config.ConfigPath,
					fmt.Sprintf("duplicate phase id: '%s'", phase.ID))
			}
			seenIDs[phase.ID] = true
		}

		if phase.Name != "" {
			if seenNames[phase.Name] {
				result.AddIssue(LevelWarning, "", config.ConfigPath,
					fmt.Sprintf("duplicate phase name: '%s'", phase.Name))
			}
			seenNames[phase.Name] = true
		}
	}
}

// checkStructuralErrors surfaces config parse problems (wrong container types)
// detected in the CLI layer as validation errors.
func (v *Validator) checkStructuralErrors(config *ConfigData, result *ValidationResult) {
	for _, msg := range config.StructuralErrors {
		result.AddIssue(LevelError, "", config.ConfigPath, msg)
	}
}

// checkConfigScopes validates each scope entry has a non-empty paths array.
func (v *Validator) checkConfigScopes(config *ConfigData, result *ValidationResult) {
	for name, scope := range config.Scopes {
		label := scopeLabel(name, scope.Description)
		if scope.Paths == nil {
			result.AddIssue(LevelError, "", config.ConfigPath,
				fmt.Sprintf("%s is missing required field: paths", label))
		} else if len(scope.Paths) == 0 {
			result.AddIssue(LevelError, "", config.ConfigPath,
				fmt.Sprintf("%s has empty paths array", label))
		}
	}
}

// scopeLabel returns a formatted label like "scope 'name' (description)" or "scope 'name'".
func scopeLabel(name, description string) string {
	if description != "" {
		return fmt.Sprintf("scope '%s' (%s)", name, description)
	}
	return fmt.Sprintf("scope '%s'", name)
}

var knownConfigKeys = map[string]bool{
	"dir":            true,
	"task-dir":       true,
	"web":            true,
	"scopes":         true,
	"sync":           true,
	"ignore":         true,
	"workflow":       true,
	"todos":          true,
	"id":             true,
	"phases":         true,
	"effort":         true,
	"worklogs":       true,
	"worktree_scope": true,
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

// checkWorkflowValue validates the workflow config value.
func (v *Validator) checkWorkflowValue(config *ConfigData, result *ValidationResult) {
	if config.Workflow == "" {
		return
	}
	valid := map[string]bool{"solo": true, "pr-review": true}
	if !valid[config.Workflow] {
		result.AddIssue(LevelError, "", config.ConfigPath,
			fmt.Sprintf("invalid workflow value: '%s' (valid values: solo, pr-review)", config.Workflow))
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
					fmt.Sprintf("touches references undefined scope: '%s'. Add it to the scopes list in .taskmd.yaml", scope))
			}
		}
	}

	return result
}

// ValidatePhasesAgainstConfig warns when tasks reference phases not defined in config.
// Skips validation if knownPhases is nil or empty.
func (v *Validator) ValidatePhasesAgainstConfig(tasks []*model.Task, knownPhases map[string]bool) *ValidationResult {
	result := &ValidationResult{
		Issues: make([]ValidationIssue, 0),
	}
	if len(knownPhases) == 0 {
		return result
	}

	reported := make(map[string]bool)
	for _, task := range tasks {
		if task.Phase == "" {
			continue
		}
		if !knownPhases[task.Phase] && !reported[task.Phase] {
			reported[task.Phase] = true
			result.AddIssue(LevelWarning, task.ID, task.FilePath,
				fmt.Sprintf("phase references undefined phase: '%s'. Add it to the phases list in .taskmd.yaml", task.Phase))
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
