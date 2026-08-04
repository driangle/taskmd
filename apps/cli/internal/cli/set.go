package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/graph"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/taskfile"
	"github.com/driangle/taskmd/sdk/go/verify"
)

var (
	setTaskID        string
	setStatus        string
	setPriority      string
	setEffort        string
	setType          string
	setOwner         string
	setParent        string
	setDone          bool
	setDryRun        bool
	setVerify        bool
	setAddTags       []string
	setRemoveTags    []string
	setAddPRs        []string
	setRemovePRs     []string
	setAddTouches    []string
	setRemoveTouches []string
	setDependsOn     string
	setPhase         string
)

var setCmd = &cobra.Command{
	Use:        "set [task-id]",
	SuggestFor: []string{"edit", "modify", "change", "update"},
	Short:      "Set a task's frontmatter fields",
	Long: `Set modifies frontmatter fields (status, priority, effort, tags) of a task file.

The task is identified by a positional argument or --task-id (exact match only).

Examples:
  taskmd set cli-049 --status completed
  taskmd set cli-049 --priority high --effort large
  taskmd set cli-049 --done
  taskmd set cli-049 --add-tag backend --add-tag api
  taskmd set cli-049 --remove-tag deprecated
  taskmd set cli-049 --add-touches cli/graph --add-touches cli/output
  taskmd set cli-049 --remove-touches cli/graph
  taskmd set --task-id cli-049 --status completed   # --task-id also works`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSet,
}

func init() {
	rootCmd.AddCommand(setCmd)

	setCmd.Flags().StringVar(&setTaskID, "task-id", "", "task ID to update (required)")
	setCmd.Flags().StringVar(&setStatus, "status", "", "new status (pending, in-progress, completed, in-review, blocked, cancelled)")
	setCmd.Flags().StringVar(&setPriority, "priority", "", "new priority (low, medium, high, critical)")
	setCmd.Flags().StringVar(&setEffort, "effort", "", "new effort (small, medium, large)")
	setCmd.Flags().StringVar(&setType, "type", "", "work type (feature, bug, improvement, chore, docs)")
	setCmd.Flags().StringVar(&setOwner, "owner", "", "owner/assignee of the task")
	setCmd.Flags().StringVar(&setParent, "parent", "", "parent task ID (use empty string to clear)")
	setCmd.Flags().BoolVar(&setDone, "done", false, "mark task as completed (alias for --status completed)")
	setCmd.Flags().BoolVar(&setDryRun, "dry-run", false, "preview changes without writing to disk")
	setCmd.Flags().BoolVar(&setVerify, "verify", false, "run verification checks before completing a task")
	setCmd.Flags().StringArrayVar(&setAddTags, "add-tag", nil, "add a tag (repeatable)")
	setCmd.Flags().StringArrayVar(&setRemoveTags, "remove-tag", nil, "remove a tag (repeatable)")
	setCmd.Flags().StringArrayVar(&setAddPRs, "add-pr", nil, "add a PR URL (repeatable)")
	setCmd.Flags().StringArrayVar(&setRemovePRs, "remove-pr", nil, "remove a PR URL (repeatable)")
	setCmd.Flags().StringArrayVar(&setAddTouches, "add-touches", nil, "add a scope identifier to touches (repeatable)")
	setCmd.Flags().StringArrayVar(&setRemoveTouches, "remove-touches", nil, "remove a scope identifier from touches (repeatable)")
	setCmd.Flags().StringVar(&setDependsOn, "depends-on", "", "set dependencies (comma-separated IDs, e.g. 010,015)")
	setCmd.Flags().StringVar(&setPhase, "phase", "", "phase name (use empty string to clear)")
}

func resolveSetTaskID(cmd *cobra.Command, args []string) (string, error) {
	positional := ""
	if len(args) > 0 {
		positional = args[0]
	}
	flagVal := setTaskID

	if positional != "" && flagVal != "" && positional != flagVal {
		return "", fmt.Errorf("conflicting task ID: positional argument %q and --task-id %q differ", positional, flagVal)
	}

	if positional != "" {
		return positional, nil
	}
	if flagVal != "" {
		return flagVal, nil
	}
	return "", fmt.Errorf("task ID required: provide as positional argument or --task-id flag")
}

func runSet(cmd *cobra.Command, args []string) error {
	taskID, err := resolveSetTaskID(cmd, args)
	if err != nil {
		return err
	}

	req, err := buildSetRequest(cmd)
	if err != nil {
		return err
	}

	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	debugLog("scan directory: %s", scanDir)
	debugLog("found %d task(s)", len(tasks))

	task := findExactMatch(taskID, tasks)
	if task == nil {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if dupes := findDuplicatesByID(taskID, tasks); len(dupes) > 1 {
		return fmt.Errorf("refusing to modify task %s: found %d files with this ID\n%s\nRun 'taskmd deduplicate' to fix",
			taskID, len(dupes), formatDuplicatePaths(dupes))
	}

	if req.Dependencies != nil {
		if err := validateDependencies(task, *req.Dependencies, tasks); err != nil {
			return err
		}
	}

	applyTerminalDateLogic(task, &req)

	if err := runSetVerification(task, req); err != nil {
		return err
	}

	changes := buildChangeLog(task, req)

	if setDryRun {
		printSetConfirmation(task, changes)
		r := getRenderer()
		fmt.Println("\n" + formatWarning("Dry run — no changes made.", r))
		return nil
	}

	if err := taskfile.UpdateTaskFile(task.FilePath, req); err != nil {
		return err
	}

	printSetConfirmation(task, changes)
	return nil
}

// setStringField sets a pointer field if the value is non-empty.
func setStringField(dst **string, value string) {
	if value != "" {
		*dst = &value
	}
}

func resolveDoneFlag(cmd *cobra.Command) error {
	if setDone && cmd.Flags().Changed("status") {
		return fmt.Errorf("--done and --status are mutually exclusive")
	}
	if setDone {
		flags := GetGlobalFlags()
		if flags.Workflow == "pr-review" {
			setStatus = string(model.StatusInReview)
		} else {
			setStatus = string(model.StatusCompleted)
		}
	}
	return nil
}

// applyTerminalDateLogic auto-sets or clears completed_at/cancelled_at based on status transitions.
func applyTerminalDateLogic(task *model.Task, req *taskfile.UpdateRequest) {
	if req.Status == nil {
		return
	}
	today := time.Now().Format("2006-01-02")
	newStatus := model.Status(*req.Status)
	wasTerminal := task.Status.IsResolved()

	switch {
	case newStatus == model.StatusCompleted:
		req.Completed = &today
		req.RemoveFields = append(req.RemoveFields, "cancelled_at")
	case newStatus == model.StatusCancelled:
		req.CancelledAt = &today
		req.RemoveFields = append(req.RemoveFields, "completed_at")
	case wasTerminal:
		// Reopening: task was completed/cancelled, now moving to a non-terminal status.
		req.RemoveFields = append(req.RemoveFields, "completed_at", "cancelled_at")
	}
}

func buildSetRequest(cmd *cobra.Command) (taskfile.UpdateRequest, error) {
	if err := resolveDoneFlag(cmd); err != nil {
		return taskfile.UpdateRequest{}, err
	}

	var req taskfile.UpdateRequest

	setStringField(&req.Status, setStatus)
	setStringField(&req.Priority, setPriority)
	setStringField(&req.Effort, setEffort)
	setStringField(&req.Type, setType)
	setStringField(&req.Owner, setOwner)

	if cmd.Flags().Changed("parent") {
		req.Parent = &setParent
	}
	if cmd.Flags().Changed("phase") {
		req.Phase = &setPhase
	}

	if len(setAddTags) > 0 {
		req.AddTags = setAddTags
	}
	if len(setRemoveTags) > 0 {
		req.RemTags = setRemoveTags
	}
	if len(setAddPRs) > 0 {
		req.AddPRs = setAddPRs
	}
	if len(setRemovePRs) > 0 {
		req.RemPRs = setRemovePRs
	}
	if len(setAddTouches) > 0 {
		req.AddTouches = setAddTouches
	}
	if len(setRemoveTouches) > 0 {
		req.RemTouches = setRemoveTouches
	}

	if cmd.Flags().Changed("depends-on") {
		deps := parseCommaSeparatedIDs(setDependsOn)
		req.Dependencies = &deps
	}

	if err := validateSetEnums(req); err != nil {
		return taskfile.UpdateRequest{}, err
	}

	if !hasUpdates(req) {
		return taskfile.UpdateRequest{}, fmt.Errorf("nothing to update: provide --status, --priority, --effort, --type, --owner, --parent, --phase, --done, --add-tag, --remove-tag, --add-pr, --remove-pr, --add-touches, --remove-touches, or --depends-on")
	}

	return req, nil
}

func hasUpdates(req taskfile.UpdateRequest) bool {
	hasScalar := req.Status != nil || req.Priority != nil || req.Effort != nil ||
		req.Type != nil || req.Owner != nil || req.Parent != nil || req.Phase != nil
	hasTags := len(req.AddTags) > 0 || len(req.RemTags) > 0
	hasPRs := len(req.AddPRs) > 0 || len(req.RemPRs) > 0
	hasTouches := len(req.AddTouches) > 0 || len(req.RemTouches) > 0
	hasDeps := req.Dependencies != nil
	return hasScalar || hasTags || hasPRs || hasTouches || hasDeps
}

type changeEntry struct {
	field    string
	oldValue string
	newValue string
}

func listChangeEntry(field string, current, add, remove []string) *changeEntry {
	if len(add) == 0 && len(remove) == 0 {
		return nil
	}
	newValues := taskfile.ComputeNewTags(current, add, remove)
	return &changeEntry{
		field:    field,
		oldValue: "[" + strings.Join(current, ", ") + "]",
		newValue: "[" + strings.Join(newValues, ", ") + "]",
	}
}

func scalarChangeEntry(field, oldValue string, newValue *string) *changeEntry {
	if newValue == nil {
		return nil
	}
	return &changeEntry{field: field, oldValue: oldValue, newValue: *newValue}
}

func buildChangeLog(task *model.Task, req taskfile.UpdateRequest) []changeEntry {
	var changes []changeEntry
	for _, sc := range []struct {
		field    string
		oldValue string
		newValue *string
	}{
		{"status", string(task.Status), req.Status},
		{"priority", string(task.Priority), req.Priority},
		{"effort", string(task.Effort), req.Effort},
		{"type", string(task.Type), req.Type},
		{"owner", task.Owner, req.Owner},
		{"parent", task.Parent, req.Parent},
		{"phase", task.Phase, req.Phase},
	} {
		if ce := scalarChangeEntry(sc.field, sc.oldValue, sc.newValue); ce != nil {
			changes = append(changes, *ce)
		}
	}
	if ce := terminalDateChangeEntry("completed_at", task.Completed, req.Completed, req.RemoveFields); ce != nil {
		changes = append(changes, *ce)
	}
	if ce := terminalDateChangeEntry("cancelled_at", task.CancelledAt, req.CancelledAt, req.RemoveFields); ce != nil {
		changes = append(changes, *ce)
	}

	for _, entry := range []struct {
		field   string
		current []string
		add     []string
		remove  []string
	}{
		{"tags", task.Tags, req.AddTags, req.RemTags},
		{"pr", task.PRs, req.AddPRs, req.RemPRs},
		{"touches", task.Touches, req.AddTouches, req.RemTouches},
	} {
		if ce := listChangeEntry(entry.field, entry.current, entry.add, entry.remove); ce != nil {
			changes = append(changes, *ce)
		}
	}

	if req.Dependencies != nil {
		changes = append(changes, changeEntry{
			field:    "dependencies",
			oldValue: "[" + strings.Join(task.Dependencies, ", ") + "]",
			newValue: "[" + strings.Join(*req.Dependencies, ", ") + "]",
		})
	}

	return changes
}

func terminalDateChangeEntry(field string, oldTime model.FlexibleTime, newValue *string, removeFields []string) *changeEntry {
	old := ""
	if !oldTime.IsZero() {
		old = oldTime.Format("2006-01-02")
	}
	if newValue != nil {
		return &changeEntry{field: field, oldValue: old, newValue: *newValue}
	}
	for _, f := range removeFields {
		if f == field && old != "" {
			return &changeEntry{field: field, oldValue: old, newValue: "(cleared)"}
		}
	}
	return nil
}

func printSetConfirmation(task *model.Task, changes []changeEntry) {
	r := getRenderer()
	fmt.Printf("Updated task %s (%s):\n", formatTaskID(task.ID, r), task.Title)
	for _, c := range changes {
		old := c.oldValue
		if old == "" {
			old = "(unset)"
		}
		fmt.Printf("  %s: %s -> %s\n",
			formatLabel(c.field, r),
			formatDim(old, r),
			colorizeFieldValue(c.field, c.newValue, r),
		)
	}
}

func colorizeFieldValue(field, value string, r *lipgloss.Renderer) string {
	switch field {
	case "status":
		return formatStatus(value, r)
	case "priority":
		return formatPriority(value, r)
	case "effort":
		return formatEffort(value, r)
	default:
		return value
	}
}

// runSetVerification runs verify checks if --verify is set and status is being set to completed.
func runSetVerification(task *model.Task, req taskfile.UpdateRequest) error {
	if !setVerify {
		return nil
	}
	isCompleting := req.Status != nil && *req.Status == string(model.StatusCompleted)
	if !isCompleting {
		return nil
	}
	if len(task.Verify) == 0 {
		return nil
	}

	if errs := model.ValidateVerifySteps(task.Verify); len(errs) > 0 {
		return fmt.Errorf("invalid verify steps:\n  %s", strings.Join(errs, "\n  "))
	}

	flags := GetGlobalFlags()
	projectRoot := resolveProjectRoot()
	opts := verify.Options{
		ProjectRoot: projectRoot,
		FailFast:    true,
		Timeout:     60 * time.Second,
		Verbose:     flags.Verbose,
		LogFunc: func(format string, args ...any) {
			if !flags.Quiet {
				fmt.Fprintf(os.Stderr, format+"\n", args...)
			}
		},
	}

	vResult := verify.Run(task.Verify, opts)
	printVerifyTable(vResult)

	if vResult.HasFailures() {
		return fmt.Errorf("verification failed: %d check(s) failed — status change aborted", vResult.Failed)
	}
	return nil
}

// parseCommaSeparatedIDs splits a comma-separated string into a slice of trimmed IDs.
// An empty string returns an empty slice (used to clear the field).
func parseCommaSeparatedIDs(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	var ids []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			ids = append(ids, p)
		}
	}
	return ids
}

// validateDependencies checks that all dependency IDs exist and that
// setting them would not create a circular dependency.
func validateDependencies(task *model.Task, depIDs []string, tasks []*model.Task) error {
	tasksByID := make(map[string]*model.Task, len(tasks))
	for _, t := range tasks {
		tasksByID[t.ID] = t
	}

	// Check existence.
	for _, id := range depIDs {
		if _, ok := tasksByID[id]; !ok {
			return fmt.Errorf("dependency %q not found", id)
		}
	}

	// Self-dependency check.
	for _, id := range depIDs {
		if id == task.ID {
			return fmt.Errorf("task cannot depend on itself: %s", id)
		}
	}

	// Check for circular dependencies by temporarily setting the new deps.
	origDeps := task.Dependencies
	task.Dependencies = depIDs
	defer func() { task.Dependencies = origDeps }()

	g := graph.NewGraph(tasks)
	if cycles := g.DetectCycles(); len(cycles) > 0 {
		return fmt.Errorf("circular dependency detected: %s", formatCycle(cycles[0]))
	}

	return nil
}

func formatCycle(cycle []string) string {
	return strings.Join(cycle, " -> ")
}
