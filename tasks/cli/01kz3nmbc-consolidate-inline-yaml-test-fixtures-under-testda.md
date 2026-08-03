---
id: "01kz3nmbc"
title: "Consolidate inline YAML test fixtures under testdata/"
status: pending
priority: medium
effort: medium
parent: "01kz3c24m"
phase: critical-feedback
dependencies: ["01kz3nm85"]
tags: ["cli", "testing", "tech-debt"]
created_at: 2026-08-03
---

# Consolidate inline YAML test fixtures under testdata/

## Objective

Move the ~298 inline YAML task fixtures that are hand-written across the CLI test files into
shared fixtures under `internal/cli/testdata/`, loaded through the harness from `01kz3nm85`. This
removes the largest single source of duplicated bytes in the test suite and makes fixtures
reviewable in one place.

## Tasks

- [ ] Inventory the inline task-YAML fixtures and cluster them into a small set of reusable
      canonical task sets (e.g. a standard 3-task dependency chain, a mixed-status set, a
      phases set, a scoped-projects set)
- [ ] Create fixture files under `internal/cli/testdata/` (one task file per fixture, or grouped
      fixture dirs the temp-repo builder can copy wholesale)
- [ ] Extend the harness temp-repo builder to seed a repo from a named `testdata/` fixture set
- [ ] Replace inline fixtures in the pilot files with fixture references to validate the loader
- [ ] Keep genuinely one-off fixtures inline where a shared fixture would obscure the test's intent
      (document the judgment call, don't force-share)

## Acceptance Criteria

- Common inline YAML fixtures are consolidated under `internal/cli/testdata/`
- The harness can seed a temp repo from a named fixture set
- `go test ./internal/cli/...` passes and coverage does not drop vs. baseline
