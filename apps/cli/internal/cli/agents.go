package cli

import (
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Manage agent configurations for task management tools",
	Long: `Commands for managing agent-specific configurations.

Agents are AI assistants like Claude Code, Gemini, or Codex that can help
you work with taskmd tasks. Each agent may have its own configuration format.

Available subcommands:
  init     - Initialize agent configuration files

Examples:
  taskmd agents init
  taskmd agents init --claude --gemini
  taskmd agents --help`,
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}
