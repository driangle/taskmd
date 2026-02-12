package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
)

var (
	updateTaskID   string
	updateStatus   string
	updatePriority string
	updateEffort   string
	updateDone     bool
)

var validStatuses = map[string]bool{
	string(model.StatusPending):    true,
	string(model.StatusInProgress): true,
	string(model.StatusCompleted):  true,
	string(model.StatusBlocked):    true,
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
	Long: `Update modifies frontmatter fields (status, priority, effort) of a task file.

The task is identified by --task-id (exact match only).

Examples:
  taskmd update --task-id cli-049 --status completed
  taskmd update --task-id cli-049 --priority high --effort large
  taskmd update --task-id cli-049 --done`,
	Args: cobra.NoArgs,
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVar(&updateTaskID, "task-id", "", "task ID to update (required)")
	updateCmd.Flags().StringVar(&updateStatus, "status", "", "new status (pending, in-progress, completed, blocked)")
	updateCmd.Flags().StringVar(&updatePriority, "priority", "", "new priority (low, medium, high, critical)")
	updateCmd.Flags().StringVar(&updateEffort, "effort", "", "new effort (small, medium, large)")
	updateCmd.Flags().BoolVar(&updateDone, "done", false, "mark task as completed (alias for --status completed)")

	_ = updateCmd.MarkFlagRequired("task-id")
}

func runUpdate(cmd *cobra.Command, _ []string) error {
	updates, err := validateUpdateFlags(cmd)
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

	return applyUpdates(task, updates)
}

// fieldUpdate represents a single frontmatter field change.
type fieldUpdate struct {
	key      string
	newValue string
}

func validateUpdateFlags(cmd *cobra.Command) ([]fieldUpdate, error) {
	if updateDone && cmd.Flags().Changed("status") {
		return nil, fmt.Errorf("--done and --status are mutually exclusive")
	}

	if updateDone {
		updateStatus = string(model.StatusCompleted)
	}

	var updates []fieldUpdate

	if updateStatus != "" {
		if !validStatuses[updateStatus] {
			return nil, fmt.Errorf("invalid status: %q (valid: pending, in-progress, completed, blocked)", updateStatus)
		}
		updates = append(updates, fieldUpdate{key: "status", newValue: updateStatus})
	}

	if updatePriority != "" {
		if !validPriorities[updatePriority] {
			return nil, fmt.Errorf("invalid priority: %q (valid: low, medium, high, critical)", updatePriority)
		}
		updates = append(updates, fieldUpdate{key: "priority", newValue: updatePriority})
	}

	if updateEffort != "" {
		if !validEfforts[updateEffort] {
			return nil, fmt.Errorf("invalid effort: %q (valid: small, medium, large)", updateEffort)
		}
		updates = append(updates, fieldUpdate{key: "effort", newValue: updateEffort})
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("nothing to update: provide --status, --priority, --effort, or --done")
	}

	return updates, nil
}

func applyUpdates(task *model.Task, updates []fieldUpdate) error {
	content, err := os.ReadFile(task.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read task file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	openIdx, closeIdx := findFrontmatterBounds(lines)
	if openIdx < 0 || closeIdx < 0 {
		return fmt.Errorf("task file has no valid frontmatter: %s", task.FilePath)
	}

	changes := buildChangeLog(task, updates)

	for i := openIdx + 1; i < closeIdx; i++ {
		for j := range updates {
			prefix := updates[j].key + ":"
			if strings.HasPrefix(strings.TrimSpace(lines[i]), prefix) {
				lines[i] = updates[j].key + ": " + updates[j].newValue
				break
			}
		}
	}

	if err := os.WriteFile(task.FilePath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	printUpdateConfirmation(task, changes)
	return nil
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

func buildChangeLog(task *model.Task, updates []fieldUpdate) []changeEntry {
	oldValues := map[string]string{
		"status":   string(task.Status),
		"priority": string(task.Priority),
		"effort":   string(task.Effort),
	}

	changes := make([]changeEntry, 0, len(updates))
	for _, u := range updates {
		changes = append(changes, changeEntry{
			field:    u.key,
			oldValue: oldValues[u.key],
			newValue: u.newValue,
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
