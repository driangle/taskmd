---
id: "040"
title: "TUI command flags (--focus, --filter, --group-by)"
status: pending
priority: medium
effort: small
dependencies: ["037", "039"]
tags:
  - cli
  - go
  - tui
created: 2026-02-08
---

# TUI Command Flags

## Objective

Add CLI flags to the `tui` command for pre-configuring the initial view state.

## Tasks

- [ ] Add `--focus <task-id>` flag to start with a specific task selected
- [ ] Add `--filter <expr>` flag to pre-apply a filter (e.g. `status=pending`)
- [ ] Add `--group-by <field>` flag to group tasks in the list
- [ ] Add `--readonly` flag to disable any future editing features
- [ ] Add tests for flag parsing and initial state

## Acceptance Criteria

- `taskmd tui --focus T3` starts with task T3 selected
- `taskmd tui --filter status=pending` shows only pending tasks
- `taskmd tui --group-by priority` groups tasks by priority
- `--readonly` flag is accepted and stored for future use
