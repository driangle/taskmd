---
id: "048"
title: "Add init command to scaffold CLAUDE.md from template"
status: completed
priority: medium
effort: small
dependencies: ["036"]
tags:
  - cli
  - dx
  - claude-integration
  - mvp
created: 2026-02-10
---

# Add `init` Command to Scaffold CLAUDE.md from Template

## Objective

Add a `taskmd init` command that outputs the CLAUDE.md template to the user's project, enabling Claude Code to understand and work with their taskmd files out of the box.

## Context

Task 036 created the CLAUDE.md template at `docs/templates/CLAUDE.md`. Currently users must manually copy this file. The `init` command provides a standard CLI workflow for bootstrapping taskmd integration in a project -- `init` is the established convention across CLIs (git, npm, go mod, etc.).

The template should be embedded in the binary so users don't need a local copy of the taskmd source.

## Tasks

- [x] Create `internal/cli/init.go` with the `init` cobra command
- [x] Embed `docs/templates/CLAUDE.md` using Go's `embed` package
- [x] Write the template to `CLAUDE.md` in the target directory (default: current working directory)
- [x] Add `--dir` flag to specify an output directory
- [x] Add `--force` flag to overwrite an existing CLAUDE.md
- [x] Refuse to overwrite without `--force` and print a clear message
- [x] Add `--stdout` flag to print the template to stdout instead of writing a file
- [x] Print a success message with the path of the created file
- [x] Create `internal/cli/init_test.go` with comprehensive tests
- [x] Run `make lint` and `make test` to verify

## Acceptance Criteria

- `taskmd init` writes `CLAUDE.md` to the current directory
- `taskmd init --dir ./my-project` writes to a specific directory
- `taskmd init --force` overwrites an existing file
- `taskmd init --stdout` prints the template without writing a file
- Running without `--force` when `CLAUDE.md` exists returns an error
- Template content matches `docs/templates/CLAUDE.md`
- All tests pass, lint passes

## Test Cases

- Happy path: writes CLAUDE.md to a temp directory
- Refuses to overwrite existing file without `--force`
- Overwrites with `--force`
- `--stdout` prints to stdout, does not create a file
- `--dir` writes to the specified directory
- `--dir` with non-existent directory returns an error
- Verify written content matches embedded template

## References

- `docs/templates/CLAUDE.md` -- the template to embed
- `internal/cli/root.go` -- command registration pattern
- Task 036 -- created the template
