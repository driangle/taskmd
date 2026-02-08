---
id: "archive-020"
title: "Task list view"
status: pending
priority: high
effort: large
dependencies: ["archive-017", "archive-019"]
tags:
  - cli
  - go
  - tui
  - archived
created: 2026-02-08
---

# Task List View

## Objective

Display the scanned tasks in a scrollable table/list within the TUI. This is the primary view users interact with.

## Tasks

- [ ] Create a `TaskListModel` bubbletea component in `internal/tui/views/`
- [ ] Render tasks as rows: ID, status indicator, title, priority, group, tags
- [ ] Support a grouped view mode (`g` to toggle): tasks grouped under group headers
- [ ] Implement keyboard navigation: j/k or arrow keys to move selection
- [ ] Implement scrolling for lists longer than the terminal height
- [ ] Use lipgloss to style rows (highlight selected, dim completed, color by priority)
- [ ] Show a summary line (e.g. "12 tasks: 5 pending, 4 in-progress, 3 done")
- [ ] Wire the scanner output into the task list model
- [ ] Handle empty state (no tasks found)

## Acceptance Criteria

- Running `taskmd` in a directory with `.md` tasks shows them in a styled list
- Keyboard navigation works smoothly
- Selected row is visually highlighted
- Summary line shows correct counts
- Group column is visible in the task list
- Grouped view mode shows tasks under group headers
- Empty directory shows a helpful message
