package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/sdk/go/model"
)

var (
	statusFormat     string
	statusExact      bool
	statusThreshold  float64
	statusMinimal    bool
	statusStatusline bool
	statusScope      string
)

// statusStdinReader is the reader used for interactive selection prompts.
// Override in tests to simulate user input.
var statusStdinReader io.Reader = os.Stdin

var statusCmd = &cobra.Command{
	Use:        "status [query]",
	SuggestFor: []string{"progress"},
	Short:      "Show in-progress tasks or get metadata for a specific task",
	Long: `Without arguments, status shows all in-progress tasks.
With a query argument, it displays the frontmatter metadata of a specific task
(without body content, resolved dependency info, context files, or worklog data).

Matching uses the same logic as 'get' (ID, title, file path, fuzzy).

Examples:
  # Show all in-progress tasks
  taskmd status

  # Compact output for shell statuslines
  taskmd status --statusline

  # Filter by scope
  taskmd status --scope cli

  # Look up a specific task
  taskmd status 042
  taskmd status "Setup project"
  taskmd status 042 --format json
  taskmd status sho --exact`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)

	statusCmd.Flags().StringVar(&statusFormat, "format", "text", "output format (text, json, yaml)")
	statusCmd.Flags().BoolVar(&statusExact, "exact", false, "disable fuzzy matching, exact only")
	statusCmd.Flags().Float64Var(&statusThreshold, "threshold", 0.6, "fuzzy match sensitivity (0.0-1.0)")
	statusCmd.Flags().BoolVar(&statusMinimal, "minimal", false, "show only task metadata, skip children")
	statusCmd.Flags().BoolVar(&statusStatusline, "statusline", false, "compact output for Claude Code statusline")
	statusCmd.Flags().StringVar(&statusScope, "scope", "", "filter by group/directory; supports wildcards (e.g. cli, cli*)")
}

// statusChild represents a child task in the recursive children tree.
type statusChild struct {
	ID       string        `json:"id" yaml:"id"`
	Title    string        `json:"title" yaml:"title"`
	Status   string        `json:"status" yaml:"status"`
	Children []statusChild `json:"children,omitempty" yaml:"children,omitempty"`
}

// statusOutput is the lightweight metadata struct for JSON/YAML output.
type statusOutput struct {
	ID           string        `json:"id" yaml:"id"`
	Title        string        `json:"title" yaml:"title"`
	Status       string        `json:"status" yaml:"status"`
	Priority     string        `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort       string        `json:"effort,omitempty" yaml:"effort,omitempty"`
	Tags         []string      `json:"tags" yaml:"tags"`
	Owner        string        `json:"owner,omitempty" yaml:"owner,omitempty"`
	Parent       string        `json:"parent,omitempty" yaml:"parent,omitempty"`
	Created      string        `json:"created,omitempty" yaml:"created,omitempty"`
	Dependencies []string      `json:"dependencies" yaml:"dependencies"`
	Blocked      *bool         `json:"blocked,omitempty" yaml:"blocked,omitempty"`
	BlockedBy    []string      `json:"blocked_by,omitempty" yaml:"blocked_by,omitempty"`
	Group        string        `json:"group,omitempty" yaml:"group,omitempty"`
	FilePath     string        `json:"file_path" yaml:"file_path"`
	Children     []statusChild `json:"children,omitempty" yaml:"children,omitempty"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runStatusList()
	}
	return runStatusSingle(args[0])
}

func runStatusList() error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}
	makeFilePathsRelative(tasks, scanDir)

	filters := []string{"status=in-progress"}
	if statusScope != "" {
		filters = append(filters, "group="+statusScope)
	}

	filtered, err := applyFilters(tasks, filters)
	if err != nil {
		return fmt.Errorf("filter failed: %w", err)
	}

	if len(filtered) == 0 {
		return outputStatusListEmpty()
	}

	if statusStatusline {
		return outputStatusline(filtered, os.Stdout)
	}

	return outputStatusListFormatted(tasks, filtered)
}

func outputStatusListEmpty() error {
	switch {
	case statusStatusline:
		return nil
	case statusFormat == "json":
		return WriteJSON(os.Stdout, []statusOutput{})
	case statusFormat == "yaml":
		return WriteYAML(os.Stdout, []statusOutput{})
	default:
		fmt.Fprintln(os.Stderr, "No tasks currently in progress.")
		return nil
	}
}

func outputStatusListFormatted(tasks, filtered []*model.Task) error {
	var childrenIndex map[string][]*model.Task
	if !statusMinimal {
		childrenIndex = buildChildrenIndex(tasks)
	}

	tasksByID := buildTasksByIDMap(tasks)

	outputs := make([]statusOutput, 0, len(filtered))
	for _, task := range filtered {
		outputs = append(outputs, buildStatusOutputFromTask(task, childrenIndex, tasksByID))
	}

	switch statusFormat {
	case "text":
		return outputStatusListText(outputs, os.Stdout)
	case "json":
		return WriteJSON(os.Stdout, outputs)
	case "yaml":
		return WriteYAML(os.Stdout, outputs)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", statusFormat)
	}
}

func runStatusSingle(query string) error {
	flags := GetGlobalFlags()
	scanDir := ResolveScanDir(nil)

	tasks, err := scanTasks(scanDir, flags)
	if err != nil {
		return err
	}
	makeFilePathsRelative(tasks, scanDir)

	// Swap stdin reader for fuzzy selection prompts
	origReader := getStdinReader
	getStdinReader = statusStdinReader
	defer func() { getStdinReader = origReader }()

	task, err := resolveTask(query, tasks, statusExact, statusThreshold)
	if err != nil {
		return err
	}

	var childrenIndex map[string][]*model.Task
	if !statusMinimal {
		childrenIndex = buildChildrenIndex(tasks)
	}

	tasksByID := buildTasksByIDMap(tasks)
	out := buildStatusOutputFromTask(task, childrenIndex, tasksByID)

	switch statusFormat {
	case "text":
		return outputStatusText(out, os.Stdout)
	case "json":
		return WriteJSON(os.Stdout, out)
	case "yaml":
		return WriteYAML(os.Stdout, out)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", statusFormat)
	}
}

func outputStatusline(tasks []*model.Task, w io.Writer) error {
	task := tasks[0]
	line := fmt.Sprintf("#%s %s", task.ID, task.Title)
	if len(tasks) > 1 {
		line += fmt.Sprintf(" (+%d more)", len(tasks)-1)
	}

	fmt.Fprintln(w, line)
	return nil
}

func outputStatusListText(outputs []statusOutput, w io.Writer) error {
	for i, out := range outputs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if err := outputStatusText(out, w); err != nil {
			return err
		}
	}
	return nil
}

func buildStatusOutputFromTask(
	task *model.Task,
	childrenIndex map[string][]*model.Task,
	tasksByID map[string]*model.Task,
) statusOutput {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	out := statusOutput{
		ID:           task.ID,
		Title:        task.Title,
		Status:       string(task.Status),
		Priority:     string(task.Priority),
		Effort:       string(task.Effort),
		Tags:         task.Tags,
		Owner:        task.Owner,
		Parent:       task.Parent,
		Created:      created,
		Dependencies: task.Dependencies,
		Group:        task.Group,
		FilePath:     task.FilePath,
	}
	if len(task.Dependencies) > 0 {
		out.BlockedBy = resolveBlockingDeps(task.Dependencies, tasksByID)
		blocked := len(out.BlockedBy) > 0
		out.Blocked = &blocked
		if !blocked {
			out.BlockedBy = nil
		}
	}
	if childrenIndex != nil {
		out.Children = collectChildrenTree(task.ID, childrenIndex, map[string]bool{task.ID: true})
	}
	return out
}

func resolveBlockingDeps(deps []string, tasksByID map[string]*model.Task) []string {
	var blocking []string
	for _, depID := range deps {
		dep, ok := tasksByID[depID]
		if !ok || dep.Status != model.StatusCompleted {
			blocking = append(blocking, depID)
		}
	}
	return blocking
}

func buildTasksByIDMap(tasks []*model.Task) map[string]*model.Task {
	m := make(map[string]*model.Task, len(tasks))
	for _, t := range tasks {
		m[t.ID] = t
	}
	return m
}

func buildChildrenIndex(tasks []*model.Task) map[string][]*model.Task {
	index := make(map[string][]*model.Task)
	for _, t := range tasks {
		if t.Parent != "" {
			index[t.Parent] = append(index[t.Parent], t)
		}
	}
	return index
}

func collectChildrenTree(taskID string, index map[string][]*model.Task, visited map[string]bool) []statusChild {
	children := index[taskID]
	if len(children) == 0 {
		return nil
	}
	result := make([]statusChild, 0, len(children))
	for _, child := range children {
		if visited[child.ID] {
			continue
		}
		visited[child.ID] = true
		sc := statusChild{
			ID:     child.ID,
			Title:  child.Title,
			Status: string(child.Status),
		}
		sc.Children = collectChildrenTree(child.ID, index, visited)
		result = append(result, sc)
	}
	return result
}

func outputStatusText(out statusOutput, w io.Writer) error {
	r := getRenderer()

	fmt.Fprintf(w, "%s %s\n", formatLabel("Task:", r), formatTaskID(out.ID, r))
	fmt.Fprintf(w, "%s %s\n", formatLabel("Title:", r), out.Title)
	fmt.Fprintf(w, "%s %s\n", formatLabel("Status:", r), formatStatus(out.Status, r))
	printStatusOptionalField(w, "Priority", out.Priority, r)
	printStatusOptionalField(w, "Effort", out.Effort, r)
	if len(out.Tags) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Tags:", r), strings.Join(out.Tags, ", "))
	}
	if out.Owner != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Owner:", r), out.Owner)
	}
	if out.Parent != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Parent:", r), out.Parent)
	}
	if out.Created != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Created:", r), out.Created)
	}
	if len(out.Dependencies) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Dependencies:", r), strings.Join(out.Dependencies, ", "))
	}
	if out.Blocked != nil {
		if *out.Blocked {
			fmt.Fprintf(w, "%s Yes (blocked by: %s)\n", formatLabel("Blocked:", r), strings.Join(out.BlockedBy, ", "))
		} else {
			fmt.Fprintf(w, "%s No\n", formatLabel("Blocked:", r))
		}
	}
	if out.Group != "" {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Group:", r), out.Group)
	}
	if len(out.Children) > 0 {
		fmt.Fprintf(w, "%s\n", formatLabel("Children:", r))
		writeChildrenTree(w, out.Children, "  ", r)
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel("File:", r), formatDim(out.FilePath, r))
	return nil
}

func writeChildrenTree(w io.Writer, children []statusChild, prefix string, r *lipgloss.Renderer) {
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├─"
		if isLast {
			connector = "└─"
		}
		fmt.Fprintf(w, "%s%s %s [%s] %s\n", prefix, connector,
			formatTaskID(child.ID, r), formatStatus(child.Status, r), child.Title)
		if len(child.Children) > 0 {
			childPrefix := prefix + "│  "
			if isLast {
				childPrefix = prefix + "   "
			}
			writeChildrenTree(w, child.Children, childPrefix, r)
		}
	}
}

func printStatusOptionalField(w io.Writer, label, value string, r *lipgloss.Renderer) {
	if value == "" {
		return
	}
	var colored string
	switch label {
	case "Priority":
		colored = formatPriority(value, r)
	case "Effort":
		colored = formatEffort(value, r)
	default:
		colored = value
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel(label+":", r), colored)
}
