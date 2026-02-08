---
id: "archive-019"
title: "TUI app shell (bubbletea)"
status: pending
priority: high
effort: medium
dependencies: ["015"]
tags:
  - cli
  - go
  - tui
  - archived
created: 2026-02-08
---

# TUI App Shell (Bubbletea)

## Objective

Set up the bubbletea application shell with a basic layout: header, main content area, and a status bar/help footer. This is the TUI foundation that views plug into.

## Tasks

- [ ] Create the root bubbletea `Model` in `internal/tui/`
- [ ] Implement `Init`, `Update`, and `View` for the app shell
- [ ] Add a header bar showing the app name and scanned directory path
- [ ] Add a footer/status bar showing key bindings (q=quit, ?=help, etc.)
- [ ] Set up lipgloss styles for the layout (borders, colors, padding)
- [ ] Handle terminal resize events
- [ ] Implement graceful quit (q / ctrl+c)
- [ ] Wire up the app shell in `cmd/taskmd/main.go`

## Acceptance Criteria

- Running `taskmd` shows the TUI with header, empty content area, and footer
- Terminal resize is handled without visual glitches
- q and ctrl+c quit cleanly
- Layout adapts to terminal width/height
