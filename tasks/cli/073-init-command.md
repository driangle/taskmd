---
id: "073"
title: "Add unified init command combining agents init and spec"
status: pending
priority: medium
effort: medium
dependencies:
  - "067"
  - "072"
tags:
  - cli
  - go
  - commands
  - dx
  - mvp
created: 2026-02-13
---

# Add Unified `taskmd init` Command

## Objective

Add a top-level `taskmd init` command that combines the functionality of `taskmd agents init` and `taskmd spec` into a single project initialization step. Users currently need two commands to bootstrap a project; this should be one.

## Context

Today, initializing a project for taskmd requires two separate commands:

```bash
taskmd agents init    # writes CLAUDE.md (or other agent configs)
taskmd spec           # writes TASKMD_SPEC.md
```

A single `taskmd init` command should run both steps, creating the agent configuration and the spec file in one go. The existing `agents init` and `spec` commands should remain available for users who only want one of the two.

## Tasks

- [ ] Create `internal/cli/project_init.go` with a `taskmd init` cobra command
- [ ] By default, run both agent config generation (Claude) and spec generation
- [ ] Support all existing agent flags (`--claude`, `--gemini`, `--codex`)
- [ ] Add `--no-spec` flag to skip writing `TASKMD_SPEC.md`
- [ ] Add `--no-agent` flag to skip writing agent configuration files
- [ ] Support `--force` flag to overwrite existing files
- [ ] Support `--stdout` flag to print all output to stdout instead of writing files
- [ ] Support `--dir` flag for target directory
- [ ] Print a summary of what was created (list of files written)
- [ ] If a file already exists and `--force` is not set, skip it with a warning (don't fail the whole command)
- [ ] Handle the existing `agents init` deprecation alias — `taskmd init` now refers to this new command, not the old agents init alias
- [ ] Create `internal/cli/project_init_test.go` with comprehensive tests
- [ ] Run `make lint` and `make test` to verify

## Acceptance Criteria

- `taskmd init` writes both `CLAUDE.md` and `TASKMD_SPEC.md` to the current directory
- `taskmd init --gemini` writes `GEMINI.md` and `TASKMD_SPEC.md`
- `taskmd init --claude --gemini` writes `CLAUDE.md`, `GEMINI.md`, and `TASKMD_SPEC.md`
- `taskmd init --no-spec` writes only the agent config file(s)
- `taskmd init --no-agent` writes only `TASKMD_SPEC.md`
- `taskmd init --force` overwrites existing files
- `taskmd init --dir ./project` writes to the specified directory
- Existing files are skipped with a warning when `--force` is not set
- `taskmd agents init` and `taskmd spec` continue to work independently
- All tests pass, lint passes

## Test Cases

- Happy path: writes both agent config and spec to a temp directory
- Agent flags: `--claude`, `--gemini`, `--codex` select which agent configs to generate
- `--no-spec` skips spec generation
- `--no-agent` skips agent config generation
- `--no-spec --no-agent` is an error (nothing to do)
- `--force` overwrites existing files
- `--dir` writes to specified directory
- Non-existent `--dir` target returns an error
- Existing files without `--force` are skipped with a warning, not a hard error
- `--stdout` prints all content to stdout without writing files
- Summary output lists all files created

## Examples

```bash
taskmd init                        # Writes CLAUDE.md + TASKMD_SPEC.md
taskmd init --gemini               # Writes GEMINI.md + TASKMD_SPEC.md
taskmd init --claude --gemini      # Writes CLAUDE.md + GEMINI.md + TASKMD_SPEC.md
taskmd init --no-spec              # Writes CLAUDE.md only
taskmd init --no-agent             # Writes TASKMD_SPEC.md only
taskmd init --force                # Overwrite existing files
taskmd init --dir ./my-project     # Write to a specific directory
```

## Implementation Notes

- Reuse the existing template logic from `internal/cli/init.go` (agents) and `internal/cli/spec.go`
- Extract shared write-file-with-force logic if it isn't already shared
- The command should be registered directly on `rootCmd`, replacing any existing `init` alias that pointed to `agents init`
