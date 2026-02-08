# Working with Taskmd Tasks

This project uses [taskmd](https://github.com/AaronJuaorth/md-task-tracker) for task management. Tasks are stored as markdown files with YAML frontmatter.

## Task File Format

Each task is a `.md` file with this structure:

```markdown
---
id: "001"
title: "Task title"
status: pending
priority: high
effort: medium
dependencies:
  - "002"
tags:
  - feature
created: 2026-01-15
---

# Task Title

## Objective

What this task accomplishes.

## Tasks

- [ ] Subtask 1
- [ ] Subtask 2

## Acceptance Criteria

- Criterion 1
```

### Frontmatter Fields

| Field | Required | Values |
|-------|----------|--------|
| `id` | Yes | Zero-padded string (`"001"`, `"042"`) |
| `title` | Yes | Brief, action-oriented description |
| `status` | Yes | `pending`, `in-progress`, `completed`, `blocked` |
| `priority` | No | `low`, `medium`, `high`, `critical` |
| `effort` | No | `small`, `medium`, `large` |
| `dependencies` | No | Array of task ID strings |
| `tags` | No | Array of lowercase, hyphen-separated strings |
| `created` | No | `YYYY-MM-DD` date |

### File Naming

Files follow the pattern `NNN-descriptive-title.md` (e.g., `015-add-user-auth.md`).

## Common CLI Commands

```bash
# List all tasks (default: table format)
taskmd list

# List with filters
taskmd list --status pending --priority high
taskmd list --tag feature --format json

# Validate task files for errors
taskmd validate

# Find the next task to work on
taskmd next

# Show project statistics
taskmd stats

# View dependency graph
taskmd graph --format ascii
taskmd graph --exclude-status completed

# Kanban board view
taskmd board

# Scan a specific directory
taskmd list --dir ./tasks
```

## Task Workflow

### Starting a Task

1. Check dependencies are met: `taskmd graph --format ascii`
2. Or use `taskmd next` to find an available task
3. Update status to `in-progress` in the task's frontmatter
4. Check off subtasks (`- [x]`) as you complete them

### Completing a Task

1. Verify all acceptance criteria are met
2. Ensure all subtasks are checked off
3. Update status to `completed`
4. Run `taskmd validate` to confirm no issues

### Task Dependencies

- Dependencies reference tasks by ID: `dependencies: ["001", "015"]`
- A task with unmet dependencies should stay `pending` or `blocked`
- Circular dependencies are invalid -- use `taskmd validate` to detect them

## Status Lifecycle

```
pending --> in-progress --> completed
  |              |
  v              v
blocked <--------
```

- `pending` - Not started
- `in-progress` - Actively being worked on
- `completed` - All acceptance criteria met
- `blocked` - Cannot proceed (explain in task body)

## Directory Organization

Tasks can be organized in subdirectories for grouping:

```
tasks/
  001-spec.md            # Root task, no group
  cli/                   # Group: "cli"
    015-scaffolding.md
    016-parsing.md
  web/                   # Group: "web"
    020-frontend.md
```

The group is inferred from the directory name unless explicitly set in frontmatter.

## Validation

Run `taskmd validate` to check for:
- Missing required fields (`id`, `title`, `status`)
- Invalid enum values
- Duplicate task IDs
- Circular dependencies
- References to non-existent tasks

## Reference

- Full specification: `docs/TASKMD_SPEC.md`
- CLI help: `taskmd --help` or `taskmd <command> --help`
