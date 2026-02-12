---
id: "069"
title: "Add --critical flag to next command"
status: pending
priority: medium
effort: small
dependencies: ["032"]
tags:
  - cli
  - go
  - commands
  - productivity
  - enhancement
created: 2026-02-12
---

# Add --critical Flag to Next Command

## Objective

Add a `--critical` flag to the `taskmd next` command that filters the recommended tasks to show only those on the critical path - tasks that directly impact the project's completion timeline.

## Problem

Users sometimes want to focus exclusively on critical path tasks to minimize project duration. While the `next` command shows critical path status in the "Reason" column, there's no way to filter to *only* critical path tasks without manually scanning the output.

## Tasks

- [ ] Add `--critical` boolean flag to the next command in `internal/cli/next.go`
- [ ] When `--critical` is set, filter tasks to only include those on the critical path
- [ ] Ensure critical path analysis is performed (may need to call graph/critical path logic)
- [ ] Update command help text and usage documentation
- [ ] Add tests in `internal/cli/next_test.go` covering:
  - Happy path: `--critical` returns only critical path tasks
  - Combination with `--filter` flag
  - Combination with `--limit` flag
  - Output in different formats (table, json, yaml)
  - Edge case: no critical path tasks available

## Acceptance Criteria

- `taskmd next --critical` shows only tasks that are on the critical path
- Critical path tasks are still ranked by priority and dependencies
- Works with other flags like `--filter`, `--limit`, and `--format`
- If no critical path tasks are actionable, shows appropriate message
- All tests pass
- Linting passes (`make lint`)

## Examples

```bash
# Show all critical path tasks
taskmd next --critical

# Show only the top critical path task
taskmd next --critical --limit 1

# Show critical path CLI tasks only
taskmd next --critical --filter tag=cli

# Machine-readable output
taskmd next --critical --format json
```

## Example Output

```
Recommended critical path tasks:

 #  ID   Title                         Priority  Effort  Reason
 1  017  Implement auth middleware      high      medium  on critical path, unblocks 3 tasks
 2  009  Update error messages          medium    small   on critical path
 3  024  Add database migrations        high      large   on critical path, unblocks 5 tasks
```

## Implementation Notes

- The flag should integrate with existing critical path analysis from the graph package
- Consider showing a message like "No critical path tasks available" if the filter yields no results
- Critical path determination should use the same logic as the graph command
- May need to ensure graph analysis is performed before filtering (check performance implications)
