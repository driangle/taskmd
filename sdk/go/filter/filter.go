package filter

import (
	"fmt"
	"slices"
	"strings"

	"github.com/driangle/taskmd/sdk/go/model"
)

// Criteria represents a single filter condition.
type Criteria struct {
	Field string
	Op    string // "=", ">", ">=", "<", "<="
	Value string
}

// ordinalFields maps field names to their ordered values (lowest to highest).
var ordinalFields = map[string][]string{
	"priority": {"low", "medium", "high", "critical"},
	"effort":   {"small", "medium", "large"},
}

// Presence sentinel values recognized uniformly across equality filters:
//   - "none" (and an empty right-hand side) match tasks whose field is unset.
//   - "any" matches tasks whose field is set.
//
// These are reserved: a field cannot be matched against a literal "none"/"any".
const (
	sentinelNone = "none"
	sentinelAny  = "any"
)

// Apply applies multiple filter expressions to tasks (AND logic).
func Apply(tasks []*model.Task, filterExprs []string) ([]*model.Task, error) {
	filters := make([]Criteria, 0, len(filterExprs))
	for _, expr := range filterExprs {
		c, err := parseExpr(expr)
		if err != nil {
			return nil, err
		}
		filters = append(filters, c)
	}

	var filtered []*model.Task
	for _, task := range tasks {
		if matchesAll(task, filters) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

// parseExpr parses a filter expression like "field=value", "field>=value", etc.
func parseExpr(expr string) (Criteria, error) {
	// Check for two-char operators first, then single-char.
	for _, op := range []string{">=", "<=", ">", "<"} {
		idx := strings.Index(expr, op)
		if idx > 0 {
			field := strings.TrimSpace(expr[:idx])
			value := strings.TrimSpace(expr[idx+len(op):])
			if value == "" {
				return Criteria{}, fmt.Errorf("invalid filter format (missing value): %s", expr)
			}
			ranks, ok := ordinalFields[field]
			if !ok {
				return Criteria{}, fmt.Errorf("operator %q is not supported for field %q (only priority and effort support ordering)", op, field)
			}
			if !slices.Contains(ranks, value) {
				return Criteria{}, fmt.Errorf("unknown %s value %q (valid: %s)", field, value, strings.Join(ranks, ", "))
			}
			return Criteria{Field: field, Op: op, Value: value}, nil
		}
	}

	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return Criteria{}, fmt.Errorf("invalid filter format (expected field=value): %s", expr)
	}
	return Criteria{
		Field: strings.TrimSpace(parts[0]),
		Op:    "=",
		Value: strings.TrimSpace(parts[1]),
	}, nil
}

func matchesAll(task *model.Task, filters []Criteria) bool {
	for _, f := range filters {
		if !matchesCriteria(task, f) {
			return false
		}
	}
	return true
}

func matchesCriteria(task *model.Task, c Criteria) bool {
	if c.Op != "=" {
		return matchesOrdinal(task, c)
	}
	return matchesEquality(task, c.Field, c.Value)
}

// matchesOrdinal handles >, >=, <, <= for ordinal fields.
func matchesOrdinal(task *model.Task, c Criteria) bool {
	v, ok := getFieldValue(task, c.Field)
	if !ok || v == "" {
		return false
	}
	ranks := ordinalFields[c.Field]
	taskRank := slices.Index(ranks, v)
	filterRank := slices.Index(ranks, c.Value)
	if taskRank < 0 || filterRank < 0 {
		return false
	}
	switch c.Op {
	case ">":
		return taskRank > filterRank
	case ">=":
		return taskRank >= filterRank
	case "<":
		return taskRank < filterRank
	case "<=":
		return taskRank <= filterRank
	default:
		return false
	}
}

func matchesEquality(task *model.Task, field, value string) bool {
	if v, ok := getFieldValue(task, field); ok {
		if matched, isSentinel := matchPresence(v != "", value); isSentinel {
			return matched
		}
		if field == "group" && strings.Contains(value, "*") {
			return MatchScope(value, v)
		}
		return v == value
	}
	switch field {
	case "title":
		return strings.Contains(strings.ToLower(task.Title), strings.ToLower(value))
	case "blocked":
		return matchBoolPresence(len(task.Dependencies) > 0, value)
	case "tag":
		return slices.Contains(task.Tags, value)
	case "touches":
		return matchesTouches(task.Touches, value)
	case "parent":
		return matchBoolOrValue(task.Parent, value)
	default:
		return false
	}
}

func matchesTouches(touches []string, value string) bool {
	if strings.Contains(value, "*") {
		for _, t := range touches {
			if MatchScope(value, t) {
				return true
			}
		}
		return false
	}
	return slices.Contains(touches, value)
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
	case "type":
		return string(task.Type), true
	case "id":
		return task.ID, true
	case "group":
		return task.Group, true
	case "owner":
		return task.Owner, true
	case "phase":
		return task.Phase, true
	default:
		return "", false
	}
}

// matchPresence evaluates the none/any presence sentinels against whether a
// field is set. An empty filter value is treated as an alias for "none". The
// isSentinel return is false when value is not a presence keyword, letting
// callers fall through to exact matching.
func matchPresence(present bool, value string) (matched, isSentinel bool) {
	switch value {
	case sentinelNone, "":
		return !present, true
	case sentinelAny:
		return present, true
	default:
		return false, false
	}
}

// matchBoolPresence matches presence sentinels plus the legacy true/false
// aliases (true→any, false→none) for fields whose only meaning is presence,
// such as blocked.
func matchBoolPresence(present bool, value string) bool {
	switch value {
	case sentinelAny, "true":
		return present
	case sentinelNone, "false", "":
		return !present
	default:
		return false
	}
}

// matchBoolOrValue matches presence sentinels and the legacy true/false aliases
// (true→any, false→none), falling back to an exact match against the field
// value. Used by parent, which carries both presence and an exact ID value.
func matchBoolOrValue(fieldValue, filterValue string) bool {
	switch filterValue {
	case sentinelAny, "true":
		return fieldValue != ""
	case sentinelNone, "false", "":
		return fieldValue == ""
	default:
		return fieldValue == filterValue
	}
}
