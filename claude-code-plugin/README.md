# taskmd Claude Code Plugin

A [Claude Code](https://claude.com/claude-code) plugin that provides taskmd skills as slash commands, so you can manage your markdown-based tasks directly within Claude Code sessions.

## Prerequisites

Install the `taskmd` CLI before using this plugin:

```bash
# Homebrew (macOS and Linux)
brew tap driangle/tap
brew install taskmd

# Or install with Go
go install github.com/driangle/taskmd/apps/cli/cmd/taskmd@latest
```

Verify it's available:

```bash
taskmd --version
```

## Installation

There are two plugins available:

| Plugin | What it provides | Requires CLI? |
|--------|-----------------|---------------|
| **taskmd** | Slash command skills (`/taskmd:do-task`, `/taskmd:next-task`, etc.) that orchestrate task workflows by invoking the `taskmd` CLI | Yes |
| **taskmd-mcp** | An MCP server that exposes task operations as tools (`list`, `get`, `next`, `search`, `set`, etc.), letting Claude call taskmd directly through the Model Context Protocol | Yes |

**taskmd** is best for interactive, human-driven workflows via slash commands. **taskmd-mcp** gives Claude direct tool access for autonomous task operations. You can install both.

First, add the taskmd marketplace:

```bash
claude plugin marketplace add driangle/taskmd
```

Then install the plugin(s):

```bash
# Install slash command skills
claude plugin install taskmd@taskmd-marketplace --scope project

# Install MCP server for direct tool access (optional)
claude plugin install taskmd-mcp@taskmd-marketplace --scope project
```

Use `--scope user` instead of `--scope project` to install across all projects.

## Available Skills

| Skill | Slash Command | Description |
|-------|--------------|-------------|
| do-task | `/taskmd:do-task <ID>` | Look up a task and start working on it |
| next-task | `/taskmd:next-task` | Find the next recommended task |
| get-task | `/taskmd:get-task <ID>` | View task details by ID or name |
| add-task | `/taskmd:add-task <description>` | Create a new task file |
| complete-task | `/taskmd:complete-task <ID>` | Mark a task as completed |
| list-tasks | `/taskmd:list-tasks` | List tasks with optional filters |
| validate | `/taskmd:validate` | Validate task files for errors |
| divide-and-conquer | `/taskmd:divide-and-conquer <ID>` | Split a large task into smaller sub-tasks |

## Usage Examples

```
# See what to work on next
/taskmd:next-task

# Start working on task 015
/taskmd:do-task 015

# List all pending tasks
/taskmd:list-tasks --status pending

# Create a new task
/taskmd:add-task Add user authentication to the API

# Mark a task as done
/taskmd:complete-task 015

# Check task files for issues
/taskmd:validate

# Split a large task into smaller ones
/taskmd:divide-and-conquer 045

# Force-split even if it seems small enough
/taskmd:divide-and-conquer 045 --force

# Look up a specific task
/taskmd:get-task 042
```

## MCP Server Integration (Optional)

For direct tool access without shelling out to the CLI, install the optional MCP plugin:

```bash
claude plugin install taskmd-mcp@taskmd-marketplace --scope project
```

The MCP server exposes task operations as tools (`list`, `get`, `next`, `search`, `context`, `set`, `validate`, `graph`), letting Claude Code call taskmd directly through the Model Context Protocol.

For other MCP-compatible clients, see the [MCP Server Guide](https://github.com/driangle/taskmd/blob/main/docs/guides/mcp-guide.md) for configuration snippets (Claude Desktop, Cursor, Windsurf, etc.).

## Troubleshooting

**"taskmd: command not found"**
The `taskmd` CLI is not installed or not in your PATH. Install it with one of the methods listed in Prerequisites above.

**"no task files found"**
Make sure you have a `tasks/` directory with `.md` files in your project. See the [taskmd Quick Start](https://github.com/driangle/taskmd#quick-start) for setup instructions.

**Skills not appearing**
Verify the plugin is installed by running `/plugins list` in Claude Code.

## Learn More

- [taskmd documentation](https://driangle.github.io/taskmd/)
- [Task file specification](https://github.com/driangle/taskmd/blob/main/docs/taskmd_specification.md)
- [GitHub repository](https://github.com/driangle/taskmd)
