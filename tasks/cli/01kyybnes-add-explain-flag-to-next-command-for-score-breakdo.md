---
title: "Add --explain flag to next command for score breakdown"
id: "01kyybnes"
status: completed
priority: medium
type: feature
tags: ["cli", "next", "scoring"]
created: "2026-08-01"
completed_at: 2026-08-01
---

# Add --explain flag to next command for score breakdown

## Objective

The `next` command scores actionable tasks using a weighted model (priority,
critical-path position, downstream impact, effort, and phase ordering — see
`sdk/go/next/next.go`), but the table output only shows a terse comma-joined
`reason` column and a flat `score`. Users can't see *how* a score was reached,
which makes it hard to understand why one task outranks another (e.g. why a
medium-priority task in an early phase beats a high-priority task with no phase,
or why the downstream bonus is scaled by a multiplier).

Add an `--explain` flag to `taskmd next` that surfaces a per-task, itemized
breakdown of every scoring component and its point contribution, so the ranking
is transparent and debuggable.

## Tasks

- [x] Change `ScoreTask` and `scorePhase` in `sdk/go/next/next.go` to return
      itemized components — a `(label, points)` pair per contribution — instead
      of (or in addition to) the flat `reasons []string`. Preserve the existing
      `Reasons` behaviour for backward compatibility.
- [x] Add a `ScoreBreakdown []ScoreComponent` field to the `Recommendation`
      struct (with `json`/`yaml` tags), populated in `buildRecommendations`.
- [x] Include the downstream-priority multiplier in the breakdown so scaled
      bonuses (critical-path and downstream) are explainable (e.g. show base,
      multiplier, and resulting points).
- [x] Add the `--explain` bool flag to `nextCmd` in `apps/cli/internal/cli/next.go`.
- [x] When `--explain` is set with `--format table`, render a per-task breakdown
      block (component lines + total) beneath each recommendation instead of the
      compact table row.
- [x] For `--format json`/`--format yaml`, emit the structured `score_breakdown`
      array alongside the existing `reasons`/`score` fields (independent of the
      `--explain` flag, or gated on it — pick one and document it).
- [x] Add comprehensive tests: unit tests for the itemized scoring in
      `sdk/go/next/next_test.go`, and CLI tests for the `--explain` output in
      `apps/cli/internal/cli/next_test.go` (table + json breakdown, and that the
      component points sum to the total score).
- [x] Update `apps/docs/guide/cli.md` and the `next` command long help/examples
      to document `--explain`.

## Acceptance Criteria

- `taskmd next --explain` prints, for each recommended task, an itemized list of
  scoring components (priority, phase, critical path, downstream, effort) with
  each component's point value, and a total that equals the task's `score`.
- The breakdown makes the downstream/critical-path multiplier visible so scaled
  bonuses are explainable.
- `taskmd next --explain --format json` (or `next --format json`) includes a
  structured `score_breakdown` field whose components sum to `score`.
- `taskmd next` without `--explain` produces byte-for-byte the same table output
  as before (no regression).
- Existing `Reasons` output is unchanged for callers that rely on it.
- New unit and CLI tests cover the itemized scoring and `--explain` output for
  table and json formats; `make test`, `make lint`, and `make build` pass.
- Docs (`apps/docs/guide/cli.md` and command help) document the new flag.
