package mcp

import (
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/driangle/taskmd/sdk/go/effort"
)

// NewServer creates an MCP server with all taskmd tools registered.
//
// efforts is the effort vocabulary the server was started with, resolved from
// the project's .taskmd.yaml. Tools accept a per-call task_dir, so a call
// targeting a different project still validates against this vocabulary — the
// same limitation that applies to every other config-driven behaviour here.
func NewServer(version string, efforts effort.Scale) *gomcp.Server {
	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "taskmd",
		Version: version,
	}, nil)

	registerListTool(server, efforts)
	registerGetTool(server)
	registerNextTool(server, efforts)
	registerSearchTool(server)
	registerContextTool(server)
	registerSetTool(server, efforts)
	registerValidateTool(server, efforts)
	registerGraphTool(server, efforts)
	registerStatusTool(server)

	return server
}
