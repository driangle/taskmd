package mcp

import (
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates an MCP server with all taskmd tools registered.
func NewServer(version string) *gomcp.Server {
	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "taskmd",
		Version: version,
	}, nil)

	registerListTool(server)
	registerGetTool(server)
	registerNextTool(server)
	registerSearchTool(server)
	registerContextTool(server)

	return server
}
