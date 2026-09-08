---
title: "Fix cross-worktree attribution dropping all local tasks in nested worktrees"
id: "01m1zqvx6"
status: completed
priority: high
type: bug
tags: ["worktree", "overlay", "validate"]
created: "2026-09-08"
effort: small
phase: worktree-support
---

# Fix cross-worktree attribution dropping all local tasks in nested worktrees

## Steps to Reproduce

User-reported against `/Users/driangle/workplace/gg/pinspot`, whose worktrees live
*under* the primary checkout rather than beside it:

```
/repo                          [main]
/repo/.claude/worktrees/wt-a   [wt-a]
```

1. `cd /repo/.claude/worktrees/wt-a`
2. `taskmd validate`

## Expected Behavior

The 128 task files in the worktree are scanned and validated.

## Actual Behavior

```
$ taskmd validate                             → ✓ All 0 task(s) are valid
$ taskmd validate --worktree-scope isolated   → ✓ All 128 task(s) are valid
```

`--worktree-scope isolated` bisects it to the cross-worktree overlay.

## Root Cause

`AttributeNestedSiblingCopies` (`apps/cli/internal/worktree/build.go`) drops any local
task whose file path lies inside *any* sibling worktree's root. That rule exists for
spec §8: a checkout nested **inside** the scan root is double-scanned, so those copies
belong to that worktree, not this one.

The prefix test never checked the direction of the nesting. Run from inside `wt-a`, the
primary checkout is a sibling whose root is an **ancestor** of every local task path, so
the test matched all 128 tasks and `Overlay.Local()` came back empty.

`validate` surfaced it most starkly because it validates `Local()` directly, but every
overlay consumer was affected (`list`, `get`, MCP, web): the local copies were excluded
from the merge entirely, so what the user saw were the primary worktree's copies
re-labelled remote-only — a status change made *in* the worktree did not register.

## Fix

Attribution now considers only siblings whose root lies strictly **inside** the scan
dir — the only ones the local scan could have double-scanned. An ancestor sibling is a
no-op. Both sides are resolved through `filepath.Abs` before comparison.

- [x] Add `nestedSiblings` / `isUnder` helpers and filter siblings before attribution
- [x] Thread `scanDir` through `AttributeNestedSiblingCopies` and `Builder.Overlay`
- [x] Regression tests in `apps/cli/internal/worktree/attribution_test.go`

## Acceptance Criteria

- [x] `taskmd validate` inside a worktree nested under the primary checkout reports the
      full local task count, matching `--worktree-scope isolated`
- [x] Spec §8 attribution still holds: a checkout nested inside the scan root has its
      double-scanned copies dropped from the local list
- [x] A sibling parked beside the repo (the common layout) is unaffected
- [x] `go test ./...`, `make lint`, and `make e2e` pass

## Environment

- OS: macOS (Darwin 25.6.0)
- Version: taskmd 0.5.1
