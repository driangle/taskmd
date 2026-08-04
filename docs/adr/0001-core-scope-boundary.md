# ADR 0001 — Core scope boundary: what belongs in taskmd

- **Status:** Accepted
- **Date:** 2026-08-04
- **Deciders:** taskmd maintainers
- **Related task:** `01kz3c24f` — Define scope boundary for sync and source-TODO scanner

## Context

taskmd's premise is small and sharp: **markdown task files plus a CLI to read and
change them.** The codebase, however, has grown two large adjacencies that sit well
outside that premise:

- **`internal/sync` (~4,700 LOC)** — bidirectional integration with Jira, Linear,
  Trello, and GitHub. This is a platform surface: HTTP clients, auth, field mapping,
  conflict resolution, and state tracking, one provider at a time. It backs
  `taskmd sync down` and `taskmd import`.
- **`internal/todos` (~1,600 LOC)** — scans *source-code comments* (`TODO`, `FIXME`,
  `HACK`, …) with language-aware parsing and git blame. This is a **different domain**
  than markdown task files: it reads code, not tasks. It backs `taskmd todos list`.

Each is self-contained and neither destabilizes the core. But together they mean a
single maintainer is carrying a platform's worth of surface area, and — more
importantly — there is no stated line that future proposals can be measured against.
"Should taskmd add X?" currently has no principled answer.

This ADR draws that line. It is a **decision, not a code change**; the moves it
implies are tracked as follow-up tasks.

## Decision

### 1. Definition of "core" vs "optional/experimental"

**Core taskmd** is everything required to treat a directory of markdown task files as
a queryable, mutable task database, and nothing more:

- **Task model & IO** — parser, scanner, validator, the frontmatter schema and spec.
- **Read views** — `list`, `get`, `next`, `board`, `stats`, `graph`, `report`,
  `metrics`, `tracks`, `phases`.
- **Mutations** — `add`, `set`, `rm`, `archive`, worklogs.
- **Config** — `.taskmd.yaml` loading, scopes, phases, ID strategy.
- **Surfaces over the core model** — the MCP server and the web UI are *views* of the
  core; they add no new domain, so they are core.

The test for core: **it reads or writes markdown task files (or presents them), and it
would still make sense if taskmd never talked to any system other than the local
filesystem.**

**Optional / adjacent** is anything that reaches outside that boundary — a network
service, a foreign task system, or a *different source domain* (like source code).
Optional code may live in-repo, but it must not be assumed by the core, must not expand
the core's dependency surface, and is held to a higher bar for new investment.

### 2. `sync` (Jira / Linear / Trello / GitHub) → **extract to a separate module**

Sync is the clearest non-core surface and the largest single maintenance liability. It
is a foreign-system integration platform that grows one provider at a time — exactly
the kind of open-ended surface the core should not carry by default.

**Decision: extract `sync` (and the provider-specific portions of `import`) into its
own Go module / plugin, behind a stable integration API.** The core defines a small,
documented "task source" contract; providers implement it in the separate module. The
default `taskmd` binary is core-only.

Rationale for a separate module over a build tag:

- It is the only option that actually removes the surface from the core's ownership
  rather than merely hiding it — the provider churn (API changes, auth, rate limits)
  lives where it belongs.
- It forces a real, minimal integration contract, which is healthy regardless.
- Build tags would keep all provider code in the core tree and in the core's
  `go.mod`; they hide the command, not the maintenance.

Cost, accepted: this is the heaviest lift. It requires defining a stable source API and
a migration. It is tracked as a follow-up task; **until it lands, `sync` stays in-tree
and supported** — this ADR states the target, it does not orphan working code.

### 3. `todos` (source-TODO scanner) → **stays in core, but capped**

`todos` reads source code, so by the strict test above it is *adjacent*, not core.
It stays in core anyway for two reasons:

- It is **small and stable** (~1,600 LOC, no network, no foreign auth).
- It has a **clear on-ramp role**: surfacing `TODO`/`FIXME` comments so they can be
  turned into real markdown tasks (`import`, the `import-todos` skill). That directly
  serves the core's job of getting work into task files.

**It is capped, not a growth area.** New languages, new markers, richer blame/analytics,
or a `todos`-centric workflow are **not** default-yes. Such proposals must justify
themselves against this boundary the same way any optional feature would, and the
preferred direction for `todos` investment is deeper integration with task creation, not
breadth as a standalone code scanner.

## Consequences

- **New feature proposals have a rule to point at.** "Does it read/write markdown task
  files, and would it make sense filesystem-only?" → core. Otherwise → optional, and it
  must earn its place. Anything resembling a new foreign-system integration goes to the
  sync module, not the core.
- **`sync` is on a path out of the core.** Pending sync/import provider tasks
  (`085` Linear sync, `086` Trello sync, `128` Trello import, `129` Linear import)
  are now **work against the future separate module**, not the core, and should be
  sequenced after the extraction lands.
- **`todos` is frozen in scope** unless a proposal clears the optional-feature bar.
- The `external-integrations` phase in `.taskmd.yaml` is re-scoped to mean "the separate
  sync module," not core work.

## Follow-up

- **Extraction task created:** split `sync` into a separate module behind a stable task-source API (see `taskmd list --tag scope`).
- No move task for `todos` — it stays in core by this decision.

## How to use this ADR

When evaluating any new feature, integration, or scanner:

1. Apply the core test in §1.
2. If it's a foreign-system integration, it belongs in the sync module (§2), not core.
3. If it's core-adjacent but small and serves task creation, it *may* be core, but
   state why it clears the cap (§3).
4. Record non-obvious scope decisions as a new ADR in this directory.
