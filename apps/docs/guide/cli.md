# CLI Guide

Complete reference for using taskmd from the command line.

## Command Reference

### Quick Reference

| Command | Description |
|---------|-------------|
| `list` | List tasks in a quick textual format |
| `get` | Get detailed information about a specific task |
| `set` | Set a task's frontmatter fields |
| `next` | Recommend what task to work on next |
| `validate` | Lint and validate tasks |
| `graph` | Export task dependency graph |
| `board` | Display tasks grouped in a kanban-like board view |
| `stats` | Show computed metrics about tasks |
| `tags` | List all tags with task counts |
| `snapshot` | Produce a frozen, machine-readable representation of tasks |
| `report` | Generate a comprehensive project report |
| `tracks` | Show parallel work tracks based on scope overlap |
| `archive` | Archive or delete completed/cancelled tasks |
| `next-id` | Show the next available task ID |
| `sync` | Sync tasks from external sources |
| `web` | Web dashboard commands |
| `init` | Initialize a project with agent configuration and spec files |
| `completion` | Generate shell completion scripts |

---

### list - View and Filter Tasks

Display tasks in various formats with filtering and sorting.

**Basic usage:**
```bash
# List all tasks
taskmd list

# List tasks in specific directory
taskmd list ./tasks

# Different output formats
taskmd list --format table   # Default
taskmd list --format json
taskmd list --format yaml
```

**Filtering:**
```bash
# Filter by status
taskmd list --filter status=pending
taskmd list --filter status=in-progress

# Filter by priority
taskmd list --filter priority=high

# Filter by multiple criteria (AND logic)
taskmd list --filter status=pending --filter priority=high

# Filter by tag
taskmd list --filter tag=cli

# Filter by effort
taskmd list --filter effort=small
```

**Sorting:**
```bash
# Sort by priority
taskmd list --sort priority

# Sort by status
taskmd list --sort status

# Sort by created date
taskmd list --sort created
```

**Custom columns:**
```bash
# Show specific columns
taskmd list --columns id,title,status

# Show more columns
taskmd list --columns id,title,status,priority,effort,deps
```

**Examples:**
```bash
# High-priority pending tasks
taskmd list --filter status=pending --filter priority=high

# Small tasks (quick wins)
taskmd list --filter effort=small --filter status=pending

# All CLI-related tasks
taskmd list --filter tag=cli --sort priority

# Export to JSON for scripting
taskmd list --format json > tasks.json
```

### validate - Check Task Files

Validate task files for errors and consistency issues.

```bash
# Validate all tasks
taskmd validate

# Validate specific directory
taskmd validate ./tasks

# Strict mode (enable warnings)
taskmd validate --strict

# JSON output
taskmd validate --format json
```

**What it checks:**
- Required fields present (id, title, status)
- Valid field values
- Duplicate task IDs
- Missing dependencies (references to non-existent tasks)
- Circular dependencies
- YAML syntax errors

**Exit codes:**
- `0` - Valid (no errors)
- `1` - Invalid (errors found)
- `2` - Valid with warnings (strict mode only)

### next - Find What to Work On

Analyze tasks and recommend the best ones to work on next.

taskmd scores tasks based on:
- **Priority**: High priority scores higher
- **Critical path**: Tasks on the critical path score higher
- **Downstream impact**: Tasks blocking many others score higher
- **Effort**: Smaller tasks get a boost (quick wins)
- **Actionability**: Only tasks with satisfied dependencies

```bash
# Get top 5 recommendations
taskmd next

# Get top 3 recommendations
taskmd next --limit 3

# Next high-priority task
taskmd next --filter priority=high

# Next small task (quick win)
taskmd next --filter effort=small --limit 1

# JSON for automation
taskmd next --format json
```

### graph - Visualize Dependencies

Export task dependency graphs in various formats.

```bash
# ASCII art (terminal-friendly)
taskmd graph --format ascii

# Mermaid diagram
taskmd graph --format mermaid

# Graphviz DOT
taskmd graph --format dot

# JSON structure
taskmd graph --format json
```

**Filtering:**
```bash
# Exclude completed tasks (default)
taskmd graph

# Include all tasks
taskmd graph --all

# Exclude specific statuses
taskmd graph --exclude-status completed --exclude-status blocked
```

**Focus on specific tasks:**
```bash
# Show task and its dependencies (upstream)
taskmd graph --root 022 --upstream

# Show task and what depends on it (downstream)
taskmd graph --root 022 --downstream

# Show full subgraph
taskmd graph --root 022
```

**Output to file:**
```bash
taskmd graph --format mermaid --out deps.mmd
taskmd graph --format dot --out deps.dot

# Generate PNG with Graphviz
taskmd graph --format dot | dot -Tpng > graph.png
```

### stats - Project Metrics

Display computed statistics about your task set.

```bash
# Show all statistics
taskmd stats

# Specific directory
taskmd stats ./tasks

# JSON output
taskmd stats --format json
```

**Metrics provided:**
- Total tasks and count by status
- Priority breakdown
- Effort breakdown
- Blocked tasks count
- Completion rate
- Critical path length
- Max dependency depth
- Average dependencies per task

### board - Kanban View

Display tasks grouped by a field in a board layout.

```bash
# Group by status (default)
taskmd board

# Group by priority
taskmd board --group-by priority

# Group by effort
taskmd board --group-by effort

# Group by tag
taskmd board --group-by tag

# Output formats
taskmd board --format md    # Markdown (default)
taskmd board --format txt   # Plain text
taskmd board --format json  # JSON
```

### snapshot - Machine-Readable Export

Produce a static, machine-readable representation for automation.

```bash
# Full snapshot (JSON)
taskmd snapshot

# Core fields only
taskmd snapshot --core

# Include derived analysis
taskmd snapshot --derived

# Output formats
taskmd snapshot --format json
taskmd snapshot --format yaml
taskmd snapshot --format md

# Grouping
taskmd snapshot --group-by status
taskmd snapshot --group-by priority

# Output to file
taskmd snapshot --out snapshot.json
```

### get - View Task Details

Display detailed information about a specific task, identified by ID, title, or file path.

**Matching priority:**
1. Exact match by task ID (case-sensitive)
2. Exact match by task title (case-insensitive)
3. Match by file path or filename
4. Fuzzy match across IDs and titles (unless `--exact` is set)

```bash
# Look up by task ID
taskmd get cli-037

# Look up by title
taskmd get "Add show command"

# Look up by file path
taskmd get tasks/cli/037-task.md

# Fuzzy search
taskmd get sho

# Strict lookup — fail if no exact match
taskmd get sho --exact
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `text` | Output format (`text`, `json`, `yaml`) |
| `--exact` | `false` | Disable fuzzy matching |
| `--threshold` | `0.6` | Fuzzy match sensitivity (0.0–1.0) |

### set - Update Task Fields

Modify a task's frontmatter fields by ID.

```bash
# Change status
taskmd set --task-id 042 --status in-progress

# Change priority and effort
taskmd set --task-id 042 --priority high --effort large

# Mark as completed (shortcut)
taskmd set --task-id 042 --done

# Preview changes without writing
taskmd set --task-id 042 --priority critical --dry-run
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-id` | *(required)* | Task ID to update |
| `--status` | | New status (`pending`, `in-progress`, `completed`, `blocked`, `cancelled`) |
| `--priority` | | New priority (`low`, `medium`, `high`, `critical`) |
| `--effort` | | New effort (`small`, `medium`, `large`) |
| `--owner` | | Owner/assignee |
| `--parent` | | Parent task ID (empty string to clear) |
| `--done` | `false` | Alias for `--status completed` |
| `--dry-run` | `false` | Preview changes without writing to disk |
| `--add-tag` | | Add a tag (repeatable) |
| `--remove-tag` | | Remove a tag (repeatable) |

**Tag management:**
```bash
# Add tags
taskmd set --task-id 042 --add-tag backend --add-tag api

# Remove a tag
taskmd set --task-id 042 --remove-tag deprecated

# Add and remove in one command
taskmd set --task-id 042 --add-tag v2 --remove-tag v1
```

### tags - List Tags

Display all tags used across task files with usage counts.

```bash
# List all tags
taskmd tags

# List tags in specific directory
taskmd tags ./tasks

# Tags used by pending tasks only
taskmd tags --filter status=pending

# JSON output
taskmd tags --format json
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--filter` | | Filter tasks before aggregating (repeatable) |

### archive - Archive Completed Tasks

Move completed or cancelled task files into an `archive/` subdirectory, or permanently delete them.

```bash
# Archive all completed tasks
taskmd archive --all-completed -y

# Archive all cancelled tasks
taskmd archive --all-cancelled -y

# Archive specific tasks by ID
taskmd archive --id 042 --id 043 -y

# Preview what would be archived
taskmd archive --all-completed --dry-run

# Permanently delete cancelled tasks
taskmd archive --all-cancelled --delete -f
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--id` | | Archive task(s) by ID (repeatable) |
| `--status` | | Archive tasks matching this status |
| `--all-completed` | `false` | Archive all completed tasks |
| `--all-cancelled` | `false` | Archive all cancelled tasks |
| `--tag` | | Archive tasks with this tag |
| `--dry-run` | `false` | Preview changes without making them |
| `--yes`, `-y` | `false` | Skip confirmation prompt |
| `--delete` | `false` | Permanently delete instead of archive |
| `--force`, `-f` | `false` | Skip confirmation for delete |

### next-id - Get Next Available ID

Scan task files and output the next available sequential ID. Finds the highest numeric ID and returns max + 1, preserving any common prefix and zero-padding.

```bash
# Get next ID
taskmd next-id

# Scan specific directory
taskmd next-id ./tasks/cli

# JSON output with metadata
taskmd next-id --format json
```

**Scripting example:**
```bash
# Create a new task file with the next ID
ID=$(taskmd next-id)
echo "---
id: \"$ID\"
title: \"My new task\"
status: pending
---" > "tasks/${ID}-my-new-task.md"
```

### report - Generate Reports

Generate a comprehensive project report combining summary statistics, task groupings, critical-path analysis, and blocked tasks.

```bash
# Markdown report to stdout
taskmd report

# HTML report to file
taskmd report --format html --out report.html

# JSON report grouped by priority
taskmd report tasks/ --group-by priority --format json

# Include dependency graph
taskmd report tasks/ --format html --include-graph --out report.html
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `md` | Output format (`md`, `html`, `json`) |
| `--group-by` | `status` | Field to group by (`status`, `priority`, `effort`, `group`, `tag`) |
| `--out`, `-o` | | Write output to file |
| `--include-graph` | `false` | Embed dependency graph in report |

### tracks - Parallel Work Tracks

Assign actionable tasks to parallel work tracks based on the `touches` frontmatter field. Tasks that share a scope are placed in separate tracks so they can be worked on without merge conflicts.

```bash
# Show work tracks
taskmd tracks

# Filter to CLI-related tasks
taskmd tracks --filter tag=cli

# Limit to top 3 tracks
taskmd tracks --limit 3

# Export track assignments
taskmd tracks --format json > tracks.json
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--filter` | | Filter tasks (repeatable) |
| `--limit` | `0` | Maximum number of tracks (0 = unlimited) |

### sync - Sync External Sources

Fetch tasks from configured external sources (e.g., GitHub Issues) and create or update local markdown task files. Configuration is read from `.taskmd.yaml`.

```bash
# Sync all configured sources
taskmd sync

# Preview without writing files
taskmd sync --dry-run

# Sync a specific source
taskmd sync --source github

# Overwrite local changes with remote data
taskmd sync --conflict remote
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview changes without writing files |
| `--source` | | Sync only the named source |
| `--conflict` | `skip` | Conflict strategy: `skip`, `remote`, or `local` |

See [Configuration](/reference/configuration#sync-configuration) for how to set up sync sources in `.taskmd.yaml`.

### web - Web Dashboard

Start the web interface server.

```bash
# Start server
taskmd web start

# Start and open browser
taskmd web start --open

# Custom port
taskmd web start --port 3000

# Specific tasks directory
taskmd web start --dir ./my-tasks --open
```

See the [Web Interface Guide](./web) for detailed web UI documentation.

## Global Flags

Available for all commands:

```bash
--config string    # Config file path
--dir string       # Task directory (default ".")
--format string    # Output format (table, json, yaml)
--verbose          # Verbose logging
--quiet            # Suppress non-essential output
--stdin            # Read from stdin instead of files
```

## Common Workflows

### Daily Task Management

```bash
# Morning: What should I work on?
taskmd next --limit 5

# Check project status
taskmd stats

# During work: Validate changes
taskmd validate

# End of day: What got done?
taskmd list --filter status=completed --sort created
```

### Weekly Planning

```bash
# Visual overview
taskmd board --group-by priority

# Identify bottlenecks
taskmd graph --exclude-status completed --format ascii

# Focus on priorities
taskmd list --filter status=pending --sort priority
```

### CI/CD Integration

```bash
# Validate in CI pipeline
if ! taskmd validate tasks/ --strict; then
    echo "Task validation failed"
    exit 1
fi

# Generate snapshot artifact
taskmd snapshot tasks/ --derived --out task-snapshot.json
```

### Scripting and Piping

```bash
# Pipe between commands
taskmd list --format json | jq '.[] | select(.priority == "high")'

# Validate from stdin
echo '---
id: "999"
title: "Test task"
status: pending
---
# Test' | taskmd validate --stdin

# Find quick wins
taskmd list tasks/ \
  --filter status=pending \
  --filter priority=high \
  --filter effort=small \
  --format json | jq -r '.[] | "\(.id): \(.title)"'
```

## Environment Variables

taskmd supports environment variables with the `TASKMD_` prefix:

```bash
export TASKMD_DIR=./tasks
export TASKMD_VERBOSE=true
```

Environment variables have lower precedence than config files and CLI flags.

## Troubleshooting

### "No tasks found"

1. Check that your tasks directory exists: `ls -la tasks/`
2. Ensure files have `.md` extension
3. Verify YAML frontmatter format
4. Run `taskmd validate tasks/` for specific errors
5. Try verbose output: `taskmd list tasks/ --verbose`

### "Invalid task format"

- Check YAML frontmatter is properly formatted
- Ensure required fields are present: `id`, `title`
- Verify status is valid: `pending`, `in-progress`, `completed`, `blocked`, `cancelled`
- Run `taskmd validate tasks/` for line-level error messages

### "Circular dependency detected"

Dependencies form a cycle (A depends on B, B depends on A). Use `taskmd graph --format ascii` to visualize the cycle and remove one dependency to break it.

### Command not found

```bash
# Check installation
which taskmd
taskmd --version

# If not found, add to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```
