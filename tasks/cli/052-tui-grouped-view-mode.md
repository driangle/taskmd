---
id: "052"
title: "TUI grouped view mode"
status: pending
priority: low
effort: medium
dependencies: []
tags:
  - cli
  - go
  - tui
  - enhancement
created: 2026-02-12
---

# TUI Grouped View Mode

## Objective

Add grouped view mode to the TUI that allows users to toggle (`g` key) to see tasks grouped under group/category headers. This enhances task organization and makes it easier to navigate large task sets.

## Context

From archive-020, the core TUI task list view is complete, but the grouped view mode feature was not implemented. This task adds that remaining functionality.

## Tasks

- [ ] Add group column to task row display (e.g., `[cli]`, `[web]`)
- [ ] Implement grouped view mode toggle (press `g` to switch)
- [ ] In grouped mode, render tasks under group headers
- [ ] Style group headers distinctly (e.g., bold, different color)
- [ ] Handle navigation in grouped mode (skip over headers)
- [ ] Update help text to show `g` toggle keybinding
- [ ] Handle tasks without groups (place in "Ungrouped" or similar section)
- [ ] Preserve selection when toggling between list/grouped views
- [ ] Add tests for grouped view rendering

## Acceptance Criteria

- Running TUI shows group information in each task row
- Pressing `g` toggles between flat list and grouped view
- In grouped view, tasks appear under their respective group headers
- Group headers are visually distinct and not selectable
- Navigation (j/k) skips over headers and moves between tasks
- Help footer shows the `g` toggle option
- Tasks without groups are handled gracefully
- Selection is preserved when toggling views

## Implementation Notes

The `App` struct already has a `groupBy` field (line 39 in `internal/tui/app.go`), suggesting this was planned. The implementation should:

1. Determine task groups (likely from file path or frontmatter)
2. Add a `groupedView` boolean flag to `App`
3. Modify `renderContent` to check `groupedView` flag
4. Create a new rendering function for grouped mode
5. Update keyboard handler to toggle on `g` press

## Examples

**Flat list view:**
```
020    ○ Stats command                      [medium] (cli, go)
021    ● Task detail view                   [high]   (cli, tui)
022    ○ Interactive TUI shell              [high]   (cli, tui)
```

**Grouped view (after pressing `g`):**
```
━━━ cli ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
020    ○ Stats command                      [medium]
022    ○ Interactive TUI shell              [high]

━━━ web ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
030    ○ Dashboard UI                       [high]
```
