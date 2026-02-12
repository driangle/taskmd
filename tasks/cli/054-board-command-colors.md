---
id: "054"
title: "Add color to board command titles and IDs"
status: pending
priority: medium
effort: small
dependencies: ["023"]
tags:
  - cli
  - go
  - enhancement
  - ux
created: 2026-02-12
---

# Add Color to Board Command Titles and IDs

## Objective

Add color formatting to the CLI board command output to improve readability and visual appeal, specifically for task titles and IDs.

## Tasks

- [ ] Choose and integrate a color library (e.g., fatih/color or charmbracelet/lipgloss)
- [ ] Add color formatting to task titles
- [ ] Add distinct color to task IDs to make them stand out
- [ ] Consider status-based coloring:
  - Green for completed tasks
  - Yellow for in-progress tasks
  - Gray or default for pending tasks
  - Red for blocked tasks
- [ ] Implement `--no-color` flag to disable colors
- [ ] Ensure colors work well in different terminal themes (light/dark)
- [ ] Update board command in `internal/cli/board.go`
- [ ] Test color output in different terminals

## Acceptance Criteria

- Task titles are visually distinct with appropriate coloring
- Task IDs stand out from other text with their own color
- Status-based coloring enhances readability
- Colors are tasteful and not overwhelming
- `--no-color` flag disables all color output
- Color output respects `NO_COLOR` environment variable
- Works correctly in both light and dark terminal themes

## Implementation Notes

Consider using:
- `github.com/fatih/color` - Simple, widely-used color library
- `github.com/charmbracelet/lipgloss` - More advanced styling with layout capabilities

The color scheme should:
- Be subtle and professional
- Enhance rather than distract
- Follow common terminal color conventions
- Respect user preferences (NO_COLOR env var)

## Examples

```bash
# Default colored output
taskmd board

# Disable colors
taskmd board --no-color

# Colors work with all grouping options
taskmd board --group-by status
taskmd board --group-by priority
```
