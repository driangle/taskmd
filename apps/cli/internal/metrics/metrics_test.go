package metrics

import (
	"testing"

	"github.com/driangle/taskmd/apps/cli/internal/model"
)

func TestCalculate_EmptyTaskList(t *testing.T) {
	tasks := []*model.Task{}
	m := Calculate(tasks)

	if m.TotalTasks != 0 {
		t.Errorf("expected TotalTasks=0, got %d", m.TotalTasks)
	}
	if m.BlockedTasksCount != 0 {
		t.Errorf("expected BlockedTasksCount=0, got %d", m.BlockedTasksCount)
	}
	if m.CriticalPathLength != 0 {
		t.Errorf("expected CriticalPathLength=0, got %d", m.CriticalPathLength)
	}
	if m.AvgDependenciesPerTask != 0 {
		t.Errorf("expected AvgDependenciesPerTask=0, got %f", m.AvgDependenciesPerTask)
	}
}

func TestCalculate_SingleTask(t *testing.T) {
	tasks := []*model.Task{
		{
			ID:           "1",
			Title:        "Task 1",
			Status:       model.StatusPending,
			Priority:     model.PriorityHigh,
			Effort:       model.EffortMedium,
			Dependencies: []string{},
		},
	}

	m := Calculate(tasks)

	if m.TotalTasks != 1 {
		t.Errorf("expected TotalTasks=1, got %d", m.TotalTasks)
	}
	if m.BlockedTasksCount != 0 {
		t.Errorf("expected BlockedTasksCount=0, got %d", m.BlockedTasksCount)
	}
	if m.TasksByStatus[model.StatusPending] != 1 {
		t.Errorf("expected 1 pending task, got %d", m.TasksByStatus[model.StatusPending])
	}
	if m.TasksByPriority[model.PriorityHigh] != 1 {
		t.Errorf("expected 1 high priority task, got %d", m.TasksByPriority[model.PriorityHigh])
	}
	if m.TasksByEffort[model.EffortMedium] != 1 {
		t.Errorf("expected 1 medium effort task, got %d", m.TasksByEffort[model.EffortMedium])
	}
	if m.CriticalPathLength != 1 {
		t.Errorf("expected CriticalPathLength=1, got %d", m.CriticalPathLength)
	}
	if m.AvgDependenciesPerTask != 0 {
		t.Errorf("expected AvgDependenciesPerTask=0, got %f", m.AvgDependenciesPerTask)
	}
}

func TestCalculate_LinearDependencyChain(t *testing.T) {
	// Task chain: 3 -> 2 -> 1
	tasks := []*model.Task{
		{
			ID:           "1",
			Title:        "Task 1",
			Status:       model.StatusCompleted,
			Priority:     model.PriorityHigh,
			Effort:       model.EffortSmall,
			Dependencies: []string{},
		},
		{
			ID:           "2",
			Title:        "Task 2",
			Status:       model.StatusInProgress,
			Priority:     model.PriorityMedium,
			Effort:       model.EffortMedium,
			Dependencies: []string{"1"},
		},
		{
			ID:           "3",
			Title:        "Task 3",
			Status:       model.StatusPending,
			Priority:     model.PriorityLow,
			Effort:       model.EffortLarge,
			Dependencies: []string{"2"},
		},
	}

	m := Calculate(tasks)

	if m.TotalTasks != 3 {
		t.Errorf("expected TotalTasks=3, got %d", m.TotalTasks)
	}
	if m.BlockedTasksCount != 2 {
		t.Errorf("expected BlockedTasksCount=2, got %d", m.BlockedTasksCount)
	}
	if m.CriticalPathLength != 3 {
		t.Errorf("expected CriticalPathLength=3, got %d", m.CriticalPathLength)
	}
	if m.MaxDependencyDepth != 3 {
		t.Errorf("expected MaxDependencyDepth=3, got %d", m.MaxDependencyDepth)
	}

	expectedAvg := 2.0 / 3.0 // (0 + 1 + 1) / 3
	if m.AvgDependenciesPerTask != expectedAvg {
		t.Errorf("expected AvgDependenciesPerTask=%.2f, got %.2f", expectedAvg, m.AvgDependenciesPerTask)
	}

	// Check status breakdown
	if m.TasksByStatus[model.StatusCompleted] != 1 {
		t.Errorf("expected 1 completed task, got %d", m.TasksByStatus[model.StatusCompleted])
	}
	if m.TasksByStatus[model.StatusInProgress] != 1 {
		t.Errorf("expected 1 in-progress task, got %d", m.TasksByStatus[model.StatusInProgress])
	}
	if m.TasksByStatus[model.StatusPending] != 1 {
		t.Errorf("expected 1 pending task, got %d", m.TasksByStatus[model.StatusPending])
	}
}

func TestCalculate_MultipleDependencies(t *testing.T) {
	// Task 4 depends on both 2 and 3
	// Task 2 depends on 1
	// Task 3 depends on 1
	// Critical path: 4 -> 2 -> 1 (length 3) or 4 -> 3 -> 1 (length 3)
	tasks := []*model.Task{
		{
			ID:           "1",
			Title:        "Task 1",
			Dependencies: []string{},
		},
		{
			ID:           "2",
			Title:        "Task 2",
			Dependencies: []string{"1"},
		},
		{
			ID:           "3",
			Title:        "Task 3",
			Dependencies: []string{"1"},
		},
		{
			ID:           "4",
			Title:        "Task 4",
			Dependencies: []string{"2", "3"},
		},
	}

	m := Calculate(tasks)

	if m.TotalTasks != 4 {
		t.Errorf("expected TotalTasks=4, got %d", m.TotalTasks)
	}
	if m.BlockedTasksCount != 3 {
		t.Errorf("expected BlockedTasksCount=3, got %d", m.BlockedTasksCount)
	}
	if m.CriticalPathLength != 3 {
		t.Errorf("expected CriticalPathLength=3, got %d", m.CriticalPathLength)
	}

	expectedAvg := 1.0 // (0 + 1 + 1 + 2) / 4
	if m.AvgDependenciesPerTask != expectedAvg {
		t.Errorf("expected AvgDependenciesPerTask=%.2f, got %.2f", expectedAvg, m.AvgDependenciesPerTask)
	}
}

func TestCalculate_DiamondDependency(t *testing.T) {
	// Diamond pattern:
	//     1
	//    / \
	//   2   3
	//    \ /
	//     4
	tasks := []*model.Task{
		{
			ID:           "1",
			Title:        "Task 1",
			Dependencies: []string{},
		},
		{
			ID:           "2",
			Title:        "Task 2",
			Dependencies: []string{"1"},
		},
		{
			ID:           "3",
			Title:        "Task 3",
			Dependencies: []string{"1"},
		},
		{
			ID:           "4",
			Title:        "Task 4",
			Dependencies: []string{"2", "3"},
		},
	}

	m := Calculate(tasks)

	if m.CriticalPathLength != 3 {
		t.Errorf("expected CriticalPathLength=3 for diamond, got %d", m.CriticalPathLength)
	}
}

func TestCalculate_ComplexGraph(t *testing.T) {
	// More complex graph with multiple paths
	tasks := []*model.Task{
		{ID: "1", Dependencies: []string{}},
		{ID: "2", Dependencies: []string{"1"}},
		{ID: "3", Dependencies: []string{"1"}},
		{ID: "4", Dependencies: []string{"2"}},
		{ID: "5", Dependencies: []string{"2", "3"}},
		{ID: "6", Dependencies: []string{"4", "5"}},
	}

	m := Calculate(tasks)

	if m.TotalTasks != 6 {
		t.Errorf("expected TotalTasks=6, got %d", m.TotalTasks)
	}
	if m.BlockedTasksCount != 5 {
		t.Errorf("expected BlockedTasksCount=5, got %d", m.BlockedTasksCount)
	}
	// Longest path: 6 -> 4 -> 2 -> 1 (length 4)
	if m.CriticalPathLength != 4 {
		t.Errorf("expected CriticalPathLength=4, got %d", m.CriticalPathLength)
	}
}

func TestCalculate_MissingDependency(t *testing.T) {
	// Task 2 references non-existent task "99"
	tasks := []*model.Task{
		{
			ID:           "1",
			Dependencies: []string{},
		},
		{
			ID:           "2",
			Dependencies: []string{"99"},
		},
	}

	m := Calculate(tasks)

	// Should not panic, just ignore missing dependency
	if m.TotalTasks != 2 {
		t.Errorf("expected TotalTasks=2, got %d", m.TotalTasks)
	}
	if m.BlockedTasksCount != 1 {
		t.Errorf("expected BlockedTasksCount=1, got %d", m.BlockedTasksCount)
	}
}
