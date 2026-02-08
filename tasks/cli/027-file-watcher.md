---
id: "027"
title: "File watcher & live reload"
status: pending
priority: medium
effort: medium
dependencies: ["017"]
tags:
  - cli
  - go
  - core
  - infrastructure
created: 2026-02-08
---

# File Watcher & Live Reload

## Objective

Implement file watching functionality to detect changes to task files and trigger reloads in the TUI and other long-running commands.

## Tasks

- [ ] Create `internal/watcher/` package for file watching
- [ ] Use `fsnotify` to watch task files
- [ ] Detect file modifications, creations, and deletions
- [ ] Debounce rapid changes (e.g., 100ms delay)
- [ ] Provide event channel for consumers
- [ ] Handle errors gracefully (file moved, deleted, permissions)
- [ ] Support watching multiple files or directories
- [ ] Add tests with temporary file system changes

## Acceptance Criteria

- File watcher detects modifications to task files
- Changes are debounced to avoid rapid re-processing
- Provides clean API for consuming file change events
- Handles edge cases (file deletion, moves, etc.)
- TUI can integrate watcher for live updates
- Tests verify watcher behavior

## Notes

This functionality is primarily for the TUI command but may be useful for other future use cases.
