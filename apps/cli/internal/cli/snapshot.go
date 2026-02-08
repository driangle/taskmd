package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	snapshotCore    bool
	snapshotDerived bool
	snapshotGroupBy string
	snapshotOut     string
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot [directory]",
	Short: "Produce a frozen, machine-readable representation of tasks",
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
	IsBlocked        *bool  `json:"is_blocked,omitempty" yaml:"is_blocked,omitempty"`
	DependencyDepth  *int   `json:"dependency_depth,omitempty" yaml:"dependency_depth,omitempty"`
	TopologicalOrder *int   `json:"topological_order,omitempty" yaml:"topological_order,omitempty"`
	OnCriticalPath   *bool  `json:"on_critical_path,omitempty" yaml:"on_critical_path,omitempty"`
}

// SnapshotOutput represents the full snapshot output
type SnapshotOutput struct {
	Tasks  []TaskSnapshot            `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Groups map[string][]TaskSnapshot `json:"groups,omitempty" yaml:"groups,omitempty"`
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

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

// buildTaskMap creates a map of task ID to task
func buildTaskMap(tasks []*model.Task) map[string]*model.Task {
	taskMap := make(map[string]*model.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	return taskMap
}

// isTaskBlocked checks if a task has unmet dependencies
func isTaskBlocked(task *model.Task, taskMap map[string]*model.Task) bool {
	for _, depID := range task.Dependencies {
		dep, exists := taskMap[depID]
		if !exists || dep.Status != model.StatusCompleted {
			return true
		}
	}
	return len(task.Dependencies) > 0 && task.Status != model.StatusCompleted
}

// calculateDepthMap calculates dependency depth for each task
func calculateDepthMap(tasks []*model.Task, taskMap map[string]*model.Task) map[string]int {
	memo := make(map[string]int)

	var getDepth func(taskID string, visited map[string]bool) int
	getDepth = func(taskID string, visited map[string]bool) int {
		if depth, ok := memo[taskID]; ok {
			return depth
		}

		if visited[taskID] {
			return 0
		}

		task, exists := taskMap[taskID]
		if !exists {
			return 0
		}

		visited[taskID] = true
		defer delete(visited, taskID)

		maxDepth := 0
		for _, depID := range task.Dependencies {
			depth := getDepth(depID, visited)
			if depth > maxDepth {
				maxDepth = depth
			}
		}

		result := maxDepth + 1
		memo[taskID] = result
		return result
	}

	for _, task := range tasks {
		getDepth(task.ID, make(map[string]bool))
	}

	return memo
}

// calculateTopologicalOrder assigns a topological order to each task
func calculateTopologicalOrder(tasks []*model.Task, taskMap map[string]*model.Task) map[string]int {
	order := make(map[string]int)
	visited := make(map[string]bool)
	counter := 0

	var visit func(taskID string)
	visit = func(taskID string) {
		if visited[taskID] {
			return
		}

		task, exists := taskMap[taskID]
		if !exists {
			return
		}

		visited[taskID] = true

		// Visit dependencies first
		for _, depID := range task.Dependencies {
			visit(depID)
		}

		// Assign order
		order[taskID] = counter
		counter++
	}

	// Visit all tasks
	for _, task := range tasks {
		visit(task.ID)
	}

	return order
}

// calculateCriticalPathTasks identifies tasks on the critical path
func calculateCriticalPathTasks(tasks []*model.Task, taskMap map[string]*model.Task) map[string]bool {
	criticalPath := make(map[string]bool)

	// Calculate depth for each task
	depthMap := calculateDepthMap(tasks, taskMap)

	// Find maximum depth
	maxDepth := 0
	for _, depth := range depthMap {
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	// Mark tasks on critical path (those with max depth)
	for taskID, depth := range depthMap {
		if depth == maxDepth {
			criticalPath[taskID] = true
			// Mark all dependencies on the path
			markCriticalPathDependencies(taskID, taskMap, depthMap, maxDepth, criticalPath)
		}
	}

	return criticalPath
}

// markCriticalPathDependencies recursively marks dependencies on critical path
func markCriticalPathDependencies(taskID string, taskMap map[string]*model.Task, depthMap map[string]int, targetDepth int, criticalPath map[string]bool) {
	task, exists := taskMap[taskID]
	if !exists {
		return
	}

	for _, depID := range task.Dependencies {
		if depthMap[depID] == targetDepth-1 {
			criticalPath[depID] = true
			markCriticalPathDependencies(depID, taskMap, depthMap, targetDepth-1, criticalPath)
		}
	}
}

// groupSnapshots groups snapshots by a field
func groupSnapshots(snapshots []TaskSnapshot, groupBy string) map[string][]TaskSnapshot {
	groups := make(map[string][]TaskSnapshot)

	for _, snapshot := range snapshots {
		var key string
		switch groupBy {
		case "status":
			key = snapshot.Status
		case "priority":
			key = snapshot.Priority
		case "effort":
			key = snapshot.Effort
		case "group":
			key = snapshot.Group
		default:
			key = "ungrouped"
		}

		if key == "" {
			key = "none"
		}

		groups[key] = append(groups[key], snapshot)
	}

	return groups
}

// outputSnapshotJSON outputs snapshot as JSON
func outputSnapshotJSON(output any, outFile *os.File) error {
	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// outputSnapshotYAML outputs snapshot as YAML
func outputSnapshotYAML(output any, outFile *os.File) error {
	encoder := yaml.NewEncoder(outFile)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(output)
}

// outputSnapshotMarkdown outputs snapshot as Markdown
func outputSnapshotMarkdown(snapshots []TaskSnapshot, outFile *os.File, groupBy string) error {
	if groupBy != "" {
		groups := groupSnapshots(snapshots, groupBy)

		// Sort group keys
		keys := make([]string, 0, len(groups))
		for key := range groups {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		// Output each group
		for _, key := range keys {
			// Capitalize first letter of key
			title := key
			if len(key) > 0 {
				title = strings.ToUpper(key[:1]) + key[1:]
			}
			fmt.Fprintf(outFile, "## %s\n\n", title)
			for _, snapshot := range groups[key] {
				outputSnapshotMarkdownTask(snapshot, outFile)
			}
			fmt.Fprintln(outFile)
		}
	} else {
		// Output all tasks
		for _, snapshot := range snapshots {
			outputSnapshotMarkdownTask(snapshot, outFile)
		}
	}

	return nil
}

// outputSnapshotMarkdownTask outputs a single task in markdown format
func outputSnapshotMarkdownTask(snapshot TaskSnapshot, outFile *os.File) {
	fmt.Fprintf(outFile, "### [%s] %s\n\n", snapshot.ID, snapshot.Title)

	if snapshot.Status != "" {
		fmt.Fprintf(outFile, "- **Status**: %s\n", snapshot.Status)
	}
	if snapshot.Priority != "" {
		fmt.Fprintf(outFile, "- **Priority**: %s\n", snapshot.Priority)
	}
	if snapshot.Effort != "" {
		fmt.Fprintf(outFile, "- **Effort**: %s\n", snapshot.Effort)
	}
	if len(snapshot.Dependencies) > 0 {
		fmt.Fprintf(outFile, "- **Dependencies**: %s\n", strings.Join(snapshot.Dependencies, ", "))
	}
	if len(snapshot.Tags) > 0 {
		fmt.Fprintf(outFile, "- **Tags**: %s\n", strings.Join(snapshot.Tags, ", "))
	}

	// Derived fields
	if snapshot.IsBlocked != nil {
		fmt.Fprintf(outFile, "- **Blocked**: %v\n", *snapshot.IsBlocked)
	}
	if snapshot.DependencyDepth != nil {
		fmt.Fprintf(outFile, "- **Depth**: %d\n", *snapshot.DependencyDepth)
	}
	if snapshot.TopologicalOrder != nil {
		fmt.Fprintf(outFile, "- **Topo Order**: %d\n", *snapshot.TopologicalOrder)
	}
	if snapshot.OnCriticalPath != nil && *snapshot.OnCriticalPath {
		fmt.Fprintf(outFile, "- **On Critical Path**: yes\n")
	}

	fmt.Fprintln(outFile)
}
