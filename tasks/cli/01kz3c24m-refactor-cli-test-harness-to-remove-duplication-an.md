---
id: "01kz3c24m"
title: "Refactor CLI test harness to remove duplication and enable parallelism"
status: completed
priority: medium
effort: large
type: improvement
phase: critical-feedback
dependencies: []
tags: [cli, testing, tech-debt]
created_at: 2026-08-03
completed_at: 2026-08-04
---

# Refactor CLI test harness to remove duplication and enable parallelism

## Objective

The Go test suite is genuine, non-gamed coverage but is structurally bloated: ~2.4:1 test:code
ratio in the CLI package, with an estimated 20-30% avoidable duplication traceable to one design
decision — tests mutate package-global flag vars and swap `os.Stdout` via `os.Pipe`. That forced
**79 near-duplicate helpers** (one `resetXFlags` / `captureXOutput` / `createXTaskFiles` per
command), **298 inline YAML fixtures** hand-written across test files, a flat non-table style
(only 56 `t.Run` subtests across 937 test functions), and **zero `t.Parallel()`** (the suite
cannot parallelize). Fix the root cause and reclaim the LOC without losing coverage.

## Tasks

- [x] Build one shared command-test harness (setup dir, run command, capture stdout/stderr, reset state)
- [x] Replace the ~79 per-command reset/capture/create helpers with the shared harness
- [x] Move the 298 inline YAML task fixtures into shared `testdata/` fixtures
- [x] Convert repetitive single-flag-variant tests into table-driven `t.Run` subtests
- [x] Where feasible, isolate global state so tests can call `t.Parallel()`
- [x] Confirm coverage is unchanged (compare `go test -cover` before/after)

## Acceptance Criteria

- Per-command boilerplate helpers are replaced by a single shared harness
- Inline YAML fixtures are consolidated under `testdata/`
- At least the read-only command tests run with `t.Parallel()`
- Test coverage percentage does not drop after the refactor

## Sub-tasks

This task was split into four focused, sequential slices (each depends on the previous):

1. **`01kz3nm85`** — Build shared CLI command-test harness _(foundation, no deps)_
2. **`01kz3nmbc`** — Consolidate inline YAML test fixtures under `testdata/` _(depends on 1)_
3. **`01kz3nmfy`** — Migrate per-command test helpers to the shared harness _(depends on 1, 2)_
4. **`01kz3nmka`** — Convert repetitive tests to table-driven `t.Run` and enable `t.Parallel()` _(depends on 3)_

The original content above is retained for reference; the acceptance criteria are collectively
satisfied by the four sub-tasks.
