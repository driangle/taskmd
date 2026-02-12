package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/driangle/md-task-tracker/apps/cli/internal/graph"
	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
	"github.com/driangle/md-task-tracker/apps/cli/internal/scanner"
)

// Scoring constants
const (
	scorePriorityCritical = 40
	scorePriorityHigh     = 30
	scorePriorityMedium   = 20
	scorePriorityLow      = 10
	scoreCriticalPath     = 15
	scorePerDownstream    = 3
	scoreDownstreamMax    = 15
	scoreEffortSmall      = 5
	scoreEffortMedium     = 2
)

// Recommendation represents a scored task recommendation
type Recommendation struct {
	Rank            int      `json:"rank" yaml:"rank"`
	ID              string   `json:"id" yaml:"id"`
	Title           string   `json:"title" yaml:"title"`
	FilePath        string   `json:"file_path" yaml:"file_path"`
	Status          string   `json:"status" yaml:"status"`
	Priority        string   `json:"priority" yaml:"priority"`
	Effort          string   `json:"effort,omitempty" yaml:"effort,omitempty"`
	Score           int      `json:"score" yaml:"score"`
	Reasons         []string `json:"reasons" yaml:"reasons"`
	DownstreamCount int      `json:"downstream_count" yaml:"downstream_count"`
	OnCriticalPath  bool     `json:"on_critical_path" yaml:"on_critical_path"`
}

var (
	nextLimit   int
	nextFilters []string
)

var nextCmd = &cobra.Command{
	Use:        "next",
	SuggestFor: []string{"pick", "suggest", "what"},
	Short:      "Recommend what task to work on next",
	Long: `Next analyzes all tasks and recommends the best ones to work on next.

Tasks are scored based on priority, critical path position, downstream impact,
and effort. Only actionable tasks (pending or in-progress with all dependencies
completed) are shown.

Examples:
  taskmd next
  taskmd next ./tasks
  taskmd next --limit 3
  taskmd next --filter tag=cli
  taskmd next --filter priority=high --format json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNext,
}

func init() {
	rootCmd.AddCommand(nextCmd)

	nextCmd.Flags().IntVar(&nextLimit, "limit", 5, "maximum number of recommendations")
	nextCmd.Flags().StringArrayVar(&nextFilters, "filter", []string{}, "filter tasks (e.g., --filter tag=cli)")
}

// hasUnmetDependencies checks if any dependency is not completed
func hasUnmetDependencies(task *model.Task, taskMap map[string]*model.Task) bool {
	for _, depID := range task.Dependencies {
		dep, exists := taskMap[depID]
		if !exists || dep.Status != model.StatusCompleted {
			return true
		}
	}
	return false
}

// isActionable returns true if the task is pending/in-progress with all deps completed
func isActionable(task *model.Task, taskMap map[string]*model.Task) bool {
	if task.Status != model.StatusPending && task.Status != model.StatusInProgress {
		return false
	}
	return !hasUnmetDependencies(task, taskMap)
}

// scoreTask computes a score and reason list for an actionable task
func scoreTask(
	task *model.Task,
	criticalPath map[string]bool,
	downstreamCounts map[string]int,
) (int, []string) {
	score := 0
	var reasons []string

	// Priority scoring
	switch task.Priority {
	case model.PriorityCritical:
		score += scorePriorityCritical
		reasons = append(reasons, "critical priority")
	case model.PriorityHigh:
		score += scorePriorityHigh
		reasons = append(reasons, "high priority")
	case model.PriorityMedium:
		score += scorePriorityMedium
	default:
		score += scorePriorityLow
	}

	// Critical path
	if criticalPath[task.ID] {
		score += scoreCriticalPath
		reasons = append(reasons, "on critical path")
	}

	// Downstream impact
	dc := downstreamCounts[task.ID]
	bonus := min(dc*scorePerDownstream, scoreDownstreamMax)
	score += bonus
	if dc > 0 {
		noun := "tasks"
		if dc == 1 {
			noun = "task"
		}
		reasons = append(reasons, fmt.Sprintf("unblocks %d %s", dc, noun))
	}

	// Effort scoring
	switch task.Effort {
	case model.EffortSmall:
		score += scoreEffortSmall
		reasons = append(reasons, "quick win")
	case model.EffortMedium:
		score += scoreEffortMedium
	}

	return score, reasons
}

//nolint:funlen // TODO: refactor to reduce length
func runNext(cmd *cobra.Command, args []string) error {
	flags := GetGlobalFlags()

	scanDir := ResolveScanDir(args)

	taskScanner := scanner.NewScanner(scanDir, flags.Verbose)
	result, err := taskScanner.Scan()
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	allTasks := result.Tasks

	// Make file paths relative to scan directory
	makeFilePathsRelative(allTasks, scanDir)

	// Build analysis on full task set
	taskMap := buildTaskMap(allTasks)
	criticalPath := calculateCriticalPathTasks(allTasks, taskMap)
	g := graph.NewGraph(allTasks)

	downstreamCounts := make(map[string]int)
	for _, task := range allTasks {
		downstreamCounts[task.ID] = len(g.GetDownstream(task.ID))
	}

	// Apply user filters to narrow candidates
	candidates := allTasks
	if len(nextFilters) > 0 {
		candidates, err = applyFilters(candidates, nextFilters)
		if err != nil {
			return fmt.Errorf("filter error: %w", err)
		}
	}

	// Filter to actionable only
	var actionable []*model.Task
	for _, task := range candidates {
		if isActionable(task, taskMap) {
			actionable = append(actionable, task)
		}
	}

	// Score each candidate
	type scored struct {
		task    *model.Task
		score   int
		reasons []string
	}
	scoredTasks := make([]scored, len(actionable))
	for i, task := range actionable {
		s, r := scoreTask(task, criticalPath, downstreamCounts)
		scoredTasks[i] = scored{task: task, score: s, reasons: r}
	}

	// Sort by score desc, ID asc (stable)
	sort.SliceStable(scoredTasks, func(i, j int) bool {
		if scoredTasks[i].score != scoredTasks[j].score {
			return scoredTasks[i].score > scoredTasks[j].score
		}
		return scoredTasks[i].task.ID < scoredTasks[j].task.ID
	})

	// Apply limit
	limit := min(nextLimit, len(scoredTasks))
	scoredTasks = scoredTasks[:limit]

	// Build recommendations
	recs := make([]Recommendation, len(scoredTasks))
	for i, st := range scoredTasks {
		recs[i] = Recommendation{
			Rank:            i + 1,
			ID:              st.task.ID,
			Title:           st.task.Title,
			FilePath:        st.task.FilePath,
			Status:          string(st.task.Status),
			Priority:        string(st.task.Priority),
			Effort:          string(st.task.Effort),
			Score:           st.score,
			Reasons:         st.reasons,
			DownstreamCount: downstreamCounts[st.task.ID],
			OnCriticalPath:  criticalPath[st.task.ID],
		}
	}

	switch flags.Format {
	case "json":
		return outputNextJSON(recs)
	case "yaml":
		return outputNextYAML(recs)
	case "table":
		return outputNextTable(recs)
	default:
		return fmt.Errorf("unsupported format: %s (supported: table, json, yaml)", flags.Format)
	}
}

func outputNextJSON(recs []Recommendation) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(recs)
}

func outputNextYAML(recs []Recommendation) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(recs)
}

func outputNextTable(recs []Recommendation) error {
	if len(recs) == 0 {
		fmt.Println("No actionable tasks found.")
		return nil
	}

	fmt.Println("Recommended tasks:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "#\tID\tTitle\tPriority\tFile\tReason")
	fmt.Fprintln(w, "-\t--\t-----\t--------\t----\t------")

	for _, rec := range recs {
		reason := strings.Join(rec.Reasons, ", ")
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n", rec.Rank, rec.ID, rec.Title, rec.Priority, rec.FilePath, reason)
	}

	return nil
}
