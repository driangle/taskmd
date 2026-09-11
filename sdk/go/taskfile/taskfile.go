package taskfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
	"gopkg.in/yaml.v3"
)

// UpdateRequest describes which fields to update. Nil pointer means "no change".
type UpdateRequest struct {
	Title        *string
	Status       *string
	Priority     *string
	Effort       *string
	Type         *string
	Owner        *string
	Parent       *string
	Tags         *[]string // replace tags entirely
	AddTags      []string  // add to existing tags
	RemTags      []string  // remove from existing tags
	AddPRs       []string  // add PR URLs
	RemPRs       []string  // remove PR URLs
	AddTouches   []string  // add scope identifiers to touches
	RemTouches   []string  // remove scope identifiers from touches
	Phase        *string
	Completed    *string   // completed date (YYYY-MM-DD)
	CancelledAt  *string   // cancelled date (YYYY-MM-DD)
	Dependencies *[]string // replace dependencies entirely
	RemoveFields []string  // field keys to remove entirely from frontmatter
	Body         *string
}

var validStatuses = map[string]bool{
	string(model.StatusPending):    true,
	string(model.StatusInProgress): true,
	string(model.StatusCompleted):  true,
	string(model.StatusInReview):   true,
	string(model.StatusBlocked):    true,
	string(model.StatusCancelled):  true,
}

var validPriorities = map[string]bool{
	string(model.PriorityLow):      true,
	string(model.PriorityMedium):   true,
	string(model.PriorityHigh):     true,
	string(model.PriorityCritical): true,
}

var validTypes = map[string]bool{
	string(model.TypeFeature):     true,
	string(model.TypeBug):         true,
	string(model.TypeImprovement): true,
	string(model.TypeChore):       true,
	string(model.TypeDocs):        true,
}

// ValidateUpdateRequest checks enum fields and returns a list of error strings.
//
// efforts supplies the project's effort vocabulary; pass effort.Scale{} for the
// default small, medium, large.
func ValidateUpdateRequest(req UpdateRequest, efforts effort.Scale) []string {
	var errs []string
	if req.Status != nil && !validStatuses[*req.Status] {
		errs = append(errs, fmt.Sprintf("invalid status: %q", *req.Status))
	}
	if req.Priority != nil && !validPriorities[*req.Priority] {
		errs = append(errs, fmt.Sprintf("invalid priority: %q", *req.Priority))
	}
	if req.Effort != nil && !efforts.Contains(*req.Effort) {
		errs = append(errs, fmt.Sprintf("invalid effort: %q (valid values: %s)", *req.Effort, efforts))
	}
	if req.Type != nil && !validTypes[*req.Type] {
		errs = append(errs, fmt.Sprintf("invalid type: %q", *req.Type))
	}
	return errs
}

// UpdateTaskFile reads a task markdown file, applies the requested changes, and writes it back.
func UpdateTaskFile(filePath string, req UpdateRequest) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read task file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	openIdx, closeIdx := FindFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return fmt.Errorf("task file has no valid frontmatter: %s", filePath)
	}

	// Apply scalar field updates within frontmatter.
	lines, closeIdx = applyScalarUpdates(lines, openIdx, closeIdx, req)

	// Remove fields requested for deletion.
	lines, closeIdx = removeFields(lines, openIdx, closeIdx, req.RemoveFields)

	// Apply tag updates.
	if req.Tags != nil {
		lines, closeIdx = setTags(lines, openIdx, closeIdx, *req.Tags)
	} else if len(req.AddTags) > 0 || len(req.RemTags) > 0 {
		currentTags := parseCurrentTags(lines, openIdx, closeIdx)
		newTags := ComputeNewTags(currentTags, req.AddTags, req.RemTags)
		lines, closeIdx = applyTagUpdates(lines, openIdx, closeIdx, currentTags, newTags)
	}

	// Apply PR updates.
	if len(req.AddPRs) > 0 || len(req.RemPRs) > 0 {
		currentPRs := parseCurrentListField(lines, openIdx, closeIdx, "pr")
		newPRs := ComputeNewTags(currentPRs, req.AddPRs, req.RemPRs)
		lines, closeIdx = applyListFieldUpdates(lines, openIdx, closeIdx, "pr", newPRs)
	}

	// Apply touches updates.
	if len(req.AddTouches) > 0 || len(req.RemTouches) > 0 {
		currentTouches := parseCurrentListField(lines, openIdx, closeIdx, "touches")
		newTouches := ComputeNewTags(currentTouches, req.AddTouches, req.RemTouches)
		lines, closeIdx = applyListFieldUpdates(lines, openIdx, closeIdx, "touches", newTouches)
	}

	// Apply dependency updates.
	if req.Dependencies != nil {
		lines, closeIdx = applyListFieldUpdates(lines, openIdx, closeIdx, "dependencies", *req.Dependencies)
	}

	// Apply body update — replace everything after closing ---.
	if req.Body != nil {
		lines = replaceBody(lines, closeIdx, *req.Body)
	}

	out := strings.Join(lines, "\n")

	// Backstop: never leave a task file taskmd cannot read back. A writer bug
	// that corrupts frontmatter would otherwise make the task vanish silently
	// from every view, with no error to explain it.
	if err := verifyFrontmatter(out); err != nil {
		return fmt.Errorf("refusing to write %s: update would produce unreadable frontmatter (%w) — this is a taskmd bug, please report it", filePath, err)
	}

	return os.WriteFile(filePath, []byte(out), 0644)
}

// verifyFrontmatter checks that the frontmatter of the about-to-be-written
// content still parses as a YAML mapping.
func verifyFrontmatter(content string) error {
	lines := strings.Split(content, "\n")
	openIdx, closeIdx := FindFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return fmt.Errorf("frontmatter delimiters missing after update")
	}

	var fm map[string]any
	return yaml.Unmarshal([]byte(strings.Join(lines[openIdx+1:closeIdx], "\n")), &fm)
}

type scalarUpdate struct {
	key   string
	value string
}

func buildScalarUpdates(req UpdateRequest) []scalarUpdate {
	var updates []scalarUpdate
	if req.Title != nil {
		updates = append(updates, scalarUpdate{key: "title", value: fmt.Sprintf("%q", *req.Title)})
	}
	if req.Status != nil {
		updates = append(updates, scalarUpdate{key: "status", value: *req.Status})
	}
	if req.Priority != nil {
		updates = append(updates, scalarUpdate{key: "priority", value: *req.Priority})
	}
	if req.Effort != nil {
		updates = append(updates, scalarUpdate{key: "effort", value: *req.Effort})
	}
	if req.Type != nil {
		updates = append(updates, scalarUpdate{key: "type", value: *req.Type})
	}
	if req.Owner != nil {
		updates = append(updates, scalarUpdate{key: "owner", value: *req.Owner})
	}
	if req.Parent != nil {
		updates = append(updates, scalarUpdate{key: "parent", value: *req.Parent})
	}
	if req.Phase != nil {
		updates = append(updates, scalarUpdate{key: "phase", value: *req.Phase})
	}
	if req.Completed != nil {
		updates = append(updates, scalarUpdate{key: "completed_at", value: *req.Completed})
	}
	if req.CancelledAt != nil {
		updates = append(updates, scalarUpdate{key: "cancelled_at", value: *req.CancelledAt})
	}
	return updates
}

// applyScalarUpdates updates or inserts scalar frontmatter fields.
func applyScalarUpdates(lines []string, openIdx, closeIdx int, req UpdateRequest) ([]string, int) {
	scalarUpdates := buildScalarUpdates(req)
	found := make([]bool, len(scalarUpdates))
	for i := openIdx + 1; i < closeIdx; i++ {
		for j, u := range scalarUpdates {
			prefix := u.key + ":"
			if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
				lines[i] = u.key + ": " + u.value
				found[j] = true
				break
			}
		}
	}

	// Insert any scalar fields that weren't found in existing frontmatter.
	for j := len(scalarUpdates) - 1; j >= 0; j-- {
		if !found[j] {
			u := scalarUpdates[j]
			lines = insertLine(lines, closeIdx, u.key+": "+u.value)
			closeIdx++
		}
	}

	return lines, closeIdx
}

// removeFields removes frontmatter lines whose key matches one of the given field names.
func removeFields(lines []string, openIdx, closeIdx int, fields []string) ([]string, int) {
	if len(fields) == 0 {
		return lines, closeIdx
	}
	removeSet := make(map[string]bool, len(fields))
	for _, f := range fields {
		removeSet[f] = true
	}
	for i := closeIdx - 1; i > openIdx; i-- {
		trimmed := strings.TrimSpace(lines[i])
		for key := range removeSet {
			if strings.HasPrefix(trimmed, key+":") {
				lines = append(lines[:i], lines[i+1:]...)
				closeIdx--
				break
			}
		}
	}
	return lines, closeIdx
}

// replaceBody replaces all content after the closing frontmatter delimiter.
func replaceBody(lines []string, closeIdx int, newBody string) []string {
	// Keep frontmatter lines including closing ---
	result := make([]string, closeIdx+1)
	copy(result, lines[:closeIdx+1])

	// Add blank line then new body
	if newBody != "" {
		result = append(result, "")
		result = append(result, strings.Split(newBody, "\n")...)
	}

	// Ensure file ends with newline
	if len(result) > 0 && result[len(result)-1] != "" {
		result = append(result, "")
	}

	return result
}

// parseCurrentTags reads the existing tags from frontmatter lines.
func parseCurrentTags(lines []string, openIdx, closeIdx int) []string {
	f, found := findListField(lines, openIdx, closeIdx, "tags")
	if !found {
		return nil
	}
	return f.values
}

// setTags replaces tags entirely with the given list.
func setTags(lines []string, openIdx, closeIdx int, tags []string) ([]string, int) {
	return applyTagUpdates(lines, openIdx, closeIdx, nil, tags)
}

// ComputeNewTags computes the resulting tag list after additions and removals.
func ComputeNewTags(current, addTags, removeTags []string) []string {
	removeSet := make(map[string]bool, len(removeTags))
	for _, t := range removeTags {
		removeSet[t] = true
	}

	var result []string
	seen := make(map[string]bool)
	for _, t := range current {
		if !removeSet[t] {
			result = append(result, t)
			seen[t] = true
		}
	}

	for _, t := range addTags {
		if !seen[t] {
			result = append(result, t)
			seen[t] = true
		}
	}

	return result
}

// applyTagUpdates modifies the lines slice to reflect the new tags.
//
// Unlike the other list fields, an empty result keeps the key as "tags: []"
// rather than removing it.
func applyTagUpdates(lines []string, openIdx, closeIdx int, _ []string, newTags []string) ([]string, int) {
	f, found := findListField(lines, openIdx, closeIdx, "tags")
	if !found {
		lines = insertLine(lines, closeIdx, FormatInlineTags(newTags))
		return lines, closeIdx + 1
	}
	return replaceListField(lines, closeIdx, f, listFieldLines("tags", f.shape, newTags, keepEmptyKey))
}

// FormatInlineTags formats tags as inline YAML: tags: ["a", "b"]
func FormatInlineTags(tags []string) string {
	if len(tags) == 0 {
		return "tags: []"
	}
	quoted := make([]string, len(tags))
	for i, t := range tags {
		quoted[i] = `"` + t + `"`
	}
	return "tags: [" + strings.Join(quoted, ", ") + "]"
}

// parseCurrentListField reads an inline YAML list field (e.g. pr: ["a", "b"]) from frontmatter.
func parseCurrentListField(lines []string, openIdx, closeIdx int, fieldName string) []string {
	f, found := findListField(lines, openIdx, closeIdx, fieldName)
	if !found {
		return nil
	}
	return f.values
}

// applyListFieldUpdates modifies the lines slice to reflect the new list values for a named field.
func applyListFieldUpdates(lines []string, openIdx, closeIdx int, fieldName string, newValues []string) ([]string, int) {
	f, found := findListField(lines, openIdx, closeIdx, fieldName)
	if !found {
		if len(newValues) == 0 {
			return lines, closeIdx
		}
		lines = insertLine(lines, closeIdx, FormatInlineList(fieldName, newValues))
		return lines, closeIdx + 1
	}
	return replaceListField(lines, closeIdx, f, listFieldLines(fieldName, f.shape, newValues, removeKey))
}

// FormatInlineList formats a named list field as inline YAML: field: ["a", "b"]
func FormatInlineList(fieldName string, values []string) string {
	if len(values) == 0 {
		return fieldName + ": []"
	}
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = `"` + v + `"`
	}
	return fieldName + ": [" + strings.Join(quoted, ", ") + "]"
}

func insertLine(lines []string, idx int, line string) []string {
	lines = append(lines, "")
	copy(lines[idx+1:], lines[idx:])
	lines[idx] = line
	return lines
}

// ReplaceID rewrites the id field in a task file's frontmatter.
func ReplaceID(filePath, newID string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	openIdx, closeIdx := FindFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return fmt.Errorf("no valid frontmatter in %s", filePath)
	}

	for i := openIdx + 1; i < closeIdx; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "id:") {
			lines[i] = fmt.Sprintf("id: %q", newID)
			return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
		}
	}

	return fmt.Errorf("id field not found in frontmatter of %s", filePath)
}

// referenceFields are the frontmatter field prefixes where task ID cross-references appear.
var referenceFields = []string{"parent:", "dependencies:"}

// ReplaceReference replaces occurrences of oldID with newID in dependency and parent
// fields within a task file's frontmatter.
func ReplaceReference(filePath, oldID, newID string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	openIdx, closeIdx := FindFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return fmt.Errorf("no valid frontmatter in %s", filePath)
	}

	changed := replaceRefsInLines(lines, openIdx, closeIdx, oldID, newID)
	if !changed {
		return nil
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

// replaceRefsInLines performs the actual replacement of oldID with newID within frontmatter lines.
func replaceRefsInLines(lines []string, openIdx, closeIdx int, oldID, newID string) bool {
	changed := false
	inDeps := false

	for i := openIdx + 1; i < closeIdx; i++ {
		trimmed := strings.TrimSpace(lines[i])

		// Check if this line is a reference field (parent: or dependencies:).
		if isReferenceField(trimmed) {
			inDeps = strings.HasPrefix(trimmed, "dependencies:")
			if strings.Contains(lines[i], oldID) {
				lines[i] = strings.Replace(lines[i], oldID, newID, 1)
				changed = true
			}
			continue
		}

		// Handle multiline dependency items.
		if inDeps && strings.HasPrefix(trimmed, "- ") {
			if strings.Contains(lines[i], oldID) {
				lines[i] = strings.Replace(lines[i], oldID, newID, 1)
				changed = true
			}
			continue
		}

		inDeps = false
	}

	return changed
}

// isReferenceField checks if a trimmed line starts with a known reference field prefix.
func isReferenceField(trimmed string) bool {
	for _, prefix := range referenceFields {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// FindFrontmatterBounds returns the line indices of the opening and closing "---" delimiters.
func FindFrontmatterBounds(lines []string) (int, int) {
	openIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if openIdx < 0 {
				openIdx = i
			} else {
				return openIdx, i
			}
		}
	}
	return -1, -1
}
