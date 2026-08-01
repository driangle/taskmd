# Worklog: 01kyyampc — Show unassigned tasks category in phases command

## 2026-08-01T00:00:00Z

Started work.

**Approach:** Add a synthetic `(unassigned)` category to `internal/cli/phases.go`
aggregating all tasks with `task.Phase == ""`. Reuse the existing `PhaseSummary`
struct and counting rules. To avoid duplicating the counting logic between real
phases and the synthetic one, extract shared `countTask` / `finalizeProgress`
helpers. Append the unassigned summary after configured phases, only when its
active task count > 0. Stable id `unassigned`, display name `(unassigned)`.

**Next:** Implement helpers, wire into runPhases, add tests, update help/docs.

## 2026-08-01T00:30:00Z

Completed.

**Implementation (`internal/cli/phases.go`):**

- Extracted shared `countTask` (folds one task into the status breakdown /
  active / done counts; cancelled tasks appear in `ByStatus` but are excluded
  from counts) and `finalizeProgress` helpers, so the synthetic category and
  configured phases share identical counting logic.
- Added `computeUnassignedSummary(tasks)` building a `PhaseSummary` from all
  `task.Phase == ""` tasks. Stable id `unassigned`, display name `(unassigned)`
  (defined as constants `unassignedPhaseID` / `unassignedPhaseName`).
- In `runPhases`, append the unassigned summary after the configured phases,
  only when `unassigned.Tasks > 0` (active count) — so a project with only
  cancelled unassigned tasks, or a fully-assigned project, shows no row.
- Orphaned-phase warning path left untouched.

**Docs:** Updated the `phases` command `Long` help text and
`apps/docs/guide/cli.md` to document the `(unassigned)` category.

**Tests (`phases_test.go`):** Added coverage for row-shown (table),
row-omitted-when-all-assigned, progress/count correctness + ordering-last +
cancelled-exclusion, row-omitted-when-only-cancelled, and YAML output. Updated
the pre-existing `TestPhases_JSONOutput` (its fixture task 005 has no phase, so
it now expects the extra unassigned summary).

**Verification:** `make test` and `make lint` pass. Manually verified table +
JSON output against a temp project — unassigned row shows correct counts with
cancelled excluded from the total but visible in `by_status`.

