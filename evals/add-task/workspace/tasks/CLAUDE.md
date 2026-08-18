# Working with Taskmd Tasks

This project uses [taskmd](https://github.com/driangle/taskmd) for task management. Tasks are markdown files with YAML frontmatter. See `TASKMD_SPEC.md` for the full field reference.

## Task File Format

```markdown
---
id: "001"
title: "Task title"
status: pending
priority: high
effort: medium
dependencies: ["002"]
tags: [feature]
created_at: 2026-01-15
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

Files follow the pattern `NNN-descriptive-title.md` (e.g., `015-add-user-auth.md`).

## CLI Commands

```bash
taskmd list                              # List all tasks
taskmd list --status pending --priority high  # Filter tasks
taskmd next                              # Find next available task
taskmd validate                          # Check for errors
taskmd graph --format ascii              # View dependency graph
taskmd board                             # Kanban board view
taskmd stats                             # Project statistics
taskmd set <id> --status in-progress     # Update task status
taskmd add "Task title"                  # Create a new task
taskmd verify <id>                       # Run acceptance checks
```

## Task Workflow

### Starting a Task

1. Run `taskmd next` or check dependencies with `taskmd graph --format ascii`
2. Update status: `taskmd set <id> --status in-progress`
3. Check off subtasks (`- [x]`) as you complete them

### Completing a Task

**Solo workflow** (default):
1. Verify all acceptance criteria are met
2. Run `taskmd verify <id>` if the task has verify checks
3. Update status: `taskmd set <id> --status completed`
4. Run `taskmd validate` to confirm no issues

**PR-review workflow** (when `workflow: pr-review` in `.taskmd.yaml`):
1. Open a pull request with your changes
2. `taskmd set <id> --status in-review --add-pr <url>`
3. Stop -- the task completes when the PR is merged

### Dependencies

- A task with unmet dependencies should stay `pending` or `blocked`
- Circular dependencies are invalid -- use `taskmd validate` to detect them

### Phases

- When introducing a new phase, add it to the `phases` list in `.taskmd.yaml` before assigning it to tasks

## Worklogs

When worklogs are enabled (`worklogs: true` in `.taskmd.yaml`), record progress with:

```bash
taskmd worklog <id> --add "Started implementation. Approach: ..."
taskmd worklog <id>            # View worklog
```

Write entries when starting, making key decisions, hitting blockers, or finishing.

## Validation

Run `taskmd validate` before committing to check for missing fields, invalid values, duplicate IDs, circular dependencies, and broken references.

## Reference

- Full specification: `TASKMD_SPEC.md`
- CLI help: `taskmd --help` or `taskmd <command> --help`
