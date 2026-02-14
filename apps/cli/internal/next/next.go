package next

import (
	"fmt"
	"sort"

	"github.com/driangle/taskmd/apps/cli/internal/filter"
	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// Scoring constants
const (
	ScorePriorityCritical = 40
	ScorePriorityHigh     = 30
	ScorePriorityMedium   = 20
	ScorePriorityLow      = 10
	ScoreCriticalPath     = 15
	ScorePerDownstream    = 3
	ScoreDownstreamMax    = 15
	ScoreEffortSmall      = 5
	ScoreEffortMedium     = 2
)

// Recommendation represents a scored task recommendation.
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

// Options controls recommendation behaviour.
type Options struct {
	Limit     int
	Filters   []string
	QuickWins bool
	Critical  bool
}

type scoredTask struct {
	task    *model.Task
	score   int
	reasons []string
}

// Recommend scores and ranks actionable tasks, returning the top recommendations.
func Recommend(tasks []*model.Task, opts Options) ([]Recommendation, error) {
	if opts.Limit <= 0 {
		opts.Limit = 5
	}

	taskMap := BuildTaskMap(tasks)
	criticalPath := CalculateCriticalPathTasks(tasks, taskMap)
	downstreamCounts := computeDownstreamCounts(tasks)

	actionable, err := filterActionable(tasks, opts, taskMap, criticalPath)
	if err != nil {
		return nil, err
	}

	scored := scoreAndSort(actionable, criticalPath, downstreamCounts)

	limit := min(opts.Limit, len(scored))
	return buildRecommendations(scored[:limit], criticalPath, downstreamCounts), nil
}

func computeDownstreamCounts(tasks []*model.Task) map[string]int {
	g := graph.NewGraph(tasks)
	counts := make(map[string]int, len(tasks))
	for _, task := range tasks {
		counts[task.ID] = len(g.GetDownstream(task.ID))
	}
	return counts
}

func filterActionable(
	tasks []*model.Task,
	opts Options,
	taskMap map[string]*model.Task,
	criticalPath map[string]bool,
) ([]*model.Task, error) {
	candidates := tasks
	if len(opts.Filters) > 0 {
		var err error
		candidates, err = filter.Apply(candidates, opts.Filters)
		if err != nil {
			return nil, fmt.Errorf("filter error: %w", err)
		}
	}

	var actionable []*model.Task
	for _, task := range candidates {
		if IsActionable(task, taskMap) {
			actionable = append(actionable, task)
		}
	}

	return applySpecialFilters(actionable, criticalPath, opts.QuickWins, opts.Critical), nil
}

func scoreAndSort(
	tasks []*model.Task,
	criticalPath map[string]bool,
	downstreamCounts map[string]int,
) []scoredTask {
	scored := make([]scoredTask, len(tasks))
	for i, task := range tasks {
		s, r := ScoreTask(task, criticalPath, downstreamCounts)
		scored[i] = scoredTask{task: task, score: s, reasons: r}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].task.ID < scored[j].task.ID
	})

	return scored
}

func buildRecommendations(
	scored []scoredTask,
	criticalPath map[string]bool,
	downstreamCounts map[string]int,
) []Recommendation {
	recs := make([]Recommendation, len(scored))
	for i, st := range scored {
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
	return recs
}

// BuildTaskMap creates a map of task ID to task.
func BuildTaskMap(tasks []*model.Task) map[string]*model.Task {
	taskMap := make(map[string]*model.Task, len(tasks))
	for _, task := range tasks {
		taskMap[task.ID] = task
	}
	return taskMap
}

// HasUnmetDependencies checks if any dependency is not completed.
func HasUnmetDependencies(task *model.Task, taskMap map[string]*model.Task) bool {
	for _, depID := range task.Dependencies {
		dep, exists := taskMap[depID]
		if !exists || dep.Status != model.StatusCompleted {
			return true
		}
	}
	return false
}

// IsActionable returns true if the task is pending/in-progress with all deps completed.
func IsActionable(task *model.Task, taskMap map[string]*model.Task) bool {
	if task.Status != model.StatusPending && task.Status != model.StatusInProgress {
		return false
	}
	return !HasUnmetDependencies(task, taskMap)
}

// ScoreTask computes a score and reason list for an actionable task.
func ScoreTask(
	task *model.Task,
	criticalPath map[string]bool,
	downstreamCounts map[string]int,
) (int, []string) {
	score := 0
	var reasons []string

	switch task.Priority {
	case model.PriorityCritical:
		score += ScorePriorityCritical
		reasons = append(reasons, "critical priority")
	case model.PriorityHigh:
		score += ScorePriorityHigh
		reasons = append(reasons, "high priority")
	case model.PriorityMedium:
		score += ScorePriorityMedium
	default:
		score += ScorePriorityLow
	}

	if criticalPath[task.ID] {
		score += ScoreCriticalPath
		reasons = append(reasons, "on critical path")
	}

	dc := downstreamCounts[task.ID]
	bonus := min(dc*ScorePerDownstream, ScoreDownstreamMax)
	score += bonus
	if dc > 0 {
		noun := "tasks"
		if dc == 1 {
			noun = "task"
		}
		reasons = append(reasons, fmt.Sprintf("unblocks %d %s", dc, noun))
	}

	switch task.Effort {
	case model.EffortSmall:
		score += ScoreEffortSmall
		reasons = append(reasons, "quick win")
	case model.EffortMedium:
		score += ScoreEffortMedium
	}

	return score, reasons
}

// CalculateCriticalPathTasks identifies tasks on the critical path.
func CalculateCriticalPathTasks(tasks []*model.Task, taskMap map[string]*model.Task) map[string]bool {
	criticalPath := make(map[string]bool)
	depthMap := calculateDepthMap(tasks, taskMap)

	maxDepth := 0
	for _, depth := range depthMap {
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	for taskID, depth := range depthMap {
		if depth == maxDepth {
			criticalPath[taskID] = true
			markCriticalPathDependencies(taskID, taskMap, depthMap, maxDepth, criticalPath)
		}
	}

	return criticalPath
}

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

func markCriticalPathDependencies(
	taskID string,
	taskMap map[string]*model.Task,
	depthMap map[string]int,
	targetDepth int,
	criticalPath map[string]bool,
) {
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

func applySpecialFilters(
	actionable []*model.Task,
	criticalPath map[string]bool,
	quickWins, critical bool,
) []*model.Task {
	result := actionable

	if quickWins {
		var filtered []*model.Task
		for _, task := range result {
			if task.Effort == model.EffortSmall {
				filtered = append(filtered, task)
			}
		}
		result = filtered
	}

	if critical {
		var filtered []*model.Task
		for _, task := range result {
			if criticalPath[task.ID] {
				filtered = append(filtered, task)
			}
		}
		result = filtered
	}

	return result
}
