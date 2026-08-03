package cli

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	snapshotFormat  string
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

Output formats: json (default), yaml, md

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

	snapshotCmd.Flags().StringVar(&snapshotFormat, "format", "json", "output format (json, yaml, md)")
	snapshotCmd.Flags().BoolVar(&snapshotCore, "core", false, "output only core fields (id, title, dependencies)")
	snapshotCmd.Flags().BoolVar(&snapshotDerived, "derived", false, "include computed/derived fields (blocked status, depth, topological order)")
	snapshotCmd.Flags().StringVar(&snapshotGroupBy, "group-by", "", "group tasks by field (status, priority, effort, type, group, phase)")
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

func runSnapshot(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	tasks, err := scanTasks(ResolveScanDir(args), flags)
	if err != nil {
		return err
	}

	snapshots := buildSnapshots(tasks)

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
	outFile, cleanup, err := openSnapshotOutput()
	if err != nil {
		return err
	}
	defer cleanup()

	// Output in requested format
	switch snapshotFormat {
	case "json":
		return outputSnapshotJSON(output, outFile)
	case "yaml":
		return outputSnapshotYAML(output, outFile)
	case "md", "markdown":
		return outputSnapshotMarkdown(snapshots, outFile, snapshotGroupBy)
	default:
		return ValidateFormat(snapshotFormat, []string{"json", "yaml", "md"})
	}
}

// buildSnapshots converts tasks into ID-sorted snapshots, computing derived
// fields when --derived is set.
func buildSnapshots(tasks []*model.Task) []TaskSnapshot {
	taskMap := buildTaskMap(tasks)

	var depthMap map[string]int
	var topoOrder map[string]int
	var criticalPathTasks map[string]bool
	if snapshotDerived {
		depthMap = calculateDepthMap(tasks, taskMap)
		topoOrder = calculateTopologicalOrder(tasks, taskMap)
		criticalPathTasks = calculateCriticalPathTasks(tasks, taskMap)
	}

	snapshots := make([]TaskSnapshot, 0, len(tasks))
	for _, task := range tasks {
		snapshot := taskToSnapshot(task, snapshotCore, snapshotDerived, depthMap, topoOrder, criticalPathTasks, taskMap)
		snapshots = append(snapshots, snapshot)
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].ID < snapshots[j].ID
	})

	return snapshots
}

// openSnapshotOutput returns the output destination and a cleanup function.
// When --out is set it creates the file; otherwise it writes to stdout.
func openSnapshotOutput() (*os.File, func(), error) {
	if snapshotOut == "" {
		return os.Stdout, func() {}, nil
	}

	f, err := os.Create(snapshotOut)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create output file: %w", err)
	}
	return f, func() { f.Close() }, nil
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
