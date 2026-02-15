package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/model"
	"github.com/driangle/taskmd/apps/cli/internal/scanner"
)

// StatusInput defines the input schema for the status tool.
type StatusInput struct {
	TaskDir string `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	TaskID  string `json:"task_id" jsonschema:"required,task ID to retrieve"`
}

// statusOutput is the lightweight metadata struct (no body, no resolved deps).
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
}

func registerStatusTool(server *gomcp.Server) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "status",
		Description: "Get lightweight metadata for a task (no body content, no resolved dependencies)",
	}, handleStatus)
}

func handleStatus(_ context.Context, _ *gomcp.CallToolRequest, input StatusInput) (*gomcp.CallToolResult, any, error) {
	if input.TaskID == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}

	taskDir := input.TaskDir
	if taskDir == "" {
		taskDir = "."
	}

	taskScanner := scanner.NewScanner(taskDir, false, nil)
	result, err := taskScanner.Scan()
	if err != nil {
		return nil, nil, fmt.Errorf("scan failed: %w", err)
	}

	task := findTaskByID(input.TaskID, result.Tasks)
	if task == nil {
		return nil, nil, fmt.Errorf("task not found: %s", input.TaskID)
	}

	out := buildStatusOutput(task)

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
