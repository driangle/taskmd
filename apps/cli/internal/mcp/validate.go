package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/effort"
	"github.com/driangle/taskmd/sdk/go/validator"
)

// ValidateInput defines the input schema for the validate tool.
type ValidateInput struct {
	TaskDir string `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Strict  bool   `json:"strict,omitempty" jsonschema:"enable strict mode for additional warnings"`
}

type validateOutput struct {
	Valid     bool            `json:"valid"`
	Errors    int             `json:"errors"`
	Warnings  int             `json:"warnings"`
	TaskCount int             `json:"task_count"`
	Issues    []validateIssue `json:"issues"`
}

type validateIssue struct {
	Level    string `json:"level"`
	TaskID   string `json:"task_id,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	Message  string `json:"message"`
}

func registerValidateTool(server *gomcp.Server, efforts effort.Scale, wt worktree.Builder) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "validate",
		Description: "Validate task files for correctness, checking required fields, enum values, dependencies, and cycles",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input ValidateInput) (*gomcp.CallToolResult, any, error) {
		return handleValidate(ctx, req, input, efforts, wt)
	})
}

func handleValidate(_ context.Context, _ *gomcp.CallToolRequest, input ValidateInput, efforts effort.Scale, wt worktree.Builder) (*gomcp.CallToolResult, any, error) {
	// Validate the local tasks after sibling-root attribution, so a checkout
	// nested in the scan root is not flagged as duplicate IDs (spec §8).
	tasks, overlay, err := scanWithOverlay(input.TaskDir, wt)
	if err != nil {
		return nil, nil, err
	}

	v := validator.NewValidator(input.Strict)
	v.SetEffortScale(efforts)
	vr := v.Validate(tasks)
	if overlay != nil {
		for _, warning := range overlay.Warnings {
			vr.AddIssue(validator.LevelWarning, warning.TaskID, "", warning.Message)
		}
	}

	out := buildValidateOutput(vr)

	data, err := json.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func buildValidateOutput(vr *validator.ValidationResult) validateOutput {
	issues := make([]validateIssue, 0, len(vr.Issues))
	for _, issue := range vr.Issues {
		issues = append(issues, validateIssue{
			Level:    string(issue.Level),
			TaskID:   issue.TaskID,
			FilePath: issue.FilePath,
			Message:  issue.Message,
		})
	}

	return validateOutput{
		Valid:     vr.IsValid(),
		Errors:    vr.Errors,
		Warnings:  vr.Warnings,
		TaskCount: vr.TaskCount,
		Issues:    issues,
	}
}
