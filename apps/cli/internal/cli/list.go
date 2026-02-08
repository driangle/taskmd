package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
)

var (
	listFilters []string
	listSort    string
	listColumns string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks in a quick textual format",
	Long: `List displays tasks in various formats with filtering and sorting support.

By default, scans the current directory and all subdirectories for markdown files
with task frontmatter. You can specify a different directory to scan.

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

	listCmd.Flags().StringArrayVar(&listFilters, "filter", []string{}, "filter tasks (can specify multiple times for AND conditions, e.g., --filter status=pending --filter priority=high)")
	listCmd.Flags().StringVar(&listSort, "sort", "", "sort by field (id, title, status, priority, effort, created)")
	listCmd.Flags().StringVar(&listColumns, "columns", "id,title,status,priority", "comma-separated list of columns to display")
}

func runList(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	// Create scanner and scan for tasks
	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks

	// Report any scan errors if verbose
	if flags.Verbose && len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: encountered %d errors during scan:\n", len(result.Errors))
		for _, scanErr := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
		}
		fmt.Fprintln(os.Stderr)
	}

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
	switch flags.Format {
	case "json":
		return outputJSON(tasks)
	case "yaml":
		return outputYAML(tasks)
	case "table":
		return outputTable(tasks, listColumns)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", flags.Format)
	}
}

// applyFilters applies multiple filter expressions to tasks (AND logic)
func applyFilters(tasks []*model.Task, filterExprs []string) ([]*model.Task, error) {
	// Parse all filter expressions first
	filters := make([]filterCriteria, 0, len(filterExprs))
	for _, expr := range filterExprs {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid filter format (expected field=value): %s", expr)
		}
		filters = append(filters, filterCriteria{
			field: strings.TrimSpace(parts[0]),
			value: strings.TrimSpace(parts[1]),
		})
	}

	// Apply all filters (AND logic)
	var filtered []*model.Task
	for _, task := range tasks {
		if matchesAllFilters(task, filters) {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

// filterCriteria represents a single filter condition
type filterCriteria struct {
	field string
	value string
}

// matchesAllFilters checks if a task matches all filter criteria (AND logic)
func matchesAllFilters(task *model.Task, filters []filterCriteria) bool {
	for _, filter := range filters {
		if !matchesFilter(task, filter.field, filter.value) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a task matches a filter criterion
func matchesFilter(task *model.Task, field, value string) bool {
	switch field {
	case "status":
		return string(task.Status) == value
	case "priority":
		return string(task.Priority) == value
	case "effort":
		return string(task.Effort) == value
	case "id":
		return task.ID == value
	case "title":
		return strings.Contains(strings.ToLower(task.Title), strings.ToLower(value))
	case "group":
		return task.Group == value
	case "blocked":
		// A task is blocked if it has dependencies
		isBlocked := len(task.Dependencies) > 0
		return (value == "true" && isBlocked) || (value == "false" && !isBlocked)
	case "tag":
		// Check if task has the specified tag
		for _, tag := range task.Tags {
			if tag == value {
				return true
			}
		}
		return false
	default:
		return false
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
		return fmt.Errorf("unsupported sort field: %s (supported: id, title, status, priority, effort, created)", sortField)
	}

	return nil
}

// outputJSON outputs tasks as JSON
func outputJSON(tasks []*model.Task) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}

// outputYAML outputs tasks as YAML
func outputYAML(tasks []*model.Task) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(tasks)
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
			row[i] = getColumnValue(task, col)
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	return nil
}

// getColumnValue extracts the value for a specific column from a task
func getColumnValue(task *model.Task, column string) string {
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
	case "created":
		if task.Created.IsZero() {
			return ""
		}
		return task.Created.Format("2006-01-02")
	case "deps":
		if len(task.Dependencies) == 0 {
			return ""
		}
		return strings.Join(task.Dependencies, ",")
	case "tags":
		if len(task.Tags) == 0 {
			return ""
		}
		return strings.Join(task.Tags, ",")
	default:
		return ""
	}
}
