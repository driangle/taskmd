---
title: "Add --strict-priority flag to next command"
id: "01kyy9qxg"
status: pending
priority: medium
type: feature
tags: ["cli", "next"]
created: "2026-08-01"
---

# Add --strict-priority flag to next command

## Objective

Add a `--strict-priority` flag to the `next` command that guarantees tasks are
tiered by priority as the primary sort key, mirroring the existing
`--strict-phases` flag.

Today priority dominates scoring but does not *guarantee* ordering: because
graph, effort, and phase bonuses stack, a lower-priority task can outrank a
higher-priority one (e.g. a `medium` task with phase + critical-path +
downstream + small-effort bonuses can score 80, outranking a bonus-less
`critical` task at 40). This flag gives users a hard guarantee that no
lower-priority task is ranked above an actionable higher-priority one. The
existing per-task score is retained as the tiebreaker *within* each priority
tier, so the smart ordering still applies inside a level.

Unlike `--priority <value>` (which *filters* the candidate set, dropping
non-matching tasks), `--strict-priority` keeps all actionable tasks and only
changes their ordering.

## Interaction with `--strict-phases`

When both `--strict-phases` and `--strict-priority` are set, **phase is checked
first, then priority**: earlier-phase tasks always rank ahead, and within the
same phase, higher-priority tasks rank ahead, with score breaking ties within a
priority tier. This ordering must be documented in the flag help text and the
command's `Long` description.

Note: `priorityWeight` (`sdk/go/next/next.go`) maps both `low` and unset
priority to weight `1`, so they tie under strict-priority and fall through to
the score/ID tiebreak — this is acceptable and intended.

## Tasks

- [ ] Add `StrictPriority bool` to `next.Options` in `sdk/go/next/next.go`
- [ ] Extend the sort in `scoreAndSort` to apply the priority tier (using
      existing `priorityWeight`) after the strict-phases check and before the
      score comparison
- [ ] Add `nextStrictPriority` flag var and register `--strict-priority` in
      `internal/cli/next.go`, wiring it into `next.Options`
- [ ] Document the flag and its interaction with `--strict-phases` in the flag
      help text and the `next` command `Long` description / examples
- [ ] Add tests: happy path (priority tiers ordered strictly), score-as-tiebreak
      within a tier, low/unset tie behavior, and the `--strict-phases` +
      `--strict-priority` interaction (phase primary, priority secondary)
- [ ] Update docs if `next` flags are documented in `apps/docs`

## Acceptance Criteria

- `taskmd next --strict-priority` ranks all actionable tasks strictly by
  priority tier (critical > high > medium > low/unset), with the existing score
  breaking ties within each tier
- `--priority` filtering behavior is unchanged and independent of the new flag
- With both `--strict-phases` and `--strict-priority`, phase is the primary sort
  key and priority is secondary; this is covered by a test
- Flag help text and the command `Long` description document the flag and the
  phase-then-priority interaction
- All new behavior is covered by tests per the CLI testing policy; `make test`
  and `make lint` pass
