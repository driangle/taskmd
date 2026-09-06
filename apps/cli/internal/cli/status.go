package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	statusFormat     string
	statusExact      bool
	statusThreshold  float64
	statusMinimal    bool
	statusStatusline bool
	statusScope      string
)

// statusStdinReader is the reader used for interactive selection prompts.
// Override in tests to simulate user input.
var statusStdinReader io.Reader = os.Stdin

var statusCmd = &cobra.Command{
	Use:        "status [query]",
	SuggestFor: []string{"progress"},
	Short:      "Show in-progress tasks or get metadata for a specific task",
	Long: `Without arguments, status shows all in-progress tasks.
With a query argument, it displays the frontmatter metadata of a specific task
(without body content, resolved dependency info, context files, or worklog data).

Matching uses the same logic as 'get' (ID, title, file path, fuzzy).

Examples:
  # Show all in-progress tasks
  taskmd status

  # Compact output for shell statuslines
  taskmd status --statusline

  # Filter by scope
  taskmd status --scope cli

  # Look up a specific task
  taskmd status 042
  taskmd status "Setup project"
  taskmd status 042 --format json
  taskmd status sho --exact`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&statusFormat, "format", "text", "output format (text, json, yaml)")
	statusCmd.Flags().BoolVar(&statusExact, "exact", false, "disable fuzzy matching, exact only")
	statusCmd.Flags().Float64Var(&statusThreshold, "threshold", 0.6, "fuzzy match sensitivity (0.0-1.0)")
	statusCmd.Flags().BoolVar(&statusMinimal, "minimal", false, "show only task metadata, skip children")
	statusCmd.Flags().BoolVar(&statusStatusline, "statusline", false, "compact output for Claude Code statusline")
	statusCmd.Flags().StringVar(&statusScope, "scope", "", "filter by group/directory; supports wildcards (e.g. cli, cli*)")
}

func runStatus(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runStatusList()
	}
	return runStatusSingle(args[0])
}

func runStatusList() error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	// The overlay is built before paths are relativized — the merge stats
	// files by path.
	local, overlay, err := scanTasksWithOverlay(scanDir, flags)
	if err != nil {
		return err
	}
	makeFilePathsRelative(local, scanDir)

	// --statusline reads local files only, even with the overlay active: it
	// answers "what am I working on in *this* checkout", and a task claimed in
	// a sibling worktree is another agent's work. See status_overlay.go.
	tasks := local
	if !statusStatusline {
		tasks = statusViewTasks(local, overlay)
	}

	filtered, err := applyFilters(tasks, statusListFilters())
	if err != nil {
		return fmt.Errorf("filter failed: %w", err)
	}

	if len(filtered) == 0 {
		return outputStatusListEmpty()
	}

	if statusStatusline {
		return outputStatusline(filtered, os.Stdout)
	}

	return outputStatusListFormatted(tasks, filtered, overlay)
}

// statusListFilters is the filter set the no-argument status view applies:
// in-progress tasks, narrowed to --scope when given.
func statusListFilters() []string {
	filters := []string{"status=in-progress"}
	if statusScope != "" {
		filters = append(filters, "group="+statusScope)
	}
	return filters
}

func outputStatusListEmpty() error {
	switch {
	case statusStatusline:
		return nil
	case statusFormat == "json":
		return WriteJSON(os.Stdout, []statusOutput{})
	case statusFormat == "yaml":
		return WriteYAML(os.Stdout, []statusOutput{})
	default:
		fmt.Fprintln(os.Stderr, "No tasks currently in progress.")
		return nil
	}
}

func outputStatusListFormatted(tasks, filtered []*model.Task, overlay *worktreeOverlay) error {
	var childrenIndex map[string][]*model.Task
	if !statusMinimal {
		childrenIndex = buildChildrenIndex(tasks)
	}

	tasksByID := buildTasksByIDMap(tasks)

	outputs := make([]statusOutput, 0, len(filtered))
	for _, task := range filtered {
		out := buildStatusOutputFromTask(task, childrenIndex, tasksByID)
		annotateStatusProvenance(&out, overlay)
		outputs = append(outputs, out)
	}

	switch statusFormat {
	case "text":
		return outputStatusListText(outputs, os.Stdout)
	case "json":
		return WriteJSON(os.Stdout, outputs)
	case "yaml":
		return WriteYAML(os.Stdout, outputs)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", statusFormat)
	}
}

func runStatusSingle(query string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	local, overlay, err := scanTasksWithOverlay(scanDir, flags)
	if err != nil {
		return err
	}
	makeFilePathsRelative(local, scanDir)

	// Swap stdin reader for fuzzy selection prompts
	origReader := getStdinReader
	getStdinReader = statusStdinReader
	defer func() { getStdinReader = origReader }()

	// Sibling-only tasks are part of the merged read view, so they are
	// addressable here too — the same rule get follows.
	tasks := statusViewTasks(local, overlay)

	task, err := resolveTask(query, tasks, statusExact, statusThreshold)
	if err != nil {
		return err
	}

	var childrenIndex map[string][]*model.Task
	if !statusMinimal {
		childrenIndex = buildChildrenIndex(tasks)
	}

	tasksByID := buildTasksByIDMap(tasks)
	out := buildStatusOutputFromTask(task, childrenIndex, tasksByID)
	annotateStatusProvenance(&out, overlay)

	switch statusFormat {
	case "text":
		return outputStatusText(out, os.Stdout)
	case "json":
		return WriteJSON(os.Stdout, out)
	case "yaml":
		return WriteYAML(os.Stdout, out)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", statusFormat)
	}
}
