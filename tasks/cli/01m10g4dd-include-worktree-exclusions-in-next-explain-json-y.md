---
title: "Include worktree exclusions in next --explain JSON/YAML output"
id: "01m10g4dd"
status: completed
priority: low
type: feature
tags: ["worktrees"]
created: "2026-08-27"
completed_at: 2026-08-27
---

# Include worktree exclusions in next --explain JSON/YAML output

## Objective

Spec §4 (`docs/specs/worktree-support.md`): with the overlay active,
`next --explain` names the worktree that excludes a task. Today
`printWorktreeExclusions` is only reached from the `table` branch of
`runNextCommand` (`apps/cli/internal/cli/next.go` around lines 171–201), so
`--format json` / `--format yaml` — the formats agents actually consume — lose
the exclusion provenance entirely.

## Tasks

- [x] Add worktree exclusions to the JSON/YAML `--explain` payload as
      structured fields (e.g. an `excluded` list with `id`, `reason`,
      `worktree`, `branch`), not preformatted strings
- [x] Keep the table output unchanged
- [x] Emit the fields only when the overlay is active and `--explain` is set,
      so existing consumers see no shape change otherwise
- [x] Unit tests: `next --explain --format json` and `--format yaml` with a
      sibling-excluded task assert the exclusion entry; without `--explain` or
      without overlay the field is absent

## Acceptance Criteria

- `taskmd next --explain --format json` includes which worktree/branch
  excluded each skipped task, matching the information the table view prints
- Output shape without `--explain`, or in non-overlay repos, is unchanged
- Tests pass: `make test`, `make lint`
