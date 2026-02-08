package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/driangle/md-task-tracker/apps/cli/internal/graph"
	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	graphFormat        string
	graphRoot          string
	graphFocus         string
	graphUpstream      bool
	graphDownstream    bool
	graphOut           string
	graphExcludeStatus []string
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph [directory]",
	Short: "Export task dependency graph",
	Long: `Export task dependency graphs in various formats for visualization and analysis.

Supported formats:
  - mermaid: Mermaid diagram syntax (default)
  - dot: Graphviz DOT format
  - ascii: ASCII art tree
  - json: JSON graph structure

Examples:
  taskmd graph > deps.mmd
  taskmd graph --format dot | dot -Tpng > graph.png
  taskmd graph --format ascii
  taskmd graph --root 022 --downstream
  taskmd graph --focus 022 --format mermaid
  taskmd graph --exclude-status completed --format ascii
  cat tasks.md | taskmd graph --stdin --format mermaid`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGraph,
}

func init() {
	rootCmd.AddCommand(graphCmd)

	graphCmd.Flags().StringVar(&graphFormat, "format", "mermaid", "output format (mermaid, dot, ascii, json)")
	graphCmd.Flags().StringVar(&graphRoot, "root", "", "start graph from specific task ID")
	graphCmd.Flags().StringVar(&graphFocus, "focus", "", "highlight specific task ID")
	graphCmd.Flags().BoolVar(&graphUpstream, "upstream", false, "show only dependencies (ancestors)")
	graphCmd.Flags().BoolVar(&graphDownstream, "downstream", false, "show only dependents (descendants)")
	graphCmd.Flags().StringSliceVar(&graphExcludeStatus, "exclude-status", []string{}, "exclude tasks with status (completed, pending, in-progress, blocked)")
	graphCmd.Flags().StringVarP(&graphOut, "out", "o", "", "write output to file instead of stdout")
}

func runGraph(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	// Validate conflicting flags
	if graphUpstream && graphDownstream {
		return fmt.Errorf("cannot use both --upstream and --downstream")
	}

	// Determine scan directory
	scanDir := "."
	if len(args) > 0 {
		scanDir = args[0]
	}

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

	// Filter tasks by status if requested
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

		// Build a map of remaining task IDs for dependency cleanup
		remainingTaskIDs := make(map[string]bool)
		for _, task := range tasks {
			remainingTaskIDs[task.ID] = true
		}

		// Clean up dependencies that reference filtered-out tasks
		for _, task := range tasks {
			if len(task.Dependencies) > 0 {
				cleanedDeps := make([]string, 0, len(task.Dependencies))
				for _, depID := range task.Dependencies {
					if remainingTaskIDs[depID] {
						cleanedDeps = append(cleanedDeps, depID)
					}
				}
				task.Dependencies = cleanedDeps
			}
		}
	}

	// Build graph
	g := graph.NewGraph(tasks)

	// Filter graph based on flags
	if graphRoot != "" {
		// Validate root task exists
		if _, exists := g.TaskMap[graphRoot]; !exists {
			return fmt.Errorf("root task %s not found", graphRoot)
		}

		// Filter tasks based on direction
		var filteredIDs map[string]bool
		if graphDownstream {
			// Show tasks that depend on root
			filteredIDs = g.GetDownstream(graphRoot)
		} else if graphUpstream {
			// Show tasks that root depends on
			filteredIDs = g.GetUpstream(graphRoot)
		} else {
			// Show both upstream and downstream
			upstream := g.GetUpstream(graphRoot)
			downstream := g.GetDownstream(graphRoot)
			filteredIDs = make(map[string]bool)
			for id := range upstream {
				filteredIDs[id] = true
			}
			for id := range downstream {
				filteredIDs[id] = true
			}
		}

		// Always include the root task itself
		filteredIDs[graphRoot] = true

		// Create filtered graph
		g = g.FilterTasks(filteredIDs)
	} else if graphUpstream || graphDownstream {
		return fmt.Errorf("--upstream and --downstream require --root")
	}

	// Validate focus task exists if specified
	if graphFocus != "" {
		if _, exists := g.TaskMap[graphFocus]; !exists {
			return fmt.Errorf("focus task %s not found", graphFocus)
		}
	}

	// Generate output based on format
	var output string
	switch graphFormat {
	case "mermaid":
		output = g.ToMermaid(graphFocus)
	case "dot":
		output = g.ToDot(graphFocus)
	case "ascii":
		// For ASCII, use root if specified, otherwise show all roots
		rootID := graphRoot
		showDownstream := !graphUpstream // Default to downstream for ASCII
		if graphUpstream {
			showDownstream = false
		}
		output = g.ToASCII(rootID, showDownstream)
	case "json":
		jsonData := g.ToJSON()
		jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		output = string(jsonBytes) + "\n"
	default:
		return fmt.Errorf("unsupported format: %s (supported: mermaid, dot, ascii, json)", graphFormat)
	}

	// Determine output destination
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

	// Write output
	_, err = outFile.WriteString(output)
	if err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print warning about cycles if any detected (only in verbose mode or for JSON format)
	cycles := g.DetectCycles()
	if len(cycles) > 0 {
		if flags.Verbose || graphFormat == "json" {
			if graphOut == "" && graphFormat != "json" {
				// Only print to stderr if not outputting JSON to stdout
				fmt.Fprintf(os.Stderr, "\nWarning: detected %d circular dependencies:\n", len(cycles))
				for i, cycle := range cycles {
					fmt.Fprintf(os.Stderr, "  Cycle %d: %v\n", i+1, cycle)
				}
			}
		}
	}

	return nil
}
