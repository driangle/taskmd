package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	snapshotCore    bool
	snapshotDerived bool
	snapshotGroupBy string
	snapshotOut     string
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:        "snapshot",
	SuggestFor: []string{"save", "backup", "export"},
	Short:      "Produce a frozen, machine-readable representation of tasks",
	Long: `Snapshot produces a static, machine-readable representation of tasks
for CI/CD pipelines and automation.

By default, outputs all task data in JSON format. Use --core to output only
essential fields, or --derived to include computed dependency analysis.

Supported output formats: json, yaml, md

Examples:
  taskmd snapshot > snapshot.json
  taskmd snapshot --format yaml --out snapshot.yaml
  taskmd snapshot --core --format json
  taskmd snapshot --derived --group-by status
  cat tasks.md | taskmd snapshot --stdin`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSnapshot,
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	snapshotCmd.Flags().BoolVar(&snapshotCore, "core", false, "output only core fields (id, title, dependencies)")
	snapshotCmd.Flags().BoolVar(&snapshotDerived, "derived", false, "include computed/derived fields (blocked status, depth, topological order)")
	snapshotCmd.Flags().StringVar(&snapshotGroupBy, "group-by", "", "group tasks by field (status, priority, effort, group)")
	snapshotCmd.Flags().StringVarP(&snapshotOut, "out", "o", "", "write output to file instead of stdout")
}

// TaskSnapshot represents a task with core or derived fields
type TaskSnapshot struct {
	// Core fields (always included unless --core is used)
	ID           string   `json:"id" yaml:"id"`
	Title        string   `json:"title" yaml:"title"`
	Status       string   `json:"status,omitempty" yaml:"status,omitempty"`
	Priority     string   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort       string   `json:"effort,omitempty" yaml:"effort,omitempty"`
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Tags         []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Group        string   `json:"group,omitempty" yaml:"group,omitempty"`
	Created      string   `json:"created,omitempty" yaml:"created,omitempty"`
	FilePath     string   `json:"file_path,omitempty" yaml:"file_path,omitempty"`

	// Derived fields (only included with --derived)
	IsBlocked        *bool `json:"is_blocked,omitempty" yaml:"is_blocked,omitempty"`
	DependencyDepth  *int  `json:"dependency_depth,omitempty" yaml:"dependency_depth,omitempty"`
	TopologicalOrder *int  `json:"topological_order,omitempty" yaml:"topological_order,omitempty"`
	OnCriticalPath   *bool `json:"on_critical_path,omitempty" yaml:"on_critical_path,omitempty"`
}

// SnapshotOutput represents the full snapshot output
type SnapshotOutput struct {
	Tasks  []TaskSnapshot            `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Groups map[string][]TaskSnapshot `json:"groups,omitempty" yaml:"groups,omitempty"`
}

//nolint:funlen // TODO: refactor to reduce length
func runSnapshot(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

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

	// Build task map for derived fields
	taskMap := buildTaskMap(tasks)

	// Calculate derived fields if requested
	var depthMap map[string]int
	var topoOrder map[string]int
	var criticalPathTasks map[string]bool

	if snapshotDerived {
		depthMap = calculateDepthMap(tasks, taskMap)
		topoOrder = calculateTopologicalOrder(tasks, taskMap)
		criticalPathTasks = calculateCriticalPathTasks(tasks, taskMap)
	}

	// Convert tasks to snapshots
	snapshots := make([]TaskSnapshot, 0, len(tasks))
	for _, task := range tasks {
		snapshot := taskToSnapshot(task, snapshotCore, snapshotDerived, depthMap, topoOrder, criticalPathTasks, taskMap)
		snapshots = append(snapshots, snapshot)
	}

	// Sort snapshots by ID
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].ID < snapshots[j].ID
	})

	// Prepare output
	var output any
	if snapshotGroupBy != "" {
		output = SnapshotOutput{
			Groups: groupSnapshots(snapshots, snapshotGroupBy),
		}
	} else {
		output = SnapshotOutput{
			Tasks: snapshots,
		}
	}

	// Determine output destination
	var outFile *os.File
	if snapshotOut != "" {
		f, err := os.Create(snapshotOut)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		outFile = f
	} else {
		outFile = os.Stdout
	}

	// Output in requested format
	switch flags.Format {
	case "json":
		return outputSnapshotJSON(output, outFile)
	case "yaml":
		return outputSnapshotYAML(output, outFile)
	case "md", "markdown":
		return outputSnapshotMarkdown(snapshots, outFile, snapshotGroupBy)
	default:
		return fmt.Errorf("unsupported format: %s (supported: json, yaml, md)", flags.Format)
	}
}

// taskToSnapshot converts a model.Task to TaskSnapshot
func taskToSnapshot(
	task *model.Task,
	coreOnly bool,
	includeDerived bool,
	depthMap map[string]int,
	topoOrder map[string]int,
	criticalPath map[string]bool,
	taskMap map[string]*model.Task,
) TaskSnapshot {
	snapshot := TaskSnapshot{
		ID:    task.ID,
		Title: task.Title,
	}

	// Add non-core fields unless --core is specified
	if !coreOnly {
		snapshot.Status = string(task.Status)
		snapshot.Priority = string(task.Priority)
		snapshot.Effort = string(task.Effort)
		snapshot.Tags = task.Tags
		snapshot.Group = task.Group
		if !task.Created.IsZero() {
			snapshot.Created = task.Created.Format("2006-01-02")
		}
		snapshot.FilePath = task.FilePath
	}

	// Always include dependencies
	snapshot.Dependencies = task.Dependencies

	// Add derived fields if requested
	if includeDerived {
		// Is blocked: has unmet dependencies
		isBlocked := isTaskBlocked(task, taskMap)
		snapshot.IsBlocked = &isBlocked

		// Dependency depth
		if depth, ok := depthMap[task.ID]; ok {
			snapshot.DependencyDepth = &depth
		}

		// Topological order
		if order, ok := topoOrder[task.ID]; ok {
			snapshot.TopologicalOrder = &order
		}

		// On critical path
		if onPath, ok := criticalPath[task.ID]; ok {
			snapshot.OnCriticalPath = &onPath
		}
	}

	return snapshot
}
