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
	if v, ok := getFieldValue(task, field); ok {
		return v == value
	}
	switch field {
	case "title":
		return strings.Contains(strings.ToLower(task.Title), strings.ToLower(value))
	case "blocked":
		isBlocked := len(task.Dependencies) > 0
		return (value == "true" && isBlocked) || (value == "false" && !isBlocked)
	case "tag":
		return slices.Contains(task.Tags, value)
	case "parent":
		return matchBoolOrValue(task.Parent, value)
	default:
		return false
	}
}

// getFieldValue returns the string value for simple equality fields.
func getFieldValue(task *model.Task, field string) (string, bool) {
	switch field {
	case "status":
		return string(task.Status), true
	case "priority":
		return string(task.Priority), true
	case "effort":
		return string(task.Effort), true
	case "id":
		return task.ID, true
	case "group":
		return task.Group, true
	case "owner":
		return task.Owner, true
	default:
		return "", false
	}
}

// matchBoolOrValue matches "true"/"false" as presence check, or exact value.
func matchBoolOrValue(fieldValue, filterValue string) bool {
	if filterValue == "true" {
		return fieldValue != ""
	}
	if filterValue == "false" {
		return fieldValue == ""
	}
	return fieldValue == filterValue
}
