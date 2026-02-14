---
id: "074"
title: "Remove interactive TUI feature"
status: completed
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

- [x] Delete `internal/tui/app.go` (main TUI application)
- [x] Delete `internal/tui/app_test.go` (TUI tests)
- [x] Delete `internal/tui/styles.go` (Lipgloss styles)
- [x] Delete the `internal/tui/` directory
- [x] Delete `internal/cli/tui.go` (cobra command registration)
- [x] Remove any TUI command registration from `internal/cli/root.go` (if wired there) — not present, self-registered via init()
- [x] Remove the `internal/watcher/` package if it is only used by the TUI — kept, also used by web server

### Dependency Cleanup

- [x] Remove `github.com/charmbracelet/bubbletea` from `go.mod`
- [x] Remove `github.com/charmbracelet/lipgloss` from `go.mod` (if not used elsewhere) — kept, used by CLI styling
- [x] Remove `github.com/charmbracelet/glamour` from `go.mod` (if not used elsewhere)
- [x] Remove `github.com/fsnotify/fsnotify` from `go.mod` (if only used by watcher) — kept, watcher used by web server
- [x] Run `go mod tidy` to clean up transitive dependencies
- [x] Verify `go.sum` is updated

### Documentation Removal

- [x] Delete `docs/guides/tui-guide.md`
- [x] Remove TUI references from `README.md` (if any)
- [x] Remove TUI references from `PLAN.md` (if any) — none found
- [x] Remove TUI references from `CLAUDE.md` (if any) — none found
- [x] Remove TUI references from any other documentation (cli-guide.md, quickstart.md, web-guide.md, docs site guide/tui.md, vitepress config, apps/docs/index.md, apps/docs/getting-started)

### Related Task Cleanup

- [x] Cancel task 059 (TUI grouped view mode) — task file does not exist, never created

### Verification

- [x] Run `go build ./...` — project compiles without TUI
- [x] Run `go test ./...` — all remaining tests pass
- [x] Run `make lint` — no lint errors
- [x] Verify `taskmd --help` no longer lists `tui` command
- [x] Verify no broken imports or dead references remain

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
