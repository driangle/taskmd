---
id: "015"
title: "Go CLI scaffolding (taskmd)"
status: pending
priority: high
effort: small
dependencies: []
tags:
  - setup
  - cli
  - go
created: 2026-02-08
---

# Go CLI Scaffolding (taskmd)

## Objective

Initialize the Go CLI project for taskmd under `apps/cli/`. This is the terminal-based interface that scans the current directory for markdown task files, builds an in-memory representation, and displays them in an interactive TUI. File changes are automatically detected and reflected in the UI.

## Tasks

- [ ] Initialize Go module under `apps/cli/` (`go mod init`)
- [ ] Set up directory structure:
  - `cmd/taskmd/main.go` — entrypoint
  - `internal/` — core application logic
- [ ] Add core dependencies:
  - `github.com/charmbracelet/bubbletea` — TUI framework
  - `github.com/charmbracelet/lipgloss` — terminal styling
  - `github.com/charmbracelet/glamour` — markdown rendering
  - `github.com/fsnotify/fsnotify` — file watching
  - `github.com/yuin/goldmark` — markdown parsing
- [ ] Create a minimal `main.go` that compiles and runs
- [ ] Add a `Makefile` or build script for `go build -o taskmd`
- [ ] Verify `go run ./cmd/taskmd` executes without errors

## Acceptance Criteria

- `go build ./cmd/taskmd` produces a working binary
- `go run ./cmd/taskmd` executes without errors
- All core dependencies are resolved
- Directory structure is in place for future development
