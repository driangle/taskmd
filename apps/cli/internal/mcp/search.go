package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/search"
)

// SearchInput defines the input schema for the search tool.
type SearchInput struct {
	TaskDir string `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Query   string `json:"query" jsonschema:"required,search query for full-text search across task titles and bodies"`
}

func registerSearchTool(server *gomcp.Server, wt worktree.Builder) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "search",
		Description: "Full-text search across task titles and bodies, returning matches with snippets",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input SearchInput) (*gomcp.CallToolResult, any, error) {
		return handleSearch(ctx, req, input, wt)
	})
}

func handleSearch(_ context.Context, _ *gomcp.CallToolRequest, input SearchInput, wt worktree.Builder) (*gomcp.CallToolResult, any, error) {
	if input.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}

	tasks, overlay, err := scanWithOverlay(input.TaskDir, wt)
	if err != nil {
		return nil, nil, err
	}

	results := search.Search(effectiveOrLocal(tasks, overlay), input.Query)

	data, err := json.Marshal(results)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}
