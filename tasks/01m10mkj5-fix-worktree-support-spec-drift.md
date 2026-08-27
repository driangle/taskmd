---
title: "Fix worktree-support spec drift"
id: "01m10mkj5"
status: completed
priority: low
type: chore
tags: ["worktrees", "docs"]
created: "2026-08-27"
completed_at: 2026-08-27
---

# Fix worktree-support spec drift

## Objective

An audit of `docs/specs/worktree-support.md` against the code found small
places where the spec and the shipped implementation disagree. None are bugs —
update the spec (or note deliberate deviations) so it reads true.

## Tasks

- [x] §4 command table lists a `metrics` command that does not exist in the
      CLI (`sdk/go/metrics` is a library consumed by `stats` and `report`) —
      remove the row or reword it
- [x] §5 claims single-worktree repos incur "zero extra git invocations beyond
      the one identity probe", but `DiscoverSiblings` always runs
      `git worktree list` when inside a repo — correct the claim (or note it
      as an accepted cost)
- [x] §3/§4: note that sibling-only tasks are not yet addressable by `get`
      and that `next --explain` provenance is table-only, cross-referencing
      the tasks that close those gaps (01m10n2qw, 01m10g4dd) — or drop this
      item if those land first
      (the `next --explain` half is dropped: 01m10g4dd is completed and §4
      already documents the structured `excluded` array)
- [x] §7: "the overlay applies per-repo when active" — no longer drift; task
      01m10rq61 implemented it and expanded the bullet with the per-project
      activation and override precedence
- [x] Minor code/spec naming drift worth a one-line note each:
      `gitmeta.Worktree` carries an extra `TasksDir` field; `bare` worktrees
      are also filtered; the overlay type is `worktree.Task` aliased to
      `OverlayTask` in the CLI rather than a literal `OverlayTask` declaration

## Acceptance Criteria

- Every statement in `docs/specs/worktree-support.md` matches the shipped code
  or explicitly marks itself as pending with a task reference
- `docs/taskmd_specification.md` is untouched (no `make sync-spec` needed) —
  this is spec-doc-only drift
- `taskmd validate` passes
