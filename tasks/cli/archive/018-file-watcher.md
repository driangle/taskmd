---
id: "archive-018"
title: "File watcher & live reload"
status: pending
priority: high
effort: medium
dependencies: ["archive-017"]
tags:
  - cli
  - go
  - core
  - archived
created: 2026-02-08
---

# File Watcher & Live Reload

## Objective

Watch the scanned directory for changes to `.md` files and automatically update the in-memory task store. Changes should be pushed to the TUI for live updates.

## Tasks

- [ ] Integrate `fsnotify` to watch the scanned directory tree
- [ ] Handle file events: create, modify, delete, rename
- [ ] On create/modify: re-parse the affected file and update the task store
- [ ] On delete: remove the task from the store
- [ ] Debounce rapid file changes (e.g. editor save-then-rename patterns)
- [ ] Expose a channel or callback mechanism for the TUI to receive update notifications
- [ ] Handle watcher errors (permission denied, too many watches) gracefully
- [ ] Write tests for add/modify/delete scenarios

## Acceptance Criteria

- Editing a `.md` file on disk updates the in-memory task within ~200ms
- Creating a new `.md` file adds it to the store
- Deleting a `.md` file removes it from the store
- Rapid edits don't cause duplicates or races
- TUI receives a notification on each change
