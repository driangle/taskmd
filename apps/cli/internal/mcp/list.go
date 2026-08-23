package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/filter"
	"github.com/driangle/taskmd/sdk/go/model"
)

// ListInput defines the input schema for the list tool.
type ListInput struct {
	TaskDir string   `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Filters []string `json:"filters,omitempty" jsonschema:"filter expressions, e.g. status=pending, priority=high"`
	Sort    string   `json:"sort,omitempty" jsonschema:"sort field: id, title, status, priority, effort, created"`
}

func registerListTool(server *gomcp.Server, efforts effort.Scale, wt worktree.Builder) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list",
		Description: "List and filter tasks in a taskmd project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input ListInput) (*gomcp.CallToolResult, any, error) {
		return handleList(ctx, req, input, efforts, wt)
	})
}

func handleList(_ context.Context, _ *gomcp.CallToolRequest, input ListInput, efforts effort.Scale, wt worktree.Builder) (*gomcp.CallToolResult, any, error) {
	tasks, overlay, err := scanWithOverlay(input.TaskDir, wt)
	if err != nil {
		return nil, nil, err
	}

	var out any
	if overlay != nil {
		out, err = overlayListRows(overlay, input, efforts)
	} else {
		out, err = listRows(tasks, input, efforts)
	}
	if err != nil {
		return nil, nil, err
	}

	data, err := json.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

// listRows applies the list tool's filters and sort to a plain task list.
func listRows(tasks []*model.Task, input ListInput, efforts effort.Scale) ([]*model.Task, error) {
	if len(input.Filters) > 0 {
		filtered, err := filter.Apply(tasks, input.Filters, efforts)
		if err != nil {
			return nil, fmt.Errorf("filter error: %w", err)
		}
		tasks = filtered
	}
	if input.Sort != "" {
		if err := sortTasks(tasks, input.Sort, efforts); err != nil {
			return nil, err
		}
	}
	return tasks, nil
}

// overlayListRows filters and sorts on shallow copies with the effective
// status substituted (so status filters match the merged view), then maps the
// surviving copies back to their overlay tasks so the output carries the
// additive provenance fields.
func overlayListRows(overlay *worktree.Overlay, input ListInput, efforts effort.Scale) ([]*worktree.Task, error) {
	copies := make([]*model.Task, len(overlay.Tasks))
	index := make(map[*model.Task]*worktree.Task, len(overlay.Tasks))
	for i, ot := range overlay.Tasks {
		effective := *ot.Task
		effective.Status = ot.EffectiveStatus
		copies[i] = &effective
		index[&effective] = ot
	}

	filtered, err := listRows(copies, input, efforts)
	if err != nil {
		return nil, err
	}

	rows := make([]*worktree.Task, 0, len(filtered))
	for _, t := range filtered {
		rows = append(rows, index[t])
	}
	return rows, nil
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
