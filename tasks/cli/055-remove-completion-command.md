---
id: "055"
title: "Remove completion CLI command"
status: pending
priority: low
effort: small
dependencies: []
tags:
  - cli
  - go
  - cleanup
created: 2026-02-12
---

# Remove Completion CLI Command

## Objective

Remove the `completion` CLI command from taskmd.

## Tasks

- [ ] Locate the completion command implementation
- [ ] Remove the completion command file/code
- [ ] Remove any references to completion in documentation
- [ ] Update help text if completion is mentioned
- [ ] Remove completion command from root command registration
- [ ] Test that CLI works without the completion command
- [ ] Update any tests that reference the completion command

## Acceptance Criteria

- `taskmd completion` command no longer exists
- `taskmd --help` does not list completion command
- No broken references or imports remain
- All tests pass after removal
- CLI functions normally without the completion command

## Implementation Notes

Typical locations to check:
- `internal/cli/completion.go` (if it exists)
- Command registration in `internal/cli/root.go`
- Any test files referencing completion
- Documentation in README or help text

## Examples

```bash
# After removal, this should show command not found
taskmd completion

# Help should not list completion
taskmd --help
```
