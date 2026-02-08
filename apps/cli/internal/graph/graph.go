package graph

import (
	"fmt"
	"sort"
	"strings"

	"github.com/driangle/md-task-tracker/apps/cli/internal/model"
)

// Graph represents a task dependency graph
type Graph struct {
	Tasks        []*model.Task
	TaskMap      map[string]*model.Task
	Adjacency    map[string][]string // task ID -> list of dependent task IDs
	RevAdjacency map[string][]string // task ID -> list of dependency task IDs
}

// NewGraph creates a new graph from a list of tasks
func NewGraph(tasks []*model.Task) *Graph {
	g := &Graph{
		Tasks:        tasks,
		TaskMap:      make(map[string]*model.Task),
		Adjacency:    make(map[string][]string),
		RevAdjacency: make(map[string][]string),
	}

	// Build task map
	for _, task := range tasks {
		g.TaskMap[task.ID] = task
	}

	// Build adjacency lists
	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			// task depends on depID
			// so depID -> task in forward adjacency (depID blocks task)
			g.Adjacency[depID] = append(g.Adjacency[depID], task.ID)
			// task -> depID in reverse adjacency
			g.RevAdjacency[task.ID] = append(g.RevAdjacency[task.ID], depID)
		}
	}

	return g
}

// GetDownstream returns all tasks that depend on the given task (transitively)
func (g *Graph) GetDownstream(taskID string) map[string]bool {
	visited := make(map[string]bool)
	var visit func(id string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		for _, dependentID := range g.Adjacency[id] {
			visit(dependentID)
		}
	}
	visit(taskID)
	delete(visited, taskID) // Don't include the root task itself
	return visited
}

// GetUpstream returns all tasks that the given task depends on (transitively)
func (g *Graph) GetUpstream(taskID string) map[string]bool {
	visited := make(map[string]bool)
	var visit func(id string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		for _, depID := range g.RevAdjacency[id] {
			visit(depID)
		}
	}
	visit(taskID)
	delete(visited, taskID) // Don't include the root task itself
	return visited
}

// DetectCycles finds all cycles in the graph
//
//nolint:gocognit // TODO: refactor to reduce complexity
func (g *Graph) DetectCycles() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(taskID string) bool
	dfs = func(taskID string) bool {
		visited[taskID] = true
		recStack[taskID] = true
		path = append(path, taskID)

		for _, depID := range g.RevAdjacency[taskID] {
			if !visited[depID] {
				if dfs(depID) {
					return true
				}
			} else if recStack[depID] {
				// Found a cycle
				cycleStart := -1
				for i, id := range path {
					if id == depID {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[taskID] = false
		return false
	}

	for _, task := range g.Tasks {
		if !visited[task.ID] {
			dfs(task.ID)
		}
	}

	return cycles
}

// FilterTasks creates a subgraph with only the specified task IDs
func (g *Graph) FilterTasks(taskIDs map[string]bool) *Graph {
	filtered := []*model.Task{}
	for _, task := range g.Tasks {
		if taskIDs[task.ID] {
			filtered = append(filtered, task)
		}
	}
	return NewGraph(filtered)
}

// ToMermaid generates a Mermaid diagram
func (g *Graph) ToMermaid(focusTaskID string) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Sort tasks by ID for consistent output
	sortedTasks := make([]*model.Task, len(g.Tasks))
	copy(sortedTasks, g.Tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].ID < sortedTasks[j].ID
	})

	// Define nodes
	for _, task := range sortedTasks {
		nodeStyle := ""
		if task.ID == focusTaskID {
			nodeStyle = ":::focus"
		} else {
			switch task.Status {
			case model.StatusCompleted:
				nodeStyle = ":::completed"
			case model.StatusInProgress:
				nodeStyle = ":::inprogress"
			case model.StatusBlocked:
				nodeStyle = ":::blocked"
			}
		}

		// Escape special characters in title
		title := strings.ReplaceAll(task.Title, "\"", "&quot;")
		sb.WriteString(fmt.Sprintf("    %s[\"%s: %s\"]%s\n", task.ID, task.ID, title, nodeStyle))
	}

	// Define edges
	edges := make(map[string]bool) // Track unique edges
	for _, task := range sortedTasks {
		for _, depID := range task.Dependencies {
			edgeKey := depID + "->" + task.ID
			if !edges[edgeKey] {
				edges[edgeKey] = true
				sb.WriteString(fmt.Sprintf("    %s --> %s\n", depID, task.ID))
			}
		}
	}

	// Add styles
	sb.WriteString("\n")
	sb.WriteString("    classDef focus fill:#ff6b6b,stroke:#c92a2a,color:#fff\n")
	sb.WriteString("    classDef completed fill:#51cf66,stroke:#2f9e44,color:#000\n")
	sb.WriteString("    classDef inprogress fill:#ffd43b,stroke:#fab005,color:#000\n")
	sb.WriteString("    classDef blocked fill:#868e96,stroke:#495057,color:#fff\n")

	return sb.String()
}

// ToDot generates a Graphviz DOT format
func (g *Graph) ToDot(focusTaskID string) string {
	var sb strings.Builder
	sb.WriteString("digraph tasks {\n")
	sb.WriteString("    rankdir=TB;\n")
	sb.WriteString("    node [shape=box, style=rounded];\n\n")

	// Sort tasks by ID for consistent output
	sortedTasks := make([]*model.Task, len(g.Tasks))
	copy(sortedTasks, g.Tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].ID < sortedTasks[j].ID
	})

	// Define nodes
	for _, task := range sortedTasks {
		color := "lightgray"
		if task.ID == focusTaskID {
			color = "red"
		} else {
			switch task.Status {
			case model.StatusCompleted:
				color = "lightgreen"
			case model.StatusInProgress:
				color = "yellow"
			case model.StatusBlocked:
				color = "gray"
			}
		}

		// Escape special characters
		title := strings.ReplaceAll(task.Title, "\"", "\\\"")
		label := fmt.Sprintf("%s: %s", task.ID, title)
		sb.WriteString(fmt.Sprintf("    %s [label=\"%s\", fillcolor=%s, style=filled];\n", task.ID, label, color))
	}

	sb.WriteString("\n")

	// Define edges
	edges := make(map[string]bool)
	for _, task := range sortedTasks {
		for _, depID := range task.Dependencies {
			edgeKey := depID + "->" + task.ID
			if !edges[edgeKey] {
				edges[edgeKey] = true
				sb.WriteString(fmt.Sprintf("    %s -> %s;\n", depID, task.ID))
			}
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// ToASCII generates an ASCII tree representation
//
//nolint:gocognit,gocyclo,funlen // TODO: refactor to reduce complexity
func (g *Graph) ToASCII(rootTaskID string, downstream bool) string {
	var sb strings.Builder

	visited := make(map[string]bool)

	var printTree func(taskID string, prefix string, isLast bool)
	printTree = func(taskID string, prefix string, isLast bool) {
		if visited[taskID] {
			// Already visited, just show reference
			task, exists := g.TaskMap[taskID]
			if exists {
				connector := "├── "
				if isLast {
					connector = "└── "
				}
				sb.WriteString(fmt.Sprintf("%s%s[%s] %s (see above)\n", prefix, connector, taskID, task.Title))
			}
			return
		}

		visited[taskID] = true

		task, exists := g.TaskMap[taskID]
		if !exists {
			return
		}

		// Print current node
		connector := "├── "
		extension := "│   "
		if isLast {
			connector = "└── "
			extension = "    "
		}

		status := ""
		switch task.Status {
		case model.StatusCompleted:
			status = " ✓"
		case model.StatusInProgress:
			status = " ⋯"
		case model.StatusBlocked:
			status = " ⊗"
		}

		if prefix == "" {
			sb.WriteString(fmt.Sprintf("[%s] %s%s\n", taskID, task.Title, status))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s[%s] %s%s\n", prefix, connector, taskID, task.Title, status))
		}

		// Get children based on direction
		var children []string
		if downstream {
			children = g.Adjacency[taskID]
		} else {
			children = g.RevAdjacency[taskID]
		}

		// Sort children for consistent output
		sort.Strings(children)

		// Print children
		for i, childID := range children {
			isLastChild := i == len(children)-1
			childPrefix := prefix
			if prefix != "" || len(children) > 1 || i > 0 {
				childPrefix = prefix + extension
			}
			printTree(childID, childPrefix, isLastChild)
		}
	}

	if rootTaskID != "" {
		printTree(rootTaskID, "", true)
	} else {
		// Print all root nodes (tasks with no dependencies)
		roots := []string{}
		for _, task := range g.Tasks {
			if downstream {
				// For downstream view, roots are tasks with no dependencies
				if len(task.Dependencies) == 0 {
					roots = append(roots, task.ID)
				}
			} else {
				// For upstream view, roots are tasks that nothing depends on
				if len(g.Adjacency[task.ID]) == 0 {
					roots = append(roots, task.ID)
				}
			}
		}

		sort.Strings(roots)

		if len(roots) == 0 {
			// No roots found, might be cycles - just print all
			for _, task := range g.Tasks {
				roots = append(roots, task.ID)
			}
			sort.Strings(roots)
		}

		for i, rootID := range roots {
			isLast := i == len(roots)-1
			printTree(rootID, "", isLast)
			if !isLast {
				sb.WriteString("\n")
			}
		}
	}

	return sb.String()
}

// ToJSON generates a JSON graph structure
func (g *Graph) ToJSON() map[string]any {
	nodes := []map[string]any{}
	edges := []map[string]string{}

	// Sort tasks by ID
	sortedTasks := make([]*model.Task, len(g.Tasks))
	copy(sortedTasks, g.Tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].ID < sortedTasks[j].ID
	})

	// Build nodes
	for _, task := range sortedTasks {
		node := map[string]any{
			"id":     task.ID,
			"title":  task.Title,
			"status": string(task.Status),
		}
		if task.Priority != "" {
			node["priority"] = string(task.Priority)
		}
		if task.Group != "" {
			node["group"] = task.Group
		}
		nodes = append(nodes, node)
	}

	// Build edges
	edgeSet := make(map[string]bool)
	for _, task := range sortedTasks {
		for _, depID := range task.Dependencies {
			edgeKey := depID + "->" + task.ID
			if !edgeSet[edgeKey] {
				edgeSet[edgeKey] = true
				edges = append(edges, map[string]string{
					"from": depID,
					"to":   task.ID,
				})
			}
		}
	}

	// Detect cycles
	cycles := g.DetectCycles()
	cyclesData := [][]string{}
	if len(cycles) > 0 {
		cyclesData = cycles
	}

	result := map[string]any{
		"nodes": nodes,
		"edges": edges,
	}

	if len(cyclesData) > 0 {
		result["cycles"] = cyclesData
	}

	return result
}
