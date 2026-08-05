package taskfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/driangle/taskmd/sdk/go/lock"
	"github.com/driangle/taskmd/sdk/go/model"
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

var validEfforts = map[string]bool{
	string(model.EffortSmall):  true,
	string(model.EffortMedium): true,
	string(model.EffortLarge):  true,
}

var validTypes = map[string]bool{
	string(model.TypeFeature):     true,
	string(model.TypeBug):         true,
	string(model.TypeImprovement): true,
	string(model.TypeChore):       true,
	string(model.TypeDocs):        true,
}

// ValidateUpdateRequest checks enum fields and returns a list of error strings.
func ValidateUpdateRequest(req UpdateRequest) []string {
	var errs []string
	if req.Status != nil && !validStatuses[*req.Status] {
		errs = append(errs, fmt.Sprintf("invalid status: %q", *req.Status))
	}
	if req.Priority != nil && !validPriorities[*req.Priority] {
		errs = append(errs, fmt.Sprintf("invalid priority: %q", *req.Priority))
	}
	if req.Effort != nil && !validEfforts[*req.Effort] {
		errs = append(errs, fmt.Sprintf("invalid effort: %q", *req.Effort))
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

	return writeFileAtomic(filePath, []byte(strings.Join(lines, "\n")), 0o644)
}

// UpdateTaskFileLocked acquires the per-task lock for taskID (scoped to scanDir)
// and then performs UpdateTaskFile within the critical section. This serializes
// concurrent writers to the same task across processes (parallel agents, the
// web server, and CLI invocations) sharing a task directory, preventing lost
// updates.
func UpdateTaskFileLocked(scanDir, filePath, taskID string, req UpdateRequest) error {
	l, err := lock.Acquire(lock.TaskLockPath(scanDir, taskID), lock.DefaultTimeout)
	if err != nil {
		return fmt.Errorf("lock task %s: %w", taskID, err)
	}
	defer l.Release()
	return UpdateTaskFile(filePath, req)
}

// writeFileAtomic writes data to path atomically: it writes to a temporary file
// in the same directory, fsyncs it, then renames it over the target. Readers
// therefore never observe a partially written file — they see either the old
// contents or the complete new contents.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-"+filepath.Base(path)+"-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
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
	for i := openIdx + 1; i < closeIdx; i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "tags:") {
			continue
		}
		if strings.Contains(lines[i], "[") {
			return parseInlineTags(strings.TrimSpace(lines[i]))
		}
		return parseMultilineTags(lines, i+1, closeIdx)
	}
	return nil
}

func parseInlineTags(line string) []string {
	inner := line[strings.Index(line, "[")+1 : strings.LastIndex(line, "]")]
	if strings.TrimSpace(inner) == "" {
		return nil
	}
	parts := strings.Split(inner, ",")
	var tags []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"'`)
		if p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

func parseMultilineTags(lines []string, start, closeIdx int) []string {
	var tags []string
	for j := start; j < closeIdx; j++ {
		lt := strings.TrimSpace(lines[j])
		if strings.HasPrefix(lt, "- ") {
			tags = append(tags, strings.TrimPrefix(lt, "- "))
		} else {
			break
		}
	}
	return tags
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
func applyTagUpdates(lines []string, openIdx, closeIdx int, _ []string, newTags []string) ([]string, int) {
	tagsLineIdx := -1
	for i := openIdx + 1; i < closeIdx; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "tags:") {
			tagsLineIdx = i
			break
		}
	}

	if tagsLineIdx < 0 {
		tagLine := FormatInlineTags(newTags)
		lines = insertLine(lines, closeIdx, tagLine)
		closeIdx++
		return lines, closeIdx
	}

	// Detect inline vs multiline format.
	if strings.Contains(lines[tagsLineIdx], "[") {
		lines[tagsLineIdx] = FormatInlineTags(newTags)
		return lines, closeIdx
	}

	// Multiline format
	removeStart := tagsLineIdx + 1
	removeEnd := removeStart
	for removeEnd < closeIdx && strings.HasPrefix(strings.TrimSpace(lines[removeEnd]), "- ") {
		removeEnd++
	}

	var newTagLines []string
	for _, t := range newTags {
		newTagLines = append(newTagLines, "  - "+t)
	}

	before := lines[:removeStart]
	after := lines[removeEnd:]
	result := make([]string, 0, len(before)+len(newTagLines)+len(after))
	result = append(result, before...)
	result = append(result, newTagLines...)
	result = append(result, after...)

	closeIdx += len(newTagLines) - (removeEnd - removeStart)
	return result, closeIdx
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
	prefix := fieldName + ":"
	for i := openIdx + 1; i < closeIdx; i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
			continue
		}
		if strings.Contains(lines[i], "[") {
			return parseInlineTags(strings.TrimSpace(lines[i]))
		}
		return parseMultilineTags(lines, i+1, closeIdx)
	}
	return nil
}

// applyListFieldUpdates modifies the lines slice to reflect the new list values for a named field.
func applyListFieldUpdates(lines []string, openIdx, closeIdx int, fieldName string, newValues []string) ([]string, int) {
	prefix := fieldName + ":"
	lineIdx := -1
	for i := openIdx + 1; i < closeIdx; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
			lineIdx = i
			break
		}
	}

	if lineIdx < 0 {
		if len(newValues) == 0 {
			return lines, closeIdx
		}
		newLine := FormatInlineList(fieldName, newValues)
		lines = insertLine(lines, closeIdx, newLine)
		closeIdx++
		return lines, closeIdx
	}

	// Inline format
	if strings.Contains(lines[lineIdx], "[") {
		if len(newValues) == 0 {
			// Remove the field line entirely
			lines = append(lines[:lineIdx], lines[lineIdx+1:]...)
			closeIdx--
		} else {
			lines[lineIdx] = FormatInlineList(fieldName, newValues)
		}
		return lines, closeIdx
	}

	// Multiline format
	removeStart := lineIdx + 1
	removeEnd := removeStart
	for removeEnd < closeIdx && strings.HasPrefix(strings.TrimSpace(lines[removeEnd]), "- ") {
		removeEnd++
	}

	if len(newValues) == 0 {
		// Remove field key line and all item lines
		lines = append(lines[:lineIdx], lines[removeEnd:]...)
		closeIdx -= removeEnd - lineIdx
		return lines, closeIdx
	}

	var newItemLines []string
	for _, v := range newValues {
		newItemLines = append(newItemLines, "  - "+v)
	}

	before := lines[:removeStart]
	after := lines[removeEnd:]
	result := make([]string, 0, len(before)+len(newItemLines)+len(after))
	result = append(result, before...)
	result = append(result, newItemLines...)
	result = append(result, after...)

	closeIdx += len(newItemLines) - (removeEnd - removeStart)
	return result, closeIdx
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
			return writeFileAtomic(filePath, []byte(strings.Join(lines, "\n")), 0o644)
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

	return writeFileAtomic(filePath, []byte(strings.Join(lines, "\n")), 0o644)
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
