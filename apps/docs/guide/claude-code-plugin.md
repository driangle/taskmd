# Claude Code Plugin

Use taskmd directly inside [Claude Code](https://claude.com/claude-code) with slash commands. The plugin provides skills for listing, creating, completing, and validating tasks without leaving your Claude session.

## Prerequisites

The `taskmd` CLI must be installed and available in your PATH. See [Installation](/getting-started/installation) for setup instructions.

## Installation

Install from the Claude Code plugin marketplace:

```
/plugins add taskmd
```

Or install directly from a local clone:

```
/plugins add /path/to/taskmd/claude-code-plugin
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
| `/taskmd:validate` | Validate task files for errors |

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
/taskmd:validate
/taskmd:validate --format json
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
