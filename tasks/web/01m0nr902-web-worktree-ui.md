---
title: "Web UI: render worktree provenance and surface the write guard"
id: "01m0nr902"
status: pending
priority: medium
type: feature
tags: ["worktrees", "web", "frontend"]
created: "2026-08-22"
dependencies: ["01m0n67bx"]
effort: small
phase: worktree-support
---

# Web UI: render worktree provenance and surface the write guard

## Objective

Render the worktree overlay in the React frontend: provenance on tasks, per-worktree
detail, an active-overlay indicator, and a visible error when a mutation is blocked
by the sibling-only write guard. Consumes the API fields added by `01m0n67bx`.
Spec: `docs/specs/worktree-support.md` §9.

## Tasks

- [ ] Extend the TypeScript API types with the provenance fields (`effective_status`, `effective_owner`, `worktree`, `branch`, `remote_only`)
- [ ] Worktree badge on task list rows and board cards when a task's winning copy is remote
- [ ] Task detail page: per-worktree copies section (status/owner/branch per copy) when copies differ, mirroring `get`
- [ ] Header indicator when the overlay is active ("worktree `agent-b` — 3 siblings")
- [ ] Surface the sibling-only guard error from mutation responses as a visible failure (toast/inline error), never a silent no-op
- [ ] Component tests for badge, detail section, and guard-error rendering

## Acceptance Criteria

- In a multi-worktree repo, list/board/detail show provenance and the header indicator; in a single-worktree repo the UI is unchanged
- A blocked status edit shows the guard error naming the worktree
- `pnpm run typeCheck` and `pnpm run lint` pass; readonly mode is unaffected
