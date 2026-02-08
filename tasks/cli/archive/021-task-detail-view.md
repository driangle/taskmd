---
id: "021"
title: "Task detail view"
status: pending
priority: medium
effort: medium
dependencies: ["020"]
tags:
  - cli
  - go
  - tui
created: 2026-02-08
---

# Task Detail View

## Objective

When a user selects a task and presses enter, show a full detail view with rendered markdown content using glamour.

## Tasks

- [ ] Create a `TaskDetailModel` bubbletea component
- [ ] Render task metadata (status, priority, effort, dependencies, tags, created date) in a styled header
- [ ] Render the markdown body using glamour for terminal-friendly markdown output
- [ ] Implement scrolling for long task descriptions
- [ ] Add keybinding to return to the list view (esc / backspace)
- [ ] Show the file path of the task at the bottom

## Acceptance Criteria

- Pressing enter on a task shows its full detail view
- Markdown content renders with formatting (headers, lists, code blocks)
- Scrolling works for long content
- Esc returns to the list view
- File path is visible
