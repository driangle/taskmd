---
id: "038"
title: "TUI task detail view with markdown rendering"
status: completed
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

- [x] Create a task detail component
- [x] Render task metadata header (ID, status, priority, effort, dependencies, tags, created date)
- [x] Render markdown body using glamour for terminal-friendly output
- [x] Implement scrolling for long task descriptions
- [x] Add keybinding to return to list view (`Esc` / `Backspace`)
- [x] Show file path of the task at the bottom
- [x] Update footer key hints for detail view context
- [x] Add tests for detail view rendering and navigation

## Acceptance Criteria

- Pressing Enter on a task shows its full detail view
- Markdown content renders with formatting (headers, lists, code blocks)
- Scrolling works for long content
- Esc returns to the list view
- File path is visible
