---
id: "025"
title: "CLI polish & error handling"
status: pending
priority: low
effort: medium
dependencies: ["021", "022", "023", "024"]
tags:
  - cli
  - go
  - polish
created: 2026-02-08
---

# CLI Polish & Error Handling

## Objective

Final polish pass on the CLI: robust error handling, help text, configuration, and user experience refinements.

## Tasks

- [ ] Add `--help` with clear usage documentation for all subcommands
- [ ] Add `--version` flag
- [ ] Graceful error messages for common issues: no `.md` files found, permission denied, invalid directory
- [ ] Add color theme support (respect `NO_COLOR` env var)
- [ ] Add `--no-watch` flag to disable file watching
- [ ] Add `--dir` flag to specify a directory other than cwd
- [ ] Ensure clean exit on SIGINT/SIGTERM (no partial output, no zombie watchers)
- [ ] Add config file support (`~/.taskmd/config.yaml`) for default flags
- [ ] Test on macOS, Linux (and Windows if feasible)
- [ ] Write a README with installation and usage instructions

## Acceptance Criteria

- Help text is clear and complete
- Errors produce human-readable messages (no raw stack traces)
- `NO_COLOR` is respected
- Clean shutdown on all signal types
- Works on macOS and Linux
