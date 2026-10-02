---
title: "Add a .taskmd.yaml setting for where next ranks unphased tasks"
id: "01m3ygtgq"
status: completed
priority: low
type: feature
tags: ["cli", "next", "config"]
created: "2026-10-02"
dependencies: ["01m3xf0rs"]
effort: small
completed_at: 2026-10-02
---

# Add a .taskmd.yaml setting for where next ranks unphased tasks

## Objective

Since 01m3xf0rs, strict phase ordering is the default for `next`, and a task
with no `phase` joins the *current* phase tier (the earliest phase with
actionable tasks) and competes there on score. Some projects use "unphased" to
mean "unscheduled" and want those tasks ranked after all phased work, which was
the old `--strict-phases` behaviour. Today the only workaround is a catch-all
last phase (e.g. `backlog`) assigned task by task.

Add a project-level setting so users can choose where unphased tasks rank:

```yaml
# .taskmd.yaml
next:
  unphased: current   # default — join the current phase tier
  # unphased: last    # rank after every configured phase, before unknown-phase tasks
```

## Design notes

- **SDK** (`sdk/go/next`): add an option (e.g. `Options.UnphasedPlacement`)
  whose zero value is today's behaviour, so the change is additive (patch
  bump). `assignPhaseTiers` puts unphased tasks at tier `len(phases)` for
  `last`, with unknown-phase tasks still after them.
- **CLI layer** (`apps/cli`, per ADR 0006): read and validate the config key,
  then pass it on every `next` surface: CLI `next`, `next --all-projects` (from
  each project's own `.taskmd.yaml`), the MCP `next` tool, web `/api/next`, and
  the static export.
- It only matters under strict phase ordering. With `--strict-phases=false` or
  no `phases` configured, it has no effect.
- Check the config key name against existing `.taskmd.yaml` conventions before
  settling on `next.unphased`.

## Tasks

- [x] SDK: add the unphased-placement option to `next.Options`, honour it in
      `assignPhaseTiers`, and add tests for both placements
- [x] Config: parse and validate the key (reject unknown values with a clear
      error), defaulting to `current`
- [x] Wire it through CLI `next`, `--all-projects` (per-project config), MCP
      `next`, web `/api/next` and the static export
- [x] CLI tests covering both values, the default, and an invalid value
- [x] Document the setting in the `.taskmd.yaml` config reference, the
      `apps/docs/guide/cli.md` "Phase ordering" section, and the `next` long help

## Acceptance Criteria

- With `next.unphased: last`, `taskmd next` ranks every unphased task after all
  actionable tasks in configured phases, and before tasks with an unknown phase
- With the key absent or set to `current`, ordering is identical to today's
- An invalid value produces a clear config error
- MCP `next`, web `/api/next`, the static export and `next --all-projects`
  honour the setting the same way the CLI does
- Projects without `phases`, and runs with `--strict-phases=false`, are unaffected
