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
	setTaskID     string
	setStatus     string
	setPriority   string
	setEffort     string
	setDone       bool
	setDryRun     bool
	setAddTags    []string
	setRemoveTags []string
)

var setCmd = &cobra.Command{
	Use:        "set",
	SuggestFor: []string{"edit", "modify", "change"},
	Short:      "Set a task's frontmatter fields",
	Long: `Set modifies frontmatter fields (status, priority, effort, tags) of a task file.

The task is identified by --task-id (exact match only).

Examples:
  taskmd set --task-id cli-049 --status completed
  taskmd set --task-id cli-049 --priority high --effort large
  taskmd set --task-id cli-049 --done
  taskmd set --task-id cli-049 --add-tag backend --add-tag api
  taskmd set --task-id cli-049 --remove-tag deprecated`,
	Args: cobra.NoArgs,
	RunE: runSet,
}

// Deprecated: use "set" instead.
var updateCmd = &cobra.Command{
	Use:        "update",
	Short:      "Update a task's frontmatter fields (deprecated: use 'set')",
	Args:       cobra.NoArgs,
	RunE:       runSet,
	Hidden:     true,
	Deprecated: "use 'set' instead",
}

func init() {
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(updateCmd)

	for _, cmd := range []*cobra.Command{setCmd, updateCmd} {
		cmd.Flags().StringVar(&setTaskID, "task-id", "", "task ID to update (required)")
		cmd.Flags().StringVar(&setStatus, "status", "", "new status (pending, in-progress, completed, blocked, cancelled)")
		cmd.Flags().StringVar(&setPriority, "priority", "", "new priority (low, medium, high, critical)")
		cmd.Flags().StringVar(&setEffort, "effort", "", "new effort (small, medium, large)")
		cmd.Flags().BoolVar(&setDone, "done", false, "mark task as completed (alias for --status completed)")
		cmd.Flags().BoolVar(&setDryRun, "dry-run", false, "preview changes without writing to disk")
		cmd.Flags().StringArrayVar(&setAddTags, "add-tag", nil, "add a tag (repeatable)")
		cmd.Flags().StringArrayVar(&setRemoveTags, "remove-tag", nil, "remove a tag (repeatable)")

		_ = cmd.MarkFlagRequired("task-id")
	}
}

func runSet(cmd *cobra.Command, _ []string) error {
	req, err := buildSetRequest(cmd)
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

	debugLog("scan directory: %s", scanDir)
	debugLog("found %d task(s)", len(result.Tasks))

	task := findExactMatch(setTaskID, result.Tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", setTaskID)
	}

	changes := buildChangeLog(task, req)

	if setDryRun {
		printSetConfirmation(task, changes)
		fmt.Println("\nDry run — no changes made.")
		return nil
	}

	if err := taskfile.UpdateTaskFile(task.FilePath, req); err != nil {
		return err
	}

	printSetConfirmation(task, changes)
	return nil
}

func buildSetRequest(cmd *cobra.Command) (taskfile.UpdateRequest, error) {
	if setDone && cmd.Flags().Changed("status") {
		return taskfile.UpdateRequest{}, fmt.Errorf("--done and --status are mutually exclusive")
	}

	if setDone {
		setStatus = string(model.StatusCompleted)
	}

	var req taskfile.UpdateRequest

	if setStatus != "" {
		req.Status = &setStatus
	}
	if setPriority != "" {
		req.Priority = &setPriority
	}
	if setEffort != "" {
		req.Effort = &setEffort
	}
	if len(setAddTags) > 0 {
		req.AddTags = setAddTags
	}
	if len(setRemoveTags) > 0 {
		req.RemTags = setRemoveTags
	}

	// Validate enum values with suggestions
	if err := validateSetEnums(req); err != nil {
		return taskfile.UpdateRequest{}, err
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

func printSetConfirmation(task *model.Task, changes []changeEntry) {
	fmt.Printf("Updated task %s (%s):\n", task.ID, task.Title)
	for _, c := range changes {
		old := c.oldValue
		if old == "" {
			old = "(unset)"
		}
		fmt.Printf("  %s: %s -> %s\n", c.field, old, c.newValue)
	}
}
