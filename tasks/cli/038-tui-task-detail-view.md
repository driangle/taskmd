---
id: "038"
title: "TUI task detail view with markdown rendering"
status: pending
priority: critical
effort: medium
dependencies: ["037"]
tags:
  - cli
  - go
  - tui
  - mvp
created: 2026-02-08
---

# TUI Task Detail View with Markdown Rendering

## Objective

When a user selects a task and presses Enter, show a full detail view with task metadata and rendered markdown body.

## Tasks

- [ ] Create a task detail component
- [ ] Render task metadata header (ID, status, priority, effort, dependencies, tags, created date)
- [ ] Render markdown body using glamour for terminal-friendly output
- [ ] Implement scrolling for long task descriptions
- [ ] Add keybinding to return to list view (`Esc` / `Backspace`)
- [ ] Show file path of the task at the bottom
- [ ] Update footer key hints for detail view context
- [ ] Add tests for detail view rendering and navigation

## Acceptance Criteria

- Pressing Enter on a task shows its full detail view
- Markdown content renders with formatting (headers, lists, code blocks)
- Scrolling works for long content
- Esc returns to the list view
- File path is visible
