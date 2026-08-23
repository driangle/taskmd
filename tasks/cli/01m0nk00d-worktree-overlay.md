---
title: "Worktree overlay merge layer wired into next and list"
id: "01m0nk00d"
status: completed
priority: critical
type: feature
tags: ["worktrees", "overlay", "next"]
created: "2026-08-22"
dependencies: ["01m0n8cpm"]
effort: large
phase: worktree-support
completed_at: 2026-08-23
---

# Worktree overlay merge layer wired into next and list

## Objective

Build the cross-worktree overlay — the merge layer that gives every worktree visibility
into sibling worktrees' task state — and wire it into `next` and `list`. This is the
core fix for double-assignment: a task set `in-progress` in worktree A stops being
recommended in worktrees B and C. Spec: `docs/specs/worktree-support.md` §3–5.

## Tasks

- [x] Add `OverlayTask` wrapper (`EffectiveStatus`, `EffectiveOwner`, `Worktree`, `Branch`, `LocalOnly`, `RemoteOnly`) generalizing the `ProjectTask` pattern from `all_projects.go`
- [x] Merge rule: local copy is the base for all content; effective status is the max across copies by the ladder `pending < blocked < in-progress < in-review < cancelled < completed`; mtime breaks ties for the winning copy; remote winners carry provenance (worktree basename + branch)
- [x] Scan sibling worktrees through the existing `newTaskScanner` seam in `internal/cli/scan.go`
- [x] Activation: `worktrees: auto|true|false` config key (default `auto` — active only in multi-worktree repos), persistent `--worktrees` global flag, `TASKMD_WORKTREES` env
- [x] `next`: recommend against effective status; `--explain` names the excluding worktree
- [x] `list`: `WORKTREE` column rendered only when the overlay is active and at least one task is annotated; `--status` filters on effective status; sibling-only tasks included and marked
- [x] `set` guard: an ID resolving only to a sibling copy fails with an error naming the worktree and branch (codifies task `01kzdpvr1`)
- [x] One-line warning when copies of a task are in divergent terminal states (`completed` vs `cancelled`)
- [x] Unit tests with injected worktree discovery (merge rules, ladder, tie-break, activation matrix); e2e tests with real `git worktree add` covering the double-assignment scenario

## Acceptance Criteria

- E2E: task set `in-progress` in worktree A → `taskmd next` in worktree B does not recommend it, and `list` in B shows it with provenance
- E2E: `taskmd set` on a task that exists only in a sibling worktree fails with the guard message; no file outside the current worktree is ever written
- Single-worktree repos and non-git directories: byte-identical behavior to today under `worktrees: auto`
- `worktrees: false` restores today's behavior entirely; `--worktrees` overrides config per invocation
- Copies of the same ID in sibling worktrees are never reported as duplicate-ID errors
