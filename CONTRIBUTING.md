# Contributing to taskmd

Thanks for your interest in improving taskmd! This guide covers how to
contribute and — just as importantly — **what belongs in the core project and
what doesn't**. Reading the [Scope & non-goals](#scope--non-goals) section
before opening a large PR will save everyone time.

For local setup, build commands, testing, and coding conventions, see
[`CLAUDE.md`](./CLAUDE.md) — it's the development handbook and this file does
not duplicate it.

## Ways to contribute

- **Bug reports & fixes** — always welcome.
- **Docs** — corrections and clarifications to `docs/` and `apps/docs/`.
- **Features** — please open an issue first for anything non-trivial, so we can
  confirm it fits taskmd's scope before you invest in it.
- **Skills, plugins, and integrations** — see
  [Extending taskmd](#extending-taskmd-without-changing-core).

## Scope & non-goals

taskmd is a **task format plus a CLI that reads and manages those task files.**
Tasks are Markdown files with frontmatter; the CLI helps you create, update,
understand, prioritize, and validate your work.

The boundary that defines core isn't read-vs-write — plenty of commands write.
It's **what state taskmd is allowed to change: its own task files, and nothing
else.**

- `next`, `tracks`, `graph`, `stats`, `validate` and friends **read** task
  metadata (and sometimes generic Git) and tell you something useful.
- `add`, `set`, `rm`, `archive` **write task files** — the data taskmd owns.
- Nothing in core writes state taskmd doesn't own: not your Git history, not
  branches or worktrees, not external systems.

### What taskmd core does *not* do

These are deliberate non-goals. PRs that add them to core will be declined on
scope (not quality) grounds:

- **It does not orchestrate execution.** taskmd does not create branches,
  create or manage Git worktrees, run your build/test commands, or drive an
  agent through a task. It tells you *what* is workable; *how* you execute is
  yours.
- **It does not mutate Git state.** No commits, no branches, no worktrees, no
  `git config` writes. Reading generic Git state (e.g. `git status`,
  `git diff --name-only`) to *inform* a command is fine; writing Git state is
  not.
- **It does not hard-depend on third-party tools.** Core depends on Git and its
  own toolchain (Go for the CLI, TypeScript for web). Integrations with
  specific external tools (worktree managers, PR tooling, other trackers) live
  behind clear boundaries — see below — and are never a hard dependency of the
  core CLI.
- **It stays in its toolchain.** The CLI is Go, the web app is TypeScript.
  New languages/runtimes in core (e.g. bundled Python scripts) need a
  compelling, discussed reason and a full CI/lint/format story — assume the
  answer is "no" unless agreed in an issue first.

If you're unsure which side of the line a feature falls on, ask in an issue
before building. A good litmus test: *does this only read tasks/Git and write
task files?* If yes, it may fit core. If it *changes* Git state, spawns
processes to do work, or requires a specific external tool, it belongs in an
integration.

## Extending taskmd without changing core

Most "I wish taskmd could also…" ideas are better — and ship faster — as
something layered on top:

- **Agent skills** — the `claude-code-plugin/` skills are thin wrappers that
  shell out to the `taskmd` CLI. Great for encoding a workflow. Keep them thin;
  if a skill grows real logic, that logic probably wants to be a CLI command or
  an SDK function instead.
- **The SDK** — `sdk/go` exposes taskmd's scanning, scoring, and track logic so
  other tools can build on the same primitives the CLI uses.
- **The MCP server** — for programmatic, tool-based access to task data.
- **Standalone integrations** — anything that mutates Git, drives worktrees, or
  depends on an external tool belongs in its own plugin/repo (e.g. a
  hypothetical `taskmd-worktrunk`). We're happy to link well-maintained
  community integrations from the docs.

## Pull requests

- Open an issue first for non-trivial features (scope check).
- Follow the testing and code-quality requirements in
  [`CLAUDE.md`](./CLAUDE.md): tests for new behavior, `make test` / `make e2e` /
  `make lint` green, and updated docs where relevant.
- Keep PRs focused. Smaller, single-purpose changes get reviewed faster.
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit
  messages (`feat:`, `fix:`, `docs:`, …).

## Task tracking

taskmd dogfoods itself: project work lives in `tasks/`. If you're picking up an
existing task, use the CLI to keep its status current (`taskmd set <id>
--status in-progress`). See [`CLAUDE.md`](./CLAUDE.md) for the full task
workflow.
