package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/scanner"
	"github.com/driangle/taskmd/apps/cli/internal/search"
)

// SearchInput defines the input schema for the search tool.
type SearchInput struct {
	TaskDir string `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Query   string `json:"query" jsonschema:"required,search query for full-text search across task titles and bodies"`
}

func registerSearchTool(server *gomcp.Server) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "search",
		Description: "Full-text search across task titles and bodies, returning matches with snippets",
	}, handleSearch)
}

func handleSearch(_ context.Context, _ *gomcp.CallToolRequest, input SearchInput) (*gomcp.CallToolResult, any, error) {
	if input.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
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

	results := search.Search(result.Tasks, input.Query)

	data, err := json.Marshal(results)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}
