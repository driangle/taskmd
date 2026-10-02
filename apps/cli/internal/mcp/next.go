package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/sdk/go/next"
	"github.com/driangle/taskmd/sdk/go/scanner"
)

// NextInput defines the input schema for the next tool.
type NextInput struct {
	TaskDir   string   `json:"task_dir,omitempty" jsonschema:"task directory to scan, defaults to current directory"`
	Limit     int      `json:"limit,omitempty" jsonschema:"max number of recommendations to return, defaults to 5"`
	Filters   []string `json:"filters,omitempty" jsonschema:"filter expressions, e.g. priority=high, tag=mvp"`
	QuickWins bool     `json:"quick_wins,omitempty" jsonschema:"only show tasks at the lowest configured effort"`
	Critical  bool     `json:"critical,omitempty" jsonschema:"only show tasks on the critical path"`
}

func registerNextTool(server *gomcp.Server, cfg Config) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "next",
		Description: "Get ranked task recommendations based on priority, dependencies, and critical path analysis",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, input NextInput) (*gomcp.CallToolResult, any, error) {
		return handleNext(ctx, req, input, cfg)
	})
}

func handleNext(_ context.Context, _ *gomcp.CallToolRequest, input NextInput, cfg Config) (*gomcp.CallToolResult, any, error) {
	tasks, overlay, err := scanWithOverlay(input.TaskDir, cfg.Worktrees)
	if err != nil {
		return nil, nil, err
	}

	archivedTasks, err := scanner.NewScanner(resolveTaskDir(input.TaskDir), false, nil).ScanArchive()
	if err != nil {
		return nil, nil, fmt.Errorf("archive scan failed: %w", err)
	}

	var excluded map[string]string
	if overlay != nil {
		tasks, excluded = overlay.RecommendationInputs()
	}

	opts := next.Options{
		Limit:         input.Limit,
		Filters:       input.Filters,
		QuickWins:     input.QuickWins,
		Critical:      input.Critical,
		ArchivedTasks: archivedTasks,
		PhaseOrder:    cfg.PhaseOrder,
		StrictPhases:  true,
		Efforts:       cfg.Efforts,
		Excluded:      excluded,
	}

	recs, err := next.Recommend(tasks, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("recommendation failed: %w", err)
	}

	data, err := json.Marshal(recs)
	if err != nil {
		return nil, nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}
