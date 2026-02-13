---
id: "032"
title: "next command - Suggest the best task to pick up"
status: completed
priority: medium
effort: medium
dependencies: ["018", "021"]
tags:
  - cli
  - go
  - commands
  - productivity
  - mvp
created: 2026-02-08
---

# Next Command - Suggest the Best Task to Pick Up

## Objective

Implement a `taskmd next` command that recommends which task(s) a user should work on next by combining dependency analysis, priority, and critical-path information into a single ranked output.

## Problem

Today, determining the next best task requires combining output from multiple commands (`list --filter`, `snapshot --derived`, `graph`) and mentally cross-referencing blocked status, priority, and critical path. There is no single command that answers "what should I work on next?"

## Tasks

- [x]Create `internal/cli/next.go` for the next command
- [x]Filter to actionable tasks: status is `pending` or `in-progress`, and not blocked (all dependencies completed)
- [x]Rank actionable tasks using a scoring heuristic that considers:
  - Priority (high > medium > low)
  - Critical path membership (prefer tasks on the critical path)
  - Dependency depth (prefer tasks that unblock the most downstream work)
  - Effort (optional tiebreaker — prefer smaller effort for quick wins)
- [x]Display the top-N recommendations (default 5, configurable via `--limit`)
- [x]Show why each task was recommended (e.g., "high priority, on critical path, unblocks 3 tasks")
- [x]Support output formats: `table` (default), `json`, `yaml`
- [x]Support `--filter` flag to narrow scope (e.g., `--filter tag=cli`)
- [x]Add comprehensive tests in `internal/cli/next_test.go`

## Acceptance Criteria

- `taskmd next` prints a ranked list of recommended tasks with reasoning
- Blocked tasks are never recommended
- Completed tasks are never recommended
- `--limit N` controls how many suggestions are shown
- `--format json` produces machine-readable output
- `--filter` can narrow the candidate pool
- Works with both directory scanning and `--stdin`

## Examples

```bash
# Show top 5 recommended tasks
taskmd next

# Show only the single best task
taskmd next --limit 1

# Filter to CLI tasks only
taskmd next --filter tag=cli

# Machine-readable output
taskmd next --format json

# From stdin
cat tasks.md | taskmd next --stdin
```

## Example Output

```
Recommended tasks:

 #  ID   Title                         Priority  Reason
 1  017  Implement auth middleware      high      on critical path, unblocks 3 tasks
 2  023  Add input validation           high      unblocks 2 tasks
 3  009  Update error messages          medium    on critical path
 4  031  Add retry logic                medium    no blockers, quick win (small effort)
 5  012  Refactor config loading        low       unblocks 1 task
```
