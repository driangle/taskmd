package next

import (
	"testing"

	"github.com/driangle/taskmd/sdk/go/model"
)

func makeTask(id string, status model.Status, priority model.Priority, deps []string) *model.Task {
	return &model.Task{
		ID:           id,
		Title:        "Task " + id,
		Status:       status,
		Priority:     priority,
		Dependencies: deps,
	}
}

func makeTaskWithTouches(id string, status model.Status, priority model.Priority, deps []string, touches []string) *model.Task {
	return &model.Task{
		ID:           id,
		Title:        "Task " + id,
		Status:       status,
		Priority:     priority,
		Dependencies: deps,
		Touches:      touches,
	}
}

func makeTaskWithPhase(id string, status model.Status, priority model.Priority, phase string) *model.Task {
	return &model.Task{
		ID:       id,
		Title:    "Task " + id,
		Status:   status,
		Priority: priority,
		Phase:    phase,
	}
}

func makeTaskWithParent(id string, status model.Status, priority model.Priority, parent string) *model.Task {
	return &model.Task{
		ID:       id,
		Title:    "Task " + id,
		Status:   status,
		Priority: priority,
		Parent:   parent,
	}
}

func TestRecommend_ArchivedCompletedDepSatisfied(t *testing.T) {
	// Task 002 depends on 001, but 001 is archived and completed.
	// 002 should be actionable.
	tasks := []*model.Task{
		makeTask("002", model.StatusPending, model.PriorityHigh, []string{"001"}),
	}
	archived := []*model.Task{
		makeTask("001", model.StatusCompleted, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{
		Limit:         10,
		ArchivedTasks: archived,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 1 {
		t.Fatalf("Expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].ID != "002" {
		t.Errorf("Expected task 002, got %s", recs[0].ID)
	}
}

func TestRecommend_ArchivedNonCompletedDepBlocks(t *testing.T) {
	// Task 002 depends on 001, which is archived but still pending.
	// 002 should be blocked.
	tasks := []*model.Task{
		makeTask("002", model.StatusPending, model.PriorityHigh, []string{"001"}),
	}
	archived := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{
		Limit:         10,
		ArchivedTasks: archived,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 0 {
		t.Errorf("Expected 0 recommendations (dep not completed), got %d", len(recs))
	}
}

func TestRecommend_ArchivedTasksNotRecommended(t *testing.T) {
	// Archived tasks should never appear in recommendations, even if actionable.
	tasks := []*model.Task{
		makeTask("002", model.StatusPending, model.PriorityHigh, nil),
	}
	archived := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{
		Limit:         10,
		ArchivedTasks: archived,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, rec := range recs {
		if rec.ID == "001" {
			t.Error("Archived task 001 should not appear in recommendations")
		}
	}

	if len(recs) != 1 || recs[0].ID != "002" {
		t.Errorf("Expected only task 002, got %v", recs)
	}
}

func TestRecommend_ActiveTaskPrecedenceOverArchived(t *testing.T) {
	// If the same ID exists in both active and archived, active wins.
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil),
		makeTask("002", model.StatusPending, model.PriorityMedium, []string{"001"}),
	}
	// Archived version has status=completed, but active version is pending.
	// Task 002 depends on 001 — since active 001 is pending, 002 should be blocked.
	archived := []*model.Task{
		makeTask("001", model.StatusCompleted, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{
		Limit:         10,
		ArchivedTasks: archived,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 001 is active+pending → actionable. 002 depends on 001 (pending) → blocked.
	if len(recs) != 1 {
		t.Fatalf("Expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].ID != "001" {
		t.Errorf("Expected task 001, got %s", recs[0].ID)
	}
}

func TestScoreTask_LowChainDoesNotOutscoreMediumTask(t *testing.T) {
	// A low-priority task unblocking 5 low-priority tasks should NOT outscore
	// a standalone medium-priority task.
	criticalPath := map[string]bool{"low1": true}
	downstreamInfo := map[string]DownstreamInfo{
		"low1": {Count: 5, MaxPriority: model.PriorityLow},
		"med1": {Count: 0},
	}

	lowTask := &model.Task{ID: "low1", Priority: model.PriorityLow}
	medTask := &model.Task{ID: "med1", Priority: model.PriorityMedium}

	lowScore, _ := ScoreTask(lowTask, criticalPath, downstreamInfo)
	medScore, _ := ScoreTask(medTask, map[string]bool{}, downstreamInfo)

	if lowScore >= medScore {
		t.Errorf("Low-priority task with all-low downstream chain (score=%d) should not outscore standalone medium task (score=%d)",
			lowScore, medScore)
	}
}

func TestScoreTask_MixedChainGetsFullDownstreamBonus(t *testing.T) {
	// A low-priority task that unblocks a high-priority task should still get
	// a full downstream bonus (multiplier = 1.0).
	downstreamInfo := map[string]DownstreamInfo{
		"low1": {Count: 1, MaxPriority: model.PriorityHigh},
	}

	task := &model.Task{ID: "low1", Priority: model.PriorityLow}
	score, _ := ScoreTask(task, map[string]bool{}, downstreamInfo)

	// Expected: base low (10) + full downstream bonus (1 * 3 * 1.0 = 3) = 13
	expectedScore := ScorePriorityLow + 1*ScorePerDownstream
	if score != expectedScore {
		t.Errorf("Mixed chain score = %d, want %d (full downstream bonus for high-priority downstream)", score, expectedScore)
	}
}

func TestScoreTask_HighChainPreservesExistingBehavior(t *testing.T) {
	// A task on the critical path with high-priority downstream tasks should
	// get full bonuses (same as before the priority-aware change).
	criticalPath := map[string]bool{"t1": true}
	downstreamInfo := map[string]DownstreamInfo{
		"t1": {Count: 3, MaxPriority: model.PriorityCritical},
	}

	task := &model.Task{ID: "t1", Priority: model.PriorityHigh}
	score, _ := ScoreTask(task, criticalPath, downstreamInfo)

	// Expected: high (30) + critical path (15 * 1.0) + downstream (min(9,15) * 1.0 = 9) = 54
	expectedScore := ScorePriorityHigh + ScoreCriticalPath + min(3*ScorePerDownstream, ScoreDownstreamMax)
	if score != expectedScore {
		t.Errorf("High/critical chain score = %d, want %d", score, expectedScore)
	}
}

func TestScoreTaskBreakdown_ComponentsSumToScore(t *testing.T) {
	criticalPath := map[string]bool{"t1": true}
	cases := []struct {
		name           string
		task           *model.Task
		downstreamInfo map[string]DownstreamInfo
	}{
		{
			name:           "high on critical path with downstream and effort",
			task:           &model.Task{ID: "t1", Priority: model.PriorityHigh, Effort: model.EffortSmall},
			downstreamInfo: map[string]DownstreamInfo{"t1": {Count: 3, MaxPriority: model.PriorityCritical}},
		},
		{
			name:           "low with scaled downstream",
			task:           &model.Task{ID: "t2", Priority: model.PriorityLow},
			downstreamInfo: map[string]DownstreamInfo{"t2": {Count: 5, MaxPriority: model.PriorityLow}},
		},
		{
			name:           "medium standalone",
			task:           &model.Task{ID: "t3", Priority: model.PriorityMedium},
			downstreamInfo: map[string]DownstreamInfo{},
		},
		{
			name:           "critical with medium effort",
			task:           &model.Task{ID: "t4", Priority: model.PriorityCritical, Effort: model.EffortMedium},
			downstreamInfo: map[string]DownstreamInfo{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cp := map[string]bool{}
			if criticalPath[tc.task.ID] {
				cp = criticalPath
			}
			components := ScoreTaskBreakdown(tc.task, cp, tc.downstreamInfo)
			score, _ := ScoreTask(tc.task, cp, tc.downstreamInfo)

			sum := 0
			for _, c := range components {
				sum += c.Points
			}
			if sum != score {
				t.Errorf("components sum = %d, want score %d (components: %+v)", sum, score, components)
			}
		})
	}
}

func TestScoreTaskBreakdown_ScaledBonusesExposeBaseAndMultiplier(t *testing.T) {
	// Low-priority downstream chain scales bonuses by 0.5 (medium max) — verify
	// the base and multiplier are surfaced and that Points = int(base * mult).
	criticalPath := map[string]bool{"t1": true}
	downstreamInfo := map[string]DownstreamInfo{
		"t1": {Count: 3, MaxPriority: model.PriorityMedium},
	}
	task := &model.Task{ID: "t1", Priority: model.PriorityHigh}

	components := ScoreTaskBreakdown(task, criticalPath, downstreamInfo)

	var crit, down *ScoreComponent
	for i := range components {
		switch components[i].Label {
		case "critical path":
			crit = &components[i]
		case "downstream (unblocks 3 tasks)":
			down = &components[i]
		}
	}

	if crit == nil {
		t.Fatal("expected a critical path component")
	}
	if crit.Base != ScoreCriticalPath || crit.Multiplier != 0.5 {
		t.Errorf("critical path base/mult = %d/%.2f, want %d/0.50", crit.Base, crit.Multiplier, ScoreCriticalPath)
	}
	if crit.Points != int(float64(crit.Base)*crit.Multiplier) {
		t.Errorf("critical path points = %d, want %d", crit.Points, int(float64(crit.Base)*crit.Multiplier))
	}

	if down == nil {
		t.Fatal("expected a downstream component")
	}
	if down.Base != min(3*ScorePerDownstream, ScoreDownstreamMax) || down.Multiplier != 0.5 {
		t.Errorf("downstream base/mult = %d/%.2f, want %d/0.50", down.Base, down.Multiplier, min(3*ScorePerDownstream, ScoreDownstreamMax))
	}
	if down.Points != int(float64(down.Base)*down.Multiplier) {
		t.Errorf("downstream points = %d, want %d", down.Points, int(float64(down.Base)*down.Multiplier))
	}
}

func TestScoreTaskBreakdown_FlatComponentsHaveNoMultiplier(t *testing.T) {
	task := &model.Task{ID: "t1", Priority: model.PriorityHigh, Effort: model.EffortSmall}
	components := ScoreTaskBreakdown(task, map[string]bool{}, map[string]DownstreamInfo{})

	if len(components) != 2 {
		t.Fatalf("expected 2 components (priority, effort), got %d: %+v", len(components), components)
	}
	for _, c := range components {
		if c.Multiplier != 0 || c.Base != 0 {
			t.Errorf("flat component %q has base/mult %d/%.2f, want 0/0", c.Label, c.Base, c.Multiplier)
		}
	}
	if components[0].Label != "priority: high" || components[0].Points != ScorePriorityHigh {
		t.Errorf("priority component = %+v", components[0])
	}
	if components[1].Label != "effort: small" || components[1].Points != ScoreEffortSmall {
		t.Errorf("effort component = %+v", components[1])
	}
}

func TestBuildPhaseScore(t *testing.T) {
	order := []string{"v0.2", "v0.3", "v0.4"}

	comp, reason := buildPhaseScore(makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v0.3"), order)
	if comp == nil {
		t.Fatal("expected a phase component")
	}
	if comp.Points != ScorePhaseBase-ScorePhaseDecay {
		t.Errorf("phase points = %d, want %d", comp.Points, ScorePhaseBase-ScorePhaseDecay)
	}
	if comp.Label != "phase v0.3" || reason != "phase v0.3" {
		t.Errorf("label/reason = %q/%q, want phase v0.3", comp.Label, reason)
	}

	if comp, _ := buildPhaseScore(makeTask("001", model.StatusPending, model.PriorityMedium, nil), order); comp != nil {
		t.Errorf("no-phase task should yield nil component, got %+v", comp)
	}
}

func TestRecommend_MediumTaskRanksAboveLowChain(t *testing.T) {
	// Integration test: an unblocked medium-priority task should rank higher than
	// a low-priority task whose entire downstream chain is low priority.
	//
	// low1 (low, no deps) -> low2 (low, depends on low1) -> low3 -> low4 -> low5
	// med1 (medium, no deps, standalone)
	tasks := []*model.Task{
		makeTask("low1", model.StatusPending, model.PriorityLow, nil),
		makeTask("low2", model.StatusPending, model.PriorityLow, []string{"low1"}),
		makeTask("low3", model.StatusPending, model.PriorityLow, []string{"low2"}),
		makeTask("low4", model.StatusPending, model.PriorityLow, []string{"low3"}),
		makeTask("low5", model.StatusPending, model.PriorityLow, []string{"low4"}),
		makeTask("med1", model.StatusPending, model.PriorityMedium, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find positions
	var medRank, lowRank int
	for _, rec := range recs {
		if rec.ID == "med1" {
			medRank = rec.Rank
		}
		if rec.ID == "low1" {
			lowRank = rec.Rank
		}
	}

	if medRank == 0 {
		t.Fatal("med1 not found in recommendations")
	}
	if lowRank == 0 {
		t.Fatal("low1 not found in recommendations")
	}
	if medRank >= lowRank {
		t.Errorf("Medium task (rank %d) should rank above low task with all-low downstream chain (rank %d)", medRank, lowRank)
	}
}

func TestCalculateCriticalPathTasks_IgnoresCompletedDependencies(t *testing.T) {
	// Scenario: tasks with completed dependencies should not have inflated depth.
	//
	// Graph:
	//   A (completed, no deps)
	//   B (pending, depends on A)  — A is done, so B's remaining depth is 1
	//   C (pending, no deps)       — depth 1
	//   D (pending, depends on C)  — C is pending, real remaining chain depth 2
	//
	// The only real remaining dependency chain is C → D.
	// B should NOT be on the critical path because its dependency A is completed.
	tasks := []*model.Task{
		{ID: "A", Status: model.StatusCompleted, Dependencies: nil},
		{ID: "B", Status: model.StatusPending, Dependencies: []string{"A"}},
		{ID: "C", Status: model.StatusPending, Dependencies: nil},
		{ID: "D", Status: model.StatusPending, Dependencies: []string{"C"}},
	}

	taskMap := BuildTaskMap(tasks)
	criticalPath := CalculateCriticalPathTasks(tasks, taskMap)

	// C and D should be on the critical path (the only real remaining chain)
	if !criticalPath["C"] {
		t.Error("Expected task C to be on critical path")
	}
	if !criticalPath["D"] {
		t.Error("Expected task D to be on critical path")
	}

	// B should NOT be on the critical path — its dependency A is already completed
	if criticalPath["B"] {
		t.Error("Task B should NOT be on critical path: its dependency A is completed")
	}

	// A is completed and should not be on the critical path either
	if criticalPath["A"] {
		t.Error("Completed task A should NOT be on critical path")
	}
}

func TestCalculateCriticalPathTasks_PendingChainIsCritical(t *testing.T) {
	// When all tasks in a chain are pending, the longest chain is the critical path.
	//
	// Graph:
	//   001 (pending, no deps)         — depth 1
	//   002 (pending, depends on 001)  — depth 2
	//   003 (pending, depends on 002)  — depth 3
	//   004 (pending, no deps)         — depth 1
	//
	// Critical path: 001 → 002 → 003
	tasks := []*model.Task{
		{ID: "001", Status: model.StatusPending, Dependencies: nil},
		{ID: "002", Status: model.StatusPending, Dependencies: []string{"001"}},
		{ID: "003", Status: model.StatusPending, Dependencies: []string{"002"}},
		{ID: "004", Status: model.StatusPending, Dependencies: nil},
	}

	taskMap := BuildTaskMap(tasks)
	criticalPath := CalculateCriticalPathTasks(tasks, taskMap)

	for _, id := range []string{"001", "002", "003"} {
		if !criticalPath[id] {
			t.Errorf("Expected task %s to be on critical path", id)
		}
	}

	if criticalPath["004"] {
		t.Error("Task 004 should NOT be on critical path (shorter parallel path)")
	}
}

func TestCalculateCriticalPathTasks_MixedCompletedPendingChain(t *testing.T) {
	// A longer chain where early tasks are completed should have reduced effective depth.
	//
	// Graph:
	//   A (completed, no deps)
	//   B (completed, depends on A)
	//   C (pending, depends on B)    — B is done, so C's remaining depth is 1
	//   D (pending, no deps)         — depth 1
	//   E (pending, depends on D)    — depth 2
	//   F (pending, depends on E)    — depth 3
	//
	// Remaining chain D → E → F is longer than just C.
	// Critical path should be D → E → F only.
	tasks := []*model.Task{
		{ID: "A", Status: model.StatusCompleted, Dependencies: nil},
		{ID: "B", Status: model.StatusCompleted, Dependencies: []string{"A"}},
		{ID: "C", Status: model.StatusPending, Dependencies: []string{"B"}},
		{ID: "D", Status: model.StatusPending, Dependencies: nil},
		{ID: "E", Status: model.StatusPending, Dependencies: []string{"D"}},
		{ID: "F", Status: model.StatusPending, Dependencies: []string{"E"}},
	}

	taskMap := BuildTaskMap(tasks)
	criticalPath := CalculateCriticalPathTasks(tasks, taskMap)

	// D → E → F is the real critical path
	for _, id := range []string{"D", "E", "F"} {
		if !criticalPath[id] {
			t.Errorf("Expected task %s to be on critical path", id)
		}
	}

	// C should NOT be on critical path (only 1 remaining step, shorter than D→E→F)
	if criticalPath["C"] {
		t.Error("Task C should NOT be on critical path (shorter remaining chain)")
	}

	// Completed tasks should not be on critical path
	if criticalPath["A"] {
		t.Error("Completed task A should NOT be on critical path")
	}
	if criticalPath["B"] {
		t.Error("Completed task B should NOT be on critical path")
	}
}

func TestHasIncompleteChildren(t *testing.T) {
	tests := []struct {
		name     string
		task     *model.Task
		children []*model.Task
		expected bool
	}{
		{
			name:     "no children",
			task:     makeTask("P1", model.StatusPending, model.PriorityMedium, nil),
			children: nil,
			expected: false,
		},
		{
			name: "all children completed",
			task: makeTask("P1", model.StatusPending, model.PriorityMedium, nil),
			children: []*model.Task{
				makeTaskWithParent("C1", model.StatusCompleted, model.PriorityMedium, "P1"),
				makeTaskWithParent("C2", model.StatusCompleted, model.PriorityMedium, "P1"),
			},
			expected: false,
		},
		{
			name: "one pending child",
			task: makeTask("P1", model.StatusPending, model.PriorityMedium, nil),
			children: []*model.Task{
				makeTaskWithParent("C1", model.StatusCompleted, model.PriorityMedium, "P1"),
				makeTaskWithParent("C2", model.StatusPending, model.PriorityMedium, "P1"),
			},
			expected: true,
		},
		{
			name: "cancelled child counts as resolved",
			task: makeTask("P1", model.StatusPending, model.PriorityMedium, nil),
			children: []*model.Task{
				makeTaskWithParent("C1", model.StatusCancelled, model.PriorityMedium, "P1"),
			},
			expected: false,
		},
		{
			name: "in-progress child is incomplete",
			task: makeTask("P1", model.StatusPending, model.PriorityMedium, nil),
			children: []*model.Task{
				makeTaskWithParent("C1", model.StatusInProgress, model.PriorityMedium, "P1"),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			childrenMap := BuildChildrenMap(tt.children)
			got := HasIncompleteChildren(tt.task, childrenMap)
			if got != tt.expected {
				t.Errorf("HasIncompleteChildren() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsActionable_WithChildren(t *testing.T) {
	parent := makeTask("P1", model.StatusPending, model.PriorityMedium, nil)
	child := makeTaskWithParent("C1", model.StatusPending, model.PriorityMedium, "P1")

	tasks := []*model.Task{parent, child}
	taskMap := BuildTaskMap(tasks)
	childrenMap := BuildChildrenMap(tasks)

	// Parent with incomplete child should not be actionable
	if IsActionable(parent, taskMap, childrenMap) {
		t.Error("Parent with incomplete child should not be actionable")
	}

	// Child itself should be actionable (no deps, no children of its own)
	if !IsActionable(child, taskMap, childrenMap) {
		t.Error("Child task should be actionable")
	}
}

func TestRecommend_ParentExcludedWhenChildrenIncomplete(t *testing.T) {
	parent := makeTask("P1", model.StatusPending, model.PriorityHigh, nil)
	child := makeTaskWithParent("C1", model.StatusPending, model.PriorityMedium, "P1")

	tasks := []*model.Task{parent, child}

	recs, err := Recommend(tasks, Options{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, rec := range recs {
		if rec.ID == "P1" {
			t.Error("Parent P1 should not be recommended while child C1 is incomplete")
		}
	}

	if len(recs) != 1 || recs[0].ID != "C1" {
		t.Errorf("Expected only child C1, got %v", recs)
	}
}

func TestRecommend_ParentIncludedWhenAllChildrenResolved(t *testing.T) {
	parent := makeTask("P1", model.StatusPending, model.PriorityHigh, nil)
	childCompleted := makeTaskWithParent("C1", model.StatusCompleted, model.PriorityMedium, "P1")
	childCancelled := makeTaskWithParent("C2", model.StatusCancelled, model.PriorityMedium, "P1")

	tasks := []*model.Task{parent, childCompleted, childCancelled}

	recs, err := Recommend(tasks, Options{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, rec := range recs {
		if rec.ID == "P1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Parent P1 should be recommended when all children are resolved")
	}
}

func TestRecommend_ScopeFiltering(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithTouches("001", model.StatusPending, model.PriorityHigh, nil, []string{"web", "api"}),
		makeTaskWithTouches("002", model.StatusPending, model.PriorityMedium, nil, []string{"cli"}),
		makeTaskWithTouches("003", model.StatusPending, model.PriorityLow, nil, []string{"web"}),
		makeTask("004", model.StatusPending, model.PriorityHigh, nil), // no touches
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 2 {
		t.Fatalf("Expected 2 recommendations for scope 'web', got %d", len(recs))
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}
	if !ids["001"] || !ids["003"] {
		t.Errorf("Expected tasks 001 and 003, got %v", recs)
	}
}

func TestRecommend_ScopeNoMatches(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithTouches("001", model.StatusPending, model.PriorityHigh, nil, []string{"web"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 0 {
		t.Errorf("Expected 0 recommendations for non-matching scope, got %d", len(recs))
	}
}

func TestRecommend_ScopeWithoutScopeUnchanged(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithTouches("001", model.StatusPending, model.PriorityHigh, nil, []string{"web"}),
		makeTask("002", model.StatusPending, model.PriorityMedium, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 2 {
		t.Errorf("Without scope, expected all 2 actionable tasks, got %d", len(recs))
	}
}

func TestRecommend_ScopeCombinedWithQuickWins(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithTouches("001", model.StatusPending, model.PriorityHigh, nil, []string{"web"}),
		makeTaskWithTouches("002", model.StatusPending, model.PriorityMedium, nil, []string{"web"}),
		makeTaskWithTouches("003", model.StatusPending, model.PriorityLow, nil, []string{"cli"}),
	}
	tasks[0].Effort = model.EffortLarge
	tasks[1].Effort = model.EffortSmall
	tasks[2].Effort = model.EffortSmall

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "web", QuickWins: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only task 002 matches: scope=web AND effort=small
	if len(recs) != 1 {
		t.Fatalf("Expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].ID != "002" {
		t.Errorf("Expected task 002, got %s", recs[0].ID)
	}
}

func TestRecommend_ScopeExpandsDependencies(t *testing.T) {
	// Task 002 (touches: ["web"]) depends on task 001 (no touches).
	// With --scope web, task 001 should also appear because it blocks a web task.
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil),                                           // no touches, but blocks 002
		makeTaskWithTouches("002", model.StatusPending, model.PriorityMedium, []string{"001"}, []string{"web"}), // touches web, depends on 001
		makeTask("003", model.StatusPending, model.PriorityLow, nil),                                            // unrelated, no touches
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "web"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Task 001 should be included (blocks a web task via dependency component).
	// Task 002 is blocked (dep 001 pending), so only 001 is actionable.
	// Task 003 is unrelated and should be excluded.
	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}

	if !ids["001"] {
		t.Errorf("Expected task 001 (blocking dependency of web task) to be included, got %v", ids)
	}
	if ids["003"] {
		t.Errorf("Task 003 (unrelated) should not appear in scope=web results")
	}
}

func TestRecommend_ScopeExactSkipsExpansion(t *testing.T) {
	// With ScopeExact=true, only tasks that directly touch the scope should appear.
	// Task 001 blocks a web task but doesn't touch "web" itself.
	tasks := []*model.Task{
		makeTask("001", model.StatusPending, model.PriorityHigh, nil),
		makeTaskWithTouches("002", model.StatusPending, model.PriorityMedium, []string{"001"}, []string{"web"}),
		makeTaskWithTouches("003", model.StatusPending, model.PriorityLow, nil, []string{"web"}),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "web", ScopeExact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only task 003 is actionable AND directly touches web.
	// Task 002 touches web but is blocked. Task 001 doesn't touch web.
	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}

	if ids["001"] {
		t.Errorf("With ScopeExact, task 001 (no touches) should not appear")
	}
	if len(recs) != 1 || recs[0].ID != "003" {
		t.Errorf("Expected only task 003, got %v", ids)
	}
}

func TestRecommend_ScopeWildcard(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithTouches("001", model.StatusPending, model.PriorityHigh, nil, []string{"cli/graph"}),
		makeTaskWithTouches("002", model.StatusPending, model.PriorityMedium, nil, []string{"cli/next"}),
		makeTaskWithTouches("003", model.StatusPending, model.PriorityLow, nil, []string{"web/board"}),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "cli/*", ScopeExact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 2 {
		t.Fatalf("Expected 2 recommendations for scope 'cli/*', got %d", len(recs))
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}
	if !ids["001"] || !ids["002"] {
		t.Errorf("Expected tasks 001 and 002, got %v", ids)
	}
	if ids["003"] {
		t.Errorf("Task 003 (web/board) should not match cli/*")
	}
}

func TestRecommend_ScopeWildcardExpanded(t *testing.T) {
	// Task 001 touches cli/graph. Task 004 depends on 001 but touches nothing.
	// With wildcard scope "cli/*" and expansion, 004 should also appear.
	tasks := []*model.Task{
		makeTaskWithTouches("001", model.StatusPending, model.PriorityHigh, nil, []string{"cli/graph"}),
		makeTaskWithTouches("003", model.StatusPending, model.PriorityLow, nil, []string{"web/board"}),
		makeTask("004", model.StatusPending, model.PriorityMedium, []string{"001"}),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Scope: "cli/*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}
	if !ids["001"] {
		t.Errorf("Task 001 (cli/graph) should match cli/*")
	}
	if ids["003"] {
		t.Errorf("Task 003 (web/board) should not match cli/*")
	}
}

func TestScorePhase_FirstPhaseGetsFullBonus(t *testing.T) {
	task := makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v0.2")
	order := []string{"v0.2", "v0.3", "v0.4"}

	score, reasons := scorePhase(task, order)

	if score != ScorePhaseBase {
		t.Errorf("First phase score = %d, want %d", score, ScorePhaseBase)
	}
	if len(reasons) != 1 || reasons[0] != "phase v0.2" {
		t.Errorf("reasons = %v, want [phase v0.2]", reasons)
	}
}

func TestScorePhase_SecondPhaseGetsDecayedBonus(t *testing.T) {
	task := makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v0.3")
	order := []string{"v0.2", "v0.3", "v0.4"}

	score, _ := scorePhase(task, order)

	expected := ScorePhaseBase - ScorePhaseDecay
	if score != expected {
		t.Errorf("Second phase score = %d, want %d", score, expected)
	}
}

func TestScorePhase_NoPhaseGetsNoBonus(t *testing.T) {
	task := makeTask("001", model.StatusPending, model.PriorityMedium, nil)
	order := []string{"v0.2", "v0.3"}

	score, reasons := scorePhase(task, order)

	if score != 0 {
		t.Errorf("No phase score = %d, want 0", score)
	}
	if len(reasons) != 0 {
		t.Errorf("reasons = %v, want empty", reasons)
	}
}

func TestScorePhase_UnknownPhaseGetsNoBonus(t *testing.T) {
	task := makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v9.9")
	order := []string{"v0.2", "v0.3"}

	score, _ := scorePhase(task, order)

	if score != 0 {
		t.Errorf("Unknown phase score = %d, want 0", score)
	}
}

func TestScorePhase_NoPhaseOrderConfigured(t *testing.T) {
	task := makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v0.2")

	score, _ := scorePhase(task, nil)

	if score != 0 {
		t.Errorf("No phase order score = %d, want 0", score)
	}
}

func TestScorePhase_LatePhasePositionZeroBonus(t *testing.T) {
	task := makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v0.7")
	order := []string{"v0.2", "v0.3", "v0.4", "v0.5", "v0.6", "v0.7"}

	score, _ := scorePhase(task, order)

	// Position 5: bonus = 25 - (5 * 5) = 0
	if score != 0 {
		t.Errorf("Late position phase score = %d, want 0", score)
	}
}

func TestRecommend_PhaseFilter(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithPhase("001", model.StatusPending, model.PriorityHigh, "v0.2"),
		makeTaskWithPhase("002", model.StatusPending, model.PriorityMedium, "v0.3"),
		makeTask("003", model.StatusPending, model.PriorityLow, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Phase: "v0.2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 1 || recs[0].ID != "001" {
		t.Errorf("Expected only task 001 with phase v0.2, got %v", recs)
	}
}

func TestRecommend_PhaseOrderAffectsRanking(t *testing.T) {
	tasks := []*model.Task{
		makeTaskWithPhase("001", model.StatusPending, model.PriorityMedium, "v0.3"),
		makeTaskWithPhase("002", model.StatusPending, model.PriorityMedium, "v0.2"),
	}

	recs, err := Recommend(tasks, Options{
		Limit:      10,
		PhaseOrder: []string{"v0.2", "v0.3"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) < 2 {
		t.Fatalf("Expected 2 recommendations, got %d", len(recs))
	}
	// Task 002 (v0.2, first phase) should rank above task 001 (v0.3, second)
	if recs[0].ID != "002" {
		t.Errorf("Expected task 002 (v0.2) to rank first, got %s", recs[0].ID)
	}
}

func TestRecommend_StrictPriority_HigherPriorityRanksFirst(t *testing.T) {
	// A medium task laden with bonuses (small effort + critical-path + downstream)
	// normally outscores a bonus-less critical task. --strict-priority must still
	// rank the critical task first.
	med := &model.Task{ID: "med", Title: "Medium", Status: model.StatusPending, Priority: model.PriorityMedium, Effort: model.EffortSmall}
	// dep (high) depends on med, making med a critical-path node with a
	// high-priority downstream (full-strength bonuses). dep itself is blocked.
	dep := makeTask("dep", model.StatusPending, model.PriorityHigh, []string{"med"})
	crit := makeTask("crit", model.StatusPending, model.PriorityCritical, nil)
	tasks := []*model.Task{med, dep, crit}

	// Sanity: without strict-priority, med outscores crit.
	loose, err := Recommend(tasks, Options{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loose[0].ID != "med" {
		t.Fatalf("precondition: expected med to rank first without strict-priority, got %s", loose[0].ID)
	}

	recs, err := Recommend(tasks, Options{Limit: 10, StrictPriority: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs[0].ID != "crit" {
		t.Errorf("Expected critical task first under --strict-priority, got %s", recs[0].ID)
	}
	if recs[1].ID != "med" {
		t.Errorf("Expected medium task second under --strict-priority, got %s", recs[1].ID)
	}
}

func TestRecommend_StrictPriority_ScoreBreaksTieWithinTier(t *testing.T) {
	// Two medium tasks: one has a small-effort bonus, so within the medium tier
	// it must rank first via the score tiebreak.
	plain := &model.Task{ID: "plain", Title: "Plain", Status: model.StatusPending, Priority: model.PriorityMedium}
	quick := &model.Task{ID: "quick", Title: "Quick", Status: model.StatusPending, Priority: model.PriorityMedium, Effort: model.EffortSmall}
	tasks := []*model.Task{plain, quick}

	recs, err := Recommend(tasks, Options{Limit: 10, StrictPriority: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recs[0].ID != "quick" {
		t.Errorf("Within the medium tier, expected higher-scoring 'quick' first, got %s", recs[0].ID)
	}
}

func TestRecommend_StrictPriority_LowAndUnsetTie(t *testing.T) {
	// low and unset priority both map to weight 1, so they tie under
	// strict-priority and fall through to the score/ID tiebreak.
	low := makeTask("b-low", model.StatusPending, model.PriorityLow, nil)
	unset := makeTask("a-unset", model.StatusPending, model.Priority(""), nil)
	tasks := []*model.Task{low, unset}

	recs, err := Recommend(tasks, Options{Limit: 10, StrictPriority: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Equal weight and equal score → ID tiebreak: "a-unset" < "b-low".
	if recs[0].ID != "a-unset" {
		t.Errorf("Expected ID tiebreak (a-unset first) for tied low/unset priorities, got %s", recs[0].ID)
	}
}

func TestRecommend_StrictPhasesAndPriority_PhasePrimary(t *testing.T) {
	// v0.2 (first phase) with a low-priority task; v0.3 (later phase) with a
	// critical task. With both flags, phase is primary: the low v0.2 task ranks
	// ahead of the critical v0.3 task.
	lowEarly := makeTaskWithPhase("low-early", model.StatusPending, model.PriorityLow, "v0.2")
	critLate := makeTaskWithPhase("crit-late", model.StatusPending, model.PriorityCritical, "v0.3")
	// Second v0.2 task, high priority, to confirm priority is secondary within a phase.
	highEarly := makeTaskWithPhase("high-early", model.StatusPending, model.PriorityHigh, "v0.2")
	tasks := []*model.Task{lowEarly, critLate, highEarly}

	recs, err := Recommend(tasks, Options{
		Limit:          10,
		PhaseOrder:     []string{"v0.2", "v0.3"},
		StrictPhases:   true,
		StrictPriority: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("Expected 3 recommendations, got %d", len(recs))
	}
	// Phase primary: both v0.2 tasks before the v0.3 task.
	if recs[2].ID != "crit-late" {
		t.Errorf("Expected critical v0.3 task last (phase primary), got %s", recs[2].ID)
	}
	// Priority secondary within v0.2: high before low.
	if recs[0].ID != "high-early" || recs[1].ID != "low-early" {
		t.Errorf("Within v0.2 expected high-early then low-early, got %s then %s", recs[0].ID, recs[1].ID)
	}
}

// makeTaskWithParentDeps builds a task that has both a parent and dependencies.
func makeTaskWithParentDeps(id string, status model.Status, priority model.Priority, parent string, deps []string) *model.Task {
	return &model.Task{
		ID:           id,
		Title:        "Task " + id,
		Status:       status,
		Priority:     priority,
		Parent:       parent,
		Dependencies: deps,
	}
}

func TestRecommend_RootLeafReturnsUpstreamOnly(t *testing.T) {
	// R depends on A; A is actionable. X is unrelated and actionable.
	// --root R should return only A (R is blocked by A; X is not reachable).
	tasks := []*model.Task{
		makeTask("R", model.StatusPending, model.PriorityHigh, []string{"A"}),
		makeTask("A", model.StatusPending, model.PriorityMedium, nil),
		makeTask("X", model.StatusPending, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Root: "R"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 1 || recs[0].ID != "A" {
		t.Fatalf("Expected only upstream task A, got %v", recs)
	}
}

func TestRecommend_RootParentReturnsActionableSubtasks(t *testing.T) {
	// P is a parent with two pending children; X is unrelated.
	// --root P should return the actionable subtasks (P itself is blocked by
	// its incomplete children, X is not reachable).
	tasks := []*model.Task{
		makeTask("P", model.StatusPending, model.PriorityHigh, nil),
		makeTaskWithParent("C1", model.StatusPending, model.PriorityMedium, "P"),
		makeTaskWithParent("C2", model.StatusPending, model.PriorityMedium, "P"),
		makeTask("X", model.StatusPending, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Root: "P"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := map[string]bool{}
	for _, rec := range recs {
		ids[rec.ID] = true
	}
	if len(recs) != 2 || !ids["C1"] || !ids["C2"] {
		t.Fatalf("Expected actionable subtasks C1 and C2, got %v", recs)
	}
}

func TestRecommend_RootNestedSubtree(t *testing.T) {
	// P -> C -> G (grandchild). Only the deepest leaf G is actionable.
	// --root P should reach the whole subtree and return G.
	tasks := []*model.Task{
		makeTask("P", model.StatusPending, model.PriorityHigh, nil),
		makeTaskWithParent("C", model.StatusPending, model.PriorityMedium, "P"),
		makeTaskWithParentDeps("G", model.StatusPending, model.PriorityLow, "C", nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Root: "P"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 1 || recs[0].ID != "G" {
		t.Fatalf("Expected deepest subtask G, got %v", recs)
	}
}

func TestRecommend_RootUnknownIDErrors(t *testing.T) {
	tasks := []*model.Task{
		makeTask("A", model.StatusPending, model.PriorityHigh, nil),
	}

	_, err := Recommend(tasks, Options{Limit: 10, Root: "NOPE"})
	if err == nil {
		t.Fatal("Expected error for unknown root ID, got nil")
	}
	if got := err.Error(); got != "root task NOPE not found" {
		t.Errorf("Expected 'root task NOPE not found', got %q", got)
	}
}

func TestRecommend_RootCombinedWithFilter(t *testing.T) {
	// R depends on A (high) and B (low), both actionable.
	// --root R with a priority filter should narrow to A only.
	tasks := []*model.Task{
		makeTask("R", model.StatusPending, model.PriorityHigh, []string{"A", "B"}),
		makeTask("A", model.StatusPending, model.PriorityHigh, nil),
		makeTask("B", model.StatusPending, model.PriorityLow, nil),
	}

	recs, err := Recommend(tasks, Options{
		Limit:   10,
		Root:    "R",
		Filters: []string{"priority=high"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 1 || recs[0].ID != "A" {
		t.Fatalf("Expected only high-priority upstream task A, got %v", recs)
	}
}

func TestRecommend_RootIncludesActionableRootItself(t *testing.T) {
	// A leaf root with satisfied dependencies is itself actionable and
	// should be returned.
	tasks := []*model.Task{
		makeTask("R", model.StatusPending, model.PriorityHigh, []string{"A"}),
		makeTask("A", model.StatusCompleted, model.PriorityMedium, nil),
		makeTask("X", model.StatusPending, model.PriorityHigh, nil),
	}

	recs, err := Recommend(tasks, Options{Limit: 10, Root: "R"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) != 1 || recs[0].ID != "R" {
		t.Fatalf("Expected root R itself, got %v", recs)
	}
}
