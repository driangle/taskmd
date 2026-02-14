---
id: "074"
title: "Remove interactive TUI feature"
status: pending
priority: medium
effort: medium
dependencies: []
tags:
  - cli
  - go
  - cleanup
  - mvp
created: 2026-02-14
---

# Remove Interactive TUI Feature

## Objective

Remove the entire TUI (Terminal User Interface) feature from taskmd. The project only needs CLI and web interfaces. This reduces maintenance burden, simplifies dependencies, and shrinks the binary size.

## Context

The TUI was built using the Charm ecosystem (Bubble Tea, Lipgloss, Glamour) and provides an interactive terminal dashboard. With the web interface available for visual task browsing and the CLI for quick operations, the TUI is redundant. Removing it eliminates ~1,600 lines of application code, a 733-line documentation guide, and several transitive dependencies.

## Tasks

### Code Removal

- [ ] Delete `internal/tui/app.go` (main TUI application)
- [ ] Delete `internal/tui/app_test.go` (TUI tests)
- [ ] Delete `internal/tui/styles.go` (Lipgloss styles)
- [ ] Delete the `internal/tui/` directory
- [ ] Delete `internal/cli/tui.go` (cobra command registration)
- [ ] Remove any TUI command registration from `internal/cli/root.go` (if wired there)
- [ ] Remove the `internal/watcher/` package if it is only used by the TUI

### Dependency Cleanup

- [ ] Remove `github.com/charmbracelet/bubbletea` from `go.mod`
- [ ] Remove `github.com/charmbracelet/lipgloss` from `go.mod` (if not used elsewhere)
- [ ] Remove `github.com/charmbracelet/glamour` from `go.mod` (if not used elsewhere)
- [ ] Remove `github.com/fsnotify/fsnotify` from `go.mod` (if only used by watcher)
- [ ] Run `go mod tidy` to clean up transitive dependencies
- [ ] Verify `go.sum` is updated

### Documentation Removal

- [ ] Delete `docs/guides/tui-guide.md`
- [ ] Remove TUI references from `README.md` (if any)
- [ ] Remove TUI references from `PLAN.md` (if any)
- [ ] Remove TUI references from `CLAUDE.md` (if any)
- [ ] Remove TUI references from any other documentation

### Related Task Cleanup

- [ ] Cancel task 059 (TUI grouped view mode) — set status to `cancelled`

### Verification

- [ ] Run `go build ./...` — project compiles without TUI
- [ ] Run `go test ./...` — all remaining tests pass
- [ ] Run `make lint` — no lint errors
- [ ] Verify `taskmd --help` no longer lists `tui` command
- [ ] Verify no broken imports or dead references remain

## Acceptance Criteria

- `taskmd tui` command no longer exists
- `taskmd --help` does not list `tui`, `ui`, `interactive`, or `dashboard`
- No TUI-related source files remain in the codebase
- Charm dependencies (bubbletea, lipgloss, glamour) are removed from `go.mod` (unless used elsewhere)
- All tests pass, lint passes, project builds cleanly
- TUI documentation is removed
- Task 059 is cancelled

## Implementation Notes

Files and directories to remove:
- `internal/tui/` (entire directory: `app.go`, `app_test.go`, `styles.go`)
- `internal/cli/tui.go` (cobra command)
- `internal/watcher/` (if TUI-only — check for other usages first)
- `docs/guides/tui-guide.md`

Dependencies to evaluate for removal:
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/glamour`
- `github.com/fsnotify/fsnotify`
- `github.com/muesli/termenv` (likely a transitive dep, handled by `go mod tidy`)

Check before removing dependencies — grep for imports across the codebase to confirm they aren't used by CLI or web code.
