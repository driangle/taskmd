package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/model"
)

// StatusInput defines the input schema for the status tool.
type StatusInput struct {
	TaskDir string `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	TaskID  string `json:"task_id" jsonschema:"required,task ID to retrieve"`
}

// statusOutput is the lightweight metadata struct (no body, no resolved
// deps). The embedded provenance fields are populated only when the worktree
// overlay is active.
type statusOutput struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Status       string   `json:"status"`
	Priority     string   `json:"priority,omitempty"`
	Effort       string   `json:"effort,omitempty"`
	Tags         []string `json:"tags"`
	Owner        string   `json:"owner,omitempty"`
	Parent       string   `json:"parent,omitempty"`
	Created      string   `json:"created,omitempty"`
	Dependencies []string `json:"dependencies"`
	Group        string   `json:"group,omitempty"`
	FilePath     string   `json:"file_path"`
	provenanceFields
}

func registerStatusTool(server *gomcp.Server, wt worktree.Builder) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "status",
		Description: "Get lightweight metadata for a task (no body content, no resolved dependencies)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input StatusInput) (*gomcp.CallToolResult, any, error) {
		return handleStatus(ctx, req, input, wt)
	})
}

func handleStatus(_ context.Context, _ *gomcp.CallToolRequest, input StatusInput, wt worktree.Builder) (*gomcp.CallToolResult, any, error) {
	if input.TaskID == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}

	tasks, overlay, err := scanWithOverlay(input.TaskDir, wt)
	if err != nil {
		return nil, nil, err
	}

	task := findTaskByID(input.TaskID, tasks)
	if task == nil && overlay != nil {
		if ot := overlay.Get(input.TaskID); ot != nil {
			task = ot.Task
		}
	}
	if task == nil {
		return nil, nil, fmt.Errorf("task not found: %s", input.TaskID)
	}

	out := buildStatusOutput(task)
	out.provenanceFields = provenanceFor(overlay, input.TaskID)

	data, err := json.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func buildStatusOutput(task *model.Task) statusOutput {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	return statusOutput{
		ID:           task.ID,
		Title:        task.Title,
		Status:       string(task.Status),
		Priority:     string(task.Priority),
		Effort:       string(task.Effort),
		Tags:         task.Tags,
		Owner:        task.Owner,
		Parent:       task.Parent,
		Created:      created,
		Dependencies: task.Dependencies,
		Group:        task.Group,
		FilePath:     task.FilePath,
	}
}
