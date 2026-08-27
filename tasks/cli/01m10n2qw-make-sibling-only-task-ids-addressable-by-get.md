---
title: "Make sibling-only task IDs addressable by get"
id: "01m10n2qw"
status: pending
priority: medium
type: feature
tags: ["worktrees"]
created: "2026-08-27"
---

# Make sibling-only task IDs addressable by get

## Objective

Spec §3 (`docs/specs/worktree-support.md`) says tasks that exist only in a
sibling worktree are "appended to read views". `list`, `board`, `stats`, etc.
include them, but `get` resolves IDs against `overlay.Local()` only
(`apps/cli/internal/cli/get.go` around line 77), so `taskmd get <id>` on a
sibling-only ID reports "not found" even though `list` just displayed it.

`get` is a read view: it should show the sibling copy (content from that
worktree's file, clearly marked remote-only with worktree/branch provenance).
Mutations (`set`, `rm`, `archive`) keep the existing guard — this task changes
reads only.

## Tasks

- [ ] Extend `get`'s ID resolution to fall back to the overlay's remote-only
      tasks when no local copy matches
- [ ] Render remote-only provenance in the table output (worktree root
      basename + branch, consistent with the `Worktrees:` section and the
      `list` column) and include `worktree` / `branch` / `remote_only` in
      JSON/YAML output
- [ ] Keep exact-match/prefix-match semantics identical to local resolution;
      an ambiguous prefix across local + sibling copies should error the same
      way as an ambiguous local prefix
- [ ] Unit tests: get on a sibling-only ID (table, JSON, YAML), get on a local
      ID unchanged, ambiguous prefix spanning worktrees
- [ ] Verify the web/MCP task-detail paths — the web handler already serves
      overlay rows; confirm sibling-only detail works there or note the gap

## Acceptance Criteria

- `taskmd get <id>` succeeds for a task that exists only in a sibling
  worktree, showing that copy's content with visible worktree/branch
  provenance and a remote-only marker
- JSON/YAML output carries `remote_only: true` plus `worktree` and `branch`
- Behavior for local tasks, non-overlay repos, and `worktree_scope: isolated`
  is unchanged
- `taskmd set` on the same sibling-only ID still fails with the guard error
- Tests pass: `make test`, `make lint`
