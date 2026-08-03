---
id: "01kz3c24f"
title: "Define scope boundary for sync and source-TODO scanner"
status: pending
priority: high
effort: medium
type: improvement
phase: critical-feedback
dependencies: []
tags: [scope, architecture, decision]
created_at: 2026-08-03
---

# Define scope boundary for sync and source-TODO scanner

## Objective

taskmd's core is "markdown task files + a CLI to read them," but the codebase has expanded well
past that: `internal/sync` (~4,700 LOC) imports from Jira / Linear / Trello / GitHub, and
`internal/todos` scans **source-code comments** for TODO/FIXME — a different domain from
markdown task files. Each is self-contained and doesn't destabilize the core, but together they
mean one maintainer (52 commits, ~4 months) is carrying a platform's worth of surface. This
task is a decision, not a code change: draw an explicit line between "core" and "optional," so
future work has a boundary to respect.

## Tasks

- [ ] Write down the definition of "core taskmd" vs "optional/experimental" modules
- [ ] Decide the home for `sync` (Jira/Linear/Trello/GitHub): core, optional build tag, or plugin
- [ ] Decide the home for the source-TODO scanner (`todos`): core or optional
- [ ] Capture the decision in a short ADR or a section in `AGENTS.md`
- [ ] Add follow-up tasks for any move (e.g. gate `sync` behind a build tag or split into a
      separate module) if that is the decision
- [ ] Align the `.taskmd.yaml` phases (external-integrations, etc.) with the decided scope

## Acceptance Criteria

- A written, discoverable statement of what belongs in core vs optional exists
- `sync` and `todos` each have a decided, documented home
- New feature proposals can be evaluated against the stated scope boundary
