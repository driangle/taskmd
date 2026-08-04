package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/search"
)

var (
	searchFormat  string
	searchFilters []string
	searchSort    string
	searchLimit   int
)

var searchCmd = &cobra.Command{
	Use:        "search <query>",
	SuggestFor: []string{"find", "grep"},
	Short:      "Full-text search across task titles and bodies",
	Long: `Search performs case-insensitive full-text search across all task titles
and markdown body content. Results show where the match was found and a
context snippet.

Output formats: table (default), json, yaml

Multiple --filter flags are combined with AND logic.

Examples:
  taskmd search "authentication"
  taskmd search deploy --format json
  taskmd search "bug fix" --task-dir ./tasks
  taskmd search "auth" --filter priority=high
  taskmd search "deploy" --filter status=pending --sort priority --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchFormat, "format", "table", "output format (table, json, yaml)")
	searchCmd.Flags().StringArrayVar(&searchFilters, "filter", []string{}, "filter tasks (can specify multiple times for AND conditions, e.g., --filter status=pending --filter priority=high)")
	searchCmd.Flags().StringVar(&searchSort, "sort", "", "sort by field (id, title, status, priority, effort, created)")
	searchCmd.Flags().IntVar(&searchLimit, "limit", 0, "maximum number of results to display (0 = unlimited)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	query := args[0]

	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}
	makeFilePathsRelative(tasks, scanDir)

	// Apply filters before search (narrows which tasks are searched)
	if len(searchFilters) > 0 {
		tasks, err = applyFilters(tasks, searchFilters)
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}

	// Apply sorting before search (preserves order through search)
	if searchSort != "" {
		if err := sortTasks(tasks, searchSort); err != nil {
			return fmt.Errorf("sort error: %w", err)
		}
	}

	results := search.Search(tasks, query)

	// Apply limit after search (limits final result count)
	if searchLimit > 0 && searchLimit < len(results) {
		results = results[:searchLimit]
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "No tasks found matching %q\n", query)
		return nil
	}

	switch searchFormat {
	case "json":
		return WriteJSON(os.Stdout, results)
	case "yaml":
		return WriteYAML(os.Stdout, results)
	case "table":
		return outputSearchTable(results, query)
	default:
		return ValidateFormat(searchFormat, []string{"table", "json", "yaml"})
	}
}

func outputSearchTable(results []search.Result, query string) error {
	r := getRenderer()
	tw := NewTableWriter()

	tw.AddHeader([]string{"ID", "TITLE", "STATUS", "PRIORITY", "MATCH", "SNIPPET"})
	tw.AddSeparator()

	for _, res := range results {
		plain := []string{res.ID, res.Title, res.Status, res.Priority, res.MatchLocation, res.Snippet}
		colored := []string{
			formatTaskID(res.ID, r),
			res.Title,
			formatStatus(res.Status, r),
			formatPriority(res.Priority, r),
			res.MatchLocation,
			highlightMatch(res.Snippet, query, r),
		}
		tw.AddRow(plain, colored)
	}

	tw.Flush(os.Stdout)
	return nil
}

func highlightMatch(text, query string, r *lipgloss.Renderer) string {
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerText, lowerQuery)
	if idx < 0 {
		return text
	}

	style := r.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // Yellow bold
	before := text[:idx]
	match := text[idx : idx+len(query)]
	after := text[idx+len(query):]

	return before + style.Render(match) + after
}
