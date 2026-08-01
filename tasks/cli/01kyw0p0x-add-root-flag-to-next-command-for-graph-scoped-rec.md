---
title: "Add --root flag to next command for graph-scoped recommendations"
id: "01kyw0p0x"
status: completed
priority: medium
type: feature
tags: ["cli", "next", "graph"]
created: "2026-07-31"
completed_at: 2026-07-31
---

# Add --root flag to next command for graph-scoped recommendations

## Objective

Let users get a `next` recommendation restricted to the dependency graph
connected to a single task ID. Today the only subgraph-style filter on `next`
is `--scope`, which matches a task's scope/group field — there is no way to say
"given task X, what's the next actionable task among the things X depends on."

Add a single `--root <ID>` flag to `next` that intersects the actionable
candidate set with the tasks *reachable from* X: X's transitive dependency
prerequisites (upstream), X's transitive subtask subtree, and X itself. This
turns the current shell/`jq` compose (`graph --root ... | jq` ∩
`next --format json`) into one ranked command, and — critically — makes it work
whether X is a leaf task or a parent task.

### Reachable set — two relationships, not one

taskmd has two distinct relationships and `--root` must span both:

- **Dependencies** (`dependencies: [...]`) are the graph edges. `--root X` follows
  these **upstream** — X's prerequisites, the work that unblocks X. (Downstream
  dependents can't be actionable until X is done, so they add nothing and are
  omitted. No `--upstream`/`--downstream` direction flags.)
- **Parent/child** (`parent: X`) is a separate hierarchy that is *not* in the
  dependency graph. Since a parent task is itself never actionable while it has
  open children (`HasIncompleteChildren`), an upstream-only root would return
  nothing useful for a parent. So `--root X` also seeds the candidate set with
  **X's transitive subtree** (via `childrenMap`), so a parent root naturally
  yields its next actionable subtask.

Otherwise `--root` is orthogonal to existing filters and composes with them — it
only narrows the candidate set before scoring, so ranking is unchanged and there
is no new scoring logic. Reuse existing traversal (`graph` package `GetUpstream`
and the SDK's `childrenMap`/`BuildChildrenMap`) — do not add a new BFS for the
dependency direction.

## Tasks

- [x] Add `Root string` to `next.Options` in `sdk/go/next/next.go`
- [x] In `filterActionable` (next.go:~183), when `Root` is set, build the reachable
  set = `Root`'s upstream deps ∪ `Root`'s transitive subtree ∪ `Root` itself, then
  intersect the actionable set with it. Reuse `GetUpstream` and `childrenMap`;
  the subtree walk is a small recursion over `childrenMap`
- [x] Return a clear error when `Root` is set but the ID does not exist (mirror
  `graph.go:181` "root task %s not found")
- [x] Add a `--root` flag to the `next` CLI command in
  `apps/cli/internal/cli/next.go` and wire it into `next.Options`
- [x] Add one `--root` example to the command `Long` help text
- [x] Update docs: add `--root` to the `next` flags table and an example in the
  `### next` section of `apps/docs/guide/cli.md` (~line 196), documenting that the
  reachable set spans upstream deps + subtasks, and the parent-root behavior.
  Check `apps/docs/reference/taskmd_operations.md` for a `next` reference and
  update it if present
- [x] Add tests:
  - happy path: `--root` on a leaf returns only its upstream actionable tasks
  - **parent root**: `--root` on a parent returns its next actionable subtask
  - unknown root ID errors
  - `--root` combined with a filter narrows correctly

## Acceptance Criteria

- `taskmd next --root <ID>` returns ranked, actionable recommendations limited to
  the tasks reachable from `<ID>`: its upstream dependencies, its subtask subtree,
  and `<ID>` itself when actionable
- When `<ID>` is a parent task, the result includes its actionable subtasks
- An unknown `--root` ID produces a clear error
- `--root` composes with existing filters without changing their behavior; all
  output formats still work
- Docs updated: `apps/docs/guide/cli.md` documents the `--root` flag, an example,
  and the parent-root/subtask behavior (plus `taskmd_operations.md` if it covers
  `next`)
- Covered by SDK and CLI tests (including the parent-root case); existing `next`
  tests still pass; `make lint` clean
