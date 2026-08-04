---
id: "01kz3nmfy"
title: "Migrate per-command test helpers to the shared harness"
status: completed
priority: medium
effort: large
parent: "01kz3c24m"
phase: critical-feedback
dependencies: ["01kz3nm85", "01kz3nmbc"]
tags: ["cli", "testing", "tech-debt"]
created_at: 2026-08-03
completed_at: 2026-08-04
---

# Migrate per-command test helpers to the shared harness

## Objective

Migrate the CLI command test files (~55 files, 939 test functions) off their per-command
boilerplate helpers and onto the shared harness (`01kz3nm85`) and consolidated fixtures
(`01kz3nmbc`). This is the bulk-churn slice of the parent refactor: delete the ~79 duplicated
`resetXFlags` / `captureXOutput` / `createXTestFiles` helpers and rewrite their call sites to use
the harness. Assertions and coverage must be preserved — this is a mechanical restructuring, not a
test rewrite.

## Tasks

- [x] Migrate command test files in batches (group by command), replacing `captureXOutput` with the
      harness run+capture and `resetXFlags`/`resetViper` with the harness reset
- [x] Replace `createXTestFiles` call sites with the harness temp-repo builder / fixture loader
- [x] Delete the now-dead per-command helpers as each file is migrated
- [x] Remove `os.Pipe`-based stdout swapping from the test files (currently ~35 files)
- [x] Run the full suite (`go test ./internal/cli/...` and `make e2e`) after each batch to catch
      regressions early

## Acceptance Criteria

- Per-command `resetX`/`captureX`/`createX` boilerplate helpers are removed in favor of the
  shared harness
- No test file swaps `os.Stdout`/`os.Pipe` directly anymore
- `go test ./internal/cli/...` and `make e2e` pass; coverage does not drop vs. baseline
