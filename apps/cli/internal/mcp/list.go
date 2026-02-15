package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/filter"
	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// ListInput defines the input schema for the list tool.
type ListInput struct {
	TaskDir string   `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Filters []string `json:"filters,omitempty" jsonschema:"filter expressions, e.g. status=pending, priority=high"`
	Sort    string   `json:"sort,omitempty" jsonschema:"sort field: id, title, status, priority, effort, created"`
}

func registerListTool(server *gomcp.Server) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list",
		Description: "List and filter tasks in a taskmd project",
	}, handleList)
}

func handleList(_ context.Context, _ *gomcp.CallToolRequest, input ListInput) (*gomcp.CallToolResult, any, error) {
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

	if input.Sort != "" {
		if err := sortTasks(tasks, input.Sort); err != nil {
			return nil, nil, err
		}
	}

	data, err := json.Marshal(tasks)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func sortTasks(tasks []*model.Task, field string) error {
	switch field {
	case "id":
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	case "title":
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].Title < tasks[j].Title })
	case "status":
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].Status < tasks[j].Status })
	case "priority":
		order := map[model.Priority]int{
			model.PriorityCritical: 0,
			model.PriorityHigh:     1,
			model.PriorityMedium:   2,
			model.PriorityLow:      3,
		}
		sort.Slice(tasks, func(i, j int) bool { return order[tasks[i].Priority] < order[tasks[j].Priority] })
	case "effort":
		order := map[model.Effort]int{
			model.EffortSmall:  0,
			model.EffortMedium: 1,
			model.EffortLarge:  2,
		}
		sort.Slice(tasks, func(i, j int) bool { return order[tasks[i].Effort] < order[tasks[j].Effort] })
	case "created":
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].Created.Before(tasks[j].Created) })
	default:
		return fmt.Errorf("unsupported sort field: %s (supported: id, title, status, priority, effort, created)", field)
	}
	return nil
}
