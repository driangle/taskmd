---
id: "071"
title: "Add --filter flag to graph command"
status: completed
priority: medium
effort: small
dependencies:
  - "022"
tags:
  - cli
  - go
  - commands
  - enhancement
  - mvp
created: 2026-02-13
---

# Add --filter Flag to Graph Command

## Objective

Add the same `--filter` flag from the `list` command to the `graph` command, allowing users to filter which tasks appear in the dependency graph using field=value expressions.

## Context

The `list` command supports `--filter status=pending --filter priority=high` (AND logic, multiple flags). The `graph` command currently only supports `--exclude-status` for filtering out tasks by status. Adding `--filter` brings parity between the two commands and enables richer graph filtering (by priority, effort, tag, group, etc.).

The filtering logic already exists in `internal/cli/list.go` (`applyFilters`, `matchesFilter`, `filterCriteria`, `matchesAllFilters`). This can be reused directly or extracted to a shared location.

## Tasks

- [ ] Extract filter functions (`applyFilters`, `matchesFilter`, `filterCriteria`, `matchesAllFilters`) from `list.go` into a shared location (e.g. `internal/cli/filter.go`) so both commands can use them
- [ ] Add `--filter` flag to `graphCmd` in `graph.go` (same semantics as list: `StringArrayVar`, multiple values AND'ed)
- [ ] Apply filters to scanned tasks in `runGraph`, before the existing `--exclude-status` logic
- [ ] Update the graph command's `Long` help text with `--filter` examples
- [ ] Clean up dependencies that reference filtered-out tasks (already done for `--exclude-status`, ensure the same cleanup applies after `--filter`)
- [ ] Add tests for `--filter` on graph command (status, priority, tag, combinations)

## Acceptance Criteria

- `taskmd graph --filter priority=high` shows only high-priority tasks in the graph
- `taskmd graph --filter tag=mvp` shows only mvp-tagged tasks
- `taskmd graph --filter status=pending --filter priority=high` combines with AND logic
- `--filter` and `--exclude-status` can be used together
- Dependencies referencing filtered-out tasks are cleaned up
- Existing graph tests still pass
- Help text documents the new flag with examples

## Examples

```bash
taskmd graph --filter priority=high --format ascii
taskmd graph --filter tag=mvp --format mermaid
taskmd graph --filter status=pending --filter effort=small
taskmd graph --filter group=web --exclude-status completed
```

## References

- Task 022 (graph command — completed)
- `internal/cli/list.go` — existing filter implementation
- `internal/cli/graph.go` — target file
