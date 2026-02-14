package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/board"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	reportFormat       string
	reportGroupBy      string
	reportOut          string
	reportIncludeGraph bool
)

var reportCmd = &cobra.Command{
	Use:   "report [path]",
	Short: "Generate a comprehensive project report",
	Long: `Generate a comprehensive report combining summary statistics, task groupings,
critical path analysis, blocked tasks, and optional dependency graphs.

Supported formats:
  - md: Rich markdown report (default)
  - html: Self-contained HTML report with inline CSS
  - json: Structured JSON report

Supported group-by fields:
  - status: Group by task status (default)
  - priority: Group by priority level
  - effort: Group by effort estimate
  - group: Group by task group
  - tag: Group by tags

Examples:
  taskmd report tasks/
  taskmd report tasks/ --format html --include-graph --out report.html
  taskmd report tasks/ --group-by priority --format json
  taskmd report --format md --out report.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runReport,
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVar(&reportFormat, "format", "md", "output format (md, html, json)")
	reportCmd.Flags().StringVar(&reportGroupBy, "group-by", "status", "field to group by (status, priority, effort, group, tag)")
	reportCmd.Flags().StringVarP(&reportOut, "out", "o", "", "write output to file instead of stdout")
	reportCmd.Flags().BoolVar(&reportIncludeGraph, "include-graph", false, "embed dependency graph in report")
}

func runReport(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	if err := ValidateFormat(reportFormat, []string{"md", "html", "json"}); err != nil {
		return err
	}

	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if flags.Verbose && len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "\nWarning: encountered %d errors during scan:\n", len(result.Errors))
		for _, scanErr := range result.Errors {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", scanErr.FilePath, scanErr.Error)
		}
		fmt.Fprintln(os.Stderr)
	}

	data, err := collectReportData(result.Tasks, reportGroupBy, reportIncludeGraph)
	if err != nil {
		return err
	}

	var outFile *os.File
	if reportOut != "" {
		f, err := os.Create(reportOut)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	switch reportFormat {
	case "md":
		return outputReportMarkdown(data, outFile)
	case "html":
		return outputReportHTML(data, outFile)
	case "json":
		return outputReportJSON(data, outFile)
	default:
		return fmt.Errorf("unsupported format: %s", reportFormat)
	}
}

// ReportTaskJSON is the JSON representation of a task in report sections.
type ReportTaskJSON struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

func taskToReportJSON(t reportTask) ReportTaskJSON {
	var deps []string
	if len(t.Dependencies) > 0 {
		deps = t.Dependencies
	}
	return ReportTaskJSON{
		ID:           t.ID,
		Title:        t.Title,
		Status:       t.Status,
		Priority:     t.Priority,
		Dependencies: deps,
	}
}

func outputReportJSON(data *reportData, outFile *os.File) error {
	type reportJSON struct {
		Summary      any               `json:"summary"`
		Groups       []board.JSONGroup `json:"groups"`
		GroupBy      string            `json:"group_by"`
		CriticalPath []ReportTaskJSON  `json:"critical_path"`
		BlockedTasks []ReportTaskJSON  `json:"blocked_tasks"`
		Graph        map[string]any    `json:"graph,omitempty"`
	}

	cpTasks := make([]ReportTaskJSON, len(data.CriticalPath))
	for i, t := range data.CriticalPath {
		cpTasks[i] = taskToReportJSON(t)
	}

	blockedTasks := make([]ReportTaskJSON, len(data.BlockedTasks))
	for i, t := range data.BlockedTasks {
		blockedTasks[i] = taskToReportJSON(t)
	}

	rj := reportJSON{
		Summary:      data.Metrics,
		Groups:       board.ToJSON(data.GroupedTasks),
		GroupBy:      data.GroupBy,
		CriticalPath: cpTasks,
		BlockedTasks: blockedTasks,
	}

	if data.IncludeGraph {
		rj.Graph = data.GraphJSON
	}

	return WriteJSON(outFile, rj)
}
