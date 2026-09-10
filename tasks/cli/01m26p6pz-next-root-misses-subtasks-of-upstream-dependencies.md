---
title: "next --root misses subtasks of upstream dependencies"
id: "01m26p6pz"
status: completed
priority: high
type: bug
tags: ["cli", "sdk", "next"]
created: "2026-09-10"
completed_at: 2026-09-10
---

# next --root misses subtasks of upstream dependencies

## Steps to Reproduce

1. Create a root task `R` that depends on task `P`.
2. Decompose `P` into subtasks (`parent: P`), all pending — e.g. `S1`, and give
   `S1` its own dependency `D` (pending, actionable).
3. Complete all of `P`'s direct `depends_on` dependencies.
4. Run `taskmd next --root R`.

Real-world instance: in the pinspot repo, `taskmd next --root 01m1sg7ra`
returns nothing. The root depends on `01m1s2dns`, which is decomposed into four
pending subtasks; the first subtask (`01m25vw8d`) is blocked by `01m1spejm`,
which is pending and actionable (it ranks #3 in plain `taskmd next`).

## Expected Behavior

`next --root R` reports the actionable work that transitively unblocks `R`:
subtasks of upstream dependencies (a parent is not complete until its children
are), and those subtasks' own dependencies — here `D` (pinspot: `01m1spejm`).

## Actual Behavior

`No actionable tasks found reachable from "R".`

`rootReachableSet` in `sdk/go/next/next.go` builds the reachable set from the
root's transitive `depends_on` closure plus the subtask subtree of the root
itself only. It never expands subtask subtrees of *upstream* tasks, so those
subtasks — and their dependencies — are invisible. Every task in the resulting
set is either completed or blocked by incomplete children, yielding zero
recommendations.

## Fix

Interleave both edge types in the upstream closure: for every reachable task,
also add its children (via the parent/child hierarchy), then those children's
`depends_on` ancestors, recursively, until a fixpoint.

## Tasks

- [x] Rewrite `rootReachableSet` to interleave `depends_on` upstream traversal
      with parent→child subtree expansion until a fixpoint
- [x] Add a regression test modeling a parent-decomposed upstream dependency
      whose subtask has its own actionable dependency
- [x] Verify against the pinspot repro: `taskmd next --root 01m1sg7ra`
      recommends `01m1spejm`

## Acceptance Criteria

- `next --root R` recommends `D` in the shape above (subtask-of-upstream's
  dependency)
- Existing `TestRecommend_Root*` tests still pass
- Cycles between deps and parents do not hang the traversal

## Environment

- OS: macOS (darwin 25.6.0)
- Version: taskmd 0.6.0
