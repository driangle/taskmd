---
id: "01m3xkmtn"
title: "Apply phase order consistently across all next surfaces"
status: completed
priority: high
effort: small
type: bug
dependencies: []
tags: ["cli", "next", "mcp", "web"]
created_at: 2026-10-02
completed_at: 2026-10-02
---

# Apply phase order consistently across all next surfaces

## Objective

Only `taskmd next` (single project) passes the project's phase order into
`next.Recommend`. Every other recommendation surface omits `PhaseOrder`, so
phases are ignored there entirely: no phase score bonus, and strict phase
ordering cannot take effect. The same task set is therefore ranked differently
depending on whether you ask the CLI, the MCP server, or the web UI.

Affected call sites:

- `apps/cli/internal/mcp/next.go` (MCP `next` tool) — no `PhaseOrder`
- `apps/cli/internal/web/handlers.go` (`/api/next`) — no `PhaseOrder`
- `apps/cli/internal/web/export.go` (static export) — no `PhaseOrder`
- `apps/cli/internal/cli/next.go` `recommendForProject` (`next --all-projects`)
  — passes `StrictPhases` but not `PhaseOrder`, so `--strict-phases` is
  currently a silent no-op there

This is a prerequisite for making strict phase ordering the default (01m3xf0rs):
flipping the default only in the CLI would widen the CLI-vs-MCP/web gap.

## Tasks

- [x] Pass the project's phase order to `next.Recommend` from the MCP `next` tool
- [x] Pass it from the web `/api/next` handler and the static export
- [x] For `--all-projects`, load each project's **own** phase order from that
      project's config (not the current directory's viper config —
      `loadPhaseOrder()` reads global viper state today)
- [x] Add tests per surface proving a phased project gets the phase bonus
      (and, for all-projects, that `--strict-phases` now orders within each project)

## Acceptance Criteria

- MCP `next`, web `/api/next`, the export, and `next --all-projects` rank a
  phased project's tasks the same way `taskmd next` does for the same inputs
- `next --all-projects --strict-phases` applies phase tiers within each project;
  cross-project merging is unchanged (phase names are project-local, so there
  is no cross-project phase ordering)
- Projects without `phases` in config are unaffected
