---
id: "01m3xf0rs"
title: "Make strict phase ordering the default for next and deprecate --strict-phases"
status: completed
priority: medium
effort: medium
type: feature
dependencies: ["01m3xkmtn"]
tags: ["cli", "next"]
created_at: 2026-10-02
completed_at: 2026-10-02
---

# Make strict phase ordering the default for next and deprecate --strict-phases

## Objective

When a project declares an ordered `phases` list in `.taskmd.yaml`, that order
is explicit user intent. Today `next` treats phase only as a score bonus, so a
high-priority task in a later phase can outrank work in an earlier phase unless
the user remembers `--strict-phases`. Make strict phase tiering the default on
every `next` surface and deprecate the flag.

Projects with no `phases` configured see no change.

## Design

- **The default lives in the CLI layer, not the SDK.** `sdk/go/next.Options.StrictPhases`
  stays an opt-in bool; `apps/cli` (CLI, MCP, web) passes `true`. Invocation
  defaults belong in `apps/cli` per ADR 0006, and this avoids a breaking SDK
  change (no `sdk-bump: minor`).
- **Deprecate, don't remove.** Keep `--strict-phases` and mark it with
  `MarkDeprecated` (it still parses, prints a warning, and is hidden from help).
  `--strict-phases=false` is the escape hatch back to score-only phase ranking.
- **Unphased tasks join the current phase tier.** Today, under strict mode,
  tasks with no `phase` sort after *every* phased task. With strict as the
  default that would bury e.g. an unphased critical bug behind all phased
  chores. Instead, a task with no `phase` is placed in the tier of the
  **earliest phase that still has actionable tasks** (the "current" phase) and
  competes there on score. Tasks whose `phase` is not in the configured list
  (a config error the validator should already flag) still sort last.
  This SDK change is behavioural only (no API change) — a patch bump, but call
  it out in the commit message.

## Tasks

- [x] SDK: place unphased tasks in the current-phase tier under `StrictPhases`;
      keep unknown-phase tasks last; update `sdk/go/next` tests
- [x] CLI: default strict phase ordering on; mark `--strict-phases` deprecated;
      honour `--strict-phases=false` as the opt-out
- [x] MCP, web `/api/next` and the static export (`web/export.go`): pass
      `StrictPhases: true` (phase order is wired by 01m3xkmtn); `next
      --all-projects` inherits the CLI flag default via `recommendForProject`
- [x] Update `next` long help: remove `--strict-phases` examples, describe the
      default tiering and the `--strict-phases=false` opt-out; update the
      `--strict-priority` text ("phase is primary" now applies by default)
- [x] Flip `TestNext_StrictPhasesOff_DefaultBehavior` to assert the new default;
      add tests for the deprecation warning, `--strict-phases=false`, and
      unphased-task placement
- [x] Update `apps/docs/guide/cli.md` flag table
- [x] Note the behaviour change for the next release notes (recorded in the
      commit body; the repo has no CHANGELOG — release notes are written from
      commit messages via `scripts/release.sh --notes-file`)

## Acceptance Criteria

- In a project with phases, `taskmd next` (no flags) never ranks an actionable
  later-phase task above an actionable earlier-phase task
- An unphased task ranks alongside current-phase tasks by score, not after all
  phased tasks
- `taskmd next --strict-phases` still works and prints a deprecation warning
- `taskmd next --strict-phases=false` restores score-only phase ranking
- MCP `next`, web `/api/next` and the static export produce the same ordering
  as the CLI default; `next --all-projects` applies strict phase tiers within
  each project by default
- Projects without `phases` configured produce identical output to before
