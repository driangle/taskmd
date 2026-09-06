## 2026-09-06T19:41:04Z

Implemented status worktree-awareness.

**Design decisions:**

1. `--statusline` stays **local-only** (user decision). It answers "what am I
   working on in *this* checkout", so a task claimed in a sibling worktree is
   another agent's work; surfacing it would assert the opposite.

2. Text vs structured output split, matching what `list` already does:
   - text renders the **effective** status (a command that just selected a task
     as in-progress must not print `Status: pending`);
   - json/yaml keep `status` as the **local** copy's and add `effective_status`.

   The second half was a mid-task correction: I initially made json report the
   effective status under `status`, then found `internal/mcp/status.go` already
   ships the local+`effective_status` convention — and taskmd-mcp is 1.x stable,
   where a changed tool signature is a major bump. Aligning the CLI was the
   right way round; `status` now means one thing across CLI, MCP, and web.

**Changed:** `runStatusList`/`runStatusSingle` now use `scanTasksWithOverlay`;
new `status_overlay.go` holds the merge wiring. `statusViewTasks` uses
`EffectiveTasks()` directly — no `resolvableTasks` call, since that list already
contains sibling-only tasks and appending would duplicate them.

**Refactor:** status.go had grown to 428 lines (over the repo's 300-line soft
guardrail, and already over before this change). Split into status.go (command +
run funcs, 203), status_output.go (output model), status_text.go (rendering),
status_overlay.go (cross-worktree). Pure moves.

**Also:** extended the shared `newSiblingWorktree` test helper to MkdirAll, so
siblings can carry nested group layouts.

**Verified:** make test, make e2e, make lint (0 issues), plus a manual
two-worktree repro. Spec §4 has a `status` row; docs/guide/cli.md updated.
