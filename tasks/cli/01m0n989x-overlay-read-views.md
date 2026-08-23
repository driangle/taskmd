---
title: "Extend worktree overlay to remaining read views and validate"
id: "01m0n989x"
status: completed
priority: high
type: feature
tags: ["worktrees", "overlay"]
created: "2026-08-22"
dependencies: ["01m0nk00d"]
effort: medium
phase: worktree-support
completed_at: 2026-08-23
---

# Extend worktree overlay to remaining read views and validate

## Objective

Extend the worktree overlay from `next`/`list` to the remaining read views and to
validation, so every presentation surface reports effective (cross-worktree) status
with provenance. Spec: `docs/specs/worktree-support.md` §4, §8.

## Tasks

- [x] `board`, `stats`, `graph`, `report`, `metrics`, `tracks`, `phases` operate on effective status when the overlay is active
- [x] `get`: show the local copy plus a `Worktrees:` section listing each copy's status/owner/branch when copies differ
- [x] `validate`: warn on divergent terminal states across worktrees; keep duplicate-ID errors scoped to a single worktree
- [x] Attribute same-ID copies whose path lies inside a sibling worktree root (non-hidden checkout inside the scan root) to that worktree instead of flagging duplicates (§8)
- [x] Unit tests per view for effective-status rendering; e2e spot-checks for `board` and `get`

## Acceptance Criteria

- A task completed only in a sibling worktree appears as completed in `board`/`stats` and as done in `graph`, with provenance where the view shows detail
- `get` on a task with diverging copies lists every worktree's status/owner/branch
- `taskmd validate` in a multi-worktree repo emits the divergent-terminal-state warning and no false duplicate-ID errors
- All views are unchanged when the overlay is inactive
