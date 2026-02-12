package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

var (
	showFormat    string
	showExact     bool
	showThreshold float64
)

// showStdinReader is the reader used for interactive selection prompts.
// Override in tests to simulate user input.
var showStdinReader io.Reader = os.Stdin

var showCmd = &cobra.Command{
	Use:        "show <query>",
	SuggestFor: []string{"view", "info", "detail", "details", "describe", "get"},
	Short:      "Show detailed information about a specific task",
	Long: `Show displays detailed information about a specific task, identified by ID or title.

Matching priority:
  1. Exact match by task ID (case-sensitive)
  2. Exact match by task title (case-insensitive)
  3. Fuzzy match across IDs and titles (unless --exact is set)

Examples:
  taskmd show cli-037
  taskmd show "Add show command"
  taskmd show sho                    # fuzzy match
  taskmd show sho --exact            # no fuzzy, returns "task not found"
  taskmd show cli-037 --format json
  taskmd show cli-037 --format yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)

	showCmd.Flags().StringVar(&showFormat, "format", "text", "output format (text, json, yaml)")
	showCmd.Flags().BoolVar(&showExact, "exact", false, "disable fuzzy matching, exact only")
	showCmd.Flags().Float64Var(&showThreshold, "threshold", 0.6, "fuzzy match sensitivity (0.0-1.0)")
}

func runShow(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()
	query := args[0]

	scanDir := ResolveScanDir(nil)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks
	makeFilePathsRelative(tasks, scanDir)

	task, err := resolveTask(query, tasks, showExact, showThreshold)
	if err != nil {
		return err
	}

	depInfo := buildDependencyInfo(task, tasks)

	return outputShow(task, depInfo, showFormat)
}

// dependencyInfo holds resolved dependency information for display.
type dependencyInfo struct {
	DependsOn []depEntry `json:"depends_on" yaml:"depends_on"`
	Blocks    []depEntry `json:"blocks" yaml:"blocks"`
}

// depEntry is a single dependency reference.
type depEntry struct {
	ID    string `json:"id" yaml:"id"`
	Title string `json:"title" yaml:"title"`
}

// resolveTask finds a task by exact match or fuzzy match.
func resolveTask(query string, tasks []*model.Task, exactOnly bool, threshold float64) (*model.Task, error) {
	if task := findExactMatch(query, tasks); task != nil {
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

	reader := bufio.NewReader(showStdinReader)
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
	return info
}

// outputShow routes to the appropriate formatter.
func outputShow(task *model.Task, deps dependencyInfo, format string) error {
	switch format {
	case "text":
		return outputShowText(task, deps, os.Stdout)
	case "json":
		return outputShowJSON(task, deps, os.Stdout)
	case "yaml":
		return outputShowYAML(task, deps, os.Stdout)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, yaml)", format)
	}
}

func outputShowText(task *model.Task, deps dependencyInfo, w io.Writer) error {
	fmt.Fprintf(w, "Task: %s\n", task.ID)
	fmt.Fprintf(w, "Title: %s\n", task.Title)
	fmt.Fprintf(w, "Status: %s\n", task.Status)
	printOptionalField(w, "Priority", string(task.Priority))
	printOptionalField(w, "Effort", string(task.Effort))
	printTags(w, task.Tags)
	if !task.Created.IsZero() {
		fmt.Fprintf(w, "Created: %s\n", task.Created.Format("2006-01-02"))
	}
	fmt.Fprintf(w, "File: %s\n", task.FilePath)
	printDescription(w, task.Body)
	printDependencies(w, deps)
	return nil
}

func printOptionalField(w io.Writer, label, value string) {
	if value != "" {
		fmt.Fprintf(w, "%s: %s\n", label, value)
	}
}

func printTags(w io.Writer, tags []string) {
	if len(tags) > 0 {
		fmt.Fprintf(w, "Tags: %s\n", strings.Join(tags, ", "))
	}
}

func printDescription(w io.Writer, body string) {
	if body == "" {
		return
	}
	separator := strings.Repeat("\u2500", 49)
	fmt.Fprintf(w, "\nDescription:\n%s\n%s\n%s\n", separator, strings.TrimSpace(body), separator)
}

func printDependencies(w io.Writer, deps dependencyInfo) {
	if len(deps.DependsOn) == 0 && len(deps.Blocks) == 0 {
		return
	}
	fmt.Fprintf(w, "\nDependencies:\n")
	if len(deps.DependsOn) > 0 {
		fmt.Fprintf(w, "  Depends on: %s\n", formatDepList(deps.DependsOn))
	}
	if len(deps.Blocks) > 0 {
		fmt.Fprintf(w, "  Blocks: %s\n", formatDepList(deps.Blocks))
	}
}

func formatDepList(entries []depEntry) string {
	parts := make([]string, len(entries))
	for i, e := range entries {
		if e.Title != "" {
			parts[i] = fmt.Sprintf("%s (%s)", e.ID, e.Title)
		} else {
			parts[i] = e.ID
		}
	}
	return strings.Join(parts, ", ")
}

// showOutput is the struct for JSON/YAML output (includes body unlike model.Task).
type showOutput struct {
	ID           string       `json:"id" yaml:"id"`
	Title        string       `json:"title" yaml:"title"`
	Status       string       `json:"status" yaml:"status"`
	Priority     string       `json:"priority,omitempty" yaml:"priority,omitempty"`
	Effort       string       `json:"effort,omitempty" yaml:"effort,omitempty"`
	Tags         []string     `json:"tags" yaml:"tags"`
	Created      string       `json:"created,omitempty" yaml:"created,omitempty"`
	FilePath     string       `json:"file_path" yaml:"file_path"`
	Content      string       `json:"content" yaml:"content"`
	Dependencies showDepsJSON `json:"dependencies" yaml:"dependencies"`
}

type showDepsJSON struct {
	DependsOn []depEntry `json:"depends_on" yaml:"depends_on"`
	Blocks    []depEntry `json:"blocks" yaml:"blocks"`
}

func buildShowOutput(task *model.Task, deps dependencyInfo) showOutput {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	return showOutput{
		ID:           task.ID,
		Title:        task.Title,
		Status:       string(task.Status),
		Priority:     string(task.Priority),
		Effort:       string(task.Effort),
		Tags:         task.Tags,
		Created:      created,
		FilePath:     task.FilePath,
		Content:      strings.TrimSpace(task.Body),
		Dependencies: showDepsJSON(deps),
	}
}

func outputShowJSON(task *model.Task, deps dependencyInfo, w io.Writer) error {
	out := buildShowOutput(task, deps)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func outputShowYAML(task *model.Task, deps dependencyInfo, w io.Writer) error {
	out := buildShowOutput(task, deps)
	enc := yaml.NewEncoder(w)
	defer enc.Close()
	return enc.Encode(out)
}
