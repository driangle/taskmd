# Claude Code Plugin

Use taskmd directly inside [Claude Code](https://claude.com/claude-code) with slash commands. The plugin provides skills for listing, creating, completing, and validating tasks without leaving your Claude session.

## Prerequisites

The `taskmd` CLI must be installed and available in your PATH. See [Installation](/getting-started/installation) for setup instructions.

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

Then install the plugin(s):

```bash
# Install slash command skills
claude plugin install taskmd@taskmd-marketplace --scope project

# Or install for all projects (user-wide)
claude plugin install taskmd@taskmd-marketplace --scope user

# Or install for the current directory only
claude plugin install taskmd@taskmd-marketplace --scope local
```

### MCP Server Plugin (Optional)

For direct tool access without shelling out to the CLI, install the MCP plugin:

```bash
claude plugin install taskmd-mcp@taskmd-marketplace --scope project
```

### CLI-free Plugin

When the `taskmd` binary cannot be installed, install **taskmd-lite** instead of **taskmd**. Its skills use the `/taskmd-lite:` prefix:

```bash
claude plugin install taskmd-lite@taskmd-marketplace --scope project
```

## Available Skills

| Slash Command | Description |
|--------------|-------------|
| `/taskmd:do-task <ID>` | Look up a task and start working on it |
| `/taskmd:next-task` | Find the next recommended task |
| `/taskmd:get-task <ID>` | View task details by ID or name |
| `/taskmd:add-task <description>` | Create a new task file |
| `/taskmd:complete-task <ID>` | Mark a task as completed |
| `/taskmd:list-tasks` | List tasks with optional filters |
| `/taskmd:validate-tasks` | Validate task files for errors |

## Usage

### Find your next task

```
/taskmd:next-task
```

Pass flags to filter results:

```
/taskmd:next-task --filter tag=mvp
/taskmd:next-task --quick-wins
/taskmd:next-task --critical --limit 3
```

### Work on a task

```
/taskmd:do-task 015
```

This looks up the task, marks it as in-progress, works through the subtasks, and marks it completed when done. For non-trivial tasks, Claude will enter plan mode first.

### List tasks

```
/taskmd:list-tasks
/taskmd:list-tasks --status pending
/taskmd:list-tasks --filter priority=high --format json
```

### View a specific task

```
/taskmd:get-task 042
/taskmd:get-task authentication
```

### Create a task

```
/taskmd:add-task Add rate limiting to the API endpoints
```

Claude will determine the next available ID, choose the appropriate subdirectory, and create the task file with proper frontmatter.

### Complete a task

```
/taskmd:complete-task 015
```

### Validate task files

```
/taskmd:validate-tasks
/taskmd:validate-tasks --format json
```

## Troubleshooting

### "taskmd: command not found"

The CLI is not installed or not in your PATH. Install it:

```bash
brew tap driangle/tap && brew install taskmd
```

Or see [Installation](/getting-started/installation) for other methods.

### "no task files found"

Ensure you have a `tasks/` directory with `.md` files in your project, or configure the default directory in `.taskmd.yaml`:

```yaml
dir: ./tasks
```

### Skills not appearing

Verify the plugin is installed:

```
/plugins list
```

## Source

The plugin source lives in [`claude-code-plugin/`](https://github.com/driangle/taskmd/tree/main/claude-code-plugin) in the taskmd repository.
