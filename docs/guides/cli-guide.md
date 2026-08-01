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
# https://github.com/driangle/taskmd/releases

# Extract and add to PATH
tar -xzf taskmd-*.tar.gz
sudo mv taskmd /usr/local/bin/
```

### Option 2: Install with Go

```bash
go install github.com/driangle/taskmd/cmd/taskmd@latest
```

Requirements:
- Go 1.22 or later

### Option 3: Build from Source

```bash
# Clone repository
git clone https://github.com/driangle/taskmd.git
cd taskmd/apps/cli

# Build CLI only
make build

# Build with embedded web interface
make build-full

# Binary will be at bin/taskmd
```

### Option 4: Homebrew (macOS and Linux)

```bash
# Add the tap
brew tap driangle/tap

# Install taskmd
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

- **Required fields**: `id`, `title`
- **Optional fields**: `status`, `priority`, `effort`, `dependencies`, `tags`, `created`
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
| `feed` | Show a chronological activity feed of task changes |
| `archive` | Archive or delete completed/cancelled tasks |
| `rm` | Delete a task file permanently |
| `deduplicate` | Detect and resolve duplicate task IDs |
| `next-id` | Show the next available task ID |
| `add` | Create a new task file with proper frontmatter |
| `search` | Full-text search across task titles and bodies |
| `templates` | List and manage task templates |
| `verify` | Run verification checks for a task |
| `status` | Show in-progress tasks or get metadata for a specific task |
| `context` | Show file context for a task |
| `worklog` | View or add worklog entries for a task |
| `import` | Import tasks from external sources |
| `spec` | Generate the taskmd specification file |
| `commit-msg` | Generate conventional commit messages from task metadata |
| `sync` | Sync tasks from external sources |
| `web` | Web dashboard commands |
| `init` | Initialize a project with agent configuration and spec files |
| `mcp` | Start MCP server for LLM tool integration |
| `todos` | Find TODO/FIXME comments in source code |
| `phases` | List project phases with progress stats |
| `projects` | List and manage registered projects |
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

# Reverse the sort order (newest first)
taskmd list --sort created --reverse

# Reverse works with any sort field; -r is shorthand
taskmd list --sort priority -r
```

**Custom columns:**
```bash
# Show specific columns
taskmd list --columns id,title,status

# Show more columns
taskmd list --columns id,title,status,priority,effort,deps
```

**Limiting results:**
```bash
# Show only 5 tasks
taskmd list --sort priority --limit 5
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--filter` | | Filter tasks (repeatable, AND logic); supports `>=`, `>`, `<=`, `<` for priority and effort |
| `--status` | | Shortcut for `--filter status=<value>` |
| `--priority` | | Shortcut for `--filter priority=<value>` |
| `--phase` | | Filter by phase |
| `--scope` | | Filter by scope; supports wildcards (e.g. `cli`, `cli*`) |
| `--sort` | | Sort by field (`id`, `title`, `status`, `priority`, `effort`, `created_at`) |
| `--reverse` / `-r` | `false` | Reverse the sort order (or the default file-scan order when `--sort` is omitted) |
| `--columns` | `id,title,status,priority,file` | Comma-separated list of columns to display |
| `--limit` | `0` | Maximum number of tasks to display (0 = unlimited) |
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |

**Examples:**
```bash
# High-priority pending tasks
taskmd list --filter status=pending --filter priority=high

# Small tasks (quick wins)
taskmd list --filter effort=small --filter status=pending

# All CLI-related tasks
taskmd list --filter tag=cli --sort priority

# Top 5 by priority
taskmd list --sort priority --limit 5

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `text` | Output format (`text`, `table`, `json`) |
| `--strict` | `false` | Enable strict validation with additional warnings |

**What it checks:**
- Required fields present (id, title)
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
- **Phase proximity**: Tasks in earlier/current phases score higher
- **Phase ordering**: Configured phase order from `.taskmd.yaml`
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

# Show only quick wins (effort: small)
taskmd next --quick-wins

# Show only critical path tasks
taskmd next --critical --limit 1
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--limit` | `5` | Maximum number of recommendations |
| `--filter` | | Filter tasks (repeatable, e.g. `--filter tag=cli`) |
| `--status` | | Shortcut for `--filter status=<value>` |
| `--priority` | | Shortcut for `--filter priority=<value>` |
| `--phase` | | Filter by phase |
| `--scope` | | Filter by scope; supports wildcards (e.g. `cli`, `cli*`) |
| `--exact` | `false` | Disable dependency expansion for `--scope` (only direct matches) |
| `--columns` | `rank,id,title,priority,effort,file,reason` | Comma-separated columns for table output |
| `--quick-wins` | `false` | Show only quick wins (effort: small) |
| `--critical` | `false` | Show only critical path tasks |
| `--strict-phases` | `false` | Enforce strict phase ordering (earlier phases always rank first) |
| `--strict-priority` | `false` | Enforce strict priority ordering (higher priority always ranks first; score breaks ties within a tier). With `--strict-phases`, phase is primary and priority secondary |

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

# Filter by task attributes (AND logic with multiple flags)
taskmd graph --filter priority=high
taskmd graph --filter priority=high --filter effort=small
taskmd graph --filter tag=cli --exclude-status completed
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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `ascii` | Output format (`mermaid`, `dot`, `ascii`, `json`) |
| `--exclude-status` | `completed` | Exclude tasks with status (repeatable) |
| `--all` | `false` | Include all tasks (overrides `--exclude-status`) |
| `--root` | | Start graph from specific task ID |
| `--upstream` | `false` | Show only dependencies (ancestors) |
| `--downstream` | `false` | Show only dependents (descendants) |
| `--focus` | | Highlight specific task ID |
| `--filter` | | Filter tasks (repeatable, AND logic) |
| `--status` | | Shortcut for `--filter status=<value>` |
| `--priority` | | Shortcut for `--filter priority=<value>` |
| `--phase` | | Filter by phase |
| `--scope` | | Filter by scope; supports wildcards (e.g. `cli`, `cli*`) |
| `--out`, `-o` | | Write output to file |

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--group-by` | | Group tasks by field (`phase`) |

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `md` | Output format (`md`, `txt`, `json`) |
| `--group-by` | `status` | Field to group by (`status`, `priority`, `effort`, `type`, `group`, `tag`, `phase`) |
| `--out`, `-o` | | Write output to file |

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `json` | Output format (`json`, `yaml`, `md`) |
| `--core` | `false` | Output only core fields (id, title, dependencies) |
| `--derived` | `false` | Include computed/derived fields (blocked status, depth, topological order) |
| `--group-by` | | Group tasks by field (`status`, `priority`, `effort`, `type`, `group`, `phase`) |
| `--out`, `-o` | | Write output to file |

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

Commands for the taskmd web dashboard.

#### web start

Start the web interface server.

**Basic usage:**
```bash
# Start server
taskmd web start

# Start and open browser
taskmd web start --open

# Custom port
taskmd web start --port 3000

# Read-only mode (disables editing)
taskmd web start --readonly

# Development mode (CORS for Vite)
taskmd web start --dev
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | Server port |
| `--open` | `false` | Open browser automatically |
| `--dev` | `false` | Enable dev mode with CORS |
| `--readonly` | `false` | Start in read-only mode (disables editing) |

**Examples:**
```bash
# Standard usage
taskmd web start --open

# Different port
taskmd web start --port 3000 --open

# Specific tasks directory
taskmd web start --task-dir ./my-tasks --open

# Read-only mode for demos or shared displays
taskmd web start --readonly --open

# Development with Vite
taskmd web start --dev --port 8080
```

#### web export

Export the dashboard as a self-contained static site. The exported site can be deployed to GitHub Pages, Netlify, S3, or any static file host.

```bash
# Export to default directory (./taskmd-export)
taskmd web export

# Export to a specific directory
taskmd web export -o ./public

# Set base path for URLs (e.g., for GitHub Pages subpath)
taskmd web export --base-path /demo/

# Export with custom task directory
taskmd web export --task-dir ./tasks -o ./site
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o`, `--output` | `./taskmd-export` | Output directory |
| `--base-path` | `/` | Base path for URLs (e.g., `/demo/`) |

See [Web User Guide](web-guide.md) for detailed web interface documentation.

### add - Create a New Task

Create a new task markdown file with proper frontmatter. The title is used to generate both the task title and the filename slug. A sequential ID is automatically assigned based on existing tasks.

**Basic usage:**
```bash
# Create a task with just a title
taskmd add "Fix the login bug"

# Set priority and tags
taskmd add "Implement OAuth" --priority high --tags backend,auth

# Create in a subdirectory group
taskmd add "Design mockups" --group design --effort large

# Open in $EDITOR after creation
taskmd add "Quick fix" --edit

# With dependencies
taskmd add "Deploy to staging" --depends-on 041,042

# Create with phase
taskmd add "Implement OAuth" --phase v0.2
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--priority` | `medium` | Task priority (`low`, `medium`, `high`, `critical`) |
| `--effort` | | Task effort (`small`, `medium`, `large`) |
| `--tags` | | Comma-separated tags |
| `--status` | `pending` | Task status (`pending`, `in-progress`, `completed`, `blocked`, `cancelled`) |
| `--owner` | | Task owner/assignee |
| `--depends-on` | | Comma-separated dependency task IDs |
| `--parent` | | Parent task ID |
| `--phase` | | Phase name |
| `--group` | | Subdirectory to create the task in |
| `--slug` | | Custom filename slug (default: auto-generated from title) |
| `--format` | `plain` | Output format (`plain`, `json`) |
| `--edit` | `false` | Open the new task in `$EDITOR` |
| `--template` | | Use a task template (e.g., `bug`, `feature`, `chore`) |

**Examples:**
```bash
# Create a high-priority task with tags
taskmd add "Implement OAuth" --priority high --tags backend,auth

# Create in a specific group
taskmd add "Design mockups" --group design --effort large

# JSON output for scripting
taskmd add "Automated task" --format json

# Open in editor immediately
taskmd add "Draft specification" --edit

# Custom filename slug
taskmd add "Fix the login bug" --slug fix-login

# Create from a template
taskmd add "Login fails on Safari" --template bug
```

### search - Full-Text Search

Perform case-insensitive full-text search across all task titles and markdown body content. Results show where the match was found and a context snippet.

**Basic usage:**
```bash
# Search for a keyword
taskmd search "authentication"

# JSON output
taskmd search deploy --format json

# Filter and sort results
taskmd search "auth" --filter priority=high
taskmd search "deploy" --filter status=pending --sort priority --limit 5

# Search in specific directory
taskmd search "bug fix" --task-dir ./tasks
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--filter` | | Filter tasks (repeatable, AND logic, e.g., `--filter status=pending --filter priority=high`) |
| `--sort` | | Sort by field (`id`, `title`, `status`, `priority`, `effort`, `created`) |
| `--limit` | `0` | Maximum number of results (0 = unlimited) |

### verify - Run Verification Checks

Run the acceptance checks defined in a task's `verify` field. Each verify step has a type:

- **bash** -- runs a shell command, reports pass/fail based on exit code
- **assert** -- displays a check for the agent to evaluate (not executed)

**Basic usage:**
```bash
# Verify a task (stops at first failure)
taskmd verify 042

# Run all checks even if some fail
taskmd verify 042 --all

# JSON output
taskmd verify 042 --format json

# Preview checks without executing
taskmd verify 042 --dry-run

# Custom timeout (seconds) per command
taskmd verify 042 --timeout 120

# --task-id flag also works
taskmd verify --task-id 042
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-id` | | Task ID to verify (alternative to positional argument) |
| `--all` | `false` | Run all checks even if one fails (default: fail-fast) |
| `--format` | `table` | Output format (`table`, `json`) |
| `--dry-run` | `false` | List checks without executing |
| `--timeout` | `60` | Per-command timeout in seconds |

**Exit codes:**
- `0` - All executable checks passed
- `1` - One or more executable checks failed

### status - Show In-Progress Tasks or Task Metadata

Without arguments, shows all in-progress tasks. With a query argument, displays only the frontmatter metadata of a task, without body content, resolved dependency info, context files, or worklog data. Use this when you just need to quickly check a task's status, priority, or other metadata.

If the task has children (other tasks with a matching `parent` field), a recursive children tree is displayed showing each child's ID, status, and title. Grandchildren and deeper descendants are shown with indentation.

Matching uses the same logic as `get` (ID, title, file path, fuzzy).

**Without arguments**, `status` shows all in-progress tasks. If no tasks are in progress, it prints an informational message to stderr. With `--format json` or `--format yaml`, it returns an empty array instead.

```bash
# Show all in-progress tasks
taskmd status

# Compact output for shell statuslines (silent when no tasks)
taskmd status --statusline

# Filter by scope
taskmd status --scope cli
```

**With a query argument**, it displays metadata for a specific task:

```bash
# Look up by task ID
taskmd status 042

# Look up by title
taskmd status "Setup project"

# JSON output
taskmd status 042 --format json

# Strict lookup (no fuzzy matching)
taskmd status sho --exact

# Metadata only, skip children tree
taskmd status 042 --minimal
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `text` | Output format (`text`, `json`, `yaml`) |
| `--exact` | `false` | Disable fuzzy matching, exact only |
| `--threshold` | `0.6` | Fuzzy match sensitivity (0.0-1.0) |
| `--minimal` | `false` | Show only task metadata, skip children |
| `--statusline` | `false` | Compact output for shell statuslines |
| `--scope` | | Filter by group/directory; supports wildcards |

### context - Show File Context

Resolve all relevant files for a task into a structured output. Files come from two sources:

1. **Scope files** -- resolved from the task's `touches` field via scope definitions in `.taskmd.yaml`
2. **Explicit files** -- listed directly in the task's `context` field

**Basic usage:**
```bash
# Show context for a task
taskmd context --task-id 042

# JSON output
taskmd context --task-id 042 --format json

# Include file contents and task body
taskmd context --task-id 042 --include-content --resolve

# Include files from dependency tasks
taskmd context --task-id 042 --include-deps

# Limit number of files
taskmd context --task-id 042 --max-files 20
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-id` | *(required)* | Task ID to build context for |
| `--format` | `text` | Output format (`text`, `json`, `yaml`) |
| `--resolve` | `false` | Expand directory paths to individual files |
| `--include-content` | `false` | Inline file contents and task body |
| `--include-deps` | `false` | Include files from direct dependency tasks |
| `--max-files` | `0` | Cap number of files (0 = unlimited) |

### worklog - View or Add Worklog Entries

View or add timestamped worklog entries for a task. Worklog files are stored at `tasks/<group>/.worklogs/<ID>.md`.

**Basic usage:**
```bash
# View worklog entries
taskmd worklog 015

# Add a new entry
taskmd worklog 015 --add "Started implementation"

# JSON output
taskmd worklog 015 --format json
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--add` | | Append a new worklog entry with the given text |
| `--format` | `text` | Output format (`text`, `json`, `yaml`) |

### import - Import Tasks from External Sources

Fetch tasks from an external source (GitHub Issues, Jira, etc.) and create local markdown task files. This is a one-time onboarding tool for populating your `tasks/` directory.

When run without `--source`, an interactive wizard guides you through setup.

**Basic usage:**
```bash
# Interactive wizard
taskmd import

# GitHub: import from a repository
taskmd import --source github --repo owner/repo

# GitHub: with auth token and filters
taskmd import --source github --project owner/repo --token-env GITHUB_TOKEN

# GitHub: filter by labels and assignee
taskmd import --source github --repo owner/repo --labels bug,critical --assignee alice

# Jira: import from a project
taskmd import --source jira --project PROJ --url https://company.atlassian.net

# Preview without writing files
taskmd import --source github --repo owner/repo --dry-run
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--source` | | Source name (`github`, `jira`, etc.). Omit for interactive wizard |
| `--project` | | Project identifier (owner/repo for GitHub, project key for Jira) |
| `--token-env` | | Environment variable name for auth token |
| `--user-env` | | Environment variable name for username (Jira) |
| `--base-url` | | API base URL (for Jira or GitHub Enterprise) |
| `--output-dir` | `./tasks` | Target directory for imported task files |
| `--filter` | | Source-specific filters as key:value pairs |
| `--dry-run` | `false` | Preview import without writing files |
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |

**GitHub-specific flags:**

| Flag | Description |
|------|-------------|
| `--repo` | Alias for `--project` (owner/repo) |
| `--labels` | Filter by labels (comma-separated) |
| `--milestone` | Filter by milestone |
| `--assignee` | Filter by assignee |

**Jira-specific flags:**

| Flag | Description |
|------|-------------|
| `--url` | Alias for `--base-url` (Jira instance URL) |
| `--jql` | Jira Query Language filter |

### spec - Generate Specification File

Generate the taskmd specification document in your project directory. The specification describes the task file format, including frontmatter fields, valid values, file naming conventions, and directory structure.

**Basic usage:**
```bash
# Write TASKMD_SPEC.md to current directory
taskmd spec

# Print spec to stdout
taskmd spec --stdout

# Write to a specific directory
taskmd spec --dir ./docs

# Overwrite existing file
taskmd spec --force
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--force` | `false` | Overwrite existing `TASKMD_SPEC.md` |
| `--stdout` | `false` | Print spec to stdout instead of writing a file |

### commit-msg - Generate Commit Messages

Generate a conventional commit message derived from task metadata.

When `--task-id` is provided, the message is generated from that task. When no `--task-id` is provided, the command inspects staged changes (`git diff --cached`) to find task files whose status changed to `completed` and generates a message from those tasks automatically.

**Basic usage:**
```bash
# Generate message for a specific task
taskmd commit-msg --task-id 042

# Use a custom commit type
taskmd commit-msg --task-id 042 --type feat

# Include completed subtasks as bullet points
taskmd commit-msg --task-id 042 --body

# Subject line only
taskmd commit-msg --task-id 042 --short

# Auto-detect completed tasks from staged changes
taskmd commit-msg

# Use with git commit
git commit -m "$(taskmd commit-msg --task-id 042)"
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-id` | | Task ID to generate the message for (omit to auto-detect) |
| `--type` | `chore` | Commit type prefix (`feat`, `fix`, `chore`, `docs`, `test`, `refactor`) |
| `--body` | `false` | Include completed subtasks as bullet points |
| `--short` | `false` | Subject line only (no body) |

### todos - Find TODO/FIXME Comments

Scan source code files recursively for marker comments (TODO, FIXME, HACK, XXX, NOTE, BUG, OPTIMIZE) and display them with file path, line number, marker type, and comment text.

Respects `.gitignore` and skips common non-source directories (`node_modules`, `.git`, `vendor`, etc.). Supports language-aware comment parsing for Go, JavaScript, TypeScript, Python, Ruby, Shell, CSS, HTML, Rust, YAML, and TOML.

**Basic usage:**
```bash
# List all TODO/FIXME comments
taskmd todos list

# Scan a specific directory
taskmd todos list --dir ./src

# Filter by marker type
taskmd todos list --marker TODO --marker FIXME

# Include only specific file patterns
taskmd todos list --include "*.go"

# Exclude specific file patterns
taskmd todos list --exclude "*.test.go"

# JSON output
taskmd todos list --format json

# Rich output with scope and git blame info
taskmd todos list --rich
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dir` | `.` | Directory to scan for source code |
| `--marker` | *(all)* | Filter by marker type (repeatable) |
| `--include` | | Include only files matching glob pattern (repeatable) |
| `--exclude` | | Exclude files matching glob pattern (repeatable) |
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--rich` | `false` | Include scope and git blame information (slower) |
| `--raw-text` | `false` | Include original source line text in output |

Exclude patterns can also be configured in `.taskmd.yaml` under `todos.exclude`. CLI `--exclude` flags are additive with config patterns.

### projects - Manage Registered Projects

List and manage globally registered projects. Projects are registered in `~/.taskmd.yaml` under the `projects` key, enabling multi-project workflows with `--project` and `--all-projects` flags.

**Basic usage:**
```bash
# List all registered projects with task stats
taskmd projects

# JSON output
taskmd projects --format json

# Register the current directory as a project
taskmd projects register

# Register with a custom ID and name
taskmd projects register --id my-project --name "My Project"

# Register a specific path
taskmd projects register --path /path/to/project

# Unregister by current directory
taskmd projects unregister

# Unregister by project ID
taskmd projects unregister --id my-project
```

**Flags (projects):**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |

**Flags (projects register):**

| Flag | Default | Description |
|------|---------|-------------|
| `--id` | *(directory basename)* | Project ID |
| `--name` | *(same as ID)* | Display name |
| `--path` | *(current directory)* | Path to register |

**Flags (projects unregister):**

| Flag | Default | Description |
|------|---------|-------------|
| `--id` | *(match by current directory)* | Project ID to remove |

**Multi-project usage:**
```bash
# Run any command against a specific registered project
taskmd list --project my-project

# Aggregate tasks from all registered projects
taskmd list --all-projects
taskmd stats --all-projects
taskmd next --all-projects
```

### completion - Generate Shell Completions

Generate shell completion scripts for bash, zsh, fish, or PowerShell.

```bash
# Bash
source <(taskmd completion bash)

# Zsh
taskmd completion zsh > "${fpath[1]}/_taskmd"

# Fish
taskmd completion fish | source

# PowerShell
taskmd completion powershell | Out-String | Invoke-Expression
```

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
| `--raw-markdown` | `false` | Display raw markdown without formatting |
| `--context` | `false` | Include context files in output |

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

Modify a task's frontmatter fields (status, priority, effort, tags, owner, parent, phase) by ID.

**Basic usage:**
```bash
# Change status
taskmd set 042 --status in-progress

# Change priority and effort
taskmd set 042 --priority high --effort large

# Mark as completed (shortcut)
taskmd set 042 --done

# Set phase
taskmd set 042 --phase v0.2

# Clear phase
taskmd set 042 --phase ""
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `[task-id]` | | Task ID as positional argument |
| `--task-id string` | | Task ID to update (alternative to positional) |
| `--status string` | | New status (`pending`, `in-progress`, `in-review`, `completed`, `blocked`, `cancelled`) |
| `--priority string` | | New priority (`low`, `medium`, `high`, `critical`) |
| `--effort string` | | New effort (`small`, `medium`, `large`) |
| `--owner string` | | Owner/assignee of the task |
| `--parent string` | | Parent task ID (empty string to clear) |
| `--phase string` | | Phase name (empty string to clear) |
| `--done` | `false` | Alias for `--status completed` |
| `--dry-run` | `false` | Preview changes without writing to disk |
| `--add-tag string` | | Add a tag (repeatable) |
| `--remove-tag string` | | Remove a tag (repeatable) |
| `--add-pr string` | | Add a PR URL (repeatable) |
| `--remove-pr string` | | Remove a PR URL (repeatable) |
| `--add-touches string` | | Add a scope identifier to touches (repeatable) |
| `--remove-touches string` | | Remove a scope identifier from touches (repeatable) |
| `--type string` | | Work type (`feature`, `bug`, `improvement`, `chore`, `docs`) |
| `--depends-on string` | | Set dependencies (comma-separated IDs, e.g. `010,015`) |
| `--verify` | `false` | Run verification checks before completing a task |

**Tag management:**
```bash
# Add tags
taskmd set 042 --add-tag backend --add-tag api

# Remove a tag
taskmd set 042 --remove-tag deprecated

# Add and remove in one command
taskmd set 042 --add-tag v2 --remove-tag v1
```

**Scope (touches) management:**
```bash
# Add scopes
taskmd set 042 --add-touches cli/graph --add-touches cli/output

# Remove a scope
taskmd set 042 --remove-touches cli/graph
```

**Examples:**
```bash
# Start working on a task
taskmd set 042 --status in-progress

# Preview changes before applying
taskmd set 042 --priority critical --dry-run

# Set owner and parent
taskmd set 042 --owner alice --parent 040

# Mark done
taskmd set 042 --done
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

### rm - Delete a Task

Permanently delete a task file by ID. Displays the task details and asks for confirmation before deleting.

**Basic usage:**
```bash
# Delete a task (with confirmation prompt)
taskmd rm 042

# Skip confirmation
taskmd rm 042 --force

# Preview what would be deleted
taskmd rm 042 --dry-run
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--force`, `-f` | `false` | Skip confirmation prompt |
| `--dry-run` | `false` | Preview what would be deleted without acting |

### deduplicate - Resolve Duplicate IDs

Detect and resolve duplicate task IDs that can occur when multiple contributors create tasks on separate branches. For each collision, the oldest task (by created date) keeps its original ID. Newer tasks get reassigned a fresh ID, with file renames and cross-reference updates applied automatically.

**Basic usage:**
```bash
# Detect and fix duplicates
taskmd deduplicate

# Scan a specific directory
taskmd deduplicate ./tasks

# Preview changes without modifying files
taskmd deduplicate --dry-run

# Skip interactive prompts for ambiguous references
taskmd deduplicate --no-interactive

# JSON output
taskmd deduplicate --format json
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview changes without modifying files |
| `--format` | `text` | Output format (`text`, `json`) |
| `--no-interactive` | `false` | Skip interactive prompts for ambiguous references |

### templates - Manage Task Templates

List and inspect task templates used by the `add` command. Templates are discovered from three sources in precedence order: project (`.taskmd/templates/`), user (`~/.taskmd/templates/`), and built-in.

**Basic usage:**
```bash
# List available templates
taskmd templates list

# JSON output
taskmd templates list --format json

# YAML output
taskmd templates list --format yaml
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |

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
| `--group-by string` | `status` | Field to group by (`status`, `priority`, `effort`, `type`, `group`, `tag`, `phase`) |
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
| `--scope string` | | Focus on a single scope; supports wildcards (e.g. `cli/graph`, `cli*`) |

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

### phases - List Project Phases

Display configured project phases with summary statistics including task counts, completion rates, and due dates. Phases are defined in `.taskmd.yaml` under the `phases` key.

**Basic usage:**
```bash
# List phases with progress
taskmd phases

# JSON output
taskmd phases --format json
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `table` | Output format (`table`, `json`, `yaml`) |

**Examples:**
```bash
# Show all phases with progress stats
taskmd phases

# Export phase data
taskmd phases --format json > phases.json

# YAML output
taskmd phases --format yaml
```

### feed - Activity Feed

Show a chronological activity feed of recent changes to task files. Uses git log to detect task creation, modification, and renames, presenting them as a time-ordered feed.

Pass a task id to scope the feed to a single task — a timeline of its status transitions, including the initial "created" event. Use `--field` to track a different frontmatter field (e.g. `priority`).

**Basic usage:**
```bash
# Show recent task activity
taskmd feed

# Show changes from the last 7 days
taskmd feed --since 7d

# Status timeline for a single task
taskmd feed cli-049
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format string` | `text` | Output format (`text`, `json`) |
| `--limit int` | `20` | Maximum number of entries to show |
| `--scope string` | | Filter to a tasks subdirectory; supports wildcards (e.g. `cli`, `cli*`) |
| `--since string` | | Show changes since (e.g. `2d`, `1w`, `2026-02-28`) |
| `--source string` | `all` | Filter by event source (`all`, `git`, `worklog`) |
| `--field string` | `status` | In single-task mode, the frontmatter field whose transitions to show |

**Examples:**
```bash
# Show recent task activity
taskmd feed

# Show changes from the last 7 days
taskmd feed --since 7d

# Limit to 10 entries
taskmd feed --limit 10

# Filter to a specific scope
taskmd feed --scope cli

# Export as JSON
taskmd feed --format json

# Status timeline for a single task
taskmd feed cli-049

# Track priority transitions instead of status
taskmd feed cli-049 --field priority
```

### sync - Sync External Sources

Commands for syncing tasks with external sources (GitHub Issues, Jira, etc.). Running `taskmd sync` alone displays usage and available subcommands.

#### sync down

Fetch tasks from configured external sources and create or update local markdown task files. Configuration is read from `.taskmd.yaml`.

**Basic usage:**
```bash
# Sync all configured sources
taskmd sync down

# Preview without writing files
taskmd sync down --dry-run

# Sync a specific source
taskmd sync down --source github
taskmd sync down --source jira
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
taskmd sync down

# Preview what would change
taskmd sync down --dry-run

# Sync only GitHub source
taskmd sync down --source github

# Sync only Jira source
taskmd sync down --source jira

# Overwrite local changes with remote data
taskmd sync down --conflict remote

# Keep local changes, ignore remote updates
taskmd sync down --conflict local
```

See the [Sync Configuration](#sync-configuration) section below for how to set up `.taskmd.yaml` for sync.

### mcp - MCP Server

Start a Model Context Protocol server over stdio for LLM tool integration.

**Basic usage:**
```bash
# Start the MCP server
taskmd mcp
```

The server exposes task operations as MCP tools (`list`, `get`, `next`, `search`, `context`, `set`, `validate`, `graph`) that any MCP-compatible client can discover and call.

See the [MCP Server Guide](mcp-guide.md) for client configuration and full tool reference.

### init - Initialize a Project

Set up a complete taskmd project in the current directory. Creates a task directory, `.taskmd.yaml` config, agent configuration files, the taskmd specification document, and built-in task templates.

When run interactively (in a terminal), prompts for any values not provided via flags. In non-interactive mode, defaults to Claude agent configuration.

**Basic usage:**
```bash
# Interactive setup (prompts for missing info)
taskmd init

# Set task directory, prompt for agents
taskmd init --task-dir ./tasks

# Claude agent config, prompt for task directory
taskmd init --claude

# Fully non-interactive
taskmd init --task-dir ./tasks --claude

# Multiple agents
taskmd init --claude --gemini

# Skip specific outputs
taskmd init --no-spec         # Skip TASKMD_SPEC.md
taskmd init --no-agent        # Skip agent configs
taskmd init --no-templates    # Skip task templates

# Overwrite existing files
taskmd init --force

# Print all content to stdout instead of writing files
taskmd init --stdout
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-dir` | `./tasks` | Task directory path to create |
| `--claude` | `false` | Initialize for Claude Code |
| `--gemini` | `false` | Initialize for Gemini |
| `--codex` | `false` | Initialize for Codex |
| `--no-spec` | `false` | Skip generating TASKMD_SPEC.md |
| `--no-agent` | `false` | Skip generating agent configuration files |
| `--no-templates` | `false` | Skip copying built-in task templates |
| `--id-strategy` | | ID generation strategy (`sequential`, `prefixed`, `random`, `ulid`) |
| `--id-prefix` | | Prefix for prefixed ID strategy |
| `--force` | `false` | Overwrite existing files |
| `--stdout` | `false` | Print all content to stdout instead of writing files |

If a file already exists and `--force` is not set, it is skipped with a warning.

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
taskmd set 042 --status in-progress

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
id:
  strategy: sequential          # "sequential", "prefixed", "random", or "ulid"
  prefix: ""                   # Required when strategy is "prefixed"
  padding: 3                   # Zero-padding width (sequential strategy)
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
taskmd list --task-dir ./other-tasks  # Overrides config dir
taskmd web start --port 8080     # Overrides config port
```

See [docs/.taskmd.yaml.example](../.taskmd.yaml.example) for a complete example with comments.

### Sync Configuration

The `sync` command reads its configuration from the `sync` section of `.taskmd.yaml`. Each source defines where to fetch tasks from, how to map fields, and where to write files.

**Example `.taskmd.yaml` with GitHub source:**

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

**Example `.taskmd.yaml` with Jira source:**

```yaml
# .taskmd.yaml
dir: ./tasks

sync:
  sources:
    - name: jira
      project: "PROJ"                        # Jira project key
      base_url: https://myteam.atlassian.net  # Jira Cloud instance URL (required)
      token_env: JIRA_API_TOKEN               # Jira API token
      user_env: JIRA_USER_EMAIL               # Jira account email (for Basic auth)
      output_dir: ./tasks/jira
      field_map:
        status:
          To Do: pending
          In Progress: in-progress
          Done: completed
        priority:
          Highest: critical
          High: high
          Medium: medium
          Low: low
          Lowest: low
        labels_to_tags: true
        assignee_to_owner: true
      filters:
        jql: 'status != "Done"'              # Additional JQL (ANDed with project)
```

Jira uses Basic authentication (email + API token). Both `token_env` and `user_env` are required. The `base_url` must point to your Jira Cloud instance. Descriptions in Jira's ADF (Atlassian Document Format) are automatically converted to Markdown.

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
alias tm='taskmd --task-dir ./tasks'
alias tmw='taskmd web start --port 8080 --open'
```

**2. Environment Variables:**
```bash
export TASKMD_DIR=./tasks
```

### Command-Line Flags

Global flags (available for all commands):

```bash
--config string       # Config file path
-d, --task-dir string # Task directory to scan (default ".")
--format string       # Output format (table, json, yaml)
--verbose             # Verbose logging
--quiet               # Suppress non-essential output
--stdin               # Read from stdin instead of files
--debug               # Enable debug output (prints to stderr)
--no-color            # Disable colored output
--project string      # Operate on a registered project by ID
--all-projects        # Aggregate tasks from all registered projects
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
4. Required fields present: `id`, `title`

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
