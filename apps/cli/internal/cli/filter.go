package cli

import (
	"fmt"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// filterCriteria represents a single filter condition
type filterCriteria struct {
	field string
	value string
}

// applyFilters applies multiple filter expressions to tasks (AND logic)
func applyFilters(tasks []*model.Task, filterExprs []string) ([]*model.Task, error) {
	// Parse all filter expressions first
	filters := make([]filterCriteria, 0, len(filterExprs))
	for _, expr := range filterExprs {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter format (expected field=value): %s", expr)
		}
		filters = append(filters, filterCriteria{
			field: strings.TrimSpace(parts[0]),
			value: strings.TrimSpace(parts[1]),
		})
	}

	// Apply all filters (AND logic)
	var filtered []*model.Task
	for _, task := range tasks {
		if matchesAllFilters(task, filters) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

// matchesAllFilters checks if a task matches all filter criteria (AND logic)
func matchesAllFilters(task *model.Task, filters []filterCriteria) bool {
	for _, filter := range filters {
		if !matchesFilter(task, filter.field, filter.value) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a task matches a filter criterion
func matchesFilter(task *model.Task, field, value string) bool {
	switch field {
	case "status":
		return string(task.Status) == value
	case "priority":
		return string(task.Priority) == value
	case "effort":
		return string(task.Effort) == value
	case "id":
		return task.ID == value
	case "title":
		return strings.Contains(strings.ToLower(task.Title), strings.ToLower(value))
	case "group":
		return task.Group == value
	case "blocked":
		// A task is blocked if it has dependencies
		isBlocked := len(task.Dependencies) > 0
		return (value == "true" && isBlocked) || (value == "false" && !isBlocked)
	case "tag":
		// Check if task has the specified tag
		for _, tag := range task.Tags {
			if tag == value {
				return true
			}
		}
		return false
	default:
		return false
	}
}
