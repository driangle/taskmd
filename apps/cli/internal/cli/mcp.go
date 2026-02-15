package cli

import (
	"context"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	taskmcp "github.com/driangle/taskmd/apps/cli/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server over stdio",
	Long: `Start a Model Context Protocol (MCP) server that communicates over stdin/stdout.

This allows LLM-based tools (Cursor, Windsurf, Copilot agents, etc.) to interact
with your taskmd project using the standard MCP protocol.

Example configuration for Claude Code (.mcp.json):
  {
    "mcpServers": {
      "taskmd": {
        "command": "taskmd",
        "args": ["mcp"]
      }
    }
  }`,
	Args: cobra.NoArgs,
	RunE: runMcp,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMcp(_ *cobra.Command, _ []string) error {
	server := taskmcp.NewServer(Version)
	return server.Run(context.Background(), &gomcp.StdioTransport{})
}
