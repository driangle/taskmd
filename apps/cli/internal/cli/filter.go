package cli

import (
	"github.com/driangle/taskmd/apps/cli/internal/filter"
	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// filterCriteria represents a single filter condition (kept for test compat).
type filterCriteria = filter.Criteria

// applyFilters delegates to the shared filter package.
func applyFilters(tasks []*model.Task, filterExprs []string) ([]*model.Task, error) {
	return filter.Apply(tasks, filterExprs)
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
	result, err := filter.Apply([]*model.Task{task}, []string{field + "=" + value})
	if err != nil {
		return false
	}
	return len(result) == 1
}
