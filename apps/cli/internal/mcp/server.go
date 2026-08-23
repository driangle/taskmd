package mcp

import (
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/effort"
)

// NewServer creates an MCP server with all taskmd tools registered.
//
// efforts is the effort vocabulary the server was started with, resolved from
// the project's .taskmd.yaml. Tools accept a per-call task_dir, so a call
// targeting a different project still validates against this vocabulary — the
// same limitation that applies to every other config-driven behaviour here.
//
// wt builds the cross-worktree overlay per scanned dir; read tools serve the
// merged view (effective status plus additive provenance fields) when it is
// active, and mutation tools return the sibling-only guard error. The zero
// value disables the overlay entirely.
func NewServer(version string, efforts effort.Scale, wt worktree.Builder) *gomcp.Server {
	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "taskmd",
		Version: version,
	}, nil)

	registerListTool(server, efforts, wt)
	registerGetTool(server, wt)
	registerNextTool(server, efforts, wt)
	registerSearchTool(server, wt)
	registerContextTool(server)
	registerSetTool(server, efforts, wt)
	registerValidateTool(server, efforts, wt)
	registerGraphTool(server, efforts, wt)
	registerStatusTool(server, wt)

	return server
}
