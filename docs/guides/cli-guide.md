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

### tui - Interactive Terminal UI

Launch an interactive terminal interface for browsing tasks.

**Basic usage:**
```bash
# Launch TUI
taskmd tui

# Specific directory
taskmd tui ./tasks
```

**Keyboard shortcuts:**

- `↑/k` - Move up
- `↓/j` - Move down
- `Enter` - View task details
- `Tab` - Switch between views
- `/` - Start search
- `Esc` - Clear search / Go back
- `q` - Quit

**Starting options:**
```bash
# Start with filter
taskmd tui --filter status=pending

# Start with specific task
taskmd tui --focus 022

# Group by field
taskmd tui --group-by priority

# Read-only mode
taskmd tui --readonly
```

**Examples:**
```bash
# Browse pending tasks
taskmd tui --filter status=pending

# Review high-priority items
taskmd tui --filter priority=high

# Start at specific task
taskmd tui --focus 025

# Grouped by status
taskmd tui --group-by status
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
# The TUI can help navigate
taskmd tui --filter status=in-progress

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
