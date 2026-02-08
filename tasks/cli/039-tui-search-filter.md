---
id: "039"
title: "TUI search and filter mode"
status: pending
priority: medium
effort: medium
dependencies: ["037"]
tags:
  - cli
  - go
  - tui
created: 2026-02-08
---

# TUI Search and Filter Mode

## Objective

Add interactive search and filtering to the TUI task list, allowing users to quickly find tasks by text or filter by status/priority/tag.

## Tasks

- [ ] Implement `/` keybinding to open a search prompt
- [ ] Filter task list in real time as user types
- [ ] Support free-text search across task title and ID
- [ ] Add `f` keybinding for a filter menu (status, priority, tag toggles)
- [ ] Show active filters in the status bar
- [ ] `Esc` to cancel search/filter and restore full list
- [ ] Add tests for search and filter logic

## Acceptance Criteria

- Pressing `/` opens a search prompt that filters tasks in real time
- Filter menu allows toggling status/priority/tag filters
- Active filters are visible in the UI
- Esc clears search and restores the full list
