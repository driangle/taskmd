package next

import (
	"fmt"
	"sort"

	"github.com/driangle/taskmd/sdk/go/filter"
	"github.com/driangle/taskmd/sdk/go/graph"
	"github.com/driangle/taskmd/sdk/go/model"
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
	ScorePhaseBase        = 25
	ScorePhaseDecay       = 5
)

// ScoreComponent is a single itemized contribution to a task's score.
// Base and Multiplier are populated for scaled bonuses (critical path,
// downstream) so the scaling is explainable; for flat components they are
// zero and omitted. Points is always the actual points contributed.
type ScoreComponent struct {
	Label      string  `json:"label" yaml:"label"`
	Points     int     `json:"points" yaml:"points"`
	Base       int     `json:"base,omitempty" yaml:"base,omitempty"`
	Multiplier float64 `json:"multiplier,omitempty" yaml:"multiplier,omitempty"`
}

// Recommendation represents a scored task recommendation.
type Recommendation struct {
	Rank            int              `json:"rank" yaml:"rank"`
	ID              string           `json:"id" yaml:"id"`
	Title           string           `json:"title" yaml:"title"`
	FilePath        string           `json:"file_path" yaml:"file_path"`
	Status          string           `json:"status" yaml:"status"`
	Priority        string           `json:"priority" yaml:"priority"`
	Effort          string           `json:"effort,omitempty" yaml:"effort,omitempty"`
	Score           int              `json:"score" yaml:"score"`
	Reasons         []string         `json:"reasons" yaml:"reasons"`
	ScoreBreakdown  []ScoreComponent `json:"score_breakdown" yaml:"score_breakdown"`
	DownstreamCount int              `json:"downstream_count" yaml:"downstream_count"`
	OnCriticalPath  bool             `json:"on_critical_path" yaml:"on_critical_path"`
}

// Options controls recommendation behaviour.
type Options struct {
	Limit          int
	Filters        []string
	QuickWins      bool
	Critical       bool
	Scope          string
	ScopeExact     bool
	Root           string
	ArchivedTasks  []*model.Task
	Phase          string
	PhaseOrder     []string
	StrictPhases   bool
	StrictPriority bool
}

type scoredTask struct {
	task       *model.Task
	score      int
	reasons    []string
	components []ScoreComponent
}

// Recommend scores and ranks actionable tasks, returning the top recommendations.
func Recommend(tasks []*model.Task, opts Options) ([]Recommendation, error) {
	if opts.Limit <= 0 {
		opts.Limit = 5
	}

	taskMap := BuildTaskMap(tasks)

	// Merge archived tasks for dependency resolution only.
	// Active tasks take precedence over archived duplicates.
	for _, at := range opts.ArchivedTasks {
		if _, exists := taskMap[at.ID]; !exists {
			taskMap[at.ID] = at
		}
	}

	childrenMap := BuildChildrenMap(tasks)
	criticalPath := CalculateCriticalPathTasks(tasks, taskMap)
	downstreamInfo := ComputeDownstreamInfo(tasks)

	actionable, err := filterActionable(tasks, opts, taskMap, childrenMap, criticalPath)
	if err != nil {
		return nil, err
	}

	scored := scoreAndSort(actionable, sortOptions{
		phaseOrder:     opts.PhaseOrder,
		strictPhases:   opts.StrictPhases,
		strictPriority: opts.StrictPriority,
	}, criticalPath, downstreamInfo)

	limit := min(opts.Limit, len(scored))
	return buildRecommendations(scored[:limit], criticalPath, downstreamInfo), nil
}

// DownstreamInfo holds both the count and max priority of downstream tasks.
type DownstreamInfo struct {
	Count       int
	MaxPriority model.Priority
}

// ComputeDownstreamInfo computes downstream counts and max priority for each task.
func ComputeDownstreamInfo(tasks []*model.Task) map[string]DownstreamInfo {
	g := graph.NewGraph(tasks)
	info := make(map[string]DownstreamInfo, len(tasks))
	for _, task := range tasks {
		downstream := g.GetDownstream(task.ID)
		maxPri := model.Priority("")
		for id := range downstream {
			if t, ok := g.TaskMap[id]; ok {
				if priorityWeight(t.Priority) > priorityWeight(maxPri) {
					maxPri = t.Priority
				}
			}
		}
		info[task.ID] = DownstreamInfo{
			Count:       len(downstream),
			MaxPriority: maxPri,
		}
	}
	return info
}

// priorityWeight returns a numeric weight for a priority level.
func priorityWeight(p model.Priority) int {
	switch p {
	case model.PriorityCritical:
		return 4
	case model.PriorityHigh:
		return 3
	case model.PriorityMedium:
		return 2
	case model.PriorityLow:
		return 1
	default:
		return 1
	}
}

// downstreamPriorityMultiplier returns a scaling factor for downstream/critical-path
// bonuses based on the max priority found in the downstream chain.
func downstreamPriorityMultiplier(maxPri model.Priority) float64 {
	switch maxPri {
	case model.PriorityCritical, model.PriorityHigh:
		return 1.0
	case model.PriorityMedium:
		return 0.5
	default:
		return 0.25
	}
}

func filterActionable(
	tasks []*model.Task,
	opts Options,
	taskMap map[string]*model.Task,
	childrenMap map[string][]*model.Task,
	criticalPath map[string]bool,
) ([]*model.Task, error) {
	var reachable map[string]bool
	if opts.Root != "" {
		if _, exists := taskMap[opts.Root]; !exists {
			return nil, fmt.Errorf("root task %s not found", opts.Root)
		}
		reachable = rootReachableSet(opts.Root, tasks, childrenMap)
	}

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
		if reachable != nil && !reachable[task.ID] {
			continue
		}
		if IsActionable(task, taskMap, childrenMap) {
			actionable = append(actionable, task)
		}
	}

	if opts.Phase != "" {
		var filtered []*model.Task
		for _, task := range actionable {
			if task.Phase == opts.Phase {
				filtered = append(filtered, task)
			}
		}
		actionable = filtered
	}

	if opts.Scope != "" {
		if opts.ScopeExact {
			actionable = filterByScope(actionable, opts.Scope)
		} else {
			actionable = filterByScopeExpanded(actionable, tasks, opts.Scope)
		}
	}

	return applySpecialFilters(actionable, criticalPath, opts.QuickWins, opts.Critical), nil
}

// rootReachableSet returns the set of task IDs reachable from root: root's
// transitive upstream dependency prerequisites, root's transitive subtask
// subtree (via the parent/child hierarchy), and root itself. Dependencies use
// the graph package's upstream traversal; subtasks recurse over childrenMap.
func rootReachableSet(root string, tasks []*model.Task, childrenMap map[string][]*model.Task) map[string]bool {
	reachable := graph.NewGraph(tasks).GetUpstream(root)
	reachable[root] = true

	var addSubtree func(id string)
	addSubtree = func(id string) {
		for _, child := range childrenMap[id] {
			if !reachable[child.ID] {
				reachable[child.ID] = true
				addSubtree(child.ID)
			}
		}
	}
	addSubtree(root)

	return reachable
}

// sortOptions groups the ordering-related recommendation options.
type sortOptions struct {
	phaseOrder     []string
	strictPhases   bool
	strictPriority bool
}

func scoreAndSort(
	tasks []*model.Task,
	opts sortOptions,
	criticalPath map[string]bool,
	downstreamInfo map[string]DownstreamInfo,
) []scoredTask {
	phaseIndex := buildPhaseIndex(opts.phaseOrder)

	scored := make([]scoredTask, len(tasks))
	for i, task := range tasks {
		comps, r := buildScore(task, criticalPath, downstreamInfo)
		if pc, pr := buildPhaseScore(task, opts.phaseOrder); pc != nil {
			comps = append(comps, *pc)
			r = append(r, pr)
		}
		scored[i] = scoredTask{task: task, score: sumComponents(comps), reasons: r, components: comps}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		// Primary key: phase tier, when strict-phases is enabled.
		if opts.strictPhases && len(opts.phaseOrder) > 0 {
			pi := taskPhaseIndex(scored[i].task, phaseIndex)
			pj := taskPhaseIndex(scored[j].task, phaseIndex)
			if pi != pj {
				return pi < pj
			}
		}
		// Secondary key: priority tier, when strict-priority is enabled.
		if opts.strictPriority {
			wi := priorityWeight(scored[i].task.Priority)
			wj := priorityWeight(scored[j].task.Priority)
			if wi != wj {
				return wi > wj
			}
		}
		// Tiebreak within a tier: score, then ID.
		if scored[i].score != scored[j].score {
			return scored[i].score > scored[j].score
		}
		return scored[i].task.ID < scored[j].task.ID
	})

	return scored
}

// buildPhaseIndex maps phase names to their position in the phase order.
func buildPhaseIndex(phaseOrder []string) map[string]int {
	idx := make(map[string]int, len(phaseOrder))
	for i, name := range phaseOrder {
		idx[name] = i
	}
	return idx
}

// taskPhaseIndex returns the phase index for a task. Tasks with no phase
// or a phase not in the order are sorted after all phased tasks.
func taskPhaseIndex(task *model.Task, phaseIndex map[string]int) int {
	if task.Phase == "" {
		return len(phaseIndex)
	}
	if idx, ok := phaseIndex[task.Phase]; ok {
		return idx
	}
	return len(phaseIndex)
}

func buildRecommendations(
	scored []scoredTask,
	criticalPath map[string]bool,
	downstreamInfo map[string]DownstreamInfo,
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
			ScoreBreakdown:  st.components,
			DownstreamCount: downstreamInfo[st.task.ID].Count,
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

// BuildChildrenMap creates a map from parent task ID to its child tasks.
func BuildChildrenMap(tasks []*model.Task) map[string][]*model.Task {
	children := make(map[string][]*model.Task)
	for _, task := range tasks {
		if task.Parent != "" {
			children[task.Parent] = append(children[task.Parent], task)
		}
	}
	return children
}

// HasIncompleteChildren returns true if the task has any children that are not resolved.
func HasIncompleteChildren(task *model.Task, childrenMap map[string][]*model.Task) bool {
	for _, child := range childrenMap[task.ID] {
		if !child.Status.IsResolved() {
			return true
		}
	}
	return false
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

// IsActionable returns true if the task is pending/in-progress with all deps completed
// and no incomplete children.
func IsActionable(task *model.Task, taskMap map[string]*model.Task, childrenMap map[string][]*model.Task) bool {
	if task.Status != model.StatusPending && task.Status != model.StatusInProgress {
		return false
	}
	if HasUnmetDependencies(task, taskMap) {
		return false
	}
	return !HasIncompleteChildren(task, childrenMap)
}

// ScoreTask computes a score and reason list for an actionable task.
func ScoreTask(
	task *model.Task,
	criticalPath map[string]bool,
	downstreamInfo map[string]DownstreamInfo,
) (int, []string) {
	components, reasons := buildScore(task, criticalPath, downstreamInfo)
	return sumComponents(components), reasons
}

// ScoreTaskBreakdown returns the itemized scoring components for an actionable
// task (priority, critical path, downstream, effort). Their points sum to the
// score returned by ScoreTask. Phase scoring is added separately by callers.
func ScoreTaskBreakdown(
	task *model.Task,
	criticalPath map[string]bool,
	downstreamInfo map[string]DownstreamInfo,
) []ScoreComponent {
	components, _ := buildScore(task, criticalPath, downstreamInfo)
	return components
}

// buildScore is the single source of truth for a task's base score: it produces
// both the itemized components and the legacy reason strings from the same
// conditionals so the two can never drift.
func buildScore(
	task *model.Task,
	criticalPath map[string]bool,
	downstreamInfo map[string]DownstreamInfo,
) ([]ScoreComponent, []string) {
	components := make([]ScoreComponent, 0, 4)
	reasons := make([]string, 0)

	priorityComp, priorityReason := priorityComponent(task)
	components = append(components, priorityComp)
	if priorityReason != "" {
		reasons = append(reasons, priorityReason)
	}

	info := downstreamInfo[task.ID]
	mult := downstreamPriorityMultiplier(info.MaxPriority)

	if criticalPath[task.ID] {
		components = append(components, ScoreComponent{
			Label:      "critical path",
			Base:       ScoreCriticalPath,
			Multiplier: mult,
			Points:     int(float64(ScoreCriticalPath) * mult),
		})
		reasons = append(reasons, "on critical path")
	}

	if dc := info.Count; dc > 0 {
		base := min(dc*ScorePerDownstream, ScoreDownstreamMax)
		noun := "tasks"
		if dc == 1 {
			noun = "task"
		}
		components = append(components, ScoreComponent{
			Label:      fmt.Sprintf("downstream (unblocks %d %s)", dc, noun),
			Base:       base,
			Multiplier: mult,
			Points:     int(float64(base) * mult),
		})
		reasons = append(reasons, fmt.Sprintf("unblocks %d %s", dc, noun))
	}

	if effortComp, effortReason, ok := effortComponent(task); ok {
		components = append(components, effortComp)
		if effortReason != "" {
			reasons = append(reasons, effortReason)
		}
	}

	return components, reasons
}

// priorityComponent returns the priority score component and its legacy reason
// (empty for medium/low, which carry no reason string).
func priorityComponent(task *model.Task) (ScoreComponent, string) {
	switch task.Priority {
	case model.PriorityCritical:
		return ScoreComponent{Label: "priority: critical", Points: ScorePriorityCritical}, "critical priority"
	case model.PriorityHigh:
		return ScoreComponent{Label: "priority: high", Points: ScorePriorityHigh}, "high priority"
	case model.PriorityMedium:
		return ScoreComponent{Label: "priority: medium", Points: ScorePriorityMedium}, ""
	default:
		return ScoreComponent{Label: "priority: low", Points: ScorePriorityLow}, ""
	}
}

// effortComponent returns the effort score component and its legacy reason.
// The bool is false when the effort contributes no points (large/unset).
func effortComponent(task *model.Task) (ScoreComponent, string, bool) {
	switch task.Effort {
	case model.EffortSmall:
		return ScoreComponent{Label: "effort: small", Points: ScoreEffortSmall}, "quick win", true
	case model.EffortMedium:
		return ScoreComponent{Label: "effort: medium", Points: ScoreEffortMedium}, "", true
	default:
		return ScoreComponent{}, "", false
	}
}

// sumComponents returns the total points across all components.
func sumComponents(components []ScoreComponent) int {
	total := 0
	for _, c := range components {
		total += c.Points
	}
	return total
}

// scorePhase computes a phase-based scoring bonus.
// Tasks from earlier phases in the configured order get higher bonuses.
func scorePhase(task *model.Task, phaseOrder []string) (int, []string) {
	comp, reason := buildPhaseScore(task, phaseOrder)
	if comp == nil {
		return 0, nil
	}
	return comp.Points, []string{reason}
}

// buildPhaseScore returns the phase score component and its reason string, or
// (nil, "") when the task has no ranked phase bonus. Earlier phases in the
// configured order get higher bonuses.
func buildPhaseScore(task *model.Task, phaseOrder []string) (*ScoreComponent, string) {
	if task.Phase == "" || len(phaseOrder) == 0 {
		return nil, ""
	}
	for i, name := range phaseOrder {
		if name == task.Phase {
			bonus := ScorePhaseBase - (i * ScorePhaseDecay)
			if bonus <= 0 {
				return nil, ""
			}
			reason := fmt.Sprintf("phase %s", task.Phase)
			return &ScoreComponent{Label: reason, Points: bonus}, reason
		}
	}
	return nil, ""
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

		// Resolved tasks represent no remaining work —
		// they should not contribute depth to the critical path.
		if task.Status.IsResolved() {
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
		depDepth, exists := depthMap[depID]
		if exists && depDepth == targetDepth-1 {
			criticalPath[depID] = true
			markCriticalPathDependencies(depID, taskMap, depthMap, targetDepth-1, criticalPath)
		}
	}
}

// filterByScope returns only tasks whose Touches field contains the given scope.
func filterByScope(tasks []*model.Task, scope string) []*model.Task {
	var filtered []*model.Task
	for _, task := range tasks {
		for _, t := range task.Touches {
			if filter.MatchScope(scope, t) {
				filtered = append(filtered, task)
				break
			}
		}
	}
	return filtered
}

// filterByScopeExpanded returns tasks related to a scope by expanding via
// dependency components. It finds seed tasks that touch the scope, identifies
// their dependency components, then includes any actionable task sharing a
// component with a seed.
func filterByScopeExpanded(actionable []*model.Task, allTasks []*model.Task, scope string) []*model.Task {
	depComponents := buildDepComponents(allTasks)

	// Find seed components from all tasks (not just actionable) that touch the scope.
	seedComponents := make(map[string]bool)
	for _, task := range allTasks {
		for _, s := range task.Touches {
			if filter.MatchScope(scope, s) {
				if comp, ok := depComponents[task.ID]; ok {
					seedComponents[comp] = true
				}
				break
			}
		}
	}

	// Include actionable tasks that either touch the scope directly or share
	// a dependency component with a seed.
	actionableSet := make(map[string]bool, len(actionable))
	for _, t := range actionable {
		actionableSet[t.ID] = true
	}

	var filtered []*model.Task
	for _, task := range actionable {
		// Direct match.
		if touchesScope(task, scope) {
			filtered = append(filtered, task)
			continue
		}
		// Component expansion.
		if comp, ok := depComponents[task.ID]; ok && seedComponents[comp] {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func touchesScope(task *model.Task, scope string) bool {
	for _, s := range task.Touches {
		if filter.MatchScope(scope, s) {
			return true
		}
	}
	return false
}

// buildDepComponents computes connected components from dependency edges
// treated as undirected. Returns a map from task ID to a representative
// component ID (the lexicographically smallest ID in the component).
func buildDepComponents(tasks []*model.Task) map[string]string {
	adj := make(map[string][]string)
	ids := make(map[string]bool)
	for _, t := range tasks {
		ids[t.ID] = true
		for _, dep := range t.Dependencies {
			adj[t.ID] = append(adj[t.ID], dep)
			adj[dep] = append(adj[dep], t.ID)
		}
	}

	visited := make(map[string]bool)
	components := make(map[string]string)
	for id := range ids {
		if visited[id] {
			continue
		}
		// BFS to find component members.
		queue := []string{id}
		visited[id] = true
		var members []string
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			members = append(members, cur)
			for _, neighbor := range adj[cur] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
		// Use lexicographic minimum as representative.
		rep := members[0]
		for _, m := range members[1:] {
			if m < rep {
				rep = m
			}
		}
		for _, m := range members {
			components[m] = rep
		}
	}
	return components
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
