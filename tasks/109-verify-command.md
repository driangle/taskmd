---
id: "109"
title: "Add verify field and command for task acceptance checks"
status: pending
priority: high
effort: medium
tags:
  - verification
  - ai
  - dx
  - mvp
touches:
  - cli
created: 2026-02-14
---

# Add Verify Field and Command for Task Acceptance Checks

## Objective

Add a `verify` frontmatter field that lets task authors define acceptance checks — either executable shell commands or natural language assertions. Pair this with a `taskmd verify --task-id <ID>` command that runs executable checks and reports pass/fail results, while surfacing natural language checks for the agent to evaluate on its own. This closes the loop between "what to do" and "how to know it's done," giving both human and AI agents a concrete way to validate their work.

## Tasks

### Specification
- [ ] Add `verify` field to the taskmd specification as an optional array of strings
- [ ] Define two entry types: executable commands (prefixed with `$ `) and natural language assertions (plain text)
- [ ] Document the field in `docs/taskmd_specification.md` with examples of both types
- [ ] Update the frontmatter schema and field summary table

### Parser
- [ ] Extend the task model to include the `verify` field
- [ ] Parse `verify` as a list of strings from frontmatter
- [ ] Classify each entry as executable (starts with `$ `) or natural language
- [ ] Preserve the field during read/write operations

### CLI Command
- [ ] Create `internal/cli/verify.go` with the `verify` command
- [ ] Accept `--task-id` flag (required) to identify the task
- [ ] Read the task's `verify` list
- [ ] For executable entries (`$ `): run in a shell subprocess, capture stdout/stderr/exit code, report pass/fail
- [ ] For natural language entries: display them clearly as checks the agent must evaluate (not executed)
- [ ] Report overall results: executable pass/fail counts + pending natural language checks
- [ ] Exit with non-zero status if any executable check fails (useful for CI/scripting)
- [ ] Support `--verbose` flag to show full command output even on success
- [ ] Support `--dry-run` flag to display all checks without executing any
- [ ] Add comprehensive tests in `internal/cli/verify_test.go`
- [ ] Register command with `rootCmd`

### Safety
- [ ] Display the commands that will be run before executing (with a confirmation prompt or `--yes` flag to skip)
- [ ] Support a timeout per command to prevent hangs (default: 60s, configurable with `--timeout`)

## Acceptance Criteria

- Tasks can define a `verify` field with both executable and natural language checks:
  ```yaml
  verify:
    - "$ go test ./internal/api/... -run TestPagination"
    - "$ curl -s localhost:8080/api/tasks?page=2 | jq '.meta.total_pages'"
    - "Pagination links appear in the API response headers"
    - "Page size defaults to 20 when not specified"
  ```
- Entries prefixed with `$ ` are treated as executable shell commands
- Plain text entries are natural language assertions for the agent to evaluate
- `taskmd verify --task-id 042` runs executable checks and displays natural language checks
- Executable checks show pass/fail with exit code; natural language checks are listed as pending for agent review
- Overall exit code is non-zero if any executable check fails
- `--dry-run` lists all checks without running any
- `--verbose` shows full output for all commands
- `--timeout` controls per-command timeout (default 60s)
- Commands are displayed before execution for transparency
- Tasks without a `verify` field produce a clear "no checks defined" message
- Specification is updated with the new field
- Tests cover pass, fail, timeout, dry-run, missing field, mixed check types, and multi-command scenarios
