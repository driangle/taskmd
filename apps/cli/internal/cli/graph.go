package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/graph"
	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	graphFormat        string
	graphRoot          string
	graphFocus         string
	graphUpstream      bool
	graphDownstream    bool
	graphOut           string
	graphExcludeStatus []string
	graphAll           bool
	graphFilters       []string
	graphScope         string
	graphStatus        string
	graphPriority      string
	graphPhase         string
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:        "graph",
	SuggestFor: []string{"deps", "dependencies", "tree"},
	Short:      "Export task dependency graph",
	Long: `Export task dependency graphs in various formats for visualization and analysis.

Supported formats:
  - mermaid: Mermaid diagram syntax (default)
  - dot: Graphviz DOT format
  - ascii: ASCII art tree
  - json: JSON graph structure

Multiple --filter flags are combined with AND logic.

Use the sentinel values 'none' and 'any' to match by field presence:
'field=none' matches tasks where the field is unset, 'field=any' where it is set.

Examples:
  taskmd graph > deps.mmd
  taskmd graph --format dot | dot -Tpng > graph.png
  taskmd graph --format ascii
  taskmd graph --root 022 --downstream
  taskmd graph --focus 022 --format mermaid
  taskmd graph --all --format ascii
  taskmd graph --filter priority=high
  taskmd graph --filter priority=high --filter effort=small
  taskmd graph --filter tag=cli --exclude-status completed
  taskmd graph --filter phase=none
  taskmd graph --status pending
  taskmd graph --priority high
  taskmd graph --phase web-ui
  taskmd graph --phase none
  taskmd graph --scope cli
  taskmd graph --scope "web*" --format mermaid

By default, completed tasks are excluded. Use --all to include them.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGraph,
}

func init() {
	rootCmd.AddCommand(graphCmd)

	graphCmd.Flags().StringVar(&graphFormat, "format", "ascii", "output format (mermaid, dot, ascii, json)")
	graphCmd.Flags().StringVar(&graphRoot, "root", "", "start graph from specific task ID")
	graphCmd.Flags().StringVar(&graphFocus, "focus", "", "highlight specific task ID")
	graphCmd.Flags().BoolVar(&graphUpstream, "upstream", false, "show only dependencies (ancestors)")
	graphCmd.Flags().BoolVar(&graphDownstream, "downstream", false, "show only dependents (descendants)")
	graphCmd.Flags().StringSliceVar(&graphExcludeStatus, "exclude-status", []string{"completed"}, "exclude tasks with status (completed, pending, in-progress, blocked, cancelled)")
	graphCmd.Flags().BoolVar(&graphAll, "all", false, "include all tasks (overrides --exclude-status)")
	graphCmd.Flags().StringVarP(&graphOut, "out", "o", "", "write output to file instead of stdout")
	graphCmd.Flags().StringArrayVar(&graphFilters, "filter", []string{}, "filter tasks (can specify multiple times for AND conditions, e.g., --filter priority=high --filter effort=small)")
	graphCmd.Flags().StringVar(&graphScope, "scope", "", "filter by scope; supports wildcards (e.g. cli, cli*)")
	graphCmd.Flags().StringVar(&graphStatus, "status", "", "shortcut for --filter status=<value>")
	graphCmd.Flags().StringVar(&graphPriority, "priority", "", "shortcut for --filter priority=<value>")
	graphCmd.Flags().StringVar(&graphPhase, "phase", "", "filter by phase (use 'none'/'any' to match tasks without/with a phase)")
}

func runGraph(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	// Validate conflicting flags
	if graphUpstream && graphDownstream {
		return fmt.Errorf("cannot use both --upstream and --downstream")
	}

	// --all overrides --exclude-status
	if graphAll {
		graphExcludeStatus = []string{}
	}

	tasks, err := scanTasks(ResolveScanDir(args), flags)
	if err != nil {
		return err
	}

	tasks, err = filterGraphTasks(tasks)
	if err != nil {
		return err
	}

	g, err := buildGraphView(tasks)
	if err != nil {
		return err
	}

	output, err := renderGraph(g)
	if err != nil {
		return err
	}

	return writeGraphOutput(output, g, flags)
}

// filterGraphTasks applies shortcut/status filters and prunes dependency
// references to tasks that were filtered out.
func filterGraphTasks(tasks []*model.Task) ([]*model.Task, error) {
	filtered := false

	shortcuts := FilterShortcuts{
		Status:   graphStatus,
		Priority: graphPriority,
		Phase:    graphPhase,
		Scope:    graphScope,
		Filters:  graphFilters,
	}
	hasShortcutFilters := shortcuts.Status != "" || shortcuts.Priority != "" ||
		shortcuts.Phase != "" || shortcuts.Scope != "" || len(shortcuts.Filters) > 0
	if hasShortcutFilters {
		var err error
		tasks, err = applyShortcutFilters(tasks, shortcuts)
		if err != nil {
			return nil, err
		}
		filtered = true
	}

	if len(graphExcludeStatus) > 0 {
		excludeMap := make(map[string]bool)
		for _, status := range graphExcludeStatus {
			excludeMap[status] = true
		}

		filteredTasks := make([]*model.Task, 0, len(tasks))
		for _, task := range tasks {
			if !excludeMap[string(task.Status)] {
				filteredTasks = append(filteredTasks, task)
			}
		}
		tasks = filteredTasks
		filtered = true
	}

	if filtered {
		pruneDanglingDependencies(tasks)
	}

	return tasks, nil
}

// pruneDanglingDependencies removes dependency references that point to tasks
// no longer present in the slice.
func pruneDanglingDependencies(tasks []*model.Task) {
	remainingTaskIDs := make(map[string]bool)
	for _, task := range tasks {
		remainingTaskIDs[task.ID] = true
	}

	for _, task := range tasks {
		if len(task.Dependencies) == 0 {
			continue
		}
		cleanedDeps := make([]string, 0, len(task.Dependencies))
		for _, depID := range task.Dependencies {
			if remainingTaskIDs[depID] {
				cleanedDeps = append(cleanedDeps, depID)
			}
		}
		task.Dependencies = cleanedDeps
	}
}

// buildGraphView builds the graph and applies root-based and focus filtering.
func buildGraphView(tasks []*model.Task) (*graph.Graph, error) {
	g := graph.NewGraph(tasks)

	if graphRoot != "" {
		if _, exists := g.TaskMap[graphRoot]; !exists {
			return nil, fmt.Errorf("root task %s not found", graphRoot)
		}
		g = g.FilterTasks(rootFilterIDs(g))
	} else if graphUpstream || graphDownstream {
		return nil, fmt.Errorf("--upstream and --downstream require --root")
	}

	if graphFocus != "" {
		if _, exists := g.TaskMap[graphFocus]; !exists {
			return nil, fmt.Errorf("focus task %s not found", graphFocus)
		}
	}

	return g, nil
}

// rootFilterIDs computes the set of task IDs to keep when filtering from a root,
// honoring the --upstream/--downstream direction flags.
func rootFilterIDs(g *graph.Graph) map[string]bool {
	var filteredIDs map[string]bool
	switch {
	case graphDownstream:
		filteredIDs = g.GetDownstream(graphRoot)
	case graphUpstream:
		filteredIDs = g.GetUpstream(graphRoot)
	default:
		filteredIDs = make(map[string]bool)
		for id := range g.GetUpstream(graphRoot) {
			filteredIDs[id] = true
		}
		for id := range g.GetDownstream(graphRoot) {
			filteredIDs[id] = true
		}
	}

	// Always include the root task itself
	filteredIDs[graphRoot] = true
	return filteredIDs
}

// renderGraph renders the graph in the requested output format.
func renderGraph(g *graph.Graph) (string, error) {
	switch graphFormat {
	case "mermaid":
		return g.ToMermaid(graphFocus), nil
	case "dot":
		return g.ToDot(graphFocus), nil
	case "ascii":
		showDownstream := !graphUpstream // Default to downstream for ASCII
		return g.ToASCII(graphRoot, showDownstream, newASCIIFormatter()), nil
	case "json":
		jsonBytes, err := json.MarshalIndent(g.ToJSON(), "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(jsonBytes) + "\n", nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: mermaid, dot, ascii, json)", graphFormat)
	}
}

// newASCIIFormatter builds the renderer-aware formatter for ASCII output.
func newASCIIFormatter() *graph.ASCIIFormatter {
	r := getRenderer()
	return &graph.ASCIIFormatter{
		FormatID: func(id string) string {
			return formatTaskID(id, r)
		},
		FormatTitle: func(title, status string) string {
			return formatTaskTitle(title, status, r)
		},
		FormatStatusIndicator: func(indicator, status string) string {
			return getStatusColor(status, r).Render(indicator)
		},
		FormatConnector: func(connector string) string {
			return formatDim(connector, r)
		},
		FormatReference: func(text string) string {
			return formatDim(text, r)
		},
	}
}

// writeGraphOutput writes the rendered graph to the destination and reports
// any detected dependency cycles.
func writeGraphOutput(output string, g *graph.Graph, flags GlobalFlags) error {
	var outFile *os.File
	if graphOut != "" {
		f, err := os.Create(graphOut)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	if _, err := outFile.WriteString(output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	reportGraphCycles(g, flags)
	return nil
}

// reportGraphCycles prints detected circular dependencies to stderr when
// verbose mode is on and output is not JSON to stdout.
func reportGraphCycles(g *graph.Graph, flags GlobalFlags) {
	cycles := g.DetectCycles()
	if len(cycles) == 0 {
		return
	}
	if !flags.Verbose && graphFormat != "json" {
		return
	}
	if graphOut != "" || graphFormat == "json" {
		return
	}

	fmt.Fprintf(os.Stderr, "\nWarning: detected %d circular dependencies:\n", len(cycles))
	for i, cycle := range cycles {
		fmt.Fprintf(os.Stderr, "  Cycle %d: %v\n", i+1, cycle)
	}
}
