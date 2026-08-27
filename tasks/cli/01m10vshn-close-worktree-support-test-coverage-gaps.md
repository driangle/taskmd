---
title: "Close worktree-support test coverage gaps"
id: "01m10vshn"
status: completed
priority: medium
type: chore
tags: ["worktrees", "testing"]
created: "2026-08-27"
completed_at: 2026-08-27
---

# Close worktree-support test coverage gaps

## Objective

An audit of `docs/specs/worktree-support.md` against the codebase found three
scenarios the spec's Testing section calls for that are only partially
covered. Each has the mechanism tested at a lower layer but no end-to-end
assertion of the actual user-facing behavior.

## Tasks

- [x] **CLI no-git degradation e2e**: the spec asks for "graceful no-git
      behavior by scrubbing PATH" against the built binary. Coverage today is
      library-level only (`apps/cli/internal/gitmeta/gitmeta_e2e_test.go`,
      `TestE2E_Degradation_NoRepoAndNoGit`). Add an e2e in
      `apps/cli/internal/e2e` that runs `taskmd next` / `taskmd list` with a
      scrubbed `PATH` (git absent) and asserts exit 0, correct local-only
      output, and no warning noise on stderr
- [x] **Re-register from a second linked worktree**: the existing no-op test
      (`apps/cli/internal/e2e/projects_worktree_test.go`, around line 71)
      re-registers from the primary. Add the spec's exact scenario: register
      from linked worktree A, then re-register from linked worktree B, assert
      the friendly "Already registered" no-op and that the stored path is
      still the primary root
- [x] **Real worktree add/remove re-enumeration**: watcher tests simulate
      membership changes with `os.Mkdir`
      (`apps/cli/internal/watcher/watcher_test.go`). Add a web live-refresh
      test that runs real `git worktree add` and `git worktree remove` and
      asserts the enumerated sibling set — and the served payload — actually
      changes (not just that onChange fired). Also assert `WatchMetaDirs()`
      returns `<common-dir>/worktrees` for a real repo and nil when the
      overlay is disabled or outside a repo

## Acceptance Criteria

- All three scenarios have tests that exercise the real binary / real git
  where the spec describes real git
- `make test` and `make e2e` pass; new e2e tests carry the `e2e` build tag and
  skip cleanly when git is unavailable (except the scrubbed-PATH test, which
  provides its own environment)
- No production code changes required — if a test exposes a real defect, file
  it as its own task rather than fixing it silently here
