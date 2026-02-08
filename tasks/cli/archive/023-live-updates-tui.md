---
id: "023"
title: "Live file updates in TUI"
status: pending
priority: medium
effort: medium
dependencies: ["018", "020"]
tags:
  - cli
  - go
  - tui
created: 2026-02-08
---

# Live File Updates in TUI

## Objective

Wire the file watcher into the TUI so that changes to markdown files on disk are immediately reflected in the task list without restarting the app.

## Tasks

- [ ] Connect the file watcher's update channel to bubbletea's event system (custom `tea.Msg`)
- [ ] On task added: insert into the list, maintaining current sort/filter
- [ ] On task modified: update the row in place, re-apply sort/filter
- [ ] On task deleted: remove from the list
- [ ] Show a brief flash/indicator when a task is updated (e.g. highlight the row briefly)
- [ ] Preserve the user's current selection and scroll position across updates
- [ ] Handle rapid successive updates without UI flicker

## Acceptance Criteria

- Editing a `.md` file in another terminal updates the TUI list within ~500ms
- Adding/deleting files updates the list
- User's selection and scroll position are preserved
- No flicker or visual artifacts during updates
