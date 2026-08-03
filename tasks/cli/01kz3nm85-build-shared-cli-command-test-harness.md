---
id: "01kz3nm85"
title: "Build shared CLI command-test harness"
status: completed
priority: medium
effort: medium
parent: "01kz3c24m"
phase: critical-feedback
dependencies: []
tags: ["cli", "testing", "tech-debt"]
created_at: 2026-08-03
completed_at: 2026-08-03
---

# Build shared CLI command-test harness

## Objective

Build one shared command-test harness for the `internal/cli` package that owns the three
responsibilities currently duplicated across ~79 per-command helpers: temp-repo setup, command
execution with stdout/stderr capture, and global-state reset. This is the foundation the rest of
the parent refactor (`01kz3c24m`) builds on — nothing else can be migrated until it exists.

The root cause being fixed: tests mutate package-global flag vars and swap `os.Stdout` via
`os.Pipe`. The harness centralizes both so the reset/capture logic lives in exactly one place.

## Tasks

- [x] Create `internal/cli/harness_test.go` with a `runCLI(t, args...)` helper that resets all
      global flag/viper state, executes the command (via `rootCmd`/`RunE`), captures stdout and
      stderr, and returns a small result struct (`{Stdout, Stderr, Err}`) — implemented as
      `taskRepo.Run(args...)` returning `cliResult{Stdout, Stderr, Err}`
- [x] Centralize the global-flag/viper reset (currently 35 `resetXFlags` + `resetViper`) into a
      single canonical reset invoked by the harness — `resetCLIState` + `resetFlagTree`
- [x] Add a temp-repo builder helper (`newTaskRepo(t, ...)`) that creates a `t.TempDir()`, writes
      task files, and points `taskDir` at it — replacing the ~50 `createXTestFiles` helpers
- [x] Prove the harness on 1–2 pilot command test files (`get_test.go` fully migrated; 2 get
      tests in `duplicate_test.go` migrated) without changing their assertions
- [x] Document the harness API with short doc comments so the migration task can follow the pattern

## Acceptance Criteria

- A single `harness_test.go` provides run+capture+reset and a temp-repo builder
- At least one command test file is migrated to the harness as a working proof
- `go test ./internal/cli/...` passes and `go test -cover` shows no coverage drop vs. baseline
- The harness API is documented well enough to drive the migration in `01kz3nmfy`
