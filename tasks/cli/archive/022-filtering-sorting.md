---
id: "archive-022"
title: "Task filtering & sorting"
status: pending
priority: medium
effort: medium
dependencies: ["archive-020"]
tags:
  - cli
  - go
  - tui
  - archived
  - post-mvp
created: 2026-02-08
---

# Task Filtering & Sorting

## Objective

Add the ability to filter and sort the task list by status, priority, tags, and other fields — both interactively in the TUI and via CLI flags.

## Tasks

- [ ] Implement filter logic in `internal/filter/` (by status, priority, tags, text search)
- [ ] Add a filter bar/prompt in the TUI (e.g. `/` to search, `f` to toggle filter menu)
- [ ] Support filtering by status: pending, in-progress, done
- [ ] Support filtering by priority: high, medium, low
- [ ] Support filtering by tag
- [ ] Support filtering by group
- [ ] Support free-text search across title and body
- [ ] Implement sorting: by priority, status, group, created date, title
- [ ] Add `s` keybinding to cycle sort order
- [ ] Show active filters in the status bar
- [ ] Support CLI flags for initial filter/sort (e.g. `taskmd --status=pending --group=cli --sort=priority`)

## Acceptance Criteria

- Pressing `/` opens a search prompt that filters tasks in real time
- Filter menu allows toggling status/priority/tag filters
- Sort order can be changed interactively
- Active filters are visible in the UI
- CLI flags set the initial filter/sort state
