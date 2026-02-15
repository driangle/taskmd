package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/taskcontext"
)

// ContextInput defines the input schema for the context tool.
type ContextInput struct {
	TaskDir        string              `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	TaskID         string              `json:"task_id" jsonschema:"required,task ID to resolve context for"`
	Scopes         map[string][]string `json:"scopes,omitempty" jsonschema:"scope definitions mapping scope names to file paths"`
	ProjectRoot    string              `json:"project_root,omitempty" jsonschema:"project root directory for resolving file paths"`
	Resolve        bool                `json:"resolve,omitempty" jsonschema:"expand directory paths to individual files"`
	IncludeContent bool                `json:"include_content,omitempty" jsonschema:"inline file contents and task body"`
	MaxFiles       int                 `json:"max_files,omitempty" jsonschema:"cap number of files returned, 0 for unlimited"`
}

func registerContextTool(server *gomcp.Server) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "context",
		Description: "Resolve relevant file paths for a task based on its touches (scopes) and explicit context fields",
	}, handleContext)
}

func handleContext(_ context.Context, _ *gomcp.CallToolRequest, input ContextInput) (*gomcp.CallToolResult, any, error) {
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

	projectRoot := input.ProjectRoot
	if projectRoot == "" {
		projectRoot = "."
	}

	opts := taskcontext.Options{
		Scopes:         taskcontext.ScopeMap(input.Scopes),
		ProjectRoot:    projectRoot,
		Resolve:        input.Resolve,
		IncludeContent: input.IncludeContent,
		MaxFiles:       input.MaxFiles,
	}

	ctxResult, err := taskcontext.Resolve(task, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("context resolution failed: %w", err)
	}

	data, err := json.Marshal(ctxResult)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}
