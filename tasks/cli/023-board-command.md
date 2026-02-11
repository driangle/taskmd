---
id: "023"
title: "board command - Grouped kanban-like view"
status: completed
priority: high
effort: medium
dependencies: ["017"]
tags:
  - cli
  - go
  - commands
  - visualization
  - mvp
created: 2026-02-08
---

# Board Command - Grouped Kanban-like View

## Objective

Implement the `board` command to generate grouped, human-readable status views similar to a kanban board.

## Tasks

- [x] Create `internal/cli/board.go` for board command
- [x] Implement `--group-by` flag with options:
  - `status` (default) - Group by task status
  - `owner` - Group by owner/assignee
  - `area` - Group by area tag
  - `milestone` - Group by milestone
  - `priority` - Group by priority
- [x] Support output formats:
  - `md` (default) - Markdown with columns
  - `txt` - Plain text
  - `json` - JSON grouped structure
- [x] Implement `--out <file>` to write to file
- [x] Display task counts per group
- [x] Sort groups logically (e.g., pending -> in-progress -> completed)

## Acceptance Criteria

- `taskmd board` displays tasks grouped by status in markdown
- `--group-by owner` groups tasks by owner field
- `--format txt` produces plain text output
- `--format json` produces structured JSON
- Shows task counts for each group
- Groups are sorted in logical order
- Works with stdin and explicit file paths

## Examples

```bash
taskmd board
taskmd board --group-by status --format md > board.md
taskmd board --group-by priority
cat tasks.md | taskmd board --stdin --group-by owner
```
