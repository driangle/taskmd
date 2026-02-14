package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// statsCmd represents the stats command
var statsFormat string

var statsCmd = &cobra.Command{
	Use:        "stats",
	SuggestFor: []string{"summary", "status", "overview"},
	Short:      "Show computed metrics about tasks",
	Long: `Stats displays computed metrics about your task set including:
- Total tasks and breakdown by status, priority, and effort
- Blocked tasks count
- Critical path length (longest dependency chain)
- Maximum dependency depth
- Average dependencies per task

By default, scans the current directory and all subdirectories for markdown files
with task frontmatter. You can specify a different directory to scan.

Output formats: table (default), json, yaml

Examples:
  taskmd stats
  taskmd stats ./tasks
  taskmd stats --format json
  taskmd stats --format yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringVar(&statsFormat, "format", "table", "output format (table, json, yaml)")
}

func runStats(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	// Create scanner and scan for tasks
	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
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

	// Calculate metrics
	m := metrics.Calculate(tasks)

	// Output in requested format
	switch statsFormat {
	case "json":
		return outputStatsJSON(m)
	case "yaml":
		return WriteYAML(os.Stdout, m)
	case "table":
		return outputStatsTable(m)
	default:
		return ValidateFormat(statsFormat, []string{"table", "json", "yaml"})
	}
}

// outputStatsJSON outputs metrics as JSON
func outputStatsJSON(m *metrics.Metrics) error {
	return WriteJSON(os.Stdout, m)
}

// outputStatsTable outputs metrics in a human-readable table format
//
//nolint:gocognit,funlen // TODO: refactor to reduce complexity
func outputStatsTable(m *metrics.Metrics) error {
	r := getRenderer()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, formatLabel("TASK STATISTICS", r))
	fmt.Fprintln(w, strings.Repeat("=", 50))
	fmt.Fprintln(w)

	// Overall stats
	fmt.Fprintf(w, "Total Tasks:\t%d\n", m.TotalTasks)
	fmt.Fprintf(w, "Blocked Tasks:\t%d\n", m.BlockedTasksCount)
	fmt.Fprintf(w, "Critical Path Length:\t%d\n", m.CriticalPathLength)
	fmt.Fprintf(w, "Max Dependency Depth:\t%d\n", m.MaxDependencyDepth)
	fmt.Fprintf(w, "Avg Dependencies/Task:\t%.2f\n", m.AvgDependenciesPerTask)
	fmt.Fprintln(w)

	// Tasks by status
	fmt.Fprintln(w, formatLabel("BY STATUS:", r))
	if len(m.TasksByStatus) > 0 {
		// Order: pending, in-progress, completed, blocked, cancelled
		statusOrder := []model.Status{
			model.StatusPending,
			model.StatusInProgress,
			model.StatusCompleted,
			model.StatusBlocked,
			model.StatusCancelled,
		}
		for _, status := range statusOrder {
			if count, ok := m.TasksByStatus[status]; ok && count > 0 {
				fmt.Fprintf(w, "  %s:\t%d\n", formatStatus(string(status), r), count)
			}
		}
	} else {
		fmt.Fprintln(w, "  (none)")
	}
	fmt.Fprintln(w)

	// Tasks by priority
	fmt.Fprintln(w, formatLabel("BY PRIORITY:", r))
	if len(m.TasksByPriority) > 0 {
		// Order: critical, high, medium, low
		priorityOrder := []model.Priority{
			model.PriorityCritical,
			model.PriorityHigh,
			model.PriorityMedium,
			model.PriorityLow,
		}
		for _, priority := range priorityOrder {
			if count, ok := m.TasksByPriority[priority]; ok && count > 0 {
				fmt.Fprintf(w, "  %s:\t%d\n", formatPriority(string(priority), r), count)
			}
		}
	} else {
		fmt.Fprintln(w, "  (none)")
	}
	fmt.Fprintln(w)

	// Tasks by effort
	fmt.Fprintln(w, formatLabel("BY EFFORT:", r))
	if len(m.TasksByEffort) > 0 {
		// Order: small, medium, large
		effortOrder := []model.Effort{
			model.EffortSmall,
			model.EffortMedium,
			model.EffortLarge,
		}
		for _, effort := range effortOrder {
			if count, ok := m.TasksByEffort[effort]; ok && count > 0 {
				fmt.Fprintf(w, "  %s:\t%d\n", formatEffort(string(effort), r), count)
			}
		}
	} else {
		fmt.Fprintln(w, "  (none)")
	}

	return nil
}
