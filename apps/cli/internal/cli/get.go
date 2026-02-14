package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	getFormat    string
	getExact     bool
	getThreshold float64
)

// getStdinReader is the reader used for interactive selection prompts.
// Override in tests to simulate user input.
var getStdinReader io.Reader = os.Stdin

var getCmd = &cobra.Command{
	Use:        "get <query>",
	SuggestFor: []string{"view", "info", "detail", "details", "describe"},
	Short:      "Get detailed information about a specific task",
	Long: `Get displays detailed information about a specific task, identified by ID, title, or file path.

Matching priority:
  1. Exact match by task ID (case-sensitive)
  2. Exact match by task title (case-insensitive)
  3. Match by file path or filename
  4. Fuzzy match across IDs and titles (unless --exact is set)

Examples:
  taskmd get cli-037
  taskmd get "Add show command"
  taskmd get tasks/cli/037-task.md  # match by file path
  taskmd get 037-task.md            # match by filename
  taskmd get 037-task               # match by filename without extension
  taskmd get sho                    # fuzzy match
  taskmd get sho --exact            # no fuzzy, returns "task not found"
  taskmd get cli-037 --format json
  taskmd get cli-037 --format yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

// Deprecated: use "get" instead.
var showCmd = &cobra.Command{
	Use:        "show <query>",
	Short:      "Show detailed information about a specific task (deprecated: use 'get')",
	Args:       cobra.ExactArgs(1),
	RunE:       runGet,
	Hidden:     true,
	Deprecated: "use 'get' instead",
}

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(showCmd)

	getCmd.Flags().StringVar(&getFormat, "format", "text", "output format (text, json, yaml)")
	getCmd.Flags().BoolVar(&getExact, "exact", false, "disable fuzzy matching, exact only")
	getCmd.Flags().Float64Var(&getThreshold, "threshold", 0.6, "fuzzy match sensitivity (0.0-1.0)")

	showCmd.Flags().StringVar(&getFormat, "format", "text", "output format (text, json, yaml)")
	showCmd.Flags().BoolVar(&getExact, "exact", false, "disable fuzzy matching, exact only")
	showCmd.Flags().Float64Var(&getThreshold, "threshold", 0.6, "fuzzy match sensitivity (0.0-1.0)")
}

func runGet(cmd *cobra.Command, args []string) error {
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

	task, err := resolveTask(query, tasks, getExact, getThreshold)
	if err != nil {
		return err
	}

	depInfo := buildDependencyInfo(task, tasks)

	return outputGet(task, depInfo, getFormat)
}

// dependencyInfo holds resolved dependency information for display.
type dependencyInfo struct {
	DependsOn []depEntry `json:"depends_on" yaml:"depends_on"`
	Blocks    []depEntry `json:"blocks" yaml:"blocks"`
	Parent    *depEntry  `json:"parent,omitempty" yaml:"parent,omitempty"`
	Children  []depEntry `json:"children,omitempty" yaml:"children,omitempty"`
}

// depEntry is a single dependency reference.
type depEntry struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`
}

// resolveTask finds a task by exact match, file path, or fuzzy match.
func resolveTask(query string, tasks []*model.Task, exactOnly bool, threshold float64) (*model.Task, error) {
	if task := findExactMatch(query, tasks); task != nil {
		return task, nil
	}

	task, err := findFilePathMatch(query, tasks)
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}

	if exactOnly {
		return nil, fmt.Errorf("task not found: %s", query)
	}

	matches := fuzzyMatchTasks(query, tasks, threshold)
	if len(matches) == 0 {
		return nil, fmt.Errorf("task not found: %s", query)
	}

	return promptSelection(query, matches)
}

// findExactMatch tries ID match (case-sensitive), then title match (case-insensitive).
func findExactMatch(query string, tasks []*model.Task) *model.Task {
	for _, t := range tasks {
		if t.ID == query {
			return t
		}
	}
	lowerQuery := strings.ToLower(query)
	for _, t := range tasks {
		if strings.ToLower(t.Title) == lowerQuery {
			return t
		}
	}
	return nil
}

// findFilePathMatch tries to match the query against task file paths and filenames.
func findFilePathMatch(query string, tasks []*model.Task) (*model.Task, error) {
	queryBase := filepath.Base(query)
	queryNoExt := strings.TrimSuffix(queryBase, ".md")

	var matches []*model.Task
	for _, t := range tasks {
		// Exact full path match — return immediately
		if t.FilePath == query {
			return t, nil
		}

		taskBase := filepath.Base(t.FilePath)
		taskNoExt := strings.TrimSuffix(taskBase, ".md")

		if taskBase == queryBase || taskNoExt == queryNoExt {
			matches = append(matches, t)
		}
	}

	switch len(matches) {
	case 0:
		return nil, nil
	case 1:
		return matches[0], nil
	default:
		return nil, fmt.Errorf("ambiguous filename %q matches multiple tasks: %s",
			query, formatAmbiguousMatches(matches))
	}
}

// formatAmbiguousMatches formats a list of tasks for an ambiguity error message.
func formatAmbiguousMatches(tasks []*model.Task) string {
	parts := make([]string, len(tasks))
	for i, t := range tasks {
		parts[i] = fmt.Sprintf("%s [%s]", t.ID, t.FilePath)
	}
	return strings.Join(parts, ", ")
}

// fuzzyMatch holds a task and its similarity score.
type fuzzyMatch struct {
	Task  *model.Task
	Score float64
}

// fuzzyMatchTasks scores all tasks against query, filters by threshold, and returns top 5.
func fuzzyMatchTasks(query string, tasks []*model.Task, threshold float64) []fuzzyMatch {
	var matches []fuzzyMatch
	for _, t := range tasks {
		score := bestFuzzyScore(query, t)
		if score >= threshold {
			matches = append(matches, fuzzyMatch{Task: t, Score: score})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})
	const maxResults = 5
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}
	return matches
}

// bestFuzzyScore returns the best similarity score between query and the task's ID or title.
func bestFuzzyScore(query string, task *model.Task) float64 {
	idScore := calculateSimilarity(query, task.ID)
	titleScore := calculateSimilarity(query, task.Title)
	if idScore > titleScore {
		return idScore
	}
	return titleScore
}

// calculateSimilarity returns a similarity score between 0.0 and 1.0.
// Substring containment scores 0.7-1.0; otherwise Levenshtein distance is used.
func calculateSimilarity(query, target string) float64 {
	lowerQuery := strings.ToLower(query)
	lowerTarget := strings.ToLower(target)

	if lowerQuery == lowerTarget {
		return 1.0
	}
	if strings.Contains(lowerTarget, lowerQuery) {
		return 0.7 + 0.3*float64(len(lowerQuery))/float64(len(lowerTarget))
	}

	maxLen := len(lowerQuery)
	if len(lowerTarget) > maxLen {
		maxLen = len(lowerTarget)
	}
	if maxLen == 0 {
		return 1.0
	}

	dist := levenshtein(lowerQuery, lowerTarget)
	return 1.0 - float64(dist)/float64(maxLen)
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev = curr
	}
	return prev[lb]
}

// promptSelection displays fuzzy matches and asks the user to pick one.
func promptSelection(query string, matches []fuzzyMatch) (*model.Task, error) {
	fmt.Fprintf(os.Stderr, "No exact match found for %q. Did you mean:\n\n", query)
	for i, m := range matches {
		fmt.Fprintf(os.Stderr, "  %d. %s: %s (%.0f%% match) [%s]\n",
			i+1, m.Task.ID, m.Task.Title, m.Score*100, m.Task.FilePath)
	}
	fmt.Fprintf(os.Stderr, "\nEnter selection (1-%d), or 0 to cancel: ", len(matches))

	reader := bufio.NewReader(getStdinReader)
	var choice int
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if _, err := fmt.Sscanf(line, "%d", &choice); err != nil || choice < 0 || choice > len(matches) {
		return nil, fmt.Errorf("invalid selection")
	}
	if choice == 0 {
		return nil, fmt.Errorf("selection cancelled")
	}
	return matches[choice-1].Task, nil
}

// buildDependencyInfo resolves depends-on and blocks lists for a task.
func buildDependencyInfo(task *model.Task, allTasks []*model.Task) dependencyInfo {
	taskMap := buildTaskMap(allTasks)
	g := graph.NewGraph(allTasks)

	var info dependencyInfo
	for _, depID := range task.Dependencies {
		entry := depEntry{ID: depID}
		if dep, ok := taskMap[depID]; ok {
			entry.Title = dep.Title
		}
		info.DependsOn = append(info.DependsOn, entry)
	}
	for _, blockedID := range g.Adjacency[task.ID] {
		entry := depEntry{ID: blockedID}
		if dep, ok := taskMap[blockedID]; ok {
			entry.Title = dep.Title
		}
		info.Blocks = append(info.Blocks, entry)
	}
	if task.Parent != "" {
		entry := depEntry{ID: task.Parent}
		if p, ok := taskMap[task.Parent]; ok {
			entry.Title = p.Title
		}
		info.Parent = &entry
	}
	for _, t := range allTasks {
		if t.Parent == task.ID {
			entry := depEntry{ID: t.ID, Title: t.Title}
			info.Children = append(info.Children, entry)
		}
	}
	return info
}

// outputGet routes to the appropriate formatter.
func outputGet(task *model.Task, deps dependencyInfo, format string) error {
	switch format {
	case "text":
		return outputGetText(task, deps, os.Stdout)
	case "json":
		return outputGetJSON(task, deps, os.Stdout)
	case "yaml":
		return outputGetYAML(task, deps, os.Stdout)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", format)
	}
}

func outputGetText(task *model.Task, deps dependencyInfo, w io.Writer) error {
	r := getRenderer()

	fmt.Fprintf(w, "%s %s\n", formatLabel("Task:", r), formatTaskID(task.ID, r))
	fmt.Fprintf(w, "%s %s\n", formatLabel("Title:", r), task.Title)
	fmt.Fprintf(w, "%s %s\n", formatLabel("Status:", r), formatStatus(string(task.Status), r))
	printOptionalField(w, "Priority", string(task.Priority), r)
	printOptionalField(w, "Effort", string(task.Effort), r)
	printTags(w, task.Tags, r)
	if deps.Parent != nil {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Parent:", r), formatDepEntry(*deps.Parent, r))
	}
	if !task.Created.IsZero() {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Created:", r), task.Created.Format("2006-01-02"))
	}
	fmt.Fprintf(w, "%s %s\n", formatLabel("File:", r), formatDim(task.FilePath, r))
	printDescription(w, task.Body)
	printDependencies(w, deps, r)
	printChildren(w, deps.Children, r)
	return nil
}

func printOptionalField(w io.Writer, label, value string, r *lipgloss.Renderer) {
	if value != "" {
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
}

func printTags(w io.Writer, tags []string, r *lipgloss.Renderer) {
	if len(tags) > 0 {
		fmt.Fprintf(w, "%s %s\n", formatLabel("Tags:", r), strings.Join(tags, ", "))
	}
}

func printDescription(w io.Writer, body string) {
	if body == "" {
		return
	}
	separator := strings.Repeat("\u2500", 49)
	fmt.Fprintf(w, "\nDescription:\n%s\n%s\n%s\n", separator, strings.TrimSpace(body), separator)
}

func printDependencies(w io.Writer, deps dependencyInfo, r *lipgloss.Renderer) {
	if len(deps.DependsOn) == 0 && len(deps.Blocks) == 0 {
		return
	}
	fmt.Fprintf(w, "\n%s\n", formatLabel("Dependencies:", r))
	if len(deps.DependsOn) > 0 {
		fmt.Fprintf(w, "  %s %s\n", formatLabel("Depends on:", r), formatDepList(deps.DependsOn, r))
	}
	if len(deps.Blocks) > 0 {
		fmt.Fprintf(w, "  %s %s\n", formatLabel("Blocks:", r), formatDepList(deps.Blocks, r))
	}
}

func formatDepEntry(e depEntry, r *lipgloss.Renderer) string {
	if e.Title != "" {
		return fmt.Sprintf("%s (%s)", formatTaskID(e.ID, r), e.Title)
	}
	return formatTaskID(e.ID, r)
}

func printChildren(w io.Writer, children []depEntry, r *lipgloss.Renderer) {
	if len(children) == 0 {
		return
	}
	fmt.Fprintf(w, "\n%s\n", formatLabel("Children:", r))
	for _, c := range children {
		fmt.Fprintf(w, "  %s\n", formatDepEntry(c, r))
	}
}

func formatDepList(entries []depEntry, r *lipgloss.Renderer) string {
	parts := make([]string, len(entries))
	for i, e := range entries {
		if e.Title != "" {
			parts[i] = fmt.Sprintf("%s (%s)", formatTaskID(e.ID, r), e.Title)
		} else {
			parts[i] = formatTaskID(e.ID, r)
		}
	}
	return strings.Join(parts, ", ")
}

// getOutput is the struct for JSON/YAML output (includes body unlike model.Task).
type getOutput struct {
	ID           string      `json:"id" yaml:"id"`
	Title        string      `json:"title" yaml:"title"`
	Status       string      `json:"status" yaml:"status"`
	Priority     string      `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort       string      `json:"effort,omitempty" yaml:"effort,omitempty"`
	Tags         []string    `json:"tags" yaml:"tags"`
	Parent       *depEntry   `json:"parent,omitempty" yaml:"parent,omitempty"`
	Created      string      `json:"created,omitempty" yaml:"created,omitempty"`
	FilePath     string      `json:"file_path" yaml:"file_path"`
	Content      string      `json:"content" yaml:"content"`
	Dependencies getDepsJSON `json:"dependencies" yaml:"dependencies"`
	Children     []depEntry  `json:"children,omitempty" yaml:"children,omitempty"`
}

type getDepsJSON struct {
	DependsOn []depEntry `json:"depends_on" yaml:"depends_on"`
	Blocks    []depEntry `json:"blocks" yaml:"blocks"`
}

func buildGetOutput(task *model.Task, deps dependencyInfo) getOutput {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	return getOutput{
		ID:       task.ID,
		Title:    task.Title,
		Status:   string(task.Status),
		Priority: string(task.Priority),
		Effort:   string(task.Effort),
		Tags:     task.Tags,
		Parent:   deps.Parent,
		Created:  created,
		FilePath: task.FilePath,
		Content:  strings.TrimSpace(task.Body),
		Dependencies: getDepsJSON{
			DependsOn: deps.DependsOn,
			Blocks:    deps.Blocks,
		},
		Children: deps.Children,
	}
}

func outputGetJSON(task *model.Task, deps dependencyInfo, w io.Writer) error {
	return WriteJSON(w, buildGetOutput(task, deps))
}

func outputGetYAML(task *model.Task, deps dependencyInfo, w io.Writer) error {
	return WriteYAML(w, buildGetOutput(task, deps))
}
