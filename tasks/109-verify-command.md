---
id: "109"
title: "Add verify field and command for task acceptance checks"
status: completed
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

Add a `verify` frontmatter field that lets task authors define typed acceptance checks. Each check is a map with a `type` field indicating what kind of check it is (e.g., `bash` for shell commands, `assert` for natural language assertions). Pair this with a `taskmd verify --task-id <ID>` command that runs executable checks and reports pass/fail results, while surfacing assertion checks for the agent to evaluate on its own. This closes the loop between "what to do" and "how to know it's done," giving both human and AI agents a concrete way to validate their work. The typed structure is extensible to future check types (e.g., `http`, `api`).

## Tasks

### Specification
- [x] Add `verify` field to the taskmd specification as an optional array of typed maps
- [x] Define initial check types: `bash` (shell commands) and `assert` (natural language assertions)
- [x] Each entry must have a `type` field; other fields depend on the type (e.g., `run` for bash, `check` for assert)
- [x] `bash` steps support an optional `dir` field (relative to project root, defaults to `.`) to control the working directory
- [x] Document the field in `docs/taskmd_specification.md` with examples of both types
- [x] Update the frontmatter schema and field summary table

### Parser
- [x] Extend the task model to include the `verify` field
- [x] Parse `verify` as a list of maps from frontmatter
- [x] Validate each entry has a `type` field and the required fields for that type
- [x] Define a `VerifyStep` struct with `Type` and type-specific fields (e.g., `Run`, `Check`, `Dir`)
- [x] Preserve the field during read/write operations

### CLI Command
- [x] Create `internal/cli/verify.go` with the `verify` command
- [x] Accept `--task-id` flag (required) to identify the task
- [x] Read the task's `verify` list
- [x] Run all commands from the project root (where `.taskmd.yaml` or the `tasks/` directory lives), regardless of the user's cwd — this makes commands like `go test ./internal/...` deterministic
- [x] Dispatch each step by `type`:
  - `bash`: run `run` field in a shell subprocess from `dir` (relative to project root, defaults to `.`), capture stdout/stderr/exit code, report pass/fail
  - `assert`: display `check` field as a check the agent must evaluate (not executed)
  - Unknown types: warn and skip
- [x] Report overall results: executable pass/fail counts + pending assertion checks
- [x] Exit with non-zero status if any executable check fails (useful for CI/scripting)
- [x] Support `--format` flag (json, table) via `GetGlobalFlags()` for consistent output across commands
- [x] JSON output: structured result with step type, status, stdout/stderr per step, and overall pass/fail
- [x] Support `--verbose` flag to show full command output even on success
- [x] Support `--dry-run` flag to display all checks without executing any
- [x] Add comprehensive tests in `internal/cli/verify_test.go`
- [x] Register command with `rootCmd`

### Integration with `set` command
- [x] Add `--verify` flag to `taskmd set` — when combined with `--status completed`, run the verify logic before applying the status change
- [x] If any bash check fails, abort the status change and exit non-zero
- [x] If the task has no `verify` field, proceed with the status change as normal (no-op for verify)

### Agent Skill
- [x] Create `.claude/skills/verify-task/SKILL.md` skill for agents to verify a task
- [x] The skill should accept a task ID, run `taskmd verify --task-id <ID> --format json`, and interpret the results
- [x] For `bash` steps: report pass/fail based on the JSON output
- [x] For `assert` steps: the agent reads each `check` assertion and evaluates whether the current codebase satisfies it
- [x] Report an overall verdict (all checks passed, or list failures)
- [x] Update `.claude/skills/do-task/SKILL.md`: use `taskmd set --task-id <ID> --status completed --verify` in the completion step so verification runs automatically
- [x] Update `.claude/skills/complete-task/SKILL.md`: use `taskmd set --task-id <ID> --status completed --verify` so completing a task always verifies first

### Safety
- [x] Log each command before executing it (no confirmation prompt — the user explicitly asked to verify)
- [x] Support a timeout per command to prevent hangs (default: 60s, configurable with `--timeout`)

## Acceptance Criteria

- Tasks can define a `verify` field with typed check steps:
  ```yaml
  verify:
    - type: bash
      run: "go test ./internal/api/... -run TestPagination"
      dir: "apps/cli"
    - type: bash
      run: "npm test"
      dir: "apps/web"
    - type: assert
      check: "Pagination links appear in the API response headers"
    - type: assert
      check: "Page size defaults to 20 when not specified"
  ```
- Each entry is a map with a `type` field that determines the check kind
- `bash` steps execute the `run` field in a shell subprocess, with optional `dir` (relative to project root, defaults to `.`)
- `assert` steps surface the `check` field for the agent to evaluate (not executed)
- `taskmd verify --task-id 042` runs bash checks and displays assert checks
- Bash checks show pass/fail with exit code; assert checks are listed as pending for agent review
- Unknown types produce a warning and are skipped
- Overall exit code is non-zero if any executable check fails
- `--format json` outputs structured results (step type, status, stdout/stderr, overall pass/fail)
- `--dry-run` lists all checks without running any
- `--verbose` shows full output for all commands
- `--timeout` controls per-command timeout (default 60s)
- All commands run from the project root, regardless of where `taskmd verify` is invoked
- Commands are logged before execution (no confirmation prompt — agents, CI, and developers all want it to just run)
- Tasks without a `verify` field produce a clear "no checks defined" message
- Specification is updated with the new field
- `taskmd set --task-id 042 --status completed --verify` runs verification before applying the status change; aborts if any bash check fails
- A `verify-task` agent skill exists that runs `taskmd verify --format json` and evaluates assert checks
- The `do-task` and `complete-task` skills use `--verify` when marking tasks completed
- Tests cover pass, fail, timeout, dry-run, JSON output, missing field, mixed check types, unknown types, and multi-step scenarios
