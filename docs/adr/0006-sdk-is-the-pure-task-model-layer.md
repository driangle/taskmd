# ADR 0006 — `sdk/go` is the pure task-model layer

- **Status:** Accepted
- **Date:** 2026-09-01
- **Related:** [ADR 0001 — Core scope boundary](0001-core-scope-boundary.md),
  [ADR 0004 — Local git metadata is core](0004-local-git-metadata-is-core.md)

## Context

ADR 0001 answers *should taskmd do this at all* (core vs. optional). It does not
answer the orthogonal question that comes up on nearly every change: **given that
this belongs in taskmd, does it go in `sdk/go` or `apps/cli`?**

That question has been answered consistently in practice but never written down, so
each contributor rediscovers it. It matters more than a filing convention because
the two are separate Go modules with independent version lines: code placed in
`sdk/go` acquires a public API, a semver obligation, and the `apps/cli/go.mod` pin
dance described in `CLAUDE.md`. Putting something there by accident is expensive to
undo, because module versions are immutable.

The de facto boundary is already sharp. A snapshot of the tree:

| | `sdk/go` | `apps/cli` |
|---|---|---|
| imports the other | never | freely |
| files importing `spf13/*` (cobra, viper) | 0 | 58 |
| non-test `fmt.Print*` / `os.Stdout` | 0 / 0 | throughout |
| non-test files importing `os/exec` | 1 | 11 |

## Decision

**`sdk/go` is the task model and the algorithms over it: pure, silent, and unaware
of how it was invoked. `apps/cli` is everything that knows a human or a process is
on the other end.**

Three tests decide placement. Code belongs in `apps/cli` if any is true:

1. **It produces output.** Writing to stdout/stderr, choosing a format
   (table/json/yaml), colorizing, or paginating. The SDK returns values; callers
   decide how to render them.
2. **It knows how it was invoked.** Flags, cobra commands, config-file resolution,
   env vars, cwd walk-up, MCP tool schemas, HTTP handlers.
3. **It reaches outside the task files.** Git, the network, a foreign task system,
   the projects registry, filesystem watching. Reading and writing markdown task
   files is the SDK's job; anything else is not.

Everything else — parsing, validating, scanning a directory, graphs, filtering,
search, recommendation, ID generation, task-file writing — belongs in `sdk/go`.

This is a layering rule, not a scope rule. Git access is squarely *core* under
ADR 0004 and still lives in `apps/cli/internal/gitmeta`, exactly as ADR 0004
instructs. Core and SDK are different axes.

### Worked example

Worktree support splits across the boundary along precisely this line:
`sdk/go/next` holds the recommendation algorithm (pure reasoning over tasks), while
`apps/cli/internal/worktree` holds the cross-worktree overlay that decides which
tasks `next` may recommend — because the overlay needs `git worktree list`. One
feature, two homes, decided by test 3.

## Consequences

- The SDK stays usable by all three surfaces (CLI, MCP server, web) without any of
  them contending over stdout or inheriting cobra.
- Adding a parameter to an SDK function is a versioned API change; the same change
  inside `apps/cli/internal` is free. Prefer the CLI side when a change is
  presentation-driven.
- A performance fix that needs to control output ordering should be solved in
  `apps/cli`, not by giving an SDK type an output writer.

## Known drift

`sdk/go/scanner` writes verbose progress to `os.Stderr` directly
(`scanner.go:80,114,128,139`) — the only violation of the silence rule in the SDK,
and load-bearing: because those writes cannot be redirected, `Builder.scanSiblings`
must scan siblings **serially** whenever `--verbose` is set, or the lines from
concurrent scans interleave nondeterministically. The parallel path is quiet-only
for this reason alone.

The fix is to give `Scanner` an injectable `io.Writer` defaulting to `os.Stderr`,
letting the concurrent path buffer per sibling and flush in worktree order. That is
an additive SDK API change (patch bump) and was deliberately deferred rather than
bundled into a performance commit. Until then, treat these four lines as debt, and
do not add new direct-to-stderr writes anywhere in `sdk/go`.

## Alternatives rejected

- **Leave it undocumented.** The boundary is consistently followed, so it looked
  self-evident — but it is only legible after reading both trees, and getting it
  wrong costs an immutable version tag.
- **Fold this into ADR 0001.** Conflates two independent axes: ADR 0001 rules on
  whether a capability belongs in taskmd, this one on which module implements it.
  A feature can be core and still not belong in the SDK (git metadata, ADR 0004).
- **Merge the modules.** Removes the question, and with it the SDK's value as a
  dependency external consumers can import without pulling in cobra and the entire
  CLI surface.
