---
id: "020"
title: "stats command - Show computed metrics"
status: completed
priority: medium
effort: small
dependencies: ["017"]
tags:
  - cli
  - go
  - commands
  - metrics
created: 2026-02-08
---

# Stats Command - Show Computed Metrics

## Objective

Implement the `stats` command to display computed metrics about the task set including totals, blocked tasks, critical path, and dependency depth.

## Tasks

- [x] Create `internal/cli/stats.go` for stats command
- [x] Create `internal/metrics/` package for metric calculations
- [x] Implement metric calculations:
  - Total tasks
  - Tasks by status (pending, in-progress, completed, blocked)
  - Tasks by priority
  - Tasks by effort
  - Blocked tasks count
  - Critical path length (longest dependency chain)
  - Maximum dependency depth
  - Average dependencies per task
- [x] Support output formats: `table` (default), `json`
- [x] Display metrics in human-readable table format
- [x] JSON format includes all raw metric data

## Acceptance Criteria

- `taskmd stats` displays key metrics in table format
- Shows total tasks broken down by status
- Calculates and displays critical path length
- Calculates maximum dependency depth
- `--format json` outputs structured metrics
- Works with stdin and explicit file paths

## Examples

```bash
taskmd stats
taskmd stats tasks.md
taskmd stats --format json
cat tasks.md | taskmd stats --stdin
```
