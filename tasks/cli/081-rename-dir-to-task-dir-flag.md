---
id: "081"
title: "Rename --dir flag to --task-dir for clarity"
status: pending
priority: medium
effort: medium
dependencies: ["033"]
tags:
  - cli
  - ux
  - consistency
  - mvp
created: 2026-02-14
---

# Rename --dir Flag to --task-dir for Clarity

## Objective

Rename the global `--dir` / `-d` flag to `--task-dir` / `-d` across all CLI commands. The current `--dir` name is ambiguous — it could refer to any directory. Using `--task-dir` makes the flag's purpose explicit: it specifies the directory containing task files.

## Tasks

- [ ] Rename the global persistent flag from `--dir` to `--task-dir` in `root.go`
- [ ] Keep `-d` as the shorthand
- [ ] Update the `GlobalFlags` struct field name if needed (e.g., `Dir` → `TaskDir`)
- [ ] Update `GetGlobalFlags()` to reflect the new field name
- [ ] Update all commands that reference the directory flag to use the new name
  - [ ] `list`
  - [ ] `validate`
  - [ ] `stats`
  - [ ] `graph`
  - [ ] `board`
  - [ ] `snapshot`
  - [ ] `next`
  - [ ] `web`
  - [ ] `show` / `get`
  - [ ] `set` / `update`
  - [ ] `export`
  - [ ] `tags`
- [ ] Keep `--dir` as a hidden deprecated alias for backward compatibility
- [ ] Update all help text and usage examples
- [ ] Update tests to use `--task-dir`
- [ ] Update CLAUDE.md examples that reference `--dir`

## Acceptance Criteria

- All commands accept `--task-dir` / `-d` as the flag for specifying the task directory
- `--dir` still works as a hidden deprecated alias (no breaking change)
- Default value remains `"."` (current working directory)
- Help text shows `--task-dir` as the primary flag
- All existing tests pass with the renamed flag
- CLAUDE.md and other documentation updated
