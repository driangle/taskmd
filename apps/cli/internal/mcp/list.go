package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/filter"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// ListInput defines the input schema for the list tool.
type ListInput struct {
	TaskDir string   `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Filters []string `json:"filters,omitempty" jsonschema:"filter expressions, e.g. status=pending, priority=high"`
	Sort    string   `json:"sort,omitempty" jsonschema:"sort field: id, title, status, priority, effort, created"`
}

func registerListTool(server *gomcp.Server, efforts effort.Scale) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list",
		Description: "List and filter tasks in a taskmd project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input ListInput) (*gomcp.CallToolResult, any, error) {
		return handleList(ctx, req, input, efforts)
	})
}

func handleList(_ context.Context, _ *gomcp.CallToolRequest, input ListInput, efforts effort.Scale) (*gomcp.CallToolResult, any, error) {
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
		tasks, err = filter.Apply(tasks, input.Filters, efforts)
		if err != nil {
			return nil, nil, fmt.Errorf("filter error: %w", err)
		}
	}

	if input.Sort != "" {
		if err := sortTasks(tasks, input.Sort, efforts); err != nil {
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

// effortRank orders an effort value by its position in the project's vocabulary.
// Unset and unrecognized values sort last, after every known value.
func effortRank(efforts effort.Scale, value model.Effort) int {
	if rank := efforts.Rank(string(value)); rank >= 0 {
		return rank
	}
	return efforts.Len()
}

func sortTasks(tasks []*model.Task, field string, efforts effort.Scale) error {
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
		sort.Slice(tasks, func(i, j int) bool {
			return effortRank(efforts, tasks[i].Effort) < effortRank(efforts, tasks[j].Effort)
		})
	case "created":
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].Created.Before(tasks[j].Created.Time) })
	default:
		return fmt.Errorf("unsupported sort field: %s (supported: id, title, status, priority, effort, created)", field)
	}
	return nil
}
