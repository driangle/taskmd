package cli

import "github.com/driangle/md-task-tracker/apps/cli/internal/model"

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
