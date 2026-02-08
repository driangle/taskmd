---
id: "019"
title: "validate command - Lint and validate tasks"
status: pending
priority: high
effort: medium
dependencies: ["017"]
tags:
  - cli
  - go
  - commands
  - validation
created: 2026-02-08
---

# Validate Command - Lint and Validate Tasks

## Objective

Implement the `validate` command to lint task files and check for common errors like missing dependencies, cycles, duplicate IDs, and invalid fields.

## Tasks

- [ ] Create `internal/cli/validate.go` for validate command
- [ ] Create `internal/validator/` package for validation logic
- [ ] Implement validation checks:
  - Missing dependencies (refs to non-existent task IDs)
  - Circular dependencies (detect cycles)
  - Duplicate task IDs
  - Invalid field values (status, priority, effort)
  - Required fields (id, title)
- [ ] Implement `--strict` flag for stricter validation
- [ ] Support output formats: `text` (default), `json`
- [ ] Return appropriate exit codes (0 = valid, 1 = errors, 2 = warnings in strict mode)
- [ ] Display clear error messages with file paths and line numbers if possible

## Acceptance Criteria

- `taskmd validate` checks all tasks and reports errors
- Detects circular dependencies
- Detects missing dependency references
- Detects duplicate IDs
- `--strict` flag enables additional checks
- `--format json` outputs structured validation results
- Exit code reflects validation status

## Examples

```bash
taskmd validate
taskmd validate tasks.md --strict
taskmd validate --format json
cat tasks.md | taskmd validate --stdin
```
