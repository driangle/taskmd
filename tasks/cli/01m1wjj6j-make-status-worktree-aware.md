---
title: "Make status worktree-aware"
id: "01m1wjj6j"
status: completed
priority: medium
type: feature
tags: ["worktrees", "cli", "overlay"]
created: "2026-09-06"
completed_at: 2026-09-06
---

# Make status worktree-aware

## Objective

`status` is the only read view that never sees the worktree overlay. Both entry
points in `apps/cli/internal/cli/status.go` (`runStatusList` at ~line 106,
`runStatusSingle` at ~line 176) call plain `scanTasks`, while every other read
view goes through a helper in `apps/cli/internal/cli/scan_overlay.go`:
`list`/`get` use `scanTasksWithOverlay`, `board`/`stats`/`graph`/`report`/
`phases` use `scanTasksEffective`, and `next`/`tracks` use
`scanActiveAndArchivedWithOverlay`.

The result is the exact stale read that `docs/specs/worktree-support.md`
§"Stale reads" was written to fix, and it is a false *negative* — the worst
shape for this command. With `001` set `in-progress` in sibling worktree
`agent-b`, from the primary checkout:

```
$ taskmd list --status in-progress
001  Alpha  in-progress  agent-b   tasks/001-alpha.md
$ taskmd board
## in-progress (1)
- [001] Alpha
$ taskmd status
No tasks currently in progress.          # <- misses it
```

Single-task lookup diverges the same way: `get 001` prints a `Worktrees:`
section with each copy's status, `status 001` prints only the local `pending`.
A sibling-only ID resolves under `get` with provenance but fails to match at all
under `status`.

This is a spec gap as much as an implementation gap: the §4 "Command behavior"
table in `docs/specs/worktree-support.md` enumerates `next`, `list`, the
status-aggregating views, `get`, the mutations, `validate`, and `mcp`/web —
`status` appears nowhere in the spec or in any prior worktree task
(`01m0n989x`, `01m10vshn`, `01m10mkj5`). Fixing the code without adding the row
just recreates the same drift.

It matters more than the command's size suggests: `status` backs `--statusline`
and the `get-task-status` skill, so it is the surface agents hit most often to
answer "what am I working on / what state is this in".

## Design decisions

**`--statusline` stays local-only.** It answers "what am I working on in *this*
checkout", so a task claimed in a sibling worktree is another agent's work and
surfacing it would assert the opposite. The full `status` view still merges.

**Text renders effective status; json/yaml keep `status` local and add
`effective_status`.** This is the split `list` already uses. Found during
implementation: `internal/mcp/status.go` already ships the local +
`effective_status` shape, and `taskmd-mcp` is 1.x stable where a changed tool
signature is a major bump — so the CLI aligned to it rather than the reverse.
`status` now means the same thing across CLI, MCP, and web.

## Tasks

- [x] Decide the `--statusline` question above; record the outcome in the spec
- [x] Add a `status` row to the §4 command-behavior table in
      `docs/specs/worktree-support.md` describing the merged behavior, including
      the `--statusline` decision
- [x] `runStatusList`: scan through the overlay so the `status=in-progress`
      filter runs against effective status and sibling-only in-progress tasks
      appear. Used `scanTasksWithOverlay` rather than `scanTasksEffective` —
      the output also needs the overlay itself for per-task provenance, which
      `scanTasksEffective` discards.
- [x] `runStatusSingle`: switch to `scanTasksWithOverlay` so sibling-only IDs
      resolve. No `resolvableTasks` call after all: `EffectiveTasks()` already
      includes sibling-only tasks, so appending them again would duplicate
      every one of them.
- [x] Add worktree provenance to `statusOutput` — a `Worktrees:` section in text
      output mirroring `get`, plus additive `worktree` / `branch` /
      `effective_status` fields in JSON/YAML (additive only; existing consumers
      must see no shape change when the overlay is inactive)
- [x] Honor the `--statusline` decision explicitly in `outputStatusline`
- [x] Unit tests in `internal/cli/status_test.go`: overlay active vs. inactive;
      sibling-only in-progress task appears in the list; sibling-only ID
      resolves in single-task mode; `--format json`/`yaml` provenance fields
      present with overlay and absent without; `--statusline` matches the
      decided behavior
- [x] E2E test with real `git worktree add` covering the reproduction above,
      alongside the existing worktree e2e coverage

## Acceptance Criteria

- From the primary checkout with a task `in-progress` in a sibling worktree,
  `taskmd status` lists it and agrees with `taskmd list --status in-progress`
  and `taskmd board`
- `taskmd status <id>` shows per-worktree copies when they diverge, matching the
  information `taskmd get <id>` prints
- A sibling-only ID resolves under `taskmd status <id>`, annotated with its
  worktree provenance
- With `worktree_scope: isolated`, or in a repo with a single worktree, output
  is byte-identical to today in every format
- `docs/specs/worktree-support.md` §4 has a `status` row
- Tests pass: `make test`, `make e2e`, `make lint`
