# ADR 0004 — Reading local git metadata is within core scope

- **Status:** Accepted
- **Date:** 2026-08-22
- **Related:** [ADR 0001 — Core scope boundary](0001-core-scope-boundary.md),
  [Worktree support spec](../specs/worktree-support.md)

## Context

ADR 0001 defines core as "reads or writes markdown task files (or presents them),
and it would still make sense if taskmd never talked to any system other than the
local filesystem." Worktree support (ADR 0005) requires taskmd to ask git questions:
which repository am I in (`rev-parse --git-common-dir`), and what worktrees does it
have (`git worktree list`). Is that a "foreign system" under ADR 0001?

The CLI already consults git in narrow ways: `.git` bounds the config walk-up in
`root.go`, `feed` and the web UI shell out to `git log`/`git show`, `commit-msg`
reads the staged diff, and `todos` uses `git check-ignore` and blame. None of these
were treated as scope violations; none were ruled on either.

## Decision

**Reading local git metadata is core.** The local git repository is part of the
local filesystem context a task directory lives in — not a foreign task system and
not a network service. Core code may invoke the `git` binary read-only to learn
about the repository containing the task files: repo boundaries, common dir,
worktree list, log/blame/diff of task files.

Three hard limits keep this from becoming a growth area:

1. **Read-only.** Core never writes to git — no commits, no branch or ref
   creation, no worktree add/remove, no config mutation. Orchestration that
   *manages* worktrees or branches stays in skills and agents.
2. **Degrade to no-op.** Every git-aware feature must behave exactly as today when
   git is absent, the directory is not a repo, or the git invocation fails. A
   plain directory of markdown files remains a fully supported deployment.
3. **Local only.** No `fetch`, `push`, `pull`, or remote inspection beyond what
   already exists (`sync/github` reads `remote get-url`, and it is leaving core
   per ADR 0001).

## Consequences

- Worktree identity resolution and discovery (ADR 0005, worktree spec) are
  in-scope core work.
- Proposals to have taskmd *write* through git — auto-committing task changes,
  claim branches, git-ref-based coordination — fail this ADR regardless of how
  useful they look; they belong in skills or an optional module.
- The existing scattered git shellouts are retroactively legitimate under limits
  1–3, and new ones should be routed through a single helper package
  (`internal/gitmeta`) rather than ad-hoc `exec.Command` calls.
