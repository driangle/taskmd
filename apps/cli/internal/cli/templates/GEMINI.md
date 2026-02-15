# Working with Taskmd Tasks

This project uses [taskmd](https://github.com/driangle/taskmd) for task management. Tasks are stored as markdown files with YAML frontmatter.

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
4. Add a worklog entry noting your approach and initial findings
5. Check off subtasks (`- [x]`) as you complete them

### Completing a Task

1. Verify all acceptance criteria are met
2. Ensure all subtasks are checked off
3. Add a final worklog entry summarizing what was done
4. Update status to `completed`
5. Run `taskmd validate` to confirm no issues

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

## Task Worklogs

Worklogs are enabled by default. To disable them, set `worklogs: false` in `.taskmd.yaml`. When worklogs are disabled, skip all worklog steps in the task workflow.

Each task can have a companion **worklog file** that records progress notes, decisions, and blockers. Worklogs live in a `.worklogs/` directory alongside the task files:

```
tasks/
  cli/
    015-add-user-auth.md
    .worklogs/
      015.md                 # Worklog for task 015
  web/
    020-frontend.md
    .worklogs/
      020.md                 # Worklog for task 020
```

### When to Write Worklog Entries

Write a worklog entry when you:
- **Start working** on a task (note your approach and initial findings)
- **Make a key decision** (record what you chose and why)
- **Hit a blocker** (describe the issue so the next session can pick up)
- **Complete a significant subtask** (note what was done and what remains)
- **Finish a session** (summarize progress, open questions, and next steps)

### Worklog Format

Each entry starts with a timestamp header. Entries are appended chronologically:

```markdown
## 2026-02-15T10:30:00Z

Started implementation of JWT authentication middleware.

**Approach:** Using `golang-jwt/jwt/v5` library. Chose HMAC-SHA256
signing since we don't need asymmetric keys for this use case.

**Completed:**
- [x] Added JWT signing utility
- [x] Created auth middleware

**Next:** Wire up login endpoint and add tests.

## 2026-02-15T14:15:00Z

Login endpoint is working. Ran into an issue with token expiry
validation -- the default clock skew tolerance was too strict for
CI environments. Set `leeway` to 10 seconds.

**Completed:**
- [x] Login endpoint with token generation
- [x] Fixed clock skew issue in token validation

**Blocked:** Need the user model changes from task 012 to land
before I can implement the password verification step.
```

### Good Worklog Practices

- **Be specific** -- "Fixed auth bug" is unhelpful; "Fixed token validation failing when clock skew exceeds 5s (set leeway to 10s)" tells the next reader exactly what happened
- **Record decisions with reasoning** -- Future readers (including yourself) will want to know *why*, not just *what*
- **Note blockers clearly** -- Call out what's blocked and which task/issue is the dependency
- **Keep entries concise** -- A worklog is a trail of breadcrumbs, not a novel

### CLI Commands

```bash
# View a task's worklog
taskmd worklog --task-id 015

# Append a new entry
taskmd worklog --task-id 015 --add "Completed login endpoint. Blocked on task 012 for user model."
```

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
