---
id: "037"
title: "TUI task list view with navigation"
status: pending
priority: high
effort: large
dependencies: ["026"]
tags:
  - cli
  - go
  - tui
created: 2026-02-08
---

# TUI Task List View with Navigation

## Objective

Replace the static summary in the TUI content area with a scrollable, navigable task list showing task metadata. This is the core interactive view.

## Tasks

- [ ] Create a task list component rendering rows: ID, status indicator, title, priority, tags
- [ ] Implement keyboard navigation: `j`/`k` or arrow keys to move selection up/down
- [ ] Implement scrolling for lists longer than the visible area
- [ ] Style rows with lipgloss (highlight selected, dim completed, color by priority)
- [ ] Show a summary line at the bottom (e.g. "12 tasks: 5 pending, 4 in-progress, 3 done")
- [ ] Handle empty state (no tasks found)
- [ ] Update footer key hints to show navigation keys
- [ ] Add tests for list navigation and rendering

## Acceptance Criteria

- `taskmd tui` shows tasks in a styled, scrollable list
- `j`/`k` and arrow keys navigate the selection
- Selected row is visually highlighted
- Summary line shows correct counts
- Empty directory shows a helpful message
