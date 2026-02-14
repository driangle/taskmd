# CLI User Guide

Complete reference for using taskmd from the command line.

## What You'll Learn

- Installation methods
- Core concepts
- All CLI commands with practical examples
- Common workflows
- Configuration options
- Tips and best practices

## Installation

### Option 1: Download Pre-built Binary

```bash
# Download from GitHub releases
# (Coming soon - see task 045)

# Extract and add to PATH
tar -xzf taskmd-*.tar.gz
sudo mv taskmd /usr/local/bin/
```

### Option 2: Install with Go

```bash
go install github.com/yourusername/md-task-tracker/cmd/taskmd@latest
```

Requirements:
- Go 1.22 or later

### Option 3: Build from Source

```bash
# Clone repository
git clone https://github.com/yourusername/md-task-tracker.git
cd md-task-tracker/apps/cli

# Build CLI only
make build

# Build with embedded web interface
make build-full

# Binary will be at bin/taskmd
```

### Option 4: Homebrew (Coming Soon)

```bash
brew install taskmd
```

### Verify Installation

```bash
taskmd --version
taskmd --help
```

## Core Concepts

### Tasks

Tasks are markdown files with YAML frontmatter. Each task has:

- **Required fields**: `id`, `title`, `status`
- **Optional fields**: `priority`, `effort`, `dependencies`, `tags`, `created`
- **Markdown body**: Rich description with objectives, subtasks, and acceptance criteria

### Task Status

- `pending` - Not started yet
- `in-progress` - Currently being worked on
- `completed` - Finished
- `blocked` - Cannot proceed (due to dependencies or external blockers)
- `cancelled` - Will not be completed (kept for historical reference)

### Dependencies

Tasks can depend on other tasks using the `dependencies` field:

```yaml
dependencies:
  - "001"  # Must complete task 001 first
  - "005"  # And task 005
```

Dependencies create a directed acyclic graph (DAG) that taskmd uses to:
- Recommend next tasks to work on
- Visualize task relationships
- Calculate critical paths
- Identify blockers

### Task Discovery

taskmd scans directories recursively for `.md` files with valid task frontmatter:

```bash
# Scan current directory
taskmd list

# Scan specific directory
taskmd list ./tasks

# Scan subdirectories
taskmd list ./tasks/cli
```

## Command Reference

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

**Basic usage:**
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

**Examples:**
```bash
# Quick validation
taskmd validate tasks/

# CI/CD integration
if ! taskmd validate tasks/ --quiet; then
    echo "Task validation failed"
    exit 1
fi

# Strict validation with warnings
taskmd validate tasks/ --strict

# Get detailed error report
taskmd validate tasks/ --format json > validation-report.json
```

### next - Find What to Work On

Analyze tasks and recommend the best ones to work on next.

**How it works:**

taskmd scores tasks based on:
- **Priority**: High priority scores higher
- **Critical path**: Tasks on the critical path score higher
- **Downstream impact**: Tasks blocking many others score higher
- **Effort**: Smaller tasks get a boost (quick wins)
- **Actionability**: Only tasks with satisfied dependencies

**Basic usage:**
```bash
# Get top 5 recommendations
taskmd next

# Get top 3 recommendations
taskmd next --limit 3

# Get all actionable tasks
taskmd next --limit 100
```

**Filtering:**
```bash
# Next high-priority task
taskmd next --filter priority=high

# Next CLI task
taskmd next --filter tag=cli

# Next small task (quick win)
taskmd next --filter effort=small --limit 1
```

**Output formats:**
```bash
# Table (default)
taskmd next

# JSON for automation
taskmd next --format json

# YAML
taskmd next --format yaml
```

**Examples:**
```bash
# Morning planning: What should I work on?
taskmd next --limit 3

# Find a quick win
taskmd next --filter effort=small --limit 1

# Focus on critical path
taskmd next --limit 5 | grep -i "critical"

# Get next high-priority backend task
taskmd next --filter priority=high --filter tag=backend
```

### graph - Visualize Dependencies

Export task dependency graphs in various formats.

**Basic usage:**
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

# Show only pending tasks
taskmd graph --exclude-status completed --exclude-status in-progress
```

**Focus on specific tasks:**
```bash
# Highlight a specific task
taskmd graph --focus 022 --format mermaid

# Show task and its dependencies (upstream)
taskmd graph --root 022 --upstream

# Show task and what depends on it (downstream)
taskmd graph --root 022 --downstream

# Show full subgraph
taskmd graph --root 022
```

**Output to file:**
```bash
# Save to file
taskmd graph --format mermaid --out deps.mmd
taskmd graph --format dot --out deps.dot
taskmd graph --format json --out graph.json
```

**Examples:**
```bash
# Quick view in terminal
taskmd graph --format ascii

# Generate PNG with Graphviz
taskmd graph --format dot | dot -Tpng > graph.png

# Mermaid for documentation
taskmd graph --format mermaid > docs/dependencies.mmd

# Find what blocks task 025
taskmd graph --root 025 --upstream --format ascii

# See impact of task 010
taskmd graph --root 010 --downstream --format ascii

# Active tasks only
taskmd graph --exclude-status completed --format mermaid
```

### stats - Project Metrics

Display computed statistics about your task set.

**Basic usage:**
```bash
# Show all statistics
taskmd stats

# Specific directory
taskmd stats ./tasks

# JSON output
taskmd stats --format json
```

**Metrics provided:**

- **Total tasks**: Count by status
- **Priority breakdown**: Tasks by priority level
- **Effort breakdown**: Tasks by effort estimate
- **Blocked tasks**: Count of blocked tasks
- **Completion rate**: Percentage complete
- **Critical path**: Longest dependency chain
- **Max depth**: Deepest dependency level
- **Avg dependencies**: Average deps per task

**Examples:**
```bash
# Quick project overview
taskmd stats

# Export for reporting
taskmd stats --format json > project-stats.json

# Check completion rate
taskmd stats | grep "Completion"

# Monitor critical path length
taskmd stats | grep "Critical path"
```

### board - Kanban View

Display tasks grouped by a field in a board/kanban layout.

**Basic usage:**
```bash
# Group by status (default)
taskmd board

# Group by priority
taskmd board --group-by priority

# Group by effort
taskmd board --group-by effort

# Group by tag
taskmd board --group-by tag
```

**Output formats:**
```bash
# Markdown (default)
taskmd board --format md

# Plain text
taskmd board --format txt

# JSON
taskmd board --format json
```

**Output to file:**
```bash
# Save board view
taskmd board --out board.md
taskmd board --group-by priority --format txt --out priority-board.txt
```

**Examples:**
```bash
# Status board (kanban)
taskmd board

# Priority planning
taskmd board --group-by priority --format txt

# Effort estimation view
taskmd board --group-by effort

# Tag-based organization
taskmd board --group-by tag --format json

# Save weekly board
taskmd board --out weekly-board-$(date +%Y-%m-%d).md
```

### snapshot - Machine-Readable Export

Produce static, machine-readable representation for automation.

**Basic usage:**
```bash
# Full snapshot (JSON)
taskmd snapshot

# Core fields only
taskmd snapshot --core

# Include derived analysis
taskmd snapshot --derived
```

**Output formats:**
```bash
# JSON (default)
taskmd snapshot --format json

# YAML
taskmd snapshot --format yaml

# Markdown
taskmd snapshot --format md
```

**Grouping:**
```bash
# Group by status
taskmd snapshot --group-by status

# Group by priority
taskmd snapshot --group-by priority

# Group by effort
taskmd snapshot --group-by effort
```

**Output to file:**
```bash
taskmd snapshot --out snapshot.json
taskmd snapshot --format yaml --out snapshot.yaml
```

**Examples:**
```bash
# CI/CD artifact
taskmd snapshot --derived --format json > ci-snapshot.json

# Backup
taskmd snapshot --out backup-$(date +%Y%m%d).json

# Core data only
taskmd snapshot --core --format yaml > minimal.yaml

# Grouped report
taskmd snapshot --group-by status --format md > status-report.md

# API data
taskmd snapshot --format json > public/api/tasks.json
```

### web - Web Dashboard

Start the web interface server.

**Basic usage:**
```bash
# Start server
taskmd web start

# Start and open browser
taskmd web start --open

# Custom port
taskmd web start --port 3000

# Development mode (CORS for Vite)
taskmd web start --dev
```

**Full command:**
```bash
taskmd web start [flags]
```

**Flags:**
- `--port int` - Server port (default 8080)
- `--open` - Open browser automatically
- `--dev` - Enable dev mode with CORS
- `--dir string` - Task directory to serve

**Examples:**
```bash
# Standard usage
taskmd web start --open

# Different port
taskmd web start --port 3000 --open

# Specific tasks directory
taskmd web start --dir ./my-tasks --open

# Development with Vite
taskmd web start --dev --port 8080
```

See [Web User Guide](web-guide.md) for detailed web interface documentation.

### get - View Task Details

Display detailed information about a specific task, identified by ID, title, or file path.

**Matching priority:**
1. Exact match by task ID (case-sensitive)
2. Exact match by task title (case-insensitive)
3. Match by file path or filename
4. Fuzzy match across IDs and titles (unless `--exact` is set)

**Basic usage:**
```bash
# Look up by task ID
taskmd get cli-037

# Look up by title
taskmd get "Add show command"

# Look up by file path
taskmd get tasks/cli/037-task.md

# Look up by filename (with or without extension)
taskmd get 037-task.md
taskmd get 037-task
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `text` | Output format (`text`, `json`, `yaml`) |
| `--exact` | `false` | Disable fuzzy matching, exact match only |
| `--threshold float` | `0.6` | Fuzzy match sensitivity (0.0–1.0) |

**Output formats:**
```bash
# Human-readable (default)
taskmd get cli-037

# JSON for scripting
taskmd get cli-037 --format json

# YAML
taskmd get cli-037 --format yaml
```

**Examples:**
```bash
# Quick task lookup
taskmd get 042

# Fuzzy search (interactive selection if multiple matches)
taskmd get sho

# Strict lookup — fail if no exact match
taskmd get sho --exact

# Pipe JSON output into jq
taskmd get 042 --format json | jq '.dependencies'
```

### set - Update Task Fields

Modify a task's frontmatter fields (status, priority, effort, tags, owner, parent) by ID.

**Basic usage:**
```bash
# Change status
taskmd set --task-id 042 --status in-progress

# Change priority and effort
taskmd set --task-id 042 --priority high --effort large

# Mark as completed (shortcut)
taskmd set --task-id 042 --done
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-id string` | *(required)* | Task ID to update |
| `--status string` | | New status (`pending`, `in-progress`, `completed`, `blocked`, `cancelled`) |
| `--priority string` | | New priority (`low`, `medium`, `high`, `critical`) |
| `--effort string` | | New effort (`small`, `medium`, `large`) |
| `--owner string` | | Owner/assignee of the task |
| `--parent string` | | Parent task ID (empty string to clear) |
| `--done` | `false` | Alias for `--status completed` |
| `--dry-run` | `false` | Preview changes without writing to disk |
| `--add-tag string` | | Add a tag (repeatable) |
| `--remove-tag string` | | Remove a tag (repeatable) |

**Tag management:**
```bash
# Add tags
taskmd set --task-id 042 --add-tag backend --add-tag api

# Remove a tag
taskmd set --task-id 042 --remove-tag deprecated

# Add and remove in one command
taskmd set --task-id 042 --add-tag v2 --remove-tag v1
```

**Examples:**
```bash
# Start working on a task
taskmd set --task-id 042 --status in-progress

# Preview changes before applying
taskmd set --task-id 042 --priority critical --dry-run

# Set owner and parent
taskmd set --task-id 042 --owner alice --parent 040

# Mark done
taskmd set --task-id 042 --done
```

### tags - List Tags

Display all tags used across task files with usage counts, sorted from most to least used.

**Basic usage:**
```bash
# List all tags
taskmd tags

# List tags in specific directory
taskmd tags ./tasks
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `table` | Output format (`table`, `json`, `yaml`) |
| `--filter string` | | Filter tasks before aggregating (repeatable, AND logic) |

**Filtering:**
```bash
# Tags used by pending tasks only
taskmd tags --filter status=pending

# Tags used by high-priority tasks
taskmd tags --filter priority=high
```

**Output formats:**
```bash
# Table (default)
taskmd tags

# JSON for scripting
taskmd tags --format json

# YAML
taskmd tags --format yaml
```

**Examples:**
```bash
# See which tags are most common
taskmd tags

# Tags on in-progress tasks only
taskmd tags --filter status=in-progress

# Export tag data
taskmd tags --format json > tags.json
```

### archive - Archive Completed Tasks

Move completed or cancelled task files into an `archive/` subdirectory, or permanently delete them. Keeps your main task list clean while preserving history.

**Basic usage:**
```bash
# Archive all completed tasks
taskmd archive --all-completed -y

# Archive all cancelled tasks
taskmd archive --all-cancelled -y

# Archive specific tasks by ID
taskmd archive --id 042 --id 043 -y
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--id string` | | Archive task(s) by ID (repeatable) |
| `--status string` | | Archive tasks matching this status |
| `--all-completed` | `false` | Archive all completed tasks |
| `--all-cancelled` | `false` | Archive all cancelled tasks |
| `--tag string` | | Archive tasks with this tag |
| `--dry-run` | `false` | Preview changes without making them |
| `--yes`, `-y` | `false` | Skip confirmation prompt |
| `--delete` | `false` | Permanently delete instead of archive |
| `--force`, `-f` | `false` | Skip confirmation for delete |

**Examples:**
```bash
# Preview what would be archived
taskmd archive --all-completed --dry-run

# Archive completed backend tasks
taskmd archive --status completed --tag backend -y

# Permanently delete cancelled tasks
taskmd archive --all-cancelled --delete -f

# Archive a specific task
taskmd archive --id 042 -y
```

### next-id - Get Next Available ID

Scan task files and output the next available sequential ID. Finds the highest numeric ID among existing tasks and returns max + 1, preserving any common prefix and zero-padding.

**Basic usage:**
```bash
# Get next ID
taskmd next-id

# Scan specific directory
taskmd next-id ./tasks/cli
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `plain` | Output format (`plain`, `json`) |

**Output formats:**
```bash
# Plain text — just the ID (default, ideal for scripting)
taskmd next-id

# JSON with metadata
taskmd next-id --format json
```

**Examples:**
```bash
# Create a new task file with the next ID
ID=$(taskmd next-id)
echo "---
id: \"$ID\"
title: \"My new task\"
status: pending
---" > "tasks/${ID}-my-new-task.md"

# Get next ID in a specific directory
taskmd next-id ./tasks/cli

# JSON output for automation
taskmd next-id --format json
```

### report - Generate Reports

Generate a comprehensive project report combining summary statistics, task groupings, critical-path analysis, blocked tasks, and optional dependency graphs.

**Basic usage:**
```bash
# Markdown report to stdout
taskmd report

# Scan specific directory
taskmd report tasks/

# HTML report to file
taskmd report --format html --out report.html
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `md` | Output format (`md`, `html`, `json`) |
| `--group-by string` | `status` | Field to group by (`status`, `priority`, `effort`, `group`, `tag`) |
| `--out`, `-o string` | | Write output to file instead of stdout |
| `--include-graph` | `false` | Embed dependency graph in report |

**Output formats:**
```bash
# Markdown (default)
taskmd report --format md

# Self-contained HTML with inline CSS
taskmd report --format html

# Structured JSON
taskmd report --format json
```

**Examples:**
```bash
# Weekly status report
taskmd report tasks/ --format md --out weekly-report.md

# HTML report with dependency graph
taskmd report tasks/ --format html --include-graph --out report.html

# Group by priority for planning
taskmd report tasks/ --group-by priority --format json

# Quick terminal report
taskmd report tasks/
```

### tracks - Parallel Work Tracks

Assign actionable tasks to parallel work tracks based on the `touches` frontmatter field. Tasks that share a scope (e.g., the same file or module) are placed in separate tracks so they can be worked on without merge conflicts.

**How it works:**

- Tasks declare which areas they affect using the `touches` frontmatter field
- Tasks sharing a scope are placed in separate tracks
- Tasks without `touches` are listed as "flexible" — they can join any track
- Scope definitions can be configured in `.taskmd.yaml` under the `scopes` key

**Basic usage:**
```bash
# Show work tracks
taskmd tracks

# Scan specific directory
taskmd tracks ./tasks
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `table` | Output format (`table`, `json`, `yaml`) |
| `--filter string` | | Filter tasks (repeatable, e.g., `--filter tag=cli`) |
| `--limit int` | `0` | Maximum number of tracks to show (0 = unlimited) |

**Examples:**
```bash
# See all work tracks
taskmd tracks

# Filter to CLI-related tasks
taskmd tracks --filter tag=cli

# Limit to top 3 tracks
taskmd tracks --limit 3

# Export track assignments
taskmd tracks --format json > tracks.json
```

### sync - Sync External Sources

Fetch tasks from configured external sources (e.g., GitHub Issues) and create or update local markdown task files. Configuration is read from `.taskmd.yaml`.

**Basic usage:**
```bash
# Sync all configured sources
taskmd sync

# Preview without writing files
taskmd sync --dry-run

# Sync a specific source
taskmd sync --source github
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview changes without writing files |
| `--source string` | | Sync only the named source |
| `--conflict string` | `skip` | Conflict resolution strategy (`skip`, `remote`, `local`) |

**Conflict strategies:**

| Strategy | Behavior |
|----------|----------|
| `skip` | Skip tasks that have local changes (default) |
| `remote` | Overwrite local changes with remote data |
| `local` | Keep local changes, ignore remote updates |

**Examples:**
```bash
# Full sync
taskmd sync

# Preview what would change
taskmd sync --dry-run

# Sync only GitHub source
taskmd sync --source github

# Overwrite local changes with remote data
taskmd sync --conflict remote

# Keep local changes, ignore remote updates
taskmd sync --conflict local
```

See the [Sync Configuration](#sync-configuration) section below for how to set up `.taskmd.yaml` for sync.

## Common Workflows

### Daily Task Management

**Morning: Plan your day**
```bash
# See what needs attention
taskmd next --limit 5

# Check project status
taskmd stats

# Review high-priority tasks
taskmd list --filter priority=high --filter status=pending
```

**During work: Track progress**
```bash
# Update task status in your editor
taskmd set --task-id 042 --status in-progress

# Validate changes
taskmd validate
```

**End of day: Review**
```bash
# Check what got done
taskmd list --filter status=completed --sort created

# See tomorrow's options
taskmd next
```

### Weekly Planning

**Monday: Week planning**
```bash
# Visual overview
taskmd board --group-by priority

# Or use web interface
taskmd web start --open

# Identify bottlenecks
taskmd graph --exclude-status completed --format ascii

# Set priorities
taskmd list --filter status=pending --sort priority
```

**Friday: Week review**
```bash
# What was completed
taskmd list --filter status=completed

# Statistics
taskmd stats

# Save snapshot
taskmd snapshot --derived --out weekly-$(date +%Y-%m-%d).json
```

### Project Initialization

```bash
# Create structure
mkdir -p my-project/tasks
cd my-project

# Create initial tasks
# (Create task files in tasks/)

# Validate structure
taskmd validate tasks/

# Visualize plan
taskmd graph tasks/ --format ascii

# Generate initial board
taskmd board tasks/ --out project-plan.md
```

### Continuous Integration

```bash
#!/bin/bash
# .github/workflows/validate-tasks.yml

# Validate all tasks
if ! taskmd validate tasks/ --strict; then
    echo "❌ Task validation failed"
    exit 1
fi

# Check for circular dependencies
if taskmd graph tasks/ --format json | jq '.cycles | length' | grep -v '^0$'; then
    echo "❌ Circular dependencies detected"
    exit 1
fi

# Generate snapshot artifact
taskmd snapshot tasks/ --derived --out task-snapshot.json

echo "✅ All task checks passed"
```

### Task Dependencies Management

**Understanding dependencies:**
```bash
# See what task 025 depends on
taskmd graph --root 025 --upstream --format ascii

# See what depends on task 010
taskmd graph --root 010 --downstream --format ascii

# Find critical path
taskmd stats | grep "Critical path"
```

**Finding actionable tasks:**
```bash
# Tasks ready to work on (no blockers)
taskmd next

# All pending tasks with satisfied dependencies
taskmd list --filter status=pending | taskmd next --limit 100
```

### Reporting and Export

**Status reports:**
```bash
# Markdown report
taskmd board --format md --out status-report.md

# JSON for external tools
taskmd snapshot --group-by status --format json > report.json

# Statistics summary
taskmd stats > project-stats.txt
```

**Visualizations:**
```bash
# Generate dependency graph PNG
taskmd graph --format dot | dot -Tpng > dependencies.png

# Mermaid for documentation
taskmd graph --format mermaid > docs/task-graph.mmd

# ASCII for terminal/logs
taskmd graph --format ascii > task-tree.txt
```

## Configuration

### Config File Support

taskmd supports `.taskmd.yaml` configuration files to set default options without repeating command-line flags.

**Supported Config Options:**

```yaml
# .taskmd.yaml
dir: ./tasks                    # Default task directory
web:
  port: 8080                   # Default web server port
  auto_open_browser: true      # Auto-open browser on web start
```

**Config File Locations:**

1. **Project-level**: `./.taskmd.yaml` (in current directory)
2. **Global**: `~/.taskmd.yaml` (in home directory)
3. **Custom**: Use `--config path/to/config.yaml`

**Precedence Order** (highest to lowest):

1. Command-line flags (explicit user intent)
2. Project-level `.taskmd.yaml` (project-specific defaults)
3. Global `~/.taskmd.yaml` (user-wide defaults)
4. Built-in defaults (fallback)

**Example Usage:**

```bash
# Create project config
cat > .taskmd.yaml <<EOF
dir: ./tasks
web:
  port: 3000
  auto_open_browser: true
EOF

# Now these commands use config defaults
taskmd list              # Uses ./tasks directory
taskmd web start        # Uses port 3000 and opens browser

# CLI flags still override config
taskmd list --dir ./other-tasks  # Overrides config dir
taskmd web start --port 8080     # Overrides config port
```

See [docs/.taskmd.yaml.example](../.taskmd.yaml.example) for a complete example with comments.

### Sync Configuration

The `sync` command reads its configuration from the `sync` section of `.taskmd.yaml`. Each source defines where to fetch tasks from, how to map fields, and where to write files.

**Example `.taskmd.yaml` with sync:**

```yaml
# .taskmd.yaml
dir: ./tasks

sync:
  sources:
    - name: github
      project: "owner/repo"
      token_env: GITHUB_TOKEN       # Environment variable holding the API token
      output_dir: ./tasks/synced     # Where to write synced task files
      field_map:
        status:
          open: pending
          closed: completed
        priority:
          urgent: critical
          high: high
          medium: medium
          low: low
        labels_to_tags: true         # Convert issue labels to task tags
        assignee_to_owner: true      # Map assignee to owner field
      filters:
        state: open                  # Only sync open issues
```

**Source fields:**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique name for this source |
| `project` | No | Project identifier (e.g., `owner/repo` for GitHub) |
| `base_url` | No | Custom API base URL |
| `token_env` | No | Environment variable name for API token |
| `user_env` | No | Environment variable name for username |
| `output_dir` | Yes | Directory where synced task files are written |
| `field_map` | No | How to map external fields to taskmd frontmatter |
| `filters` | No | Source-specific filters (e.g., `state: open`) |

**Field mapping (`field_map`):**

| Sub-field | Type | Description |
|-----------|------|-------------|
| `status` | `map[string]string` | Map external status values to taskmd statuses |
| `priority` | `map[string]string` | Map external priority values to taskmd priorities |
| `labels_to_tags` | `bool` | Convert external labels/categories to task tags |
| `assignee_to_owner` | `bool` | Map external assignee to the `owner` field |

### Alternative Configuration Methods

**1. Shell Aliases:**
```bash
# Add to ~/.bashrc or ~/.zshrc
alias tm='taskmd --dir ./tasks'
alias tmw='taskmd web start --port 8080 --open'
```

**2. Environment Variables:**
```bash
export TASKMD_DIR=./tasks
```

### Command-Line Flags

Global flags (available for all commands):

```bash
--config string    # Config file path
--dir string       # Task directory (default ".")
--format string    # Output format (table, json, yaml)
--verbose         # Verbose logging
--quiet           # Suppress non-essential output
--stdin           # Read from stdin instead of files
```

### Environment Variables

taskmd supports environment variables with the `TASKMD_` prefix:

```bash
# Override default directory
export TASKMD_DIR=./tasks

# Override verbose flag
export TASKMD_VERBOSE=true

# All flags can be set via TASKMD_FLAGNAME
# Environment variables have lower precedence than config files and CLI flags
```

## Tips and Best Practices

### Task Organization

**1. Use consistent IDs**
```markdown
# Good: Zero-padded
001-setup.md
002-feature-a.md
010-integration.md

# Bad: Inconsistent
1-setup.md
2-feature-a.md
10-integration.md
```

**2. Organize with directories**
```
tasks/
├── cli/
│   ├── 001-list-command.md
│   └── 002-graph-command.md
├── web/
│   ├── 010-board-view.md
│   └── 011-graph-view.md
└── docs/
    └── 020-user-guide.md
```

**3. Use descriptive filenames**
```markdown
# Good
015-add-user-authentication.md
016-implement-rate-limiting.md

# Bad
task1.md
todo.md
```

### Dependency Management

**1. Keep dependency chains short**
- Long chains increase project duration
- Aim for parallel work streams

**2. Identify critical path**
```bash
taskmd stats | grep "Critical path"
taskmd graph --format ascii
```

**3. Break down large tasks**
- Tasks with many dependencies are risky
- Split into smaller, parallel tasks

### Validation

**1. Validate before committing**
```bash
# Pre-commit hook
taskmd validate tasks/ --strict
```

**2. CI/CD integration**
```yaml
# GitHub Actions
- name: Validate tasks
  run: taskmd validate tasks/ --strict
```

**3. Regular validation**
```bash
# Validate often during development
alias tv='taskmd validate tasks/ --strict'
```

### Filtering and Search

**1. Use consistent tags**
```yaml
tags:
  - feature    # Not "feat", "Feature", etc.
  - backend    # Not "back-end", "server"
  - urgent     # Not "URGENT", "high-priority"
```

**2. Combine filters effectively**
```bash
# High-priority pending backend tasks
taskmd list \
  --filter priority=high \
  --filter status=pending \
  --filter tag=backend
```

**3. Save common queries as aliases**
```bash
# In your .bashrc or .zshrc
alias tnext='taskmd next --limit 3'
alias thigh='taskmd list --filter priority=high --filter status=pending'
alias tsmall='taskmd list --filter effort=small --filter status=pending'
```

### Performance

**1. Scan specific directories**
```bash
# Faster
taskmd list ./tasks/cli

# Slower (scans everything)
taskmd list .
```

**2. Use --quiet in scripts**
```bash
# Suppress unnecessary output
taskmd validate --quiet
```

**3. Limit output when needed**
```bash
# Get just what you need
taskmd next --limit 1
```

## Troubleshooting

### "No tasks found"

**Check:**
1. Directory exists: `ls -la tasks/`
2. Files have `.md` extension
3. Files have valid YAML frontmatter
4. Required fields present: `id`, `title`, `status`

**Debug:**
```bash
# Verbose output
taskmd list tasks/ --verbose

# Check specific file
head -20 tasks/001-task.md
```

### "Invalid task format"

**Run validation:**
```bash
taskmd validate tasks/
```

**Common issues:**
- Missing closing `---` in frontmatter
- Invalid YAML syntax
- Invalid status value (must be: pending, in-progress, completed, blocked)
- Duplicate task IDs

### "Circular dependency detected"

Dependencies form a cycle (A depends on B, B depends on A).

**Find the cycle:**
```bash
taskmd validate tasks/
taskmd graph --format ascii
```

**Fix:**
Remove one dependency to break the cycle.

### Command not found

**Check installation:**
```bash
which taskmd
taskmd --version
```

**If not found:**
```bash
# Verify $PATH includes installation directory
echo $PATH

# Add to PATH (example for Go install)
export PATH=$PATH:$(go env GOPATH)/bin
```

### Web server won't start

**Check port availability:**
```bash
# See if port is in use
lsof -i :8080

# Use different port
taskmd web start --port 3000
```

**Check permissions:**
```bash
# Ensure you have permission to bind to port
# Ports < 1024 require root (not recommended)
```

## Advanced Usage

### Piping and stdin

```bash
# Generate tasks programmatically
echo '---
id: "999"
title: "Test task"
status: pending
---
# Test' | taskmd validate --stdin

# Pipe between commands
taskmd list --format json | jq '.[] | select(.priority == "high")'
```

### Scripting

```bash
#!/bin/bash
# Script: find-quick-wins.sh

# Find small, high-priority pending tasks
taskmd list tasks/ \
  --filter status=pending \
  --filter priority=high \
  --filter effort=small \
  --format json | \
  jq -r '.[] | "\(.id): \(.title)"'
```

### Custom Queries

```bash
# Tasks ready to start (no dependencies blocking)
taskmd next --limit 100 --format json | \
  jq '.[] | select(.score > 50)'

# Blocked tasks with reasons
taskmd list --filter status=blocked --format json | \
  jq -r '.[] | "\(.id): \(.title) - Blocked by: \(.dependencies | join(", "))"'

# Completion rate
TOTAL=$(taskmd stats --format json | jq '.total')
COMPLETED=$(taskmd stats --format json | jq '.completed')
echo "Completion: $(($COMPLETED * 100 / $TOTAL))%"
```

## Getting Help

### Built-in Help

```bash
# General help
taskmd --help

# Command help
taskmd list --help
taskmd graph --help

# List all commands
taskmd --help | grep "Available Commands"
```

### Documentation

- **[Quick Start Guide](quickstart.md)** - Get started fast
- **[Web User Guide](web-guide.md)** - Web interface docs
- **[Task Specification](../taskmd_specification.md)** - Task format reference
- **[CLAUDE.md](../../CLAUDE.md)** - Developer documentation

### Support

- **GitHub Issues**: Report bugs and request features
- **Examples**: Check `tasks/` in the repository for real examples

---

**Next:** Check out the [Web User Guide](web-guide.md) to learn about the visual interface.
