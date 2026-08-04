---
id: "01kz3nmka"
title: "Convert repetitive tests to table-driven t.Run and enable t.Parallel"
status: completed
priority: medium
effort: medium
parent: "01kz3c24m"
phase: critical-feedback
dependencies: ["01kz3nmfy"]
tags: ["cli", "testing", "tech-debt"]
created_at: 2026-08-03
completed_at: 2026-08-04
---

# Convert repetitive tests to table-driven t.Run and enable t.Parallel

## Objective

With the harness and migration in place, collapse the flat, repetitive single-flag-variant tests
into table-driven `t.Run` subtests (the suite currently has only 56 `t.Run` across 939 test
functions) and enable `t.Parallel()` where global state has been isolated. This closes out the
parent refactor's parallelism and readability goals and gates on coverage being preserved.

## Tasks

- [x] Identify clusters of near-duplicate tests that differ only by a flag/format value
      (e.g. `--format text/json/yaml`, status filters) and convert them to table-driven `t.Run`
      — collapsed the 10 `TestAnalyzeDiff_*` funcs in `feed_diff_test.go` into one table, and the
      `get` text/json/yaml/unsupported-format funcs into `TestGet_Formats` (the cited exemplar file).
- [x] Enable `t.Parallel()` on the safely-isolatable tests: pure-function unit tests (formatting,
      colors, fuzzy scoring, tablewriter, etc.) and read-only command tests that no longer depend
      on shared mutable global state
      — parallelized all pure-function tests across `feed_diff`, `filter`, `format`, `suggest`,
      `input`, `tablewriter`, `spec_reference` and the renderer-based subset of `colors` (1 → 43
      `t.Parallel()` sites). No command test qualifies: they all still depend on shared global state
      (see next item).
- [x] Note (in the task or code comments) which tests cannot yet be parallelized because they still
      mutate package-global flag vars, so a future flag-architecture change can target them
      — documented in `harness_test.go` (the two blockers: `os.Stdout` swap in `captureOutput` and
      cobra flags bound to package-global vars reset by `resetCLIState`) and in `colors_test.go`
      (the `noColor`/`forceColor`/`NO_COLOR` mutating tests). Command tests stay serial by design
      until flag state moves off package globals into per-invocation context.
- [x] Capture a `go test -cover` baseline before the change and confirm the final coverage
      percentage is unchanged — baseline 82.2% held exactly (`-race` green).

## Acceptance Criteria

- Repetitive single-flag tests are expressed as table-driven `t.Run` subtests
- At least the read-only command tests (and pure-function unit tests) run with `t.Parallel()`
- `go test ./internal/cli/...` passes under `-race` and coverage does not drop vs. baseline
