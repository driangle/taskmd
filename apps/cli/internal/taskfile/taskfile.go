package taskfile

import (
	"fmt"
	"os"
	"strings"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// UpdateRequest describes which fields to update. Nil pointer means "no change".
type UpdateRequest struct {
	Title    *string
	Status   *string
	Priority *string
	Effort   *string
	Tags     *[]string // replace tags entirely
	AddTags  []string  // add to existing tags
	RemTags  []string  // remove from existing tags
	Body     *string
}

var validStatuses = map[string]bool{
	string(model.StatusPending):    true,
	string(model.StatusInProgress): true,
	string(model.StatusCompleted):  true,
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
	scalarUpdates := buildScalarUpdates(req)
	for i := openIdx + 1; i < closeIdx; i++ {
		for _, u := range scalarUpdates {
			prefix := u.key + ":"
			if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
				lines[i] = u.key + ": " + u.value
				break
			}
		}
	}

	// Apply tag updates.
	if req.Tags != nil {
		lines, closeIdx = setTags(lines, openIdx, closeIdx, *req.Tags)
	} else if len(req.AddTags) > 0 || len(req.RemTags) > 0 {
		currentTags := parseCurrentTags(lines, openIdx, closeIdx)
		newTags := ComputeNewTags(currentTags, req.AddTags, req.RemTags)
		lines, closeIdx = applyTagUpdates(lines, openIdx, closeIdx, currentTags, newTags)
	}

	// Apply body update — replace everything after closing ---.
	if req.Body != nil {
		lines = replaceBody(lines, closeIdx, *req.Body)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
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
	return updates
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

func insertLine(lines []string, idx int, line string) []string {
	lines = append(lines, "")
	copy(lines[idx+1:], lines[idx:])
	lines[idx] = line
	return lines
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
