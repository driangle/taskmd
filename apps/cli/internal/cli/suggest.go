package cli

import (
	"fmt"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/taskfile"
)

// validStatusValues lists all valid task status values.
var validStatusValues = []string{"pending", "in-progress", "completed", "blocked", "cancelled"}

// validPriorityValues lists all valid priority values.
var validPriorityValues = []string{"low", "medium", "high", "critical"}

// validEffortValues lists all valid effort values.
var validEffortValues = []string{"small", "medium", "large"}

// validSortFields lists all valid sort field values for the list command.
var validSortFields = []string{"id", "title", "status", "priority", "effort", "created"}

// suggestValue finds the closest match from valid options using Levenshtein distance.
// Returns the closest match, or empty string if no reasonable match exists.
func suggestValue(input string, valid []string) string {
	input = strings.ToLower(input)
	best := ""
	bestDist := len(input)/2 + 1 // threshold: must be within half the input length

	for _, v := range valid {
		d := levenshtein(input, strings.ToLower(v))
		if d < bestDist {
			bestDist = d
			best = v
		}
	}
	return best
}

// invalidValueError creates an error message with an optional suggestion for the closest valid value.
func invalidValueError(field, value string, valid []string) error {
	msg := fmt.Sprintf("invalid %s: %q (valid: %s)", field, value, strings.Join(valid, ", "))
	if suggestion := suggestValue(value, valid); suggestion != "" {
		msg += fmt.Sprintf("; did you mean %q?", suggestion)
	}
	return fmt.Errorf("%s", msg)
}

// validateSetEnums validates enum fields in the set command's UpdateRequest, providing suggestions on error.
func validateSetEnums(req taskfile.UpdateRequest) error {
	if req.Status != nil && !contains(validStatusValues, *req.Status) {
		return invalidValueError("status", *req.Status, validStatusValues)
	}
	if req.Priority != nil && !contains(validPriorityValues, *req.Priority) {
		return invalidValueError("priority", *req.Priority, validPriorityValues)
	}
	if req.Effort != nil && !contains(validEffortValues, *req.Effort) {
		return invalidValueError("effort", *req.Effort, validEffortValues)
	}
	return nil
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
