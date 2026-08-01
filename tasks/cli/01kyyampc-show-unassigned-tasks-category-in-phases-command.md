---
title: "Show unassigned tasks category in phases command"
id: "01kyyampc"
status: completed
priority: medium
type: feature
tags: ["cli", "phases"]
created: "2026-08-01"
completed_at: 2026-08-01
---

# Show unassigned tasks category in phases command

## Objective

The `phases` command (`internal/cli/phases.go`) currently has a visibility gap: tasks whose `phase` field is empty (`task.Phase == ""`) are silently skipped and never counted in any summary. The command surfaces tasks assigned to a known phase, and warns (on stderr) about tasks referencing an *undefined* phase, but tasks with no phase at all are invisible.

Add a synthetic `(unassigned)` category to the `phases` output that aggregates every task with an empty phase, so users can see how much work has not yet been scheduled into any phase. This is distinct from the existing orphaned-phase warning (which flags tasks pointing at a phase id that doesn't exist — likely a typo/config drift).

## Design Decisions

- **Reuse `PhaseSummary`** — the synthetic category uses the same struct and counting logic (cancelled tasks excluded, `Done`/`Progress` computed identically), so table, JSON, and YAML output all handle it for free.
- **Append after real phases** — the `(unassigned)` row appears last, after all configured phases.
- **Only show when non-zero** — omit the row entirely when there are no unassigned tasks, to avoid noise in a fully-assigned project.
- **Stable id for consumers** — use a fixed, non-colliding id (e.g. `unassigned`; the display name can be `(unassigned)`). Parentheses aren't valid phase-id characters, so there's no collision risk with a real phase.

## Tasks

- [x] Add a helper (e.g. `computeUnassignedSummary`) that builds a `PhaseSummary` from all tasks where `task.Phase == ""`, reusing the same counting rules as `computePhaseSummaries`
- [x] Append the unassigned summary to the summaries slice in `runPhases`, only when its task count > 0
- [x] Ensure JSON/YAML output includes the synthetic entry with a stable `id`
- [x] Confirm the table renderer displays it correctly (name, tasks, done, progress, due `-`)
- [x] Add tests in `phases_test.go`: unassigned tasks present → row shown; none present → row absent; cancelled unassigned tasks excluded from count; progress percentage correct; JSON/YAML include the entry
- [x] Update `phases` command help/Long text and any relevant docs to mention the unassigned category

## Acceptance Criteria

- Running `taskmd phases` in a project with tasks that have no `phase` shows an `(unassigned)` row with correct task count, done count, and progress
- The `(unassigned)` row is omitted when every task is assigned to a phase
- Cancelled unassigned tasks are excluded from the count (consistent with configured phases)
- JSON and YAML output include the unassigned entry with a stable, documented `id`
- The existing orphaned-phase warning behavior is unchanged
- New tests cover: presence/absence of the row, cancelled exclusion, progress calculation, and JSON/YAML output
- `make test` and `make lint` pass
