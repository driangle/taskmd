---
id: "096"
title: "Add colors to graph command output"
status: pending
priority: medium
effort: small
dependencies: ["075"]
tags:
  - cli
  - go
  - ux
  - mvp
created: 2026-02-14
---

# Add Colors to Graph Command Output

## Objective

Add color and styling to the `graph` command ASCII output, consistent with the color scheme used by other CLI commands (list, board, next, etc.). Use the existing `internal/style/` package.

## Tasks

- [ ] Color graph nodes by task status (green=completed, yellow=in-progress, gray=pending, red=blocked/cancelled)
- [ ] Bold task titles in graph node labels
- [ ] Style task IDs with the standard ID color
- [ ] Dim completed nodes to reduce visual noise
- [ ] Style edges/arrows between nodes
- [ ] Ensure `--no-color` flag and `NO_COLOR` env var disable all colors
- [ ] Verify colors render correctly in both light and dark terminal themes
- [ ] Add tests for colored vs no-color output

## Acceptance Criteria

- Graph command output uses the same color conventions as other CLI commands
- Status-based coloring matches list/board/next commands
- Colors are disabled when `--no-color` is passed or `NO_COLOR` env var is set
- Non-TTY output (pipes) automatically disables colors
- All existing graph command tests continue to pass
- New tests verify colored output behavior
