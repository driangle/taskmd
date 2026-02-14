---
id: "075"
title: "Add color and styling to CLI output"
status: completed
priority: high
effort: large
dependencies: []
tags:
  - cli
  - go
  - ux
  - mvp
created: 2026-02-14
---

# Add Color and Styling to CLI Output

## Objective

Add colors, bold text, and underlines to all CLI command output for a polished, readable user experience. Styling should be configurable via `.taskmd.yaml` and a `--no-color` CLI flag.

## Context

Currently the CLI output is plain text with no visual hierarchy. Adding ANSI colors and text styling (bold, underline, dim) makes output scannable and professional. Users should be able to disable colors for piping, CI environments, or personal preference.

## Configuration

### `.taskmd.yaml` support

```yaml
color: true  # or false to disable globally
```

### CLI flag

- `--no-color` global flag disables all styling (overrides config)
- Respect the `NO_COLOR` environment variable (see https://no-color.org/)
- Auto-detect non-TTY output (pipes) and disable colors automatically

## Tasks

### Core Infrastructure

- [ ] Add a `style` package (`internal/style/`) for centralized color/styling helpers
- [ ] Implement TTY detection — disable colors when stdout is not a terminal
- [ ] Add `--no-color` global flag to root command
- [ ] Read `color` setting from `.taskmd.yaml` config
- [ ] Respect `NO_COLOR` environment variable
- [ ] Define a consistent color palette (status colors, priority colors, accent colors)

### Command-Specific Styling

Review and apply styling to each command's output:

- [ ] **`list`** — Color-code status labels (green=completed, yellow=in-progress, gray=pending, red=cancelled), bold task titles, dim metadata
- [ ] **`next`** — Highlight the recommended task, bold task title, color priority badge
- [ ] **`graph`** — Color nodes by status in ASCII output, bold edges, dim completed nodes
- [ ] **`show`/`get`** — Bold section headers (Objective, Context, etc.), color status and priority badges, underline task title
- [ ] **`set`** — Color confirmation messages (green for success), highlight changed fields
- [ ] **`validate`** — Red for errors, yellow for warnings, green for "all valid"
- [ ] **`init`** — Green success messages, dim informational text
- [ ] **`spec`** — Bold section headers, syntax-highlight frontmatter fields
- [ ] **`archive`** — Color confirmation, dim archived task details

### Consistent Visual Elements

- [ ] Status badges: `completed` (green), `in-progress` (yellow), `pending` (gray/white), `cancelled` (red/strikethrough)
- [ ] Priority badges: `high` (red/bold), `medium` (yellow), `low` (dim)
- [ ] Effort badges: styled distinctly (e.g., `small`=green, `medium`=yellow, `large`=red)
- [ ] Task IDs: bold or distinct color for quick scanning
- [ ] Headers and labels: bold
- [ ] Counts and summaries: accent color
- [ ] Error messages: red with bold prefix
- [ ] Warning messages: yellow with bold prefix
- [ ] Success messages: green

### Testing

- [ ] Unit tests for style package (color functions, TTY detection mock)
- [ ] Tests verifying `--no-color` flag strips all ANSI codes
- [ ] Tests verifying `NO_COLOR` env var is respected
- [ ] Tests verifying pipe/non-TTY detection disables colors
- [ ] Integration tests for each styled command output

## Acceptance Criteria

- All CLI commands produce styled, colored output on TTY terminals
- `--no-color` flag disables all ANSI styling globally
- `NO_COLOR` environment variable is respected
- Non-TTY output (pipes, redirects) automatically disables colors
- `.taskmd.yaml` `color: false` disables colors globally
- Colors are consistent across commands (same status = same color everywhere)
- Output remains readable and correct with colors disabled
- All existing tests continue to pass
- New tests cover the style package and color toggle behavior

## Implementation Notes

Consider using a lightweight Go color library such as:
- `github.com/fatih/color` — simple, widely used
- `github.com/muesli/termenv` — more features, profile detection

Keep the style package thin — a small set of helper functions that wrap the chosen library. Commands should call helpers like `style.Status("completed")` or `style.Bold("title")` rather than using raw ANSI codes.

Ensure all styled output goes through the style package so the `--no-color` toggle works in one place.
