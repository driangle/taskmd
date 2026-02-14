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

Install the plugin from the Claude Code plugin marketplace:

```
/plugins add taskmd
```

Or install directly from this repository:

```
/plugins add /path/to/taskmd/claude-code-plugin
```

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

# Look up a specific task
/taskmd:get-task 042
```

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
