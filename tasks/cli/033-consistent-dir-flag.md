---
id: "033"
title: "Consistent --dir flag across all commands"
status: completed
priority: medium
effort: medium
dependencies: []
tags:
  - cli
  - go
  - ux
  - consistency
created: 2026-02-08
---

# Consistent --dir Flag Across All Commands

## Objective

Ensure every CLI command that scans a task directory accepts a `--dir` flag (with `-d` shorthand) instead of relying on a positional argument. This makes the interface consistent and easier to discover.

## Problem

Currently, most commands (`list`, `validate`, `stats`, `graph`, `board`, `snapshot`, `next`) determine their scan directory from an optional positional argument (`args[0]`), defaulting to `"."`. The `web` command is the only one that uses a `--dir` flag. This inconsistency means:

- Users must remember which commands use `--dir` vs a positional arg
- `--dir` is more explicit and self-documenting than a bare positional argument
- Shell completion and help text are clearer with a named flag

## Tasks

- [ ] Add a global `--dir` / `-d` persistent flag to `rootCmd` in `root.go`, defaulting to `"."`
- [ ] Include `Dir` in the `GlobalFlags` struct and `GetGlobalFlags()`
- [ ] Update each command's `RunE` to read the directory from `GetGlobalFlags().Dir` instead of `args[0]`
  - [ ] `list`
  - [ ] `validate`
  - [ ] `stats`
  - [ ] `graph`
  - [ ] `board`
  - [ ] `snapshot`
  - [ ] `next`
  - [ ] `web` (migrate from its local `--dir` flag to the global one)
- [ ] Keep positional arg support as a fallback for backward compatibility (positional arg overrides `--dir` if both provided)
- [ ] Update command `Use` / help text to reflect the new `--dir` flag
- [ ] Add tests verifying `--dir` works for each command
- [ ] Add test verifying positional arg still works (backward compat)

## Acceptance Criteria

- All commands that scan a task directory accept `--dir` / `-d`
- Default value is `"."` (current working directory)
- Existing positional-arg usage still works for backward compatibility
- `taskmd list --dir ./tasks` and `taskmd list ./tasks` produce the same result
- `taskmd web start --dir ./tasks` continues to work as before
- Help text for every command shows the `--dir` flag
- Tests cover both `--dir` and positional-arg usage

## Examples

```bash
# These should all be equivalent:
taskmd list ./tasks
taskmd list --dir ./tasks
taskmd list -d ./tasks

# Global flag works with any command:
taskmd validate --dir ./tasks
taskmd stats -d ./my-project/tasks
taskmd graph --dir ./tasks --format json
taskmd next --dir ./tasks --limit 3
taskmd board --dir ./tasks
taskmd snapshot --dir ./tasks
taskmd web start --dir ./tasks
```
