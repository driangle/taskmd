---
id: "018"
title: "list command - Quick textual list view"
status: completed
priority: high
effort: small
dependencies: ["017"]
tags:
  - cli
  - go
  - commands
  - mvp
created: 2026-02-08
---

# List Command - Quick Textual List View

## Objective

Implement the `list` command for quick textual listing of tasks with filtering and sorting capabilities.

## Tasks

- [x] Create `internal/cli/list.go` for list command
- [x] Implement task loading from input (file or stdin)
- [x] Support output formats:
  - `table` (default) - formatted table
  - `json` - JSON array
  - `yaml` - YAML list
- [x] Implement `--filter <expr>` for basic filtering (e.g., `status=pending`, `blocked=true`)
- [x] Implement `--sort <field>` for sorting (id, title, status, priority, etc.)
- [x] Implement `--columns` flag to customize displayed fields
- [x] Default columns: id, title, status, priority
- [x] Handle empty results gracefully

## Acceptance Criteria

- `taskmd list` displays all tasks in table format
- `taskmd list --format json` outputs valid JSON
- `taskmd list --filter status=pending` shows only pending tasks
- `taskmd list --sort priority` sorts by priority
- `taskmd list --columns id,title,status` shows only specified columns
- Works with `--stdin` and explicit file paths

## Examples

```bash
taskmd list
taskmd list tasks.md
taskmd list --filter blocked=true
taskmd list --columns id,title,deps
cat tasks.md | taskmd list --stdin --format json
```
