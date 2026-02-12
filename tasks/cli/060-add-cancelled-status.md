---
id: "060"
title: "Add cancelled status to spec and code"
status: pending
priority: medium
effort: medium
dependencies: ["001"]
tags:
  - cli
  - go
  - spec
  - enhancement
created: 2026-02-12
---

# Add Cancelled Status to Spec and Code

## Objective

Add a new task status "cancelled" to the taskmd specification and implement support for it throughout the codebase. This allows users to mark tasks that won't be completed without deleting them from the task list.

## Tasks

- [ ] Update task specification document (`docs/TASKMD_SPEC.md`)
  - Add "cancelled" to the list of valid statuses
  - Document when to use cancelled vs completed/pending
  - Add examples of cancelled status usage
- [ ] Update status validation in code
  - Add "cancelled" to valid status enum/constants
  - Update validation logic in `internal/parser/` or `internal/validator/`
- [ ] Update status filtering and display
  - Include cancelled in status filters
  - Add distinct visual indicator for cancelled tasks (e.g., strikethrough, gray color)
- [ ] Update all commands that show/filter by status:
  - `list` command - show cancelled tasks with distinct styling
  - `board` command - add cancelled column/group
  - `graph` command - include cancelled in graph visualization
  - `stats` command - count cancelled tasks separately
  - `next` command - exclude cancelled tasks from recommendations
  - `tui` command - show cancelled tasks with distinct styling
- [ ] Update `set` command to allow setting status to cancelled
- [ ] Update tests for all affected commands
- [ ] Add integration tests for cancelled status workflows
- [ ] Update CLI help text and documentation

## Acceptance Criteria

- Task files can have `status: cancelled` in frontmatter
- Validation passes for tasks with cancelled status
- `taskmd list` shows cancelled tasks with distinct styling (e.g., strikethrough)
- `taskmd board --group-by status` includes "Cancelled" column
- `taskmd stats` shows count of cancelled tasks
- `taskmd graph` includes cancelled tasks in visualization
- `taskmd next` excludes cancelled tasks from recommendations
- `taskmd set <task-id> --status cancelled` works correctly
- TUI shows cancelled tasks with distinct visual styling
- All existing tests pass
- New tests cover cancelled status scenarios

## Implementation Notes

### Status Values

Current statuses: `pending`, `in-progress`, `completed`
New status: `cancelled`

### Semantic Meaning

- **cancelled**: Task will not be completed, but kept for historical/reference purposes
  - Examples: requirements changed, superseded by another task, deprioritized permanently
- Differs from **completed**: Task was finished successfully
- Differs from deletion: Task remains in history for tracking/reference

### Visual Indicators

Consider these styling options for cancelled tasks:
- Strike-through text for task title
- Gray or muted color
- Special icon or prefix (e.g., `✗` or `⊘`)
- In board view: separate "Cancelled" column at the end

### Status Transitions

Valid transitions to cancelled:
- `pending` → `cancelled` (most common)
- `in-progress` → `cancelled` (work stopped)
- `completed` → `cancelled` (rare, but possible if incorrectly marked)

### Filtering Considerations

Should cancelled tasks be:
- ✅ Included in `--all` flag
- ✅ Included in `--status cancelled` filter
- ❌ Excluded from default `next` recommendations
- ❌ Excluded from "active work" views (unless explicitly requested)
- ✅ Included in statistics counts

## Examples

### Task File

```yaml
---
id: "042"
title: "Implement feature X"
status: cancelled
priority: low
effort: medium
tags:
  - cli
  - deprecated
created: 2026-01-15
---

# Feature X Implementation

This task was cancelled because requirements changed.
```

### CLI Usage

```bash
# Set task to cancelled
taskmd set 042 --status cancelled

# List only cancelled tasks
taskmd list --status cancelled

# List all tasks including cancelled
taskmd list --all

# Board view with cancelled column
taskmd board --group-by status
# Output shows: Pending | In Progress | Completed | Cancelled

# Stats including cancelled
taskmd stats
# Output: Pending: 5, In Progress: 3, Completed: 10, Cancelled: 2
```

## References

- Task specification: `docs/TASKMD_SPEC.md`
- Status validation: `internal/parser/` or `internal/validator/`
- List command: `internal/cli/list.go`
- Board command: `internal/cli/board.go`
