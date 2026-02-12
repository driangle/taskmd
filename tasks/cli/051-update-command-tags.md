---
id: "051"
title: "Support adding and removing tags via the update command"
status: pending
priority: medium
effort: small
dependencies: ["049"]
tags:
  - cli
  - go
  - commands
  - mvp
created: 2026-02-12
---

# Support Adding and Removing Tags via the Update Command

## Objective

Extend the `taskmd update` command to support adding and removing tags from a task's frontmatter.

## Problem

The `update` command currently supports modifying `status`, `priority`, and `effort`, but not `tags`. To add or remove a tag, users must manually edit the markdown file. This is especially tedious when bulk-tagging tasks or when integrating with scripts and automation.

## Tasks

- [ ] Add `--add-tag` flag (repeatable) to append one or more tags to the task
- [ ] Add `--remove-tag` flag (repeatable) to remove one or more tags from the task
- [ ] Prevent duplicate tags when adding (idempotent)
- [ ] Handle removing a tag that doesn't exist gracefully (no error, no-op)
- [ ] Update `applyUpdates` to handle the YAML list format for `tags:`
- [ ] Include tag changes in the confirmation output (e.g., `tags: [cli, go] -> [cli, go, mvp]`)
- [ ] Add comprehensive tests in `internal/cli/update_test.go`
- [ ] Run `make lint` and `make test` to verify

## Acceptance Criteria

- `taskmd update --task-id 050 --add-tag mvp` adds `mvp` to the tags list
- `taskmd update --task-id 050 --add-tag foo --add-tag bar` adds multiple tags in one call
- `taskmd update --task-id 050 --remove-tag ux` removes `ux` from the tags list
- `taskmd update --task-id 050 --add-tag mvp --remove-tag ux` adds and removes in a single call
- Adding a tag that already exists is a no-op (no duplicates)
- Removing a tag that doesn't exist is a no-op (no error)
- Can be combined with other update flags (`--status`, `--priority`, etc.)
- Tag-only updates count as "something to update" (no "nothing to update" error)
- Confirmation output shows the before/after tag list
- All tests pass, lint passes

## Examples

```bash
# Add a single tag
taskmd update --task-id 050 --add-tag mvp

# Add multiple tags
taskmd update --task-id 050 --add-tag foo --add-tag bar

# Remove a tag
taskmd update --task-id 050 --remove-tag ux

# Add and remove in one call
taskmd update --task-id 050 --add-tag mvp --remove-tag draft

# Combine with other flags
taskmd update --task-id 050 --status completed --add-tag done
```

## Example Output

```
Updated task 050 (Add command alias suggestions):
  tags: [cli, go, ux] -> [cli, go, mvp]
```

## Implementation Notes

- Use `StringArrayVar` for `--add-tag` and `--remove-tag` so they can be specified multiple times
- The `tags:` field in frontmatter is a YAML list -- the current line-by-line replacement in `applyUpdates` won't work directly for multi-line lists. Two approaches:
  1. Detect and rewrite the `tags:` block (inline format `tags: [a, b, c]` or multi-line `tags:\n  - a\n  - b`)
  2. Use the parsed task's `Tags` slice, compute the new list, and serialize the whole frontmatter back
- Approach 1 is simpler if all tag lists use the same format; approach 2 is more robust

## References

- `apps/cli/internal/cli/update.go` -- current update command implementation
- `apps/cli/internal/model/task.go` -- Task struct with Tags field
- `docs/taskmd_specification.md` -- frontmatter schema
