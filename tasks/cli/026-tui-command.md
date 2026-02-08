---
id: "026"
title: "tui command - Interactive terminal UI"
status: completed
priority: high
effort: large
dependencies: ["017", "027", "028"]
tags:
  - cli
  - go
  - tui
  - commands
created: 2026-02-08
---

# TUI Command - Interactive Terminal UI

## Objective

Implement the `tui` command to launch an interactive terminal interface using Bubble Tea for browsing and managing tasks.

## Tasks

- [x] Create `internal/cli/tui.go` for tui command
- [x] Create `internal/tui/` package for TUI implementation
- [ ] Implement Bubble Tea model with:
  - Task list view
  - Task detail view
  - Navigation (arrow keys, vim keys)
  - Search/filter mode
- [ ] Implement command flags:
  - `--focus <task-id>` - Start with specific task selected
  - `--filter <expr>` - Pre-apply filter
  - `--group-by <field>` - Group tasks in list
  - `--readonly` - Disable editing features
- [x] Keyboard shortcuts:
  - `j/k` or arrows - Navigate
  - `/` - Search/filter
  - `Enter` - View details
  - `q` - Quit
  - `?` - Help
- [ ] Display task metadata (id, title, status, priority, deps)
- [ ] Syntax-highlighted markdown rendering for task body
- [x] Status bar with help hints

## Acceptance Criteria

- `taskmd tui` launches interactive interface
- Task list is navigable with keyboard
- Enter key shows task details
- Search/filter mode works
- `--focus T3` starts with task T3 selected
- `--readonly` disables editing
- Markdown body is rendered nicely
- Works with stdin and explicit file paths
- Graceful error if no TTY available

## Examples

```bash
taskmd tui
taskmd tui tasks.md --focus T3
taskmd tui --filter status=pending
cat tasks.md | taskmd tui --stdin
```

## Notes

This replaces the old task 019 (TUI app shell) and consolidates TUI-related tasks.
