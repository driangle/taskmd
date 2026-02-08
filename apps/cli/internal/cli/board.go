package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	boardGroupBy string
	boardFormat  string
	boardOut     string
)

var boardCmd = &cobra.Command{
	Use:   "board [directory]",
	Short: "Display tasks grouped in a kanban-like board view",
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

// groupResult holds ordered group keys and the grouped tasks map.
type groupResult struct {
	Keys   []string
	Groups map[string][]*model.Task
}

func runBoard(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := "."
	if len(args) > 0 {
		scanDir = args[0]
	}

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

	grouped, err := groupTasks(result.Tasks, boardGroupBy)
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

func groupTasks(tasks []*model.Task, field string) (*groupResult, error) {
	groups := make(map[string][]*model.Task)

	switch field {
	case "status":
		for _, t := range tasks {
			key := string(t.Status)
			if key == "" {
				key = "(none)"
			}
			groups[key] = append(groups[key], t)
		}
		return &groupResult{
			Keys:   orderedKeys(groups, statusOrder()),
			Groups: groups,
		}, nil

	case "priority":
		for _, t := range tasks {
			key := string(t.Priority)
			if key == "" {
				key = "(none)"
			}
			groups[key] = append(groups[key], t)
		}
		return &groupResult{
			Keys:   orderedKeys(groups, priorityOrder()),
			Groups: groups,
		}, nil

	case "effort":
		for _, t := range tasks {
			key := string(t.Effort)
			if key == "" {
				key = "(none)"
			}
			groups[key] = append(groups[key], t)
		}
		return &groupResult{
			Keys:   orderedKeys(groups, effortOrder()),
			Groups: groups,
		}, nil

	case "group":
		for _, t := range tasks {
			key := t.GetGroup()
			if key == "" {
				key = "(none)"
			}
			groups[key] = append(groups[key], t)
		}
		return &groupResult{
			Keys:   sortedKeys(groups),
			Groups: groups,
		}, nil

	case "tag":
		for _, t := range tasks {
			if len(t.Tags) == 0 {
				groups["(none)"] = append(groups["(none)"], t)
			} else {
				for _, tag := range t.Tags {
					groups[tag] = append(groups[tag], t)
				}
			}
		}
		return &groupResult{
			Keys:   sortedKeys(groups),
			Groups: groups,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported group-by field: %s (supported: status, priority, effort, group, tag)", field)
	}
}

func statusOrder() []string {
	return []string{
		string(model.StatusPending),
		string(model.StatusInProgress),
		string(model.StatusBlocked),
		string(model.StatusCompleted),
	}
}

func priorityOrder() []string {
	return []string{
		string(model.PriorityCritical),
		string(model.PriorityHigh),
		string(model.PriorityMedium),
		string(model.PriorityLow),
	}
}

func effortOrder() []string {
	return []string{
		string(model.EffortSmall),
		string(model.EffortMedium),
		string(model.EffortLarge),
	}
}

// orderedKeys returns group keys in a predefined order, appending any extra keys alphabetically.
func orderedKeys(groups map[string][]*model.Task, order []string) []string {
	var keys []string
	seen := make(map[string]bool)

	for _, k := range order {
		if _, ok := groups[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}

	// Append any keys not in the predefined order (e.g., "(none)")
	var extra []string
	for k := range groups {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sort.Strings(extra)
	keys = append(keys, extra...)

	return keys
}

// sortedKeys returns map keys sorted alphabetically, with "(none)" last.
func sortedKeys(groups map[string][]*model.Task) []string {
	var keys []string
	hasNone := false
	for k := range groups {
		if k == "(none)" {
			hasNone = true
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if hasNone {
		keys = append(keys, "(none)")
	}
	return keys
}

func outputBoardMarkdown(gr *groupResult, w io.Writer) error {
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

func outputBoardText(gr *groupResult, w io.Writer) error {
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

type boardJSONGroup struct {
	Group string          `json:"group"`
	Count int             `json:"count"`
	Tasks []boardJSONTask `json:"tasks"`
}

type boardJSONTask struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority,omitempty"`
	Effort   string `json:"effort,omitempty"`
}

func outputBoardJSON(gr *groupResult, w io.Writer) error {
	var out []boardJSONGroup
	for _, key := range gr.Keys {
		tasks := gr.Groups[key]
		jTasks := make([]boardJSONTask, len(tasks))
		for i, t := range tasks {
			jTasks[i] = boardJSONTask{
				ID:       t.ID,
				Title:    t.Title,
				Status:   string(t.Status),
				Priority: string(t.Priority),
				Effort:   string(t.Effort),
			}
		}
		out = append(out, boardJSONGroup{
			Group: key,
			Count: len(tasks),
			Tasks: jTasks,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
