## 2026-08-27T06:34:06Z

Closed all three worktree-support coverage gaps. No production code changed.

**Added:**

- `apps/cli/internal/e2e/nogit_test.go` — three scrubbed-PATH e2e tests against
  the real binary: a plain non-repo dir, a repo whose sibling worktree has
  claimed a task (overlay must go inert: 001 back to pending, no `agent-b`
  provenance, `next` offers it again), and `--worktree-scope unified|isolated`
  still accepted. All assert exit 0 and empty stderr via `assertCleanRun`.
- `apps/cli/internal/e2e/projects_worktree_test.go` —
  `TestProjectsRegister_FromSecondLinkedWorktreeIsNoOp`: register from linked
  worktree A, re-register from linked worktree B, assert the friendly
  `Already registered as "myrepo"` no-op, that neither worktree path reaches
  the registry, and that `projects --format json` reports one entry at the
  primary root.
- `apps/cli/internal/web/worktree_git_e2e_test.go` — real `git worktree
  add`/`remove` driving the NewServer wiring (DataProvider + SSE broker +
  watcher). Asserts the enumerated sibling set changes, an SSE broadcast fires
  on each membership change, and the *served* `/api/tasks` payload gains and
  then loses `effective_status`/`worktree`. Plus `WatchMetaDirs()` =
  `<common-dir>/worktrees` from both primary and linked checkouts, and nil when
  disabled or outside a repo.
- `apps/cli/internal/e2e/worktree_test.go` — `requireGit` helper so the new
  git-dependent tests skip cleanly.

**Decisions:**

- Tagged the web test `e2e` (matching `gitmeta_e2e_test.go`'s precedent for
  real-git tests) rather than letting `make test` shell out to git. That
  required adding `./internal/web/...` to the `e2e` make target — the only
  non-test change in the task.
- Verified the live-refresh test is not tautological: swapping
  `dp.WatchMetaDirs()` for nil in the harness makes it fail on both the missing
  broadcast and the stale payload.

**Verified:** `go test ./...`, `make e2e`, `make lint` (and
`golangci-lint run --build-tags e2e` — the 6 findings are all pre-existing, none
in the new files). No defects surfaced, so no follow-up task filed.
