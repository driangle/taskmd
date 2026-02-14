package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// Criteria represents a single filter condition.
type Criteria struct {
	Field string
	Value string
}

// Apply applies multiple filter expressions to tasks (AND logic).
func Apply(tasks []*model.Task, filterExprs []string) ([]*model.Task, error) {
	filters := make([]Criteria, 0, len(filterExprs))
	for _, expr := range filterExprs {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter format (expected field=value): %s", expr)
		}
		filters = append(filters, Criteria{
			Field: strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		})
	}

	var filtered []*model.Task
	for _, task := range tasks {
		if matchesAll(task, filters) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func matchesAll(task *model.Task, filters []Criteria) bool {
	for _, f := range filters {
		if !matches(task, f.Field, f.Value) {
			return false
		}
	}
	return true
}

func matches(task *model.Task, field, value string) bool {
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
	case "owner":
		return task.Owner == value
	case "blocked":
		isBlocked := len(task.Dependencies) > 0
		return (value == "true" && isBlocked) || (value == "false" && !isBlocked)
	case "tag":
		return slices.Contains(task.Tags, value)
	default:
		return false
	}
}
