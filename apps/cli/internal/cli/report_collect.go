package cli

import (
	"sort"

	"github.com/driangle/taskmd/apps/cli/internal/board"
	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/metrics"
	"github.com/driangle/taskmd/apps/cli/internal/model"
)

// reportTask holds flattened task data used across report sections.
type reportTask struct {
	ID           string
	Title        string
	Status       string
	Priority     string
	Dependencies []string
}

// reportData is the format-agnostic intermediate representation of a report.
type reportData struct {
	Metrics      *metrics.Metrics
	GroupedTasks *board.GroupResult
	GroupBy      string
	CriticalPath []reportTask
	BlockedTasks []reportTask
	IncludeGraph bool
	GraphMermaid string
	GraphJSON    map[string]any
}

func collectReportData(tasks []*model.Task, groupBy string, includeGraph bool) (*reportData, error) {
	m := metrics.Calculate(tasks)

	grouped, err := board.GroupTasks(tasks, groupBy)
	if err != nil {
		return nil, err
	}

	taskMap := buildTaskMap(tasks)

	blocked := findBlockedTasks(tasks, taskMap)
	critical := findCriticalPathTasks(tasks, taskMap)

	data := &reportData{
		Metrics:      m,
		GroupedTasks: grouped,
		GroupBy:      groupBy,
		CriticalPath: critical,
		BlockedTasks: blocked,
		IncludeGraph: includeGraph,
	}

	if includeGraph {
		g := graph.NewGraph(tasks)
		data.GraphMermaid = g.ToMermaid("")
		data.GraphJSON = g.ToJSON()
	}

	return data, nil
}

func findBlockedTasks(tasks []*model.Task, taskMap map[string]*model.Task) []reportTask {
	var blocked []reportTask
	for _, t := range tasks {
		if isTaskBlocked(t, taskMap) {
			blocked = append(blocked, reportTask{
				ID:           t.ID,
				Title:        t.Title,
				Status:       string(t.Status),
				Priority:     string(t.Priority),
				Dependencies: t.Dependencies,
			})
		}
	}
	sort.Slice(blocked, func(i, j int) bool {
		return blocked[i].ID < blocked[j].ID
	})
	return blocked
}

func findCriticalPathTasks(tasks []*model.Task, taskMap map[string]*model.Task) []reportTask {
	cpIDs := calculateCriticalPathTasks(tasks, taskMap)
	depthMap := calculateDepthMap(tasks, taskMap)

	var cpTasks []reportTask
	for _, t := range tasks {
		if cpIDs[t.ID] {
			cpTasks = append(cpTasks, reportTask{
				ID:           t.ID,
				Title:        t.Title,
				Status:       string(t.Status),
				Priority:     string(t.Priority),
				Dependencies: t.Dependencies,
			})
		}
	}

	// Sort by depth ascending so the chain reads from root to leaf.
	sort.Slice(cpTasks, func(i, j int) bool {
		di := depthMap[cpTasks[i].ID]
		dj := depthMap[cpTasks[j].ID]
		if di != dj {
			return di < dj
		}
		return cpTasks[i].ID < cpTasks[j].ID
	})

	return cpTasks
}
