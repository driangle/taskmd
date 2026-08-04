package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/metrics"
	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	statsFormat  string
	statsGroupBy string
)

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
	statsCmd.Flags().StringVar(&statsGroupBy, "group-by", "", "group tasks by field (phase)")
}

func runStats(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	// Scan for tasks
	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}

	// Calculate metrics
	m := metrics.Calculate(tasks)

	if statsGroupBy != "" && statsGroupBy != "phase" {
		return fmt.Errorf("unsupported group-by field: %s (supported: phase)", statsGroupBy)
	}

	// Output in requested format
	switch statsFormat {
	case "json":
		return outputStatsJSON(m)
	case "yaml":
		return WriteYAML(os.Stdout, m)
	case "table":
		return outputStatsTable(m, statsGroupBy)
	default:
		return ValidateFormat(statsFormat, []string{"table", "json", "yaml"})
	}
}

// outputStatsJSON outputs metrics as JSON
func outputStatsJSON(m *metrics.Metrics) error {
	return WriteJSON(os.Stdout, m)
}

// outputStatsTable outputs metrics in a human-readable table format
func outputStatsTable(m *metrics.Metrics, groupBy string) error {
	r := getRenderer()

	fmt.Println(formatLabel("TASK STATISTICS", r))
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	printStatsOverall(m)

	// Tasks by status
	fmt.Println(formatLabel("BY STATUS:", r))
	printStatsBreakdownByStatus(m, r)
	fmt.Println()

	// Tasks by priority
	fmt.Println(formatLabel("BY PRIORITY:", r))
	printStatsBreakdownByPriority(m, r)
	fmt.Println()

	// Tasks by effort
	fmt.Println(formatLabel("BY EFFORT:", r))
	printStatsBreakdownByEffort(m, r)

	// Tasks by phase (shown when --group-by phase or when phases exist)
	if groupBy == "phase" || len(m.TasksByPhase) > 0 {
		fmt.Println()
		fmt.Println(formatLabel("BY PHASE:", r))
		printStatsBreakdownByPhase(m)
	}

	return nil
}

// printStatsOverall prints the top-level totals table.
func printStatsOverall(m *metrics.Metrics) {
	tw := NewTableWriter()
	tw.AddRow([]string{"Total Tasks:", fmt.Sprintf("%d", m.TotalTasks)}, []string{"Total Tasks:", fmt.Sprintf("%d", m.TotalTasks)})
	tw.AddRow([]string{"Blocked Tasks:", fmt.Sprintf("%d", m.BlockedTasksCount)}, []string{"Blocked Tasks:", fmt.Sprintf("%d", m.BlockedTasksCount)})
	tw.AddRow([]string{"Critical Path Length:", fmt.Sprintf("%d", m.CriticalPathLength)}, []string{"Critical Path Length:", fmt.Sprintf("%d", m.CriticalPathLength)})
	tw.AddRow([]string{"Max Dependency Depth:", fmt.Sprintf("%d", m.MaxDependencyDepth)}, []string{"Max Dependency Depth:", fmt.Sprintf("%d", m.MaxDependencyDepth)})
	tw.AddRow([]string{"Avg Dependencies/Task:", fmt.Sprintf("%.2f", m.AvgDependenciesPerTask)}, []string{"Avg Dependencies/Task:", fmt.Sprintf("%.2f", m.AvgDependenciesPerTask)})
	tw.Flush(os.Stdout)
	fmt.Println()
}

func printStatsBreakdownByStatus(m *metrics.Metrics, r *lipgloss.Renderer) {
	if len(m.TasksByStatus) == 0 {
		fmt.Println("  (none)")
		return
	}
	tw := NewTableWriter()
	for _, status := range []model.Status{
		model.StatusPending, model.StatusInProgress, model.StatusCompleted,
		model.StatusBlocked, model.StatusCancelled,
	} {
		if count, ok := m.TasksByStatus[status]; ok && count > 0 {
			label := fmt.Sprintf("  %s:", string(status))
			colorLabel := fmt.Sprintf("  %s:", formatStatus(string(status), r))
			val := fmt.Sprintf("%d", count)
			tw.AddRow([]string{label, val}, []string{colorLabel, val})
		}
	}
	tw.Flush(os.Stdout)
}

func printStatsBreakdownByPriority(m *metrics.Metrics, r *lipgloss.Renderer) {
	if len(m.TasksByPriority) == 0 {
		fmt.Println("  (none)")
		return
	}
	tw := NewTableWriter()
	for _, priority := range []model.Priority{
		model.PriorityCritical, model.PriorityHigh, model.PriorityMedium, model.PriorityLow,
	} {
		if count, ok := m.TasksByPriority[priority]; ok && count > 0 {
			label := fmt.Sprintf("  %s:", string(priority))
			colorLabel := fmt.Sprintf("  %s:", formatPriority(string(priority), r))
			val := fmt.Sprintf("%d", count)
			tw.AddRow([]string{label, val}, []string{colorLabel, val})
		}
	}
	tw.Flush(os.Stdout)
}

func printStatsBreakdownByPhase(m *metrics.Metrics) {
	if len(m.TasksByPhase) == 0 {
		fmt.Println("  (none)")
		return
	}
	// Sort phase names alphabetically
	names := make([]string, 0, len(m.TasksByPhase))
	for name := range m.TasksByPhase {
		names = append(names, name)
	}
	sort.Strings(names)

	tw := NewTableWriter()
	for _, name := range names {
		label := fmt.Sprintf("  %s:", name)
		val := fmt.Sprintf("%d", m.TasksByPhase[name])
		tw.AddRow([]string{label, val}, []string{label, val})
	}
	tw.Flush(os.Stdout)
}

func printStatsBreakdownByEffort(m *metrics.Metrics, r *lipgloss.Renderer) {
	if len(m.TasksByEffort) == 0 {
		fmt.Println("  (none)")
		return
	}
	tw := NewTableWriter()
	for _, effort := range []model.Effort{
		model.EffortSmall, model.EffortMedium, model.EffortLarge,
	} {
		if count, ok := m.TasksByEffort[effort]; ok && count > 0 {
			label := fmt.Sprintf("  %s:", string(effort))
			colorLabel := fmt.Sprintf("  %s:", formatEffort(string(effort), r))
			val := fmt.Sprintf("%d", count)
			tw.AddRow([]string{label, val}, []string{colorLabel, val})
		}
	}
	tw.Flush(os.Stdout)
}
