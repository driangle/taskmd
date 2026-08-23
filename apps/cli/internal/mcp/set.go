package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/model"
	"github.com/driangle/taskmd/sdk/go/scanner"
	"github.com/driangle/taskmd/sdk/go/taskfile"
)

// SetInput defines the input schema for the set tool.
type SetInput struct {
	TaskDir  string   `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	TaskID   string   `json:"task_id" jsonschema:"required,task ID to update"`
	Status   string   `json:"status,omitempty" jsonschema:"new status: pending, in-progress, completed, in-review, blocked, cancelled"`
	Priority string   `json:"priority,omitempty" jsonschema:"new priority: low, medium, high, critical"`
	Effort   string   `json:"effort,omitempty" jsonschema:"new effort; values come from the project's .taskmd.yaml, defaulting to small, medium, large"`
	Owner    string   `json:"owner,omitempty" jsonschema:"new owner/assignee"`
	Tags     []string `json:"tags,omitempty" jsonschema:"replace all tags with this list"`
	AddTags  []string `json:"add_tags,omitempty" jsonschema:"tags to add to existing tags"`
	RemTags  []string `json:"rem_tags,omitempty" jsonschema:"tags to remove from existing tags"`
	AddPRs   []string `json:"add_prs,omitempty" jsonschema:"PR URLs to add"`
	RemPRs   []string `json:"rem_prs,omitempty" jsonschema:"PR URLs to remove"`
}

func registerSetTool(server *gomcp.Server, efforts effort.Scale, wt worktree.Builder) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "set",
		Description: "Update fields on a task (status, priority, effort, owner, tags)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input SetInput) (*gomcp.CallToolResult, any, error) {
		return handleSet(ctx, req, input, efforts, wt)
	})
}

func handleSet(_ context.Context, _ *gomcp.CallToolRequest, input SetInput, efforts effort.Scale, wt worktree.Builder) (*gomcp.CallToolResult, any, error) {
	if input.TaskID == "" {
		return nil, nil, fmt.Errorf("task_id is required")
	}

	req := buildUpdateRequest(input)

	if errs := taskfile.ValidateUpdateRequest(req, efforts); len(errs) > 0 {
		return nil, nil, fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
	}

	if isEmptyRequest(req) {
		return nil, nil, fmt.Errorf("no fields to update")
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
		// Writes stay strictly local: a task that exists only in a sibling
		// worktree gets the guard error, never a redirected write.
		if guardErr := wt.SiblingGuard(input.TaskID, taskDir); guardErr != nil {
			return nil, nil, guardErr
		}
		return nil, nil, fmt.Errorf("task not found: %s", input.TaskID)
	}

	applyStatusTransitionDates(&req, task)

	if err := taskfile.UpdateTaskFile(task.FilePath, req); err != nil {
		return nil, nil, fmt.Errorf("update failed: %w", err)
	}

	out := buildSetOutput(input, req, task.FilePath)
	data, err := json.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func applyStatusTransitionDates(req *taskfile.UpdateRequest, task *model.Task) {
	if req.Status == nil {
		return
	}
	today := time.Now().Format("2006-01-02")
	newStatus := model.Status(*req.Status)
	wasTerminal := task.Status.IsResolved()
	switch {
	case newStatus == model.StatusCompleted:
		req.Completed = &today
		req.RemoveFields = append(req.RemoveFields, "cancelled_at")
	case newStatus == model.StatusCancelled:
		req.CancelledAt = &today
		req.RemoveFields = append(req.RemoveFields, "completed_at")
	case wasTerminal:
		req.RemoveFields = append(req.RemoveFields, "completed_at", "cancelled_at")
	}
}

func buildUpdateRequest(input SetInput) taskfile.UpdateRequest {
	var req taskfile.UpdateRequest
	if input.Status != "" {
		req.Status = &input.Status
	}
	if input.Priority != "" {
		req.Priority = &input.Priority
	}
	if input.Effort != "" {
		req.Effort = &input.Effort
	}
	if input.Owner != "" {
		req.Owner = &input.Owner
	}
	if input.Tags != nil {
		req.Tags = &input.Tags
	}
	req.AddTags = input.AddTags
	req.RemTags = input.RemTags
	req.AddPRs = input.AddPRs
	req.RemPRs = input.RemPRs
	return req
}

func isEmptyRequest(req taskfile.UpdateRequest) bool {
	return req.Status == nil &&
		req.Priority == nil &&
		req.Effort == nil &&
		req.Owner == nil &&
		req.Tags == nil &&
		len(req.AddTags) == 0 &&
		len(req.RemTags) == 0 &&
		len(req.AddPRs) == 0 &&
		len(req.RemPRs) == 0
}

type setOutput struct {
	TaskID   string            `json:"task_id"`
	FilePath string            `json:"file_path"`
	Updated  map[string]string `json:"updated"`
}

func buildSetOutput(input SetInput, req taskfile.UpdateRequest, filePath string) setOutput {
	updated := make(map[string]string)
	if input.Status != "" {
		updated["status"] = input.Status
	}
	if input.Priority != "" {
		updated["priority"] = input.Priority
	}
	if input.Effort != "" {
		updated["effort"] = input.Effort
	}
	if input.Owner != "" {
		updated["owner"] = input.Owner
	}
	if input.Tags != nil {
		updated["tags"] = strings.Join(input.Tags, ", ")
	}
	if len(input.AddTags) > 0 {
		updated["add_tags"] = strings.Join(input.AddTags, ", ")
	}
	if len(input.RemTags) > 0 {
		updated["rem_tags"] = strings.Join(input.RemTags, ", ")
	}
	if len(input.AddPRs) > 0 {
		updated["add_prs"] = strings.Join(input.AddPRs, ", ")
	}
	if len(input.RemPRs) > 0 {
		updated["rem_prs"] = strings.Join(input.RemPRs, ", ")
	}
	if req.Completed != nil {
		updated["completed_at"] = *req.Completed
	}
	if req.CancelledAt != nil {
		updated["cancelled_at"] = *req.CancelledAt
	}
	return setOutput{
		TaskID:   input.TaskID,
		FilePath: filePath,
		Updated:  updated,
	}
}
