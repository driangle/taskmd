## 2026-08-22T00:00:00Z (session start)

Starting task 01m0n8cpm: add `apps/cli/internal/gitmeta` package (repo identity +
worktree discovery) per docs/specs/worktree-support.md §1–2 and ADR 0004.

**Approach:** New package with `Identity`/`Resolve` (single `git rev-parse`
invocation), `Worktree`/`ListWorktrees` (`git worktree list --porcelain`),
filtering (prunable, missing roots, no `.taskmd.yaml`), and raw-yaml tasks-dir
resolution. Unit tests on fixture porcelain output; e2e tests with a real repo
and `git worktree add`. Mirroring the shellout style in feed.go/commit_msg.go.

## 2026-08-22T20:05:00Z

Implemented `apps/cli/internal/gitmeta` and finished the task.

**What was built:**

- `identity.go` — `Identity{CommonDir, WorktreeRoot, IsLinked}` + `Resolve(dir)`,
  one `git rev-parse --path-format=absolute --git-common-dir --git-dir --show-toplevel`
  invocation; `(nil, nil)` on missing git / non-repo / any git error.
- `worktrees.go` — `Worktree{Root, Branch, IsLocal, TasksDir}` + `ListWorktrees(id)`
  over `git worktree list --porcelain`; pure `parsePorcelain` + `filterWorktrees`
  split for gitless unit testing. Filters bare, prunable, missing-on-disk, and
  no-`.taskmd.yaml` roots.
- `tasksdir.go` — raw-yaml (`os.ReadFile` + `yaml.Unmarshal`, no viper) resolution
  of `dir`/`task-dir` against the worktree root, `tasks/` fallback (same pattern
  as `resolveProjectScanDir`).
- `log.go` — `Debugf`/`Warnf` no-op hook vars; the CLI wires them to
  `--debug`/`--verbose` when a later task integrates the package into commands.

**Decisions:**

- Added `TasksDir` to the `Worktree` struct (spec showed only Root/Branch/IsLocal)
  since ListWorktrees already stats `.taskmd.yaml` and downstream overlay tasks
  need the dir — one read, no second pass.
- Also skip `bare` porcelain entries explicitly (a bare root can exist on disk).
- E2E tests live in the package itself behind `//go:build e2e` (no command uses
  gitmeta yet, so binary-level e2e can't exercise it); widened `make e2e` to
  include `./internal/gitmeta/...`.
- `filepath.EvalSymlinks` in e2e fixtures — macOS `/var` → `/private/var` would
  otherwise break path equality against git's resolved output.

**Verification:** unit 86.5% coverage, unit+e2e 95.9% (≥90% core-package goal);
`make lint`, `go vet`, `make test`, `make e2e` all green.

**Open items:** none for this task. Wiring hooks + activation flag land with the
overlay tasks (01m0nk00d, 01m0n1nxx).
