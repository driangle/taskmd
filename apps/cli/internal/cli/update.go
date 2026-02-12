package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	updateTaskID     string
	updateStatus     string
	updatePriority   string
	updateEffort     string
	updateDone       bool
	updateAddTags    []string
	updateRemoveTags []string
)

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

var updateCmd = &cobra.Command{
	Use:        "update",
	SuggestFor: []string{"set", "edit", "modify", "change"},
	Short:      "Update a task's frontmatter fields",
	Long: `Update modifies frontmatter fields (status, priority, effort, tags) of a task file.

The task is identified by --task-id (exact match only).

Examples:
  taskmd update --task-id cli-049 --status completed
  taskmd update --task-id cli-049 --priority high --effort large
  taskmd update --task-id cli-049 --done
  taskmd update --task-id cli-049 --add-tag backend --add-tag api
  taskmd update --task-id cli-049 --remove-tag deprecated`,
	Args: cobra.NoArgs,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVar(&updateTaskID, "task-id", "", "task ID to update (required)")
	updateCmd.Flags().StringVar(&updateStatus, "status", "", "new status (pending, in-progress, completed, blocked, cancelled)")
	updateCmd.Flags().StringVar(&updatePriority, "priority", "", "new priority (low, medium, high, critical)")
	updateCmd.Flags().StringVar(&updateEffort, "effort", "", "new effort (small, medium, large)")
	updateCmd.Flags().BoolVar(&updateDone, "done", false, "mark task as completed (alias for --status completed)")
	updateCmd.Flags().StringArrayVar(&updateAddTags, "add-tag", nil, "add a tag (repeatable)")
	updateCmd.Flags().StringArrayVar(&updateRemoveTags, "remove-tag", nil, "remove a tag (repeatable)")

	_ = updateCmd.MarkFlagRequired("task-id")
}

// tagUpdate holds the add/remove tag lists for a tag modification.
type tagUpdate struct {
	addTags    []string
	removeTags []string
}

func runUpdate(cmd *cobra.Command, _ []string) error {
	updates, tags, err := validateUpdateFlags(cmd)
	if err != nil {
		return err
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	task := findExactMatch(updateTaskID, result.Tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", updateTaskID)
	}

	return applyUpdates(task, updates, tags)
}

// fieldUpdate represents a single frontmatter field change.
type fieldUpdate struct {
	key      string
	newValue string
}

func validateUpdateFlags(cmd *cobra.Command) ([]fieldUpdate, *tagUpdate, error) {
	if updateDone && cmd.Flags().Changed("status") {
		return nil, nil, fmt.Errorf("--done and --status are mutually exclusive")
	}

	if updateDone {
		updateStatus = string(model.StatusCompleted)
	}

	var updates []fieldUpdate

	if updateStatus != "" {
		if !validStatuses[updateStatus] {
			return nil, nil, fmt.Errorf("invalid status: %q (valid: pending, in-progress, completed, blocked, cancelled)", updateStatus)
		}
		updates = append(updates, fieldUpdate{key: "status", newValue: updateStatus})
	}

	if updatePriority != "" {
		if !validPriorities[updatePriority] {
			return nil, nil, fmt.Errorf("invalid priority: %q (valid: low, medium, high, critical)", updatePriority)
		}
		updates = append(updates, fieldUpdate{key: "priority", newValue: updatePriority})
	}

	if updateEffort != "" {
		if !validEfforts[updateEffort] {
			return nil, nil, fmt.Errorf("invalid effort: %q (valid: small, medium, large)", updateEffort)
		}
		updates = append(updates, fieldUpdate{key: "effort", newValue: updateEffort})
	}

	var tags *tagUpdate
	if len(updateAddTags) > 0 || len(updateRemoveTags) > 0 {
		tags = &tagUpdate{addTags: updateAddTags, removeTags: updateRemoveTags}
	}

	if len(updates) == 0 && tags == nil {
		return nil, nil, fmt.Errorf("nothing to update: provide --status, --priority, --effort, --done, --add-tag, or --remove-tag")
	}

	return updates, tags, nil
}

func applyUpdates(task *model.Task, updates []fieldUpdate, tags *tagUpdate) error {
	content, err := os.ReadFile(task.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read task file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	openIdx, closeIdx := findFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return fmt.Errorf("task file has no valid frontmatter: %s", task.FilePath)
	}

	changes := buildChangeLog(task, updates, tags)

	// Apply scalar field updates.
	for i := openIdx + 1; i < closeIdx; i++ {
		for j := range updates {
			prefix := updates[j].key + ":"
			if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
				lines[i] = updates[j].key + ": " + updates[j].newValue
				break
			}
		}
	}

	// Apply tag updates.
	if tags != nil {
		lines, _ = applyTagUpdates(lines, openIdx, closeIdx, task.Tags, tags)
	}

	if err := os.WriteFile(task.FilePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	printUpdateConfirmation(task, changes)
	return nil
}

// computeNewTags computes the resulting tag list after additions and removals.
func computeNewTags(current, addTags, removeTags []string) []string {
	removeSet := make(map[string]bool, len(removeTags))
	for _, t := range removeTags {
		removeSet[t] = true
	}

	// Start with current tags, filtering out removed ones.
	var result []string
	seen := make(map[string]bool)
	for _, t := range current {
		if !removeSet[t] {
			result = append(result, t)
			seen[t] = true
		}
	}

	// Add new tags (skip duplicates).
	for _, t := range addTags {
		if !seen[t] {
			result = append(result, t)
			seen[t] = true
		}
	}

	return result
}

// applyTagUpdates modifies the lines slice to reflect the new tags.
// Returns the updated lines and the new closeIdx (which may shift if lines are added/removed).
func applyTagUpdates(lines []string, openIdx, closeIdx int, currentTags []string, tags *tagUpdate) ([]string, int) {
	newTags := computeNewTags(currentTags, tags.addTags, tags.removeTags)

	// Find the tags: line within frontmatter.
	tagsLineIdx := -1
	for i := openIdx + 1; i < closeIdx; i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "tags:") {
			tagsLineIdx = i
			break
		}
	}

	if tagsLineIdx < 0 {
		// No tags: line exists — insert before closing ---.
		tagLine := formatInlineTags(newTags)
		lines = insertLine(lines, closeIdx, tagLine)
		closeIdx++
		return lines, closeIdx
	}

	// Detect inline vs multiline format.
	if strings.Contains(lines[tagsLineIdx], "[") {
		// Inline format: tags: ["a", "b"]
		lines[tagsLineIdx] = formatInlineTags(newTags)
		return lines, closeIdx
	}

	// Multiline format: tags:\n  - a\n  - b
	// Remove existing   - lines after tags:
	removeStart := tagsLineIdx + 1
	removeEnd := removeStart
	for removeEnd < closeIdx && strings.HasPrefix(strings.TrimSpace(lines[removeEnd]), "- ") {
		removeEnd++
	}

	// Build new multiline tag lines.
	var newTagLines []string
	for _, t := range newTags {
		newTagLines = append(newTagLines, "  - "+t)
	}

	// Replace the old tag item lines with new ones.
	before := lines[:removeStart]
	after := lines[removeEnd:]
	lines = make([]string, 0, len(before)+len(newTagLines)+len(after))
	lines = append(lines, before...)
	lines = append(lines, newTagLines...)
	lines = append(lines, after...)

	// Adjust closeIdx based on line count change.
	closeIdx += len(newTagLines) - (removeEnd - removeStart)
	return lines, closeIdx
}

func formatInlineTags(tags []string) string {
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

// findFrontmatterBounds returns the line indices of the opening and closing "---" delimiters.
func findFrontmatterBounds(lines []string) (int, int) {
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

type changeEntry struct {
	field    string
	oldValue string
	newValue string
}

func buildChangeLog(task *model.Task, updates []fieldUpdate, tags *tagUpdate) []changeEntry {
	oldValues := map[string]string{
		"status":   string(task.Status),
		"priority": string(task.Priority),
		"effort":   string(task.Effort),
	}

	changes := make([]changeEntry, 0, len(updates)+1)
	for _, u := range updates {
		changes = append(changes, changeEntry{
			field:    u.key,
			oldValue: oldValues[u.key],
			newValue: u.newValue,
		})
	}

	if tags != nil {
		newTags := computeNewTags(task.Tags, tags.addTags, tags.removeTags)
		changes = append(changes, changeEntry{
			field:    "tags",
			oldValue: "[" + strings.Join(task.Tags, ", ") + "]",
			newValue: "[" + strings.Join(newTags, ", ") + "]",
		})
	}

	return changes
}

func printUpdateConfirmation(task *model.Task, changes []changeEntry) {
	fmt.Printf("Updated task %s (%s):\n", task.ID, task.Title)
	for _, c := range changes {
		old := c.oldValue
		if old == "" {
			old = "(unset)"
		}
		fmt.Printf("  %s: %s -> %s\n", c.field, old, c.newValue)
	}
}
