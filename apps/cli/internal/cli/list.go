package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	listFormat   string
	listFilters  []string
	listSort     string
	listReverse  bool
	listColumns  string
	listLimit    int
	listScope    string
	listPhase    string
	listStatus   string
	listPriority string
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

Priority and effort support comparison operators (>=, >, <=, <).

Use the sentinel values 'none' and 'any' to match by field presence:
'field=none' matches tasks where the field is unset, 'field=any' where it is set.

Sorting is ascending by default. Use --reverse / -r to flip the order,
following the Unix convention (sort -r, ls -r). With no --sort, --reverse
reverses the default file-scan order.

Examples:
  taskmd list
  taskmd list ./tasks
  taskmd list --status pending
  taskmd list --status pending --priority high
  taskmd list --filter status=pending --filter priority=high
  taskmd list --filter "priority>=medium"
  taskmd list --filter "effort<large"
  taskmd list --filter phase=none
  taskmd list --filter owner=any
  taskmd list --phase none
  taskmd list --phase none --status pending
  taskmd list --sort priority
  taskmd list --sort priority --reverse
  taskmd list --sort created_at -r
  taskmd list --columns id,title,deps
  taskmd list --format json
  taskmd list --scope cli
  taskmd list --scope "web*"
  taskmd list --sort priority --limit 5`,
	Args: cobra.MaximumNArgs(1),
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listFormat, "format", "table", "output format (table, json, yaml)")
	listCmd.Flags().StringArrayVar(&listFilters, "filter", []string{}, "filter tasks (e.g., --filter status=pending --filter \"priority>=medium\"); supports >=, >, <=, < for priority and effort")
	listCmd.Flags().StringVar(&listSort, "sort", "", "sort by field (id, title, status, priority, effort, created_at)")
	listCmd.Flags().BoolVarP(&listReverse, "reverse", "r", false, "reverse the sort order (or the default file-scan order when --sort is omitted)")
	listCmd.Flags().StringVar(&listColumns, "columns", "id,title,status,priority,file", "comma-separated list of columns to display")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "maximum number of tasks to display (0 = unlimited)")
	listCmd.Flags().StringVar(&listScope, "scope", "", "filter by scope; supports wildcards (e.g. cli, cli*)")
	listCmd.Flags().StringVar(&listPhase, "phase", "", "filter by phase (use 'none'/'any' to match tasks without/with a phase)")
	listCmd.Flags().StringVar(&listStatus, "status", "", "shortcut for --filter status=<value>")
	listCmd.Flags().StringVar(&listPriority, "priority", "", "shortcut for --filter priority=<value>")
}

func runList(cmd *cobra.Command, args []string) error {
	if allProjectsFlag {
		return runListAllProjects()
	}

	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)
	debugLog("scan directory: %s", scanDir)

	// Scan for tasks; the overlay is built before paths are relativized —
	// the merge stats files by path.
	tasks, overlay, err := scanTasksWithOverlay(scanDir, flags)
	if err != nil {
		return err
	}
	debugLog("found %d task(s)", len(tasks))
	makeFilePathsRelative(tasks, scanDir)

	if overlay != nil {
		return runListOverlay(overlay)
	}

	debugLog("format: %s, sort: %q, filters: %v", listFormat, listSort, listFilters)

	tasks, err = applyListFiltersAndSort(tasks)
	if err != nil {
		return err
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

func runListAllProjects() error {
	ptasks, err := scanAllProjects()
	if err != nil {
		return err
	}

	// Extract plain tasks for filtering/sorting
	tasks := make([]*model.Task, len(ptasks))
	for i, pt := range ptasks {
		tasks[i] = pt.Task
	}

	tasks, err = applyListFiltersAndSort(tasks)
	if err != nil {
		return err
	}

	// Rebuild project task index after filtering
	taskIndex := make(map[*model.Task]*ProjectTask, len(ptasks))
	for _, pt := range ptasks {
		taskIndex[pt.Task] = pt
	}

	filtered := make([]*ProjectTask, 0, len(tasks))
	for _, t := range tasks {
		if pt, ok := taskIndex[t]; ok {
			filtered = append(filtered, pt)
		}
	}

	switch listFormat {
	case "json":
		return outputProjectJSON(filtered)
	case "yaml":
		return outputProjectYAML(filtered)
	case "table":
		return outputProjectTable(filtered, listColumns)
	default:
		return ValidateFormat(listFormat, []string{"table", "json", "yaml"})
	}
}

// projectTaskOutput is the JSON/YAML representation for --all-projects output.
type projectTaskOutput struct {
	Project string `json:"project" yaml:"project"`
	*model.Task
}

func outputProjectJSON(ptasks []*ProjectTask) error {
	out := make([]projectTaskOutput, len(ptasks))
	for i, pt := range ptasks {
		out[i] = projectTaskOutput{Project: pt.ProjectID, Task: pt.Task}
	}
	if len(out) == 0 {
		return WriteJSON(os.Stdout, []projectTaskOutput{})
	}
	return WriteJSON(os.Stdout, out)
}

func outputProjectYAML(ptasks []*ProjectTask) error {
	out := make([]projectTaskOutput, len(ptasks))
	for i, pt := range ptasks {
		out[i] = projectTaskOutput{Project: pt.ProjectID, Task: pt.Task}
	}
	return WriteYAML(os.Stdout, out)
}

func outputProjectTable(ptasks []*ProjectTask, columnsStr string) error {
	if len(ptasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}
	columns = injectProjectColumn(columns)

	r := getRenderer()
	tw := NewTableWriter()
	tw.AddHeader(columns)
	tw.AddSeparator()

	for _, pt := range ptasks {
		plain := make([]string, len(columns))
		colored := make([]string, len(columns))
		for i, col := range columns {
			if col == "project" {
				plain[i] = pt.ProjectID
				colored[i] = pt.ProjectID
			} else if col == "id" {
				plain[i] = pt.QualifiedID()
				colored[i] = formatTaskID(pt.QualifiedID(), r)
			} else {
				plain[i] = getColumnValue(pt.Task, col)
				colored[i] = colorizeColumn(pt.Task, col, r)
			}
		}
		tw.AddRow(plain, colored)
	}

	tw.Flush(os.Stdout)
	return nil
}

// applyListFiltersAndSort applies all list filters, sorting, and limit.
func applyListFiltersAndSort(tasks []*model.Task) ([]*model.Task, error) {
	var err error

	tasks, err = applyShortcutFilters(tasks, FilterShortcuts{
		Status:   listStatus,
		Priority: listPriority,
		Phase:    listPhase,
		Scope:    listScope,
		Filters:  listFilters,
	})
	if err != nil {
		return nil, err
	}

	// Apply sorting
	if listSort != "" {
		if err := sortTasks(tasks, listSort); err != nil {
			return nil, fmt.Errorf("sort error: %w", err)
		}
	}

	// Reverse ordering when requested (after sorting, before limit so
	// --limit keeps the top N of the reversed order). With no --sort this
	// reverses the default file-scan order.
	if listReverse {
		reverseTasks(tasks)
	}

	// Apply limit (after sorting)
	if listLimit > 0 && listLimit < len(tasks) {
		tasks = tasks[:listLimit]
	}

	return tasks, nil
}

// effortSortRank orders an effort value by its position in the project's
// vocabulary. Unset and unrecognized values sort last, after every known value.
func effortSortRank(scale effort.Scale, value model.Effort) int {
	if rank := scale.Rank(string(value)); rank >= 0 {
		return rank
	}
	return scale.Len()
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
		scale := resolveEffortScale()
		sort.Slice(tasks, func(i, j int) bool {
			return effortSortRank(scale, tasks[i].Effort) < effortSortRank(scale, tasks[j].Effort)
		})
	case "created", "created_at":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Created.Before(tasks[j].Created.Time)
		})
	default:
		return invalidValueError("sort field", sortField, validSortFields)
	}

	return nil
}

// reverseTasks reverses the order of the tasks slice in place.
func reverseTasks(tasks []*model.Task) {
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
}

// outputJSON outputs tasks as JSON
func outputJSON(tasks []*model.Task) error {
	if tasks == nil {
		tasks = []*model.Task{}
	}
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

	columns := strings.Split(columnsStr, ",")
	for i, col := range columns {
		columns[i] = strings.TrimSpace(col)
	}

	r := getRenderer()
	tw := NewTableWriter()
	tw.AddHeader(columns)
	tw.AddSeparator()

	for _, task := range tasks {
		plain := make([]string, len(columns))
		colored := make([]string, len(columns))
		for i, col := range columns {
			plain[i] = getColumnValue(task, col)
			colored[i] = colorizeColumn(task, col, r)
		}
		tw.AddRow(plain, colored)
	}

	tw.Flush(os.Stdout)
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
	case "created", "created_at":
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
	case "type":
		return string(task.Type)
	case "group":
		return task.Group
	case "owner":
		return task.Owner
	case "parent":
		return task.Parent
	case "phase":
		return task.Phase
	case "file":
		return task.FilePath
	default:
		return ""
	}
}
