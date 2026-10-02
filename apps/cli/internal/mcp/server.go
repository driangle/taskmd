package mcp

import (
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/apps/cli/internal/worktree"
	"github.com/driangle/taskmd/sdk/go/effort"
)

// Config is the server-wide configuration, resolved once at startup.
type Config struct {
	Version string

	// Efforts and PhaseOrder come from the .taskmd.yaml of the project the
	// server was started in. Tools accept a per-call task_dir, so a call
	// targeting a different project still uses this configuration — the same
	// limitation that applies to every other config-driven behaviour here.
	Efforts effort.Scale
	// PhaseOrder lists phase ids in configured order; nil means no phases.
	PhaseOrder []string

	// Worktrees builds the cross-worktree overlay per scanned dir; read tools
	// serve the merged view (effective status plus additive provenance fields)
	// when it is active, and mutation tools return the sibling-only guard
	// error. The zero value disables the overlay entirely.
	Worktrees worktree.Builder
}

// NewServer creates an MCP server with all taskmd tools registered.
func NewServer(cfg Config) *gomcp.Server {
	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "taskmd",
		Version: cfg.Version,
	}, nil)

	efforts, wt := cfg.Efforts, cfg.Worktrees
	registerListTool(server, efforts, wt)
	registerGetTool(server, wt)
	registerNextTool(server, cfg)
	registerSearchTool(server, wt)
	registerContextTool(server)
	registerSetTool(server, efforts, wt)
	registerValidateTool(server, efforts, wt)
	registerGraphTool(server, efforts, wt)
	registerStatusTool(server, wt)

	return server
}
