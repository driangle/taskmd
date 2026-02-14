package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/next"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// Recommendation is re-exported from the shared package.
type Recommendation = next.Recommendation

var (
	nextFormat    string
	nextLimit     int
	nextFilters   []string
	nextQuickWins bool
	nextCritical  bool
)

var nextCmd = &cobra.Command{
	Use:        "next",
	SuggestFor: []string{"pick", "suggest", "what"},
	Short:      "Recommend what task to work on next",
	Long: `Next analyzes all tasks and recommends the best ones to work on next.

Tasks are scored based on priority, critical path position, downstream impact,
and effort. Only actionable tasks (pending or in-progress with all dependencies
completed) are shown.

Output formats: table (default), json, yaml

Examples:
  taskmd next
  taskmd next ./tasks
  taskmd next --limit 3
  taskmd next --filter tag=cli
  taskmd next --filter priority=high --format json
  taskmd next --quick-wins
  taskmd next --critical --limit 1`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNext,
}

func init() {
	rootCmd.AddCommand(nextCmd)

	nextCmd.Flags().StringVar(&nextFormat, "format", "table", "output format (table, json, yaml)")
	nextCmd.Flags().IntVar(&nextLimit, "limit", 5, "maximum number of recommendations")
	nextCmd.Flags().StringArrayVar(&nextFilters, "filter", []string{}, "filter tasks (e.g., --filter tag=cli)")
	nextCmd.Flags().BoolVar(&nextQuickWins, "quick-wins", false, "show only quick wins (effort: small)")
	nextCmd.Flags().BoolVar(&nextCritical, "critical", false, "show only critical path tasks")
}

func runNext(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	allTasks := result.Tasks
	makeFilePathsRelative(allTasks, scanDir)

	recs, err := next.Recommend(allTasks, next.Options{
		Limit:     nextLimit,
		Filters:   nextFilters,
		QuickWins: nextQuickWins,
		Critical:  nextCritical,
	})
	if err != nil {
		return err
	}

	switch nextFormat {
	case "json":
		return outputNextJSON(recs)
	case "yaml":
		return outputNextYAML(recs)
	case "table":
		return outputNextTable(recs)
	default:
		return ValidateFormat(nextFormat, []string{"table", "json", "yaml"})
	}
}

func outputNextJSON(recs []Recommendation) error {
	return WriteJSON(os.Stdout, recs)
}

func outputNextYAML(recs []Recommendation) error {
	return WriteYAML(os.Stdout, recs)
}

func outputNextTable(recs []Recommendation) error {
	r := getRenderer()

	if len(recs) == 0 {
		if nextQuickWins {
			fmt.Println("No quick wins available.")
		} else if nextCritical {
			fmt.Println("No critical path tasks available.")
		} else {
			fmt.Println("No actionable tasks found.")
		}
		return nil
	}

	if nextQuickWins {
		fmt.Println(formatLabel("Recommended quick wins:", r))
	} else if nextCritical {
		fmt.Println(formatLabel("Recommended critical path tasks:", r))
	} else {
		fmt.Println(formatLabel("Recommended tasks:", r))
	}
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "#\tID\tTitle\tPriority\tFile\tReason")
	fmt.Fprintln(w, "-\t--\t-----\t--------\t----\t------")

	for _, rec := range recs {
		reason := strings.Join(rec.Reasons, ", ")
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			rec.Rank,
			formatTaskID(rec.ID, r),
			rec.Title,
			formatPriority(rec.Priority, r),
			formatDim(rec.FilePath, r),
			reason,
		)
	}

	return nil
}
