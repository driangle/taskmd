---
id: "01kz3c24b"
title: "Consolidate overlapping read-view commands"
status: cancelled
priority: medium
effort: large
type: improvement
phase: critical-feedback
dependencies: ["01kz3c24f"]
tags: [cli, scope, ux]
created_at: 2026-08-03
cancelled_at: 2026-08-04
---

# Consolidate overlapping read-view commands

## Objective

The CLI exposes many read/aggregate views over the same task set with heavily duplicated
flags: `list`, `board`, `stats`, `report`, `snapshot`, `feed`, `status`, `tags`, `phases`,
`tracks`. `report` vs `snapshot` vs `stats` in particular overlap in intent (summaries /
metrics), and filter/status/priority/phase/scope flags are re-declared per command
(e.g. `list.go:80-89`, `graph.go:85-87`). This inflates the surface a new user must learn and
the maintenance cost of every filter change. Reduce the overlap without dropping real
capability.

## Tasks

- [ ] Map each view command to its distinct value; identify true duplicates vs genuinely
      different outputs
- [ ] Decide a consolidation: e.g. fold `report`/`snapshot`/`stats` into one command with
      `--format`/`--mode`, or clearly document why each stays
- [ ] Extract shared filter flags into a reusable flag set applied across view commands
- [ ] Add deprecation aliases for any removed/renamed command so existing scripts keep working
- [ ] Update docs and help text to reflect the consolidated surface

## Acceptance Criteria

- The number of near-duplicate view commands is reduced, or each is justified in docs
- Filter flags are defined once and shared, not re-declared per command
- Removed/renamed commands still work via hidden deprecation aliases for one release
- Help output and docs reflect the new, smaller surface
