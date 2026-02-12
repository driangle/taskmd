---
id: "068"
title: "Add --quick-wins flag to next command"
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

# Add --quick-wins Flag to Next Command

## Objective

Add a `--quick-wins` flag to the `taskmd next` command that filters the recommended tasks to show only quick wins - tasks that are unblocked, have small effort, and provide immediate value.

## Problem

Users sometimes want to quickly identify tasks they can knock out in a short amount of time. Currently, the `next` command shows all actionable tasks, but there's no easy way to filter specifically for quick wins without manually scanning the effort column.

## Tasks

- [ ] Add `--quick-wins` boolean flag to the next command in `internal/cli/next.go`
- [ ] When `--quick-wins` is set, filter tasks to only include those with `effort: small`
- [ ] Update the ranking logic to prioritize quick wins within the filtered set
- [ ] Update command help text and usage documentation
- [ ] Add tests in `internal/cli/next_test.go` covering:
  - Happy path: `--quick-wins` returns only small effort tasks
  - Combination with `--filter` flag
  - Combination with `--limit` flag
  - Output in different formats (table, json, yaml)
  - Edge case: no quick wins available

## Acceptance Criteria

- `taskmd next --quick-wins` shows only tasks with `effort: small`
- Quick wins are still ranked by priority, critical path, and dependencies
- Works with other flags like `--filter`, `--limit`, and `--format`
- If no quick wins are available, shows appropriate message
- All tests pass
- Linting passes (`make lint`)

## Examples

```bash
# Show all quick wins
taskmd next --quick-wins

# Show only the top quick win
taskmd next --quick-wins --limit 1

# Show quick wins in CLI tasks only
taskmd next --quick-wins --filter tag=cli

# Machine-readable output
taskmd next --quick-wins --format json
```

## Example Output

```
Recommended quick wins:

 #  ID   Title                         Priority  Effort  Reason
 1  031  Add retry logic                medium    small   no blockers, high value
 2  045  Fix typo in help text          low       small   no blockers
 3  052  Update README example          medium    small   no blockers
```

## Implementation Notes

- The flag should filter after dependency analysis but before ranking
- Consider showing a message like "No quick wins available" if the filter yields no results
- Quick wins should still respect the existing scoring heuristic within the filtered set
