---
id: "01kz3c24r"
title: "Extract a shared scanTasks helper to remove scan boilerplate"
status: pending
priority: low
effort: small
type: improvement
phase: critical-feedback
dependencies: []
tags: [cli, refactor, dry]
created_at: 2026-08-03
---

# Extract a shared scanTasks helper to remove scan boilerplate

## Objective

The identical scan block — `scanner.NewScanner(scanDir, flags.Verbose, flags.IgnoreDirs)` +
`Scan()` + error-wrap — is repeated ~26 times across command files. A single `scanTasks(args)`
helper (resolving the scan dir, constructing the scanner, scanning, and wrapping the error)
would remove the copy-paste and give one place to evolve scan behavior. This is minor,
low-risk debt that also shrinks the per-command test surface once centralized.

## Tasks

- [ ] Add a `scanTasks(cmd, args) ([]Task, error)` (or similar) helper in the `cli` package
- [ ] Have it resolve the scan dir, build the scanner, scan, and wrap errors consistently
- [ ] Replace the ~26 duplicated scan blocks across command files with calls to the helper
- [ ] Relocate the misplaced generic `levenshtein` util out of `get.go` into a string-util home
- [ ] Run the full CLI test suite to confirm no behavior change

## Acceptance Criteria

- The duplicated scanner-construction block exists in exactly one place
- All commands scan via the shared helper with identical behavior and error wrapping
- `levenshtein` lives in a sensible utility file, not in `get.go`
- All existing tests pass
