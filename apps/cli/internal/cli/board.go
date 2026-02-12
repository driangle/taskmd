package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/driangle/md-task-tracker/apps/cli/internal/board"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
)

var (
	boardGroupBy string
	boardFormat  string
	boardOut     string
)

var boardCmd = &cobra.Command{
	Use:        "board",
	SuggestFor: []string{"kanban", "columns"},
	Short:      "Display tasks grouped in a kanban-like board view",
	Long: `Display tasks grouped by a field in a board/kanban-like view.

Supported group-by fields:
  - status: Group by task status (default)
  - priority: Group by priority level
  - effort: Group by effort estimate
  - group: Group by task group
  - tag: Group by tags (tasks may appear in multiple groups)

Supported formats:
  - md: Markdown sections (default)
  - txt: Plain text with dividers
  - json: JSON structure

Examples:
  taskmd board tasks/
  taskmd board tasks/ --group-by priority
  taskmd board tasks/ --group-by tag --format json
  taskmd board tasks/ --format txt --out board.txt`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBoard,
}

func init() {
	rootCmd.AddCommand(boardCmd)

	boardCmd.Flags().StringVar(&boardGroupBy, "group-by", "status", "field to group by (status, priority, effort, group, tag)")
	boardCmd.Flags().StringVar(&boardFormat, "format", "md", "output format (md, txt, json)")
	boardCmd.Flags().StringVarP(&boardOut, "out", "o", "", "write output to file instead of stdout")
}

func runBoard(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
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

	grouped, err := board.GroupTasks(result.Tasks, boardGroupBy)
	if err != nil {
		return err
	}

	var outFile *os.File
	if boardOut != "" {
		f, err := os.Create(boardOut)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	switch boardFormat {
	case "md":
		return outputBoardMarkdown(grouped, outFile)
	case "txt":
		return outputBoardText(grouped, outFile)
	case "json":
		return outputBoardJSON(grouped, outFile)
	default:
		return fmt.Errorf("unsupported format: %s (supported: md, txt, json)", boardFormat)
	}
}

func outputBoardMarkdown(gr *board.GroupResult, w io.Writer) error {
	for i, key := range gr.Keys {
		tasks := gr.Groups[key]
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "## %s (%d)\n\n", key, len(tasks))
		for _, t := range tasks {
			fmt.Fprintf(w, "- [%s] %s", t.ID, t.Title)
			if t.Priority != "" {
				fmt.Fprintf(w, " (priority: %s)", t.Priority)
			}
			fmt.Fprintln(w)
		}
	}
	return nil
}

func outputBoardText(gr *board.GroupResult, w io.Writer) error {
	for i, key := range gr.Keys {
		tasks := gr.Groups[key]
		if i > 0 {
			fmt.Fprintln(w)
		}
		header := fmt.Sprintf("%s (%d)", key, len(tasks))
		fmt.Fprintln(w, header)
		fmt.Fprintln(w, strings.Repeat("-", len(header)))
		for _, t := range tasks {
			fmt.Fprintf(w, "  %s  %s\n", t.ID, t.Title)
		}
	}
	return nil
}

func outputBoardJSON(gr *board.GroupResult, w io.Writer) error {
	out := board.ToJSON(gr)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
