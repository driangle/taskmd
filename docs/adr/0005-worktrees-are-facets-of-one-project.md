# ADR 0005 — Worktrees are facets of one project

- **Status:** Accepted
- **Date:** 2026-08-22
- **Related:** [ADR 0004 — Local git metadata is core](0004-local-git-metadata-is-core.md),
  [Worktree support spec](../specs/worktree-support.md),
  task `01kzdpvr1` (worktree pinning post-mortem)

## Context

Agents work in multiple git worktrees of one repository at once. Task files are
branch-local, so a claim (`set --status in-progress`) made in one worktree is
invisible to `next`/`list` in the others until branches merge — agents
double-assign themselves the same task. Separately, each worktree looks to taskmd
(and to the projects registry) like an unrelated project.

Two philosophies were on the table:

- **Tasks-as-code:** task state travels with the branch; the status change ships in
  the same PR as the work. Divergence between worktrees is *by design* and
  coordination must be layered on top.
- **Tasks-as-shared-database:** one canonical tasks dir; worktree copies are a bug
  to route around by redirecting reads/writes to the primary checkout.

Task `01kzdpvr1` already settled this in practice: an agent that wrote status into
the primary checkout instead of its own worktree caused a real defect, and pinning
the CLI to the primary was explicitly rejected. Tasks-as-code stands.

## Decision

### 1. Identity: one repo, one project

A worktree is another mount of the same project, not a new project. Project
identity is **repository identity** — the git common directory — not the checkout
path. The projects registry stores one entry per repository (the primary
worktree's path); cwd-to-project matching and `--all-projects` compare repo
identity, so N worktrees never become N projects.

### 2. Coordination: read-side overlay, not shared state

Cross-worktree awareness is achieved by **reading the other worktrees' task
files**, nothing more. Commands that scan tasks also scan sibling worktrees
(discovered via `git worktree list`) and merge per task ID: the most advanced
status wins and carries provenance. `next` recommends against this effective
status; read views may annotate it.

No shared mutable state is introduced: no lock files, no claims database, no
daemon, no git refs. taskmd remains purely a reader/writer of markdown task files
— there are just more of them to read.

### 3. Mutations stay strictly local

`set`, `add`, `rm`, and `archive` operate only on the current worktree's files. A
mutation targeting a task that exists only in a sibling worktree is an **error**
that names the worktree, never a redirected write. This codifies the `01kzdpvr1`
outcome as CLI behavior instead of skill prose.

## Alternatives rejected

- **Canonical tasks dir / write redirection** (route mutations to the primary
  checkout): decouples the status change from the branch containing the work,
  fights the divide-and-conquer model, and reintroduces the `01kzdpvr1` bug as a
  feature.
- **Claims via a git ref** (commit leases to a dedicated branch): durable and
  multi-machine, but requires core to *write* git — barred by ADR 0004 — and turns
  taskmd into git plumbing.
- **Claims sidecar in the git common dir** (`<common-dir>/taskmd/claims.yaml`,
  flock-guarded leases): the only alternative kept alive. It closes the
  seconds-wide race the overlay leaves (two agents running `next` before either
  writes `in-progress`) at the cost of a second source of truth and stale-lease
  failure modes. **Deferred**, not rejected: adopt only if the race is observed in
  practice, as an explicit `taskmd claim` command layered on top of this ADR's
  model.

## Consequences

- The worktree overlay, repo-identity registry resolution, and the sibling-write
  guard are core work; the [worktree spec](../specs/worktree-support.md) is the
  implementation contract.
- `owner` becomes coordination-relevant: `next` must not recommend in-progress
  tasks owned by someone else. This applies even in a single checkout.
- The same task ID appearing in multiple worktrees is the *expected* case, not a
  duplicate-ID error; duplicate detection applies within one worktree only.
- Skills that orchestrate worktrees (divide-and-conquer, do-task) can drop their
  defensive prose over time as the CLI enforces locality itself.
- Anything stronger than the overlay (leases, cross-machine claims) must come back
  through a new ADR.
