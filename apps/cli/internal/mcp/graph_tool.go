package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/filter"
	"github.com/driangle/taskmd/apps/cli/internal/graph"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// GraphInput defines the input schema for the graph tool.
type GraphInput struct {
	TaskDir       string   `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	RootTaskID    string   `json:"root_task_id,omitempty" jsonschema:"focus on a specific task and its dependencies/dependents"`
	ExcludeStatus []string `json:"exclude_status,omitempty" jsonschema:"exclude tasks with these statuses"`
	Filters       []string `json:"filters,omitempty" jsonschema:"filter expressions, e.g. status=pending, priority=high"`
}

func registerGraphTool(server *gomcp.Server) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "graph",
		Description: "Get the task dependency graph as JSON with nodes, edges, and cycle detection",
	}, handleGraph)
}

func handleGraph(_ context.Context, _ *gomcp.CallToolRequest, input GraphInput) (*gomcp.CallToolResult, any, error) {
	taskDir := input.TaskDir
	if taskDir == "" {
		taskDir = "."
	}

	taskScanner := scanner.NewScanner(taskDir, false, nil)
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}

	tasks := result.Tasks

	if len(input.Filters) > 0 {
		tasks, err = filter.Apply(tasks, input.Filters)
		if err != nil {
			return nil, nil, fmt.Errorf("filter error: %w", err)
		}
	}

	if len(input.ExcludeStatus) > 0 {
		tasks = excludeByStatus(tasks, input.ExcludeStatus)
	}

	g := graph.NewGraph(tasks)

	if input.RootTaskID != "" {
		if _, ok := g.TaskMap[input.RootTaskID]; !ok {
			return nil, nil, fmt.Errorf("root task not found: %s", input.RootTaskID)
		}
		upstream := g.GetUpstream(input.RootTaskID)
		downstream := g.GetDownstream(input.RootTaskID)
		combined := make(map[string]bool, len(upstream)+len(downstream)+1)
		for id := range upstream {
			combined[id] = true
		}
		for id := range downstream {
			combined[id] = true
		}
		combined[input.RootTaskID] = true
		g = g.FilterTasks(combined)
	}

	graphJSON := g.ToJSON()

	data, err := json.Marshal(graphJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func excludeByStatus(tasks []*model.Task, statuses []string) []*model.Task {
	excludeMap := make(map[string]bool, len(statuses))
	for _, s := range statuses {
		excludeMap[s] = true
	}

	filtered := make([]*model.Task, 0, len(tasks))
	for _, task := range tasks {
		if !excludeMap[string(task.Status)] {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
