package metrics

import (
	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// Metrics contains computed statistics about a task set
type Metrics struct {
	TotalTasks             int                    `json:"total_tasks"`
	TasksByStatus          map[model.Status]int   `json:"tasks_by_status"`
	TasksByPriority        map[model.Priority]int `json:"tasks_by_priority"`
	TasksByEffort          map[model.Effort]int   `json:"tasks_by_effort"`
	BlockedTasksCount      int                    `json:"blocked_tasks_count"`
	CriticalPathLength     int                    `json:"critical_path_length"`
	MaxDependencyDepth     int                    `json:"max_dependency_depth"`
	AvgDependenciesPerTask float64                `json:"avg_dependencies_per_task"`
}

// Calculate computes metrics for a set of tasks
func Calculate(tasks []*model.Task) *Metrics {
	m := &Metrics{
		TasksByStatus:   make(map[model.Status]int),
		TasksByPriority: make(map[model.Priority]int),
		TasksByEffort:   make(map[model.Effort]int),
	}

	// Build task map for dependency lookups
	taskMap := make(map[string]*model.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Count totals and categorize tasks
	totalDeps := 0
	for _, task := range tasks {
		m.TotalTasks++

		// Count by status
		m.TasksByStatus[task.Status]++

		// Count by priority
		m.TasksByPriority[task.Priority]++

		// Count by effort
		m.TasksByEffort[task.Effort]++

		// Count blocked tasks (tasks with dependencies)
		if len(task.Dependencies) > 0 {
			m.BlockedTasksCount++
		}

		totalDeps += len(task.Dependencies)
	}

	// Calculate average dependencies per task
	if m.TotalTasks > 0 {
		m.AvgDependenciesPerTask = float64(totalDeps) / float64(m.TotalTasks)
	}

	// Calculate critical path and max depth
	m.CriticalPathLength = calculateCriticalPath(tasks, taskMap)
	m.MaxDependencyDepth = calculateMaxDepth(tasks, taskMap)

	return m
}

// calculateCriticalPath finds the longest path through the dependency graph
func calculateCriticalPath(tasks []*model.Task, taskMap map[string]*model.Task) int {
	// Memoization for depth calculation
	memo := make(map[string]int)

	var getDepth func(taskID string, visited map[string]bool) int
	getDepth = func(taskID string, visited map[string]bool) int {
		// Check memo first
		if depth, ok := memo[taskID]; ok {
			return depth
		}

		// Prevent cycles
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

	maxPath := 0
	for _, task := range tasks {
		depth := getDepth(task.ID, make(map[string]bool))
		if depth > maxPath {
			maxPath = depth
		}
	}

	return maxPath
}

// calculateMaxDepth finds the maximum depth of any single task's dependency chain
func calculateMaxDepth(tasks []*model.Task, taskMap map[string]*model.Task) int {
	// This is the same as critical path for now
	// In the future, we could differentiate between:
	// - Critical path: longest path from root to leaf
	// - Max depth: deepest single task
	return calculateCriticalPath(tasks, taskMap)
}
