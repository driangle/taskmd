---
id: "01kz3c24m"
title: "Refactor CLI test harness to remove duplication and enable parallelism"
status: pending
priority: medium
effort: large
type: improvement
phase: critical-feedback
dependencies: []
tags: [cli, testing, tech-debt]
created_at: 2026-08-03
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

- [ ] Build one shared command-test harness (setup dir, run command, capture stdout/stderr, reset state)
- [ ] Replace the ~79 per-command reset/capture/create helpers with the shared harness
- [ ] Move the 298 inline YAML task fixtures into shared `testdata/` fixtures
- [ ] Convert repetitive single-flag-variant tests into table-driven `t.Run` subtests
- [ ] Where feasible, isolate global state so tests can call `t.Parallel()`
- [ ] Confirm coverage is unchanged (compare `go test -cover` before/after)

## Acceptance Criteria

- Per-command boilerplate helpers are replaced by a single shared harness
- Inline YAML fixtures are consolidated under `testdata/`
- At least the read-only command tests run with `t.Parallel()`
- Test coverage percentage does not drop after the refactor
