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

There are three plugins available:

| Plugin | What it provides | Requires CLI? |
|--------|-----------------|---------------|
| **taskmd** | Slash command skills (`/taskmd:do-task`, `/taskmd:next-task`, etc.) that orchestrate task workflows by invoking the `taskmd` CLI | Yes |
| **taskmd-mcp** | An MCP server that exposes task operations as tools (`list`, `get`, `status`, `next`, `search`, `context`, `set`, `validate`, `graph`), letting Claude call taskmd directly through the Model Context Protocol | Yes |
| **taskmd-lite** | The same slash command skills re-implemented with Claude's file tools (Read, Write, Edit, Glob, Grep), for environments where the `taskmd` binary cannot be installed | No |

**Which one?**

- You have the CLI and want slash-command workflows: install **taskmd**.
- You want Claude to call task operations as tools, for autonomous work or from other MCP clients: install **taskmd-mcp**. It complements **taskmd** rather than replacing it; the MCP server has no equivalent for workflow skills such as `do-task`, `split-task`, `divide-and-conquer`, `import-todos`, and `verify-task`, so installing both is reasonable.
- You cannot install the binary (sandboxed or restricted environment): install **taskmd-lite** on its own. It has no MCP server, and because Claude re-derives results from the files, ranking and validation are approximations of the CLI's. Do not install it alongside **taskmd**, which provides the same skills.

First, add the taskmd marketplace:

```bash
claude plugin marketplace add driangle/taskmd
```

Then install this plugin, optionally with the MCP server:

```bash
# Slash command skills (this plugin)
claude plugin install taskmd@taskmd-marketplace --scope project

# Optional: MCP server (direct tool access for Claude)
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
| update-task | `/taskmd:update-task <description>` | Update a task's fields (status, priority, title, tags, etc.) |
| list-tasks | `/taskmd:list-tasks` | List tasks with optional filters |
| validate-tasks | `/taskmd:validate-tasks` | Validate task files for errors |
| split-task | `/taskmd:split-task <ID>` | Split a large task into smaller sub-tasks |
| divide-and-conquer | `/taskmd:divide-and-conquer <ID>` | Execute a task using parallel subagents for independent workstreams |
| import-todos | `/taskmd:import-todos` | Discover TODO/FIXME comments and convert them into task files |

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

# Update task fields
/taskmd:update-task set task 042 to high priority and in-progress

# Mark a task as done
/taskmd:complete-task 015

# Check task files for issues
/taskmd:validate-tasks

# Split a large task into smaller ones
/taskmd:split-task 045

# Force-split even if it seems small enough
/taskmd:split-task 045 --force

# Look up a specific task
/taskmd:get-task 042

# Execute a task with parallel subagents
/taskmd:divide-and-conquer 045

# Import TODOs from code as tasks
/taskmd:import-todos

# Import only FIXME comments from a specific directory
/taskmd:import-todos --marker FIXME --dir ./src
```

## MCP Server Integration (Optional)

For direct tool access without shelling out to the CLI, install the optional MCP plugin:

```bash
claude plugin install taskmd-mcp@taskmd-marketplace --scope project
```

The MCP server exposes task operations as tools (`list`, `get`, `status`, `next`, `search`, `context`, `set`, `validate`, `graph`), letting Claude Code call taskmd directly through the Model Context Protocol.

For other MCP-compatible clients, see the [MCP Server Guide](https://driangle.github.io/taskmd/guide/mcp) for configuration snippets (Claude Desktop, Cursor, Windsurf, etc.).

## Versioning

This plugin is on its **own `0.x` semver line**. It does not track the taskmd CLI release
number — the three marketplace plugins version independently, so a bump here means *this
plugin* changed, not that a CLI release happened. See
[ADR 0003](https://github.com/driangle/taskmd/blob/main/docs/adr/0003-plugin-versioning-policy.md).

Pre-1.0, skill names and their arguments are not yet a stability promise:

- **Patch** — new skills, new flags on existing skills, prompt and wording fixes.
- **Minor** — a skill renamed or removed, or the arguments it accepts changed.

The authoritative version is the `version` field in
[`.claude-plugin/plugin.json`](./.claude-plugin/plugin.json). The marketplace manifest
deliberately does not repeat it.

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
