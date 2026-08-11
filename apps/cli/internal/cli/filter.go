package cli

import (
	"fmt"

	"github.com/driangle/taskmd/sdk/go/filter"
	"github.com/driangle/taskmd/sdk/go/model"
)

// filterCriteria represents a single filter condition (kept for test compat).
type filterCriteria = filter.Criteria

// applyFilters delegates to the shared filter package, supplying the project's
// configured effort vocabulary so effort comparisons accept its values.
func applyFilters(tasks []*model.Task, filterExprs []string) ([]*model.Task, error) {
	return filter.Apply(tasks, filterExprs, resolveEffortScale())
}

// FilterShortcuts holds the common shortcut filter parameters shared across commands.
type FilterShortcuts struct {
	Status   string
	Priority string
	Phase    string
	Scope    string
	Filters  []string
}

// applyShortcutFilters expands shortcut flags into filter expressions and applies
// all filters, scope, and phase filtering. Used by both list and graph commands.
func applyShortcutFilters(tasks []*model.Task, s FilterShortcuts) ([]*model.Task, error) {
	filters := append([]string{}, s.Filters...)
	if s.Status != "" {
		filters = append(filters, "status="+s.Status)
	}
	if s.Priority != "" {
		filters = append(filters, "priority="+s.Priority)
	}
	if s.Phase != "" {
		filters = append(filters, "phase="+s.Phase)
	}

	if len(filters) > 0 {
		var err error
		tasks, err = applyFilters(tasks, filters)
		if err != nil {
			return nil, fmt.Errorf("filter error: %w", err)
		}
	}

	if s.Scope != "" {
		warnUnknownScope(s.Scope)
		tasks = filterTasksByScope(tasks, s.Scope)
	}

	return tasks, nil
}

// matchesAllFilters is kept for backward-compatible tests.
func matchesAllFilters(task *model.Task, filters []filterCriteria) bool {
	for _, f := range filters {
		if !matchesFilter(task, f.Field, f.Value) {
			return false
		}
	}
	return true
}

// matchesFilter is kept for backward-compatible tests.
func matchesFilter(task *model.Task, field, value string) bool {
	// Test a single-filter list via the shared package
	result, err := filter.Apply([]*model.Task{task}, []string{field + "=" + value}, resolveEffortScale())
	if err != nil {
		return false
	}
	return len(result) == 1
}
