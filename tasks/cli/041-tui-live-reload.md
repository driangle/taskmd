---
id: "041"
title: "TUI live reload via file watcher"
status: completed
priority: high
effort: medium
dependencies: ["027", "037"]
tags:
  - cli
  - go
  - tui
  - mvp
created: 2026-02-08
completed: 2026-02-12
---

# TUI Live Reload via File Watcher

## Objective

Wire the file watcher into the TUI so that changes to markdown files on disk are immediately reflected in the task list without restarting the app.

## Tasks

- [x] Connect the file watcher to bubbletea's event system via a custom `tea.Msg`
- [x] On file change: re-scan tasks and update the list
- [x] Preserve the user's current selection and scroll position across updates
- [x] Show a brief indicator when tasks are refreshed
- [x] Handle rapid successive updates without UI flicker
- [x] Add tests for update message handling

## Acceptance Criteria

- Editing a `.md` file in another terminal updates the TUI list within ~500ms
- Adding/deleting files updates the list
- User's selection and scroll position are preserved
- No flicker or visual artifacts during updates
