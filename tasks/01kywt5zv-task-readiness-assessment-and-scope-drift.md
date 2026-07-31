---
title: "Answer \"can I start this task now?\" with a read-only readiness check"
id: "01kywt5zv"
status: pending
priority: medium
effort: medium
type: feature
tags:
  - cli
  - tracks
  - parallelism
touches:
  - "cli/next"
  - "cli/tracks"
created: "2026-07-31"
---

# Answer "can I start this task now?" with a read-only readiness check

## Objective

Give users a fast, task-centric answer to *"can I safely start **this** task
right now, or will it collide with work already in flight?"* — using only
taskmd metadata and generic, read-only Git.

This is the core-appropriate slice of the problem raised in the declined
`safe-queue` PR (#13). That PR bundled the answer with Worktrunk orchestration,
worktree creation, and branch mutation — all of which are explicit non-goals
per `CONTRIBUTING.md`. This task keeps **only** the parts taskmd should own:
reading tasks and reasoning about them.

## Background

Two of the three sub-questions are already answered today:

- **Dependency readiness** — `taskmd next` already shows only actionable tasks
  ("pending or in-progress with all dependencies completed").
- **Declared scope overlap** — `taskmd tracks` already groups actionable tasks
  into parallel lanes based on `touches` overlap.

The genuinely missing piece is a **reality check**: `next` and `tracks` trust
the `touches` frontmatter completely. They can't tell that an in-progress task
declared `touches: [api]` but is actually editing files under `src/search/`.
That declared-vs-actual **scope drift** silently invalidates the parallelism
picture, and detecting it needs only generic Git (`git status`,
`git diff --name-only`) — never worktree creation or a third-party tool.

## Design decisions to make

Decide the command surface before implementing (open for discussion):

- **Preferred:** a scope-drift check surfaced through an existing command, e.g.
  `taskmd tracks --check-drift`, since drift directly contradicts the lanes
  `tracks` computes and that command already has the scanner + scope resolution
  in hand. A lint-style exit code (0 = clean, non-zero = drift) makes it
  CI/pre-commit usable.
- Optionally, a per-task explanation on `taskmd next <id> --explain` that
  reuses the existing scoring engine to say *why (not) now* for a single task.

Whatever surface is chosen, the drift primitive (attribution + path match)
should live in `sdk/go` (e.g. `sdk/go/tracks` or a small `sdk/go/drift`) so the
web UI and MCP server can reuse it.

## Attribution (the hard part — handle honestly)

Git changes aren't labeled by task. The command must be explicit about how it
maps a changed file to a task:

- **Single working tree, one in-progress task** — attribute all changes to it
  (unambiguous).
- **Single working tree, multiple in-progress tasks** — per-task attribution is
  impossible from Git alone. Either require `--task <id>` (check one task's
  scope), or check changes against the *union* of in-progress declared scopes
  and clearly warn that per-task attribution is unavailable.
- **Worktree-per-task (if present)** — changed files in a task's worktree are
  unambiguously that task's; use `git worktree list` (generic Git, read-only).

## Tasks

- [ ] Add a read-only drift primitive to `sdk/go`: given a task's declared
      scopes (resolved via the existing scope config) and its actual changed
      files, return files that fall outside the declared scope.
- [ ] Reuse the existing scope resolution (`.taskmd.yaml` `scopes.<name>.paths`
      / `taskcontext`) rather than reimplementing path matching.
- [ ] Collect actual changed files with generic, read-only Git only
      (`git status --porcelain`, `git diff --name-only`); no mutation.
- [ ] Implement the chosen command surface (see Design decisions) with `table`,
      `json`, and `yaml` output consistent with other commands.
- [ ] Implement the attribution strategy above and make ambiguity explicit in
      the output.
- [ ] Add comprehensive tests (happy path, all formats, flags, error/edge
      cases) per the CLI testing policy in `CLAUDE.md`, plus an e2e test.
- [ ] Update docs (`docs/`, `apps/docs/`) and command help text.

## Acceptance Criteria

- A user can ask, for a single task, whether it is ready and non-conflicting,
  and get a clear answer with reasons.
- The tool flags in-progress tasks whose **actual** changed files fall outside
  their **declared** `touches`.
- The command is entirely read-only: it never creates branches or worktrees,
  never writes Git config, and never changes task status or ownership.
- It depends only on Git and taskmd's own toolchain — no third-party tools.
- Attribution ambiguity (multiple in-progress tasks in one working tree) is
  reported honestly rather than guessed.
- A lint-style exit code makes the drift check usable in CI / pre-commit.
- Output supports `table`, `json`, and `yaml`.

## Explicitly out of scope

Per `CONTRIBUTING.md` non-goals — these belong in a downstream integration, not
core:

- Creating or managing Git worktrees.
- Creating/switching branches, freezing bases, or stacked-merge orchestration.
- Writing any Git state (commits, branches, `git config`).
- Claiming tasks / changing status as part of "starting" work.
- Any dependency on Worktrunk (`wt`) or other external tools.
