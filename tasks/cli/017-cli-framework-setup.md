---
id: "017"
title: "CLI framework setup (cobra)"
status: completed
priority: high
effort: medium
dependencies: ["016"]
tags:
  - cli
  - go
  - core
  - infrastructure
created: 2026-02-08
---

# CLI Framework Setup (Cobra)

## Objective

Set up the CLI framework using cobra for subcommand architecture, implement shared flags, and establish the foundation for all taskmd commands.

## Tasks

- [x] Add `github.com/spf13/cobra` dependency
- [x] Add `github.com/spf13/viper` for configuration
- [x] Create root command structure in `cmd/taskmd/main.go`
- [x] Implement shared global flags:
  - `--stdin` - Read from stdin instead of file
  - `--format <fmt>` - Output format
  - `--quiet` - Suppress non-essential output
  - `--verbose` - Verbose logging
  - `--config <file>` - Custom config file
- [x] Create `internal/cli/` package for command implementations
- [x] Implement input resolver: defaults to `tasks.md`, supports explicit file path, supports `--stdin`
- [x] Add basic error handling and exit codes
- [x] Create help text and usage examples
- [x] Add version command

## Acceptance Criteria

- [x] `taskmd --help` displays available commands and flags
- [x] `taskmd --version` shows version info
- [x] All shared flags are registered and accessible to subcommands
- [x] Input resolution works (default file, explicit file, stdin)
- [x] Cobra framework is properly configured

## Notes

- This replaces the original directory scanner task
- All future commands will build on this foundation
