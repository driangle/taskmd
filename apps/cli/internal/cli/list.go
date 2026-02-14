package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	listFormat  string
	listFilters []string
	listSort    string
	listColumns string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:        "list",
	SuggestFor: []string{"ls", "tasks", "all"},
	Short:      "List tasks in a quick textual format",
	Long: `List displays tasks in various formats with filtering and sorting support.

By default, scans the current directory and all subdirectories for markdown files
with task frontmatter. You can specify a different directory to scan.

Output formats: table (default), json, yaml

Multiple --filter flags are combined with AND logic.

Examples:
  taskmd list
  taskmd list ./tasks
  taskmd list --filter status=pending
  taskmd list --filter status=pending --filter priority=high
  taskmd list --sort priority
  taskmd list --columns id,title,deps
  taskmd list --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listFormat, "format", "table", "output format (table, json, yaml)")
	listCmd.Flags().StringArrayVar(&listFilters, "filter", []string{}, "filter tasks (can specify multiple times for AND conditions, e.g., --filter status=pending --filter priority=high)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "sort by field (id, title, status, priority, effort, created)")
	listCmd.Flags().StringVar(&listColumns, "columns", "id,title,status,priority,file", "comma-separated list of columns to display")
}

func runList(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)
	debugLog("scan directory: %s", scanDir)

	// Create scanner and scan for tasks
	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks
	debugLog("found %d task(s)", len(tasks))

	// Make file paths relative to scan directory
	makeFilePathsRelative(tasks, scanDir)

	// Report any scan errors if verbose
	if flags.Verbose && len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: encountered %d errors during scan:\n", len(result.Errors))
		for _, scanErr := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
		}
		fmt.Fprintln(os.Stderr)
	}

	debugLog("format: %s, sort: %q, filters: %v", listFormat, listSort, listFilters)

	// Apply filters (multiple filters are AND'ed together)
	if len(listFilters) > 0 {
		tasks, err = applyFilters(tasks, listFilters)
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}

	// Apply sorting
	if listSort != "" {
		if err := sortTasks(tasks, listSort); err != nil {
			return fmt.Errorf("sort error: %w", err)
		}
	}

	// Output in requested format
	switch listFormat {
	case "json":
		return outputJSON(tasks)
	case "yaml":
		return outputYAML(tasks)
	case "table":
		return outputTable(tasks, listColumns)
	default:
		return ValidateFormat(listFormat, []string{"table", "json", "yaml"})
	}
}

// sortTasks sorts tasks by the specified field
func sortTasks(tasks []*model.Task, sortField string) error {
	switch sortField {
	case "id":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].ID < tasks[j].ID
		})
	case "title":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Title < tasks[j].Title
		})
	case "status":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Status < tasks[j].Status
		})
	case "priority":
		priorityOrder := map[model.Priority]int{
			model.PriorityCritical: 0,
			model.PriorityHigh:     1,
			model.PriorityMedium:   2,
			model.PriorityLow:      3,
		}
		sort.Slice(tasks, func(i, j int) bool {
			return priorityOrder[tasks[i].Priority] < priorityOrder[tasks[j].Priority]
		})
	case "effort":
		effortOrder := map[model.Effort]int{
			model.EffortSmall:  0,
			model.EffortMedium: 1,
			model.EffortLarge:  2,
		}
		sort.Slice(tasks, func(i, j int) bool {
			return effortOrder[tasks[i].Effort] < effortOrder[tasks[j].Effort]
		})
	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Created.Before(tasks[j].Created)
		})
	default:
		return invalidValueError("sort field", sortField, validSortFields)
	}

	return nil
}

// outputJSON outputs tasks as JSON
func outputJSON(tasks []*model.Task) error {
	return WriteJSON(os.Stdout, tasks)
}

// outputYAML outputs tasks as YAML
func outputYAML(tasks []*model.Task) error {
	return WriteYAML(os.Stdout, tasks)
}

// outputTable outputs tasks as a formatted table
func outputTable(tasks []*model.Task, columnsStr string) error {
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Parse columns
	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}

	r := getRenderer()

	// Create tab writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Write header
	fmt.Fprintln(w, strings.Join(columns, "\t"))

	// Write separator
	separators := make([]string, len(columns))
	for i := range separators {
		separators[i] = strings.Repeat("-", 10)
	}
	fmt.Fprintln(w, strings.Join(separators, "\t"))

	// Write rows
	for _, task := range tasks {
		row := make([]string, len(columns))
		for i, col := range columns {
			row[i] = colorizeColumn(task, col, r)
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
}

// colorizeColumn returns the column value with color formatting applied.
func colorizeColumn(task *model.Task, column string, r *lipgloss.Renderer) string {
	value := getColumnValue(task, column)
	switch column {
	case "id":
		return formatTaskID(value, r)
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

// makeFilePathsRelative converts absolute task file paths to paths relative to baseDir.
func makeFilePathsRelative(tasks []*model.Task, baseDir string) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return
	}
	for _, task := range tasks {
		if rel, err := filepath.Rel(absBase, task.FilePath); err == nil {
			task.FilePath = rel
		}
	}
}

// getColumnValue extracts the value for a specific column from a task
func getColumnValue(task *model.Task, column string) string {
	scalar := getScalarColumnValue(task, column)
	if scalar != "" {
		return scalar
	}
	switch column {
	case "created":
		if task.Created.IsZero() {
			return ""
		}
		return task.Created.Format("2006-01-02")
	case "deps":
		return strings.Join(task.Dependencies, ",")
	case "tags":
		return strings.Join(task.Tags, ",")
	default:
		return ""
	}
}

// getScalarColumnValue returns simple string field values.
func getScalarColumnValue(task *model.Task, column string) string {
	switch column {
	case "id":
		return task.ID
	case "title":
		return task.Title
	case "status":
		return string(task.Status)
	case "priority":
		return string(task.Priority)
	case "effort":
		return string(task.Effort)
	case "group":
		return task.Group
	case "owner":
		return task.Owner
	case "parent":
		return task.Parent
	case "file":
		return task.FilePath
	default:
		return ""
	}
}
