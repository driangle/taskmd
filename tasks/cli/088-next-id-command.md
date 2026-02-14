---
id: "088"
title: "Add next-id CLI command"
status: pending
priority: medium
effort: small
tags:
  - mvp
created: 2026-02-14
---

# Add next-id CLI Command

## Objective

Add a `next-id` CLI command that outputs the next available task ID. This simplifies task creation workflows by letting users quickly discover what ID to use for a new task file, without manually scanning existing files. The command should assume a numeric ID system (zero-padded) and handle cases where IDs may have a static alphanumeric prefix.

## Tasks

- [ ] Create `internal/cli/nextid.go` with the `next-id` command
- [ ] Scan all task files in the target directory to find the highest numeric ID
- [ ] Handle zero-padded numeric IDs (e.g., `001`, `042`, `087`)
- [ ] Handle IDs with a static prefix pattern (e.g., `WEB-001`, `CLI-042`)
- [ ] Output the next sequential ID in the same format
- [ ] Support `--dir` flag for specifying the tasks directory
- [ ] Support `--format` flag for output (plain text, json)
- [ ] Write tests in `internal/cli/nextid_test.go`

## Acceptance Criteria

- `taskmd next-id` outputs the next available numeric ID (e.g., `088`)
- Correctly handles gaps in ID sequences (uses max + 1, not gap filling)
- Works with zero-padded IDs preserving the padding width
- Handles prefixed IDs when a consistent prefix pattern exists
- Returns `001` when no existing tasks are found
- All functionality has test coverage
