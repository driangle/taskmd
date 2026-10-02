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

## Repository layout

taskmd is a monorepo. The pieces that ship, and where they live:

| Path | What it is | Versioned as |
|------|-----------|--------------|
| `apps/cli/` | The `taskmd` CLI (Go, cobra), including the MCP server and the web server; embeds the built web app and the spec | `vX.Y.Z` repo tags |
| `sdk/go/` | The pure task model as a Go module: parser, scanner, validator, graph, `next` scoring, tracks, search, task-file writing. No I/O policy, no output formatting ([ADR 0006](docs/adr/0006-sdk-is-the-pure-task-model-layer.md)) | `sdk/go/vX.Y.Z` tags, separate module |
| `apps/web/` | The web dashboard (Vite, React, TypeScript); built assets are embedded into the CLI binary | with the CLI |
| `apps/docs/` | The documentation site (VitePress), deployed to GitHub Pages | with the CLI |
| `apps/vscode/` | VS Code extension: frontmatter validation and completion for task files | with the CLI |
| `claude-code-plugin/` | Claude Code plugin: slash-command skills that drive the CLI | own `0.x` line |
| `claude-code-plugin-mcp/` | Claude Code plugin: MCP server over the CLI | own `1.x` line |
| `claude-code-plugin-lite/` | Claude Code plugin: the same skills using Claude's file tools, no binary | own `0.x` line |
| `.claude-plugin/marketplace.json` | The marketplace that lists the three plugins | — |

Supporting directories:

| Path | What it is |
|------|-----------|
| `docs/` | Canonical sources: `taskmd_specification.md` (the task format), `taskmd_operations.md`, `.taskmd.yaml.example`, agent templates. `adr/` holds decisions, `specs/` implementation contracts for accepted features, `design/` non-binding notes |
| `tasks/` | This project's own tasks (taskmd dogfoods itself); `.worklogs/` live alongside |
| `tests/conformance/` | Fixtures and expected outputs for checking that other implementations of the spec agree with the CLI |
| `evals/` | Skill benchmarks on skival. `benchmark/` is its deprecated predecessor and is kept only for reference |
| `scripts/` | `release.sh`, the SDK-pin guards, and the MCPB bundle builder |
| `.github/` | CI, docs deploy, and release workflows, plus the `taskmd-complete` action that marks a task complete when its PR merges |
| `.githooks/` | The pre-commit hook (`git config core.hooksPath .githooks` enables it) |

Two build entry points exist. The root `Makefile` runs the repo-wide checks
(`make check` covers the CLI, the SDK, the web app, the VS Code extension, and
the docs build; `make check-lite` is what the pre-commit hook runs) and the
SDK-pin tooling. `apps/cli/Makefile` has the CLI-only targets (`build`,
`install-dev`, `test`, `e2e`, `lint`, `sync-spec`). `go.work` ties `apps/cli`
and `sdk/go` together during development; see "The sdk/go pin" in
[`AGENTS.md`](./AGENTS.md) for why that matters at release time.

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

- **Agent skills** — most skills in `claude-code-plugin/` and
  `claude-code-plugin-lite/` are thin wrappers over the `taskmd` CLI (or, for
  lite, over Claude's file tools). Keep them thin; if a skill grows real task
  logic, that logic probably wants to be a CLI command or an SDK function.

  The **execution skills** are the deliberate exception to the non-goals above:
  `do-task`, `divide-and-conquer`, `verify-task`, and `complete-task` drive an
  agent through the work, run build and test commands, and `divide-and-conquer`
  creates branches and Git worktrees, commits to them, and may open a PR in
  pr-review mode. That is allowed because a plugin is not core: nothing in
  `apps/cli`, `sdk/go`, or `apps/web` does any of it, and the plugins are
  optional installs with their own version lines. Two rules keep the carve-out
  honest: a skill that mutates Git or runs the user's commands must say so in
  its `description` and in the plugin README's skill table, and it must never
  write to the primary checkout's branch without asking.
- **The SDK** — `sdk/go` exposes taskmd's scanning, scoring, and track logic so
  other tools can build on the same primitives the CLI uses.
- **The MCP server** — for programmatic, tool-based access to task data.
- **Standalone integrations** — anything else that mutates Git, drives
  worktrees, or depends on an external tool belongs in its own plugin or repo
  (the in-tree Claude Code plugins count; so would a hypothetical
  `taskmd-worktrunk`). We're happy to link well-maintained community
  integrations from the docs.

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
