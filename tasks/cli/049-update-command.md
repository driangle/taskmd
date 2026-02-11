---
id: "049"
title: "Add update command to change task status from the CLI"
status: pending
priority: high
effort: medium
dependencies: ["016"]
tags:
  - cli
  - go
  - commands
  - productivity
  - mvp
created: 2026-02-11
---

# Add `update` Command to Change Task Status from the CLI

## Objective

Implement a `taskmd update` command that modifies a task's frontmatter fields directly from the command line. The primary use case is updating task status (e.g., marking a task as `completed`), but the command should support updating other frontmatter fields as well.

## Problem

Currently, changing a task's status requires manually opening the markdown file and editing the YAML frontmatter. This is tedious and error-prone, especially when completing tasks in rapid succession. A CLI command to update status would streamline the workflow and integrate well with scripting and automation.

## Tasks

- [ ] Create `internal/cli/update.go` with the `update` cobra command
- [ ] Add `--task-id` flag (required) to specify the task to update
- [ ] Add `--status` flag to set the task status (`pending`, `in-progress`, `completed`, `blocked`)
- [ ] Add `--priority` flag to set the task priority (`high`, `medium`, `low`)
- [ ] Add `--effort` flag to set the task effort (`small`, `medium`, `large`)
- [ ] Add a shorthand `--done` flag as an alias for `--status completed`
- [ ] Locate the task file by scanning the directory and matching by task ID
- [ ] Parse the existing frontmatter, apply the changes, and write back the file preserving the markdown body
- [ ] Validate that the new status/priority/effort values are valid before writing
- [ ] Print a confirmation message showing what changed (e.g., `Task 032: status pending -> completed`)
- [ ] Support `--dir` flag for consistency with other commands
- [ ] Create `internal/cli/update_test.go` with comprehensive tests
- [ ] Run `make lint` and `make test` to verify

## Acceptance Criteria

- `taskmd update --task-id 032 --status completed` sets task 032's status to `completed`
- `taskmd update --task-id 032 --done` is a shorthand for `--status completed`
- `taskmd update --task-id 032 --priority high` changes the priority
- `taskmd update --task-id 032 --status completed --priority high` updates multiple fields at once
- Invalid status/priority/effort values produce a clear error message
- Task ID not found produces a clear error message
- The markdown body and non-updated frontmatter fields are preserved exactly
- Works with `--dir` to target a specific task directory
- All tests pass, lint passes

## Examples

```bash
# Mark a task as completed
taskmd update --task-id 032 --done

# Set status explicitly
taskmd update --task-id 032 --status in-progress

# Update multiple fields
taskmd update --task-id 032 --status completed --priority high

# Target a specific directory
taskmd update --task-id 032 --done --dir ./tasks/cli
```

## Example Output

```
Updated task 032 (next command - Suggest the best task to pick up):
  status: pending -> completed
```

```
Updated task 015 (Go CLI scaffolding):
  status: pending -> in-progress
  priority: low -> high
```

## Test Cases

- Happy path: update status of a task in a temp directory
- `--done` flag sets status to `completed`
- `--status` with each valid value (`pending`, `in-progress`, `completed`, `blocked`)
- `--priority` with each valid value (`high`, `medium`, `low`)
- `--effort` with each valid value (`small`, `medium`, `large`)
- Multiple flags combined in a single call
- Invalid status value returns an error
- Invalid priority value returns an error
- Task ID not found returns an error
- No flags provided returns an error (nothing to update)
- Markdown body is preserved after update
- Non-updated frontmatter fields are preserved after update

## Implementation Notes

- Reuse the existing scanner/parser to locate and read task files
- When writing back, preserve the original markdown body verbatim -- only modify the YAML frontmatter block
- Consider using a YAML-aware approach to avoid reordering frontmatter keys unexpectedly

## References

- `internal/cli/show.go` -- similar pattern of finding a task by ID
- `internal/parser/` -- existing frontmatter parsing logic
- `docs/taskmd_specification.md` -- valid field values
