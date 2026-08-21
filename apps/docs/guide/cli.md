# CLI Guide

Complete reference for using taskmd from the command line.

## Command Reference

### Quick Reference

| Command | Description |
|---------|-------------|
| [`list`](#list-view-and-filter-tasks) | List tasks in a quick textual format |
| [`get`](#get-view-task-details) | Get detailed information about a specific task |
| [`set`](#set-update-task-fields) | Set a task's frontmatter fields |
| [`next`](#next-find-what-to-work-on) | Recommend what task to work on next |
| [`validate`](#validate-check-task-files) | Lint and validate tasks |
| [`graph`](#graph-visualize-dependencies) | Export task dependency graph |
| [`board`](#board-kanban-view) | Display tasks grouped in a kanban-like board view |
| [`stats`](#stats-project-metrics) | Show computed metrics about tasks |
| [`tags`](#tags-list-tags) | List all tags with task counts |
| [`snapshot`](#snapshot-machine-readable-export) | Produce a frozen, machine-readable representation of tasks |
| [`report`](#report-generate-reports) | Generate a comprehensive project report |
| [`tracks`](#tracks-parallel-work-tracks) | Show parallel work tracks based on scope overlap |
| [`feed`](#feed-activity-feed) | Show a chronological activity feed of task changes |
| [`archive`](#archive-archive-completed-tasks) | Archive or delete completed/cancelled tasks |
| [`rm`](#rm-delete-a-task) | Delete a task file permanently |
| [`deduplicate`](#deduplicate-resolve-duplicate-ids) | Detect and resolve duplicate task IDs |
| [`next-id`](#next-id-get-next-available-id) | Show the next available task ID |
| [`add`](#add-create-a-new-task) | Create a new task file with proper frontmatter |
| [`search`](#search-full-text-search) | Full-text search across task titles and bodies |
| [`templates`](#templates-manage-task-templates) | List and manage task templates |
| [`verify`](#verify-run-verification-checks) | Run verification checks for a task |
| [`status`](#status-show-in-progress-tasks-or-task-metadata) | Show in-progress tasks or get metadata for a specific task |
| [`context`](#context-show-file-context) | Show file context for a task |
| [`worklog`](#worklog-view-or-add-worklog-entries) | View or add worklog entries for a task |
| [`import`](#import-import-tasks-from-external-sources) | Import tasks from external sources |
| [`spec`](#spec-generate-specification-file) | Generate the taskmd specification file |
| [`sync`](#sync-sync-external-sources) | Sync tasks from external sources |
| [`web`](#web-web-dashboard) | Web dashboard commands |
| [`init`](#init-initialize-a-project) | Initialize a project with agent configuration and spec files |
| [`commit-msg`](#commit-msg-generate-commit-messages) | Generate conventional commit messages from task metadata |
| [`mcp`](#mcp-start-mcp-server) | Start MCP server over stdio |
| [`todos`](#todos-find-todo-fixme-comments) | Find TODO/FIXME comments in source code |
| [`phases`](#phases-list-project-phases) | List project phases with progress stats |
| [`projects`](#projects-manage-registered-projects) | List and manage registered projects |
| [`completion`](#completion-generate-shell-completions) | Generate shell completion scripts |

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

# Filter by priority with comparison (>=, >, <=, <)
taskmd list --filter "priority>=medium"
taskmd list --filter "priority>low"

# Filter by multiple criteria (AND logic)
taskmd list --filter status=pending --filter priority=high

# Filter by tag
taskmd list --filter tag=cli

# Filter by effort
taskmd list --filter effort=small

# Filter by effort with comparison
taskmd list --filter "effort>=medium"
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
| `--filter` | | Filter tasks (repeatable, AND logic); `priority` and `effort` support `>=`, `>`, `<=`, `<` (effort values come from the [configured vocabulary](/reference/configuration#effort-configuration)) |
| `--phase` | | Filter tasks by phase name |
| `--scope` | | Filter by scope; supports wildcards (e.g. `cli`, `cli*`) |
| `--status` | | Shortcut for `--filter status=<value>` |
| `--priority` | | Shortcut for `--filter priority=<value>` |
| `--sort` | | Sort by field (`id`, `title`, `status`, `priority`, `effort`, `created`) |
| `--reverse` / `-r` | `false` | Reverse the sort order (or the default file-scan order when `--sort` is omitted) |
| `--columns` | `id,title,status,priority,file` | Comma-separated list of columns to display |
| `--limit` | `0` | Maximum number of tasks to display (0 = unlimited) |
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |

**Examples:**
```bash
# High-priority pending tasks
taskmd list --filter status=pending --filter priority=high

# Medium priority and above
taskmd list --filter status=pending --filter "priority>=medium"

# Small tasks (quick wins)
taskmd list --filter effort=small --filter status=pending

# All CLI-related tasks
taskmd list --filter tag=cli --sort priority

# Top 5 by priority
taskmd list --sort priority --limit 5

# Export to JSON for scripting
taskmd list --format json > tasks.json

# Filter by phase
taskmd list --phase v0.2
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

### next - Find What to Work On

Analyze tasks and recommend the best ones to work on next.

taskmd scores tasks based on:
- **Priority**: High priority scores higher
- **Critical path**: Tasks on the critical path score higher
- **Downstream impact**: Tasks blocking many others score higher
- **Effort**: Smaller tasks get a boost (quick wins); points scale with the value's position in the [configured effort vocabulary](/reference/configuration#effort-configuration)
- **Phase proximity**: Tasks in phases with nearer due dates score higher
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

# Show only quick wins (tasks at the lowest configured effort)
taskmd next --quick-wins

# Show only critical path tasks
taskmd next --critical --limit 1

# Next task for a specific phase
taskmd next --phase v0.2

# Next actionable task within the graph reachable from a task
taskmd next --root 022

# Explain how each recommendation was scored
taskmd next --explain

# JSON for automation
taskmd next --format json
```

**Explaining scores with `--explain`:**

`--explain` prints, beneath each recommendation, an itemized breakdown of every
scoring component — priority, phase, critical path, downstream impact, and
effort — with each component's point value and a total equal to the task's
score. Scaled bonuses (critical path and downstream) show their base and
multiplier, so it's clear how the downstream-priority multiplier reduces a bonus
when the downstream chain is low priority:

```
1. 003  Build CLI parser
     priority: critical              +40
     critical path                    +7  (15 × 0.50)
     downstream (unblocks 1 task)     +1  (3 × 0.50)
     ───────────────────────────────────
     total                            48
```

The structured `score_breakdown` array is **always** included in `json` and
`yaml` output (independent of `--explain`), and its component points always sum
to `score`.

**Scoping to a task's graph with `--root`:**

`--root <ID>` limits recommendations to the tasks _reachable from_ `<ID>`:

- **Upstream dependencies** — `<ID>`'s transitive prerequisites (the work that
  unblocks it), followed via `dependencies`.
- **Subtasks** — `<ID>`'s transitive subtree, followed via the `parent`
  hierarchy.
- **`<ID>` itself**, when it is actionable.

This answers "given task X, what should I work on next to make progress toward
it?". Because a parent task is never actionable while it has open children,
`--root` on a parent naturally returns that parent's next actionable subtask.
`--root` only narrows the candidate set, so scoring and ranking are unchanged
and it composes with the other filters.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--limit` | `5` | Maximum number of recommendations |
| `--filter` | | Filter tasks (repeatable, e.g. `--filter tag=cli`) |
| `--phase` | | Filter recommendations by phase name |
| `--scope` | | Filter by scope; supports wildcards (e.g. `cli`, `cli*`) |
| `--exact` | `false` | Disable dependency expansion for `--scope` (only direct matches) |
| `--root` | | Limit recommendations to tasks reachable from an ID (its upstream deps + subtasks) |
| `--status` | | Shortcut for `--filter status=<value>` |
| `--priority` | | Shortcut for `--filter priority=<value>` |
| `--columns` | `rank,id,title,priority,effort,file,reason` | Comma-separated columns for table output |
| `--explain` | `false` | Show an itemized score breakdown beneath each recommendation (table format). `score_breakdown` is always present in json/yaml |
| `--strict-phases` | `false` | Enforce strict phase ordering (earlier phases always rank first) |
| `--strict-priority` | `false` | Enforce strict priority ordering (higher priority always ranks first; score breaks ties within a tier). With `--strict-phases`, phase is primary and priority secondary |
| `--quick-wins` | `false` | Show only quick wins (tasks at the lowest configured effort) |
| `--critical` | `false` | Show only critical path tasks |


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

**Filter and highlight:**
```bash
# Filter by task attributes
taskmd graph --filter priority=high
taskmd graph --filter tag=cli --exclude-status completed

# Highlight a specific task
taskmd graph --focus 022 --format mermaid
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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |
| `--group-by` | | Group stats by field (e.g., `phase`) |

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

# Group by phase
taskmd board --group-by phase

# Output formats
taskmd board --format md    # Markdown (default)
taskmd board --format txt   # Plain text
taskmd board --format json  # JSON

# Output to file
taskmd board --out board.md
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `md` | Output format (`md`, `txt`, `json`) |
| `--group-by` | `status` | Field to group by (`status`, `priority`, `effort`, `type`, `group`, `tag`, `phase`) |
| `--out`, `-o` | | Write output to file |

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `json` | Output format (`json`, `yaml`, `md`) |
| `--core` | `false` | Output only core fields (id, title, dependencies) |
| `--derived` | `false` | Include computed/derived fields (blocked status, depth, topological order) |
| `--group-by` | | Group tasks by field (`status`, `priority`, `effort`, `type`, `group`, `phase`) |
| `--out`, `-o` | | Write output to file |

### get - View Task Details

> **Alias:** `show` is a deprecated alias for `get`. Use `get` instead.

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
| `--raw-markdown` | `false` | Display raw markdown without formatting |
| `--context` | `false` | Include context files in output |

### set - Update Task Fields

> **Alias:** `update` is a deprecated alias for `set`. Use `set` instead.

Modify a task's frontmatter fields by ID.

```bash
# Change status
taskmd set 042 --status in-progress

# Change priority and effort
taskmd set 042 --priority high --effort large

# Mark as completed (shortcut)
taskmd set 042 --done

# Preview changes without writing
taskmd set 042 --priority critical --dry-run

# Set phase
taskmd set 042 --phase v0.2

# Clear phase
taskmd set 042 --phase ""

# Rename a task
taskmd set 042 --title "A better name"
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `[task-id]` | | Task ID as positional argument |
| `--task-id` | | Task ID to update (alternative to positional) |
| `--title` | | New title (also updates the body's `#` heading) |
| `--rename` | `false` | With `--title`, rename the file to `<id>-<new-slug>.md` |
| `--status` | | New status (`pending`, `in-progress`, `in-review`, `completed`, `blocked`, `cancelled`) |
| `--priority` | | New priority (`low`, `medium`, `high`, `critical`) |
| `--effort` | | New effort (`small`, `medium`, `large`) |
| `--owner` | | Owner/assignee |
| `--parent` | | Parent task ID (empty string to clear) |
| `--phase` | | Phase name (empty string to clear) |
| `--done` | `false` | Alias for `--status completed` |
| `--dry-run` | `false` | Preview changes without writing to disk |
| `--add-tag` | | Add a tag (repeatable) |
| `--remove-tag` | | Remove a tag (repeatable) |
| `--add-pr` | | Add a PR URL (repeatable) |
| `--remove-pr` | | Remove a PR URL (repeatable) |
| `--add-touches` | | Add a scope identifier to touches (repeatable) |
| `--remove-touches` | | Remove a scope identifier from touches (repeatable) |
| `--type` | | Work type (`feature`, `bug`, `improvement`, `chore`, `docs`) |
| `--depends-on` | | Set dependencies (comma-separated IDs, e.g. `010,015`) |
| `--verify` | `false` | Run verification checks before completing a task |

**Renaming a task:**

`--title` updates the `title` frontmatter field and the matching `#` heading in the
task body. If the body heading has been customized so that it no longer matches the
old title, it is left alone and the command says so.

```bash
# Update the title and body heading; the file stays where it is
taskmd set 042 --title "A better name"

# Also rename the file to 042-a-better-name.md
taskmd set 042 --title "A better name" --rename
```

Renaming the file is opt-in because the old path may be referenced by git history,
links, or bookmarks. Tasks are always resolved by ID, so a stale filename slug is
harmless. `--rename` refuses to overwrite an existing file, and `--rename` without
`--title` is an error.

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

### rm - Delete a Task

Permanently delete a task file by ID. Displays the task details and asks for confirmation before deleting.

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

Detect and resolve duplicate task IDs that can occur when multiple contributors create tasks on separate branches.

For each collision, the oldest task (by created date) keeps its original ID. Newer tasks get reassigned a fresh ID, with file renames and cross-reference updates applied automatically.

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `plain` | Output format (`plain`, `json`) |

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

### commit-msg - Generate Commit Messages

Generate a conventional commit message derived from task metadata.

When `--task-id` is provided, the message is generated from that task. When no `--task-id` is provided, the command inspects staged changes (`git diff --cached`) to find task files whose status changed to `completed` and generates a message from those tasks automatically.

The subject line format is `type(scope): lowercase title (task ID)`, where the scope is the task's group directory (if any).

```bash
# Generate message for a specific task
taskmd commit-msg --task-id 042

# Use a custom commit type
taskmd commit-msg --task-id 042 --type feat

# Include completed subtasks as bullet points in the body
taskmd commit-msg --task-id 042 --body

# Subject line only (no body)
taskmd commit-msg --task-id 042 --short

# Auto-detect completed tasks from staged changes
taskmd commit-msg

# Use with git commit
git commit -m "$(taskmd commit-msg --task-id 042)"
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--task-id` | | Task ID to generate the message for (omit to auto-detect from staged changes) |
| `--type` | `chore` | Commit type prefix (`feat`, `fix`, `chore`, `docs`, `test`, `refactor`) |
| `--body` | `false` | Include completed subtasks (`- [x]`) as bullet points in the commit body |
| `--short` | `false` | Output the subject line only (no body) |

**Auto-detection:**

When `--task-id` is omitted, the command runs `git diff --cached` and looks for task files where `+status: completed` appears in the diff. It then generates a commit message from all matched tasks. If multiple tasks are found, the subject line lists all task IDs (e.g., `chore: complete tasks 042, 043`).

### add - Create a New Task

Create a new task markdown file with proper frontmatter. The title is used to generate both the task title and the filename slug. A sequential ID is automatically assigned based on existing tasks.

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

# Custom filename slug
taskmd add "Fix the login bug" --slug fix-login

# JSON output for scripting
taskmd add "Automated task" --format json
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

**Templates:**
```bash
# Create from a template
taskmd add "Login fails on Safari" --template bug
taskmd add "Dark mode support" --template feature --priority high
```

### templates - Manage Task Templates

List and inspect task templates used by the `add` command. Templates are discovered from three sources in precedence order: project (`.taskmd/templates/`), user (`~/.taskmd/templates/`), and built-in.

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

### search - Full-Text Search

Perform case-insensitive full-text search across all task titles and markdown body content. Results show where the match was found and a context snippet.

```bash
# Search for a keyword
taskmd search "authentication"

# JSON output
taskmd search deploy --format json

# Filter and sort results
taskmd search "auth" --filter priority=high
taskmd search "deploy" --filter status=pending --sort priority --limit 5

# YAML output
taskmd search "bug fix" --format yaml
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

Without arguments, shows all in-progress tasks. With a query argument, displays the frontmatter metadata of a specific task (without body content, resolved dependency info, context files, or worklog data).

If the task has children (other tasks with a matching `parent` field), a recursive children tree is displayed showing each child's ID, status, and title. Grandchildren and deeper descendants are shown with indentation.

Matching uses the same logic as `get` (ID, title, file path, fuzzy).

```bash
# Show all in-progress tasks
taskmd status

# Compact output for shell statuslines
taskmd status --statusline

# Filter by scope
taskmd status --scope cli

# Look up by task ID
taskmd status 042

# Look up by title
taskmd status "Setup project"

# JSON output
taskmd status 042 --format json

# YAML output
taskmd status 042 --format yaml

# Strict lookup (no fuzzy matching)
taskmd status sho --exact

# Metadata only, skip children tree
taskmd status 042 --minimal
```

**Example output (parent task):**

```
Task: 173
Title: Build e2e test suite for CLI
Status: completed
Children:
  ├─ 174 [completed] Set up e2e test foundation
  ├─ 175 [completed] E2e tests for workflows
  └─ 176 [completed] E2e tests for error handling
File: cli/173-e2e-test-suite.md
```

**Statusline examples:**

The `--statusline` flag outputs a compact format suitable for embedding in Claude Code's statusline. If multiple tasks are in progress, the first is shown with `(+N more)`.

Use `$(taskmd status --statusline)` anywhere you want to display the active task:

::: code-group

```bash [Claude Code statusline]
# In your ~/.claude/statusline-command.sh:
current_task=$(taskmd status --statusline 2>/dev/null)
if [ -n "$current_task" ]; then
  line="${line} | ${current_task}"
fi
```

```bash [tmux status-right]
# In your ~/.tmux.conf:
set -g status-right '#(taskmd status --statusline 2>/dev/null)'
```

```bash [Shell prompt (zsh)]
# In your ~/.zshrc:
RPROMPT='$(taskmd status --statusline 2>/dev/null)'
```

```bash [Starship custom module]
# In your ~/.config/starship.toml:
[custom.task]
command = "taskmd status --statusline"
when = "taskmd status --statusline"
format = "[$output]($style) "
style = "dimmed yellow"
```

:::

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `text` | Output format (`text`, `json`, `yaml`) |
| `--exact` | `false` | Disable fuzzy matching, exact only |
| `--threshold` | `0.6` | Fuzzy match sensitivity (0.0-1.0) |
| `--minimal` | `false` | Show only task metadata, skip children |
| `--statusline` | `false` | Compact output for Claude Code statusline (no-args mode) |
| `--scope` | | Filter by group/directory (no-args mode) |

### context - Show File Context

Resolve all relevant files for a task into a structured output. Files come from two sources:

1. **Scope files** -- resolved from the task's `touches` field via scope definitions in `.taskmd.yaml`
2. **Explicit files** -- listed directly in the task's `context` field

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

```bash
# View worklog entries
taskmd worklog 015

# Add a new entry
taskmd worklog 015 --add "Started implementation"

# JSON output
taskmd worklog 015 --format json

# YAML output
taskmd worklog 015 --format yaml
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--add` | | Append a new worklog entry with the given text |
| `--format` | `text` | Output format (`text`, `json`, `yaml`) |

`--add` respects the [`worklogs`](/reference/configuration#configuration-options) config key. Worklogs
are opt-in: unless `worklogs: true` is set in `.taskmd.yaml`, the command refuses to write and
explains why. Viewing is never gated.

### import - Import Tasks from External Sources

Fetch tasks from an external source (GitHub Issues, Jira, etc.) and create local markdown task files. This is a one-time onboarding tool for populating your `tasks/` directory.

When run without `--source`, an interactive wizard guides you through setup.

```bash
# Interactive wizard
taskmd import

# GitHub: import from a repository
taskmd import --source github --repo owner/repo

# GitHub: with auth token and filters
taskmd import --source github --project owner/repo --token-env GITHUB_TOKEN

# GitHub: filter by labels and assignee
taskmd import --source github --repo owner/repo --labels bug,critical --assignee alice

# GitHub: filter by milestone
taskmd import --source github --repo owner/repo --milestone "v2.0"

# Jira: import from a project
taskmd import --source jira --project PROJ --url https://company.atlassian.net

# Jira: with JQL filter
taskmd import --source jira --project PROJ --url https://company.atlassian.net --jql "assignee = currentUser()"

# Preview without writing files
taskmd import --source github --repo owner/repo --dry-run

# JSON output for scripting
taskmd import --source github --repo owner/repo --format json
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
| `--filter` | | Source-specific filters as key:value pairs (e.g. `"state:open labels:bug"`) |
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

### mcp - Start MCP Server

Start a Model Context Protocol (MCP) server that communicates over stdin/stdout. This allows LLM-based tools (Cursor, Windsurf, Copilot agents, Claude Code, etc.) to interact with your taskmd project using the standard MCP protocol.

```bash
# Start MCP server
taskmd mcp
```

**Configuration example for Claude Code (`.mcp.json`):**
```json
{
  "mcpServers": {
    "taskmd": {
      "command": "taskmd",
      "args": ["mcp"]
    }
  }
}
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
| `--group-by` | `status` | Field to group by (`status`, `priority`, `effort`, `type`, `group`, `tag`, `phase`) |
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
| `--scope` | | Focus on a single scope; supports wildcards (e.g. `cli/graph`, `cli*`) |

### phases - List Project Phases

Display configured project phases with summary statistics including task counts, completion rates, and due dates. Phases are defined in `.taskmd.yaml` under the `phases` key.

Tasks that aren't assigned to any phase are aggregated into a synthetic `(unassigned)` category (id: `unassigned`), listed after the configured phases. This row appears only when unassigned tasks exist, letting you see how much work hasn't been scheduled into a phase yet. It's distinct from the warning about tasks referencing an undefined phase id.

```bash
# List phases with progress
taskmd phases

# JSON output
taskmd phases --format json

# YAML output
taskmd phases --format yaml
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `table` | Output format (`table`, `json`, `yaml`) |

### feed - Activity Feed

Show a chronological activity feed of recent changes to task files. Uses git log to detect task creation, modification, and renames, presenting them as a time-ordered feed.

Pass a task id to scope the feed to a single task — a timeline of its status transitions, including the initial "created" event. Use `--field` to track a different frontmatter field (e.g. `priority`).

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

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `text` | Output format (`text`, `json`) |
| `--limit` | `20` | Maximum number of entries to show |
| `--scope` | | Filter to a tasks subdirectory; supports wildcards (e.g. `cli`, `cli*`) |
| `--since` | | Show changes since (e.g. `2d`, `1w`, `2026-02-28`) |
| `--source` | `all` | Filter by event source (`all`, `git`, `worklog`) |
| `--field` | `status` | In single-task mode, the frontmatter field whose transitions to show |

### sync - Sync External Sources

Commands for syncing tasks with external sources (GitHub Issues, Jira, etc.). Running `taskmd sync` alone displays usage and available subcommands.

#### sync down

Fetch tasks from configured external sources and create or update local markdown task files. Configuration is read from `.taskmd.yaml`.

```bash
# Sync all configured sources
taskmd sync down

# Preview without writing files
taskmd sync down --dry-run

# Sync a specific source
taskmd sync down --source github
taskmd sync down --source jira

# Overwrite local changes with remote data
taskmd sync down --conflict remote
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview changes without writing files |
| `--source` | | Sync only the named source |
| `--conflict` | `skip` | Conflict strategy: `skip`, `remote`, or `local` |

See [Configuration](/reference/configuration#sync-configuration) for how to set up sync sources in `.taskmd.yaml`.

### web - Web Dashboard

Commands for the taskmd web dashboard.

#### web start

Start the web interface server.

```bash
# Start server
taskmd web start

# Start and open browser
taskmd web start --open

# Custom port
taskmd web start --port 3000

# Read-only mode (disables editing)
taskmd web start --readonly

# Specific tasks directory
taskmd web start --task-dir ./my-tasks --open
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | Server port |
| `--open` | `false` | Open browser on start |
| `--dev` | `false` | Enable dev mode (CORS for Vite dev server) |
| `--readonly` | `false` | Start in read-only mode (disables editing) |

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

See the [Web Interface Guide](./web) for detailed web UI documentation.

### todos - Find TODO/FIXME Comments

Scan source code files recursively for marker comments (TODO, FIXME, HACK, XXX, NOTE, BUG, OPTIMIZE) and display them with file path, line number, marker type, and comment text.

Respects `.gitignore` and skips common non-source directories (`node_modules`, `.git`, `vendor`, etc.). Supports language-aware comment parsing for Go, JavaScript, TypeScript, Python, Ruby, Shell, CSS, HTML, Rust, YAML, and TOML.

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

### init - Initialize a Project

Set up a complete taskmd project in the current directory. Creates a task directory, `.taskmd.yaml` config, agent configuration files, the taskmd specification document, and built-in task templates.

When run interactively (in a terminal), prompts for any values not provided via flags. In non-interactive mode, defaults to Claude agent configuration.

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

### projects - Manage Registered Projects

List and manage globally registered projects. Projects are registered in `~/.taskmd.yaml` under the `projects` key, enabling multi-project workflows with `--project` and `--all-projects` flags.

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

Generate shell completion scripts for taskmd. Supports Bash, Zsh, Fish, and PowerShell.

```bash
# Bash
source <(taskmd completion bash)

# Bash (persistent - Linux)
taskmd completion bash > /etc/bash_completion.d/taskmd

# Bash (persistent - macOS with Homebrew)
taskmd completion bash > $(brew --prefix)/etc/bash_completion.d/taskmd

# Zsh (persistent)
taskmd completion zsh > "${fpath[1]}/_taskmd"

# Fish
taskmd completion fish | source

# Fish (persistent)
taskmd completion fish > ~/.config/fish/completions/taskmd.fish

# PowerShell
taskmd completion powershell | Out-String | Invoke-Expression
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `bash\|zsh\|fish\|powershell` | Yes | Shell type to generate completions for |

::: tip
After installing completions, start a new shell session for them to take effect. For Zsh, ensure `compinit` is loaded: `echo "autoload -U compinit; compinit" >> ~/.zshrc`
:::

## Global Flags

Available for all commands:

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
- Verify status is valid if present: `pending`, `in-progress`, `completed`, `blocked`, `cancelled`
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
