package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var searchFormat string

type searchResult struct {
	ID            string `json:"id" yaml:"id"`
	Title         string `json:"title" yaml:"title"`
	Status        string `json:"status" yaml:"status"`
	FilePath      string `json:"file_path" yaml:"file_path"`
	MatchLocation string `json:"match_location" yaml:"match_location"`
	Snippet       string `json:"snippet" yaml:"snippet"`
}

var searchCmd = &cobra.Command{
	Use:        "search <query>",
	SuggestFor: []string{"find", "grep"},
	Short:      "Full-text search across task titles and bodies",
	Long: `Search performs case-insensitive full-text search across all task titles
and markdown body content. Results show where the match was found and a
context snippet.

Output formats: table (default), json, yaml

Examples:
  taskmd search "authentication"
  taskmd search deploy --format json
  taskmd search "bug fix" --task-dir ./tasks`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchFormat, "format", "table", "output format (table, json, yaml)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	query := args[0]

	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks
	makeFilePathsRelative(tasks, scanDir)

	results := searchTasks(tasks, query)

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

func searchTasks(tasks []*model.Task, query string) []searchResult {
	lowerQuery := strings.ToLower(query)
	var results []searchResult

	for _, task := range tasks {
		titleMatch := strings.Contains(strings.ToLower(task.Title), lowerQuery)
		bodyMatch := strings.Contains(strings.ToLower(task.Body), lowerQuery)

		if !titleMatch && !bodyMatch {
			continue
		}

		location := matchLocation(titleMatch, bodyMatch)
		snippet := extractSnippet(task, lowerQuery, bodyMatch)

		results = append(results, searchResult{
			ID:            task.ID,
			Title:         task.Title,
			Status:        string(task.Status),
			FilePath:      task.FilePath,
			MatchLocation: location,
			Snippet:       snippet,
		})
	}

	return results
}

func matchLocation(titleMatch, bodyMatch bool) string {
	if titleMatch && bodyMatch {
		return "title,body"
	}
	if titleMatch {
		return "title"
	}
	return "body"
}

func extractSnippet(task *model.Task, lowerQuery string, bodyMatch bool) string {
	if bodyMatch {
		return extractBodySnippet(task.Body, lowerQuery)
	}
	return task.Title
}

func extractBodySnippet(body, lowerQuery string) string {
	lowerBody := strings.ToLower(body)
	idx := strings.Index(lowerBody, lowerQuery)
	if idx < 0 {
		return ""
	}

	const radius = 40

	start := max(idx-radius, 0)
	end := min(idx+len(lowerQuery)+radius, len(body))

	// Trim to word boundaries
	if start > 0 {
		if spaceIdx := strings.IndexByte(body[start:], ' '); spaceIdx >= 0 {
			start += spaceIdx + 1
		}
	}
	if end < len(body) {
		if spaceIdx := strings.LastIndexByte(body[:end], ' '); spaceIdx >= 0 {
			end = spaceIdx
		}
	}

	snippet := body[start:end]
	// Collapse whitespace
	snippet = strings.Join(strings.Fields(snippet), " ")

	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(body) {
		snippet = snippet + "..."
	}

	return snippet
}

func outputSearchTable(results []searchResult, query string) error {
	r := getRenderer()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tMATCH\tSNIPPET")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		"----------", "----------", "----------", "----------", "----------")

	for _, res := range results {
		snippet := highlightMatch(res.Snippet, query, r)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			formatTaskID(res.ID, r),
			res.Title,
			formatStatus(res.Status, r),
			res.MatchLocation,
			snippet,
		)
	}

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
