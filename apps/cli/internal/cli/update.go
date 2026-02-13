package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/taskfile"
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

func runUpdate(cmd *cobra.Command, _ []string) error {
	req, err := buildUpdateRequest(cmd)
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

	changes := buildChangeLog(task, req)

	if err := taskfile.UpdateTaskFile(task.FilePath, req); err != nil {
		return err
	}

	printUpdateConfirmation(task, changes)
	return nil
}

func buildUpdateRequest(cmd *cobra.Command) (taskfile.UpdateRequest, error) {
	if updateDone && cmd.Flags().Changed("status") {
		return taskfile.UpdateRequest{}, fmt.Errorf("--done and --status are mutually exclusive")
	}

	if updateDone {
		updateStatus = string(model.StatusCompleted)
	}

	var req taskfile.UpdateRequest

	if updateStatus != "" {
		req.Status = &updateStatus
	}
	if updatePriority != "" {
		req.Priority = &updatePriority
	}
	if updateEffort != "" {
		req.Effort = &updateEffort
	}
	if len(updateAddTags) > 0 {
		req.AddTags = updateAddTags
	}
	if len(updateRemoveTags) > 0 {
		req.RemTags = updateRemoveTags
	}

	// Validate enum values
	errs := taskfile.ValidateUpdateRequest(req)
	if len(errs) > 0 {
		return taskfile.UpdateRequest{}, fmt.Errorf("%s", errs[0])
	}

	hasScalar := req.Status != nil || req.Priority != nil || req.Effort != nil
	hasTags := len(req.AddTags) > 0 || len(req.RemTags) > 0
	if !hasScalar && !hasTags {
		return taskfile.UpdateRequest{}, fmt.Errorf("nothing to update: provide --status, --priority, --effort, --done, --add-tag, or --remove-tag")
	}

	return req, nil
}

type changeEntry struct {
	field    string
	oldValue string
	newValue string
}

func buildChangeLog(task *model.Task, req taskfile.UpdateRequest) []changeEntry {
	oldValues := map[string]string{
		"status":   string(task.Status),
		"priority": string(task.Priority),
		"effort":   string(task.Effort),
	}

	var changes []changeEntry

	if req.Status != nil {
		changes = append(changes, changeEntry{field: "status", oldValue: oldValues["status"], newValue: *req.Status})
	}
	if req.Priority != nil {
		changes = append(changes, changeEntry{field: "priority", oldValue: oldValues["priority"], newValue: *req.Priority})
	}
	if req.Effort != nil {
		changes = append(changes, changeEntry{field: "effort", oldValue: oldValues["effort"], newValue: *req.Effort})
	}

	if len(req.AddTags) > 0 || len(req.RemTags) > 0 {
		newTags := taskfile.ComputeNewTags(task.Tags, req.AddTags, req.RemTags)
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
