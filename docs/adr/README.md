# Architecture Decision Records

Durable records of significant, non-obvious decisions about taskmd's architecture and
scope. Unlike the brainstorm notes in `docs/design/`, ADRs are **accepted decisions**:
they state a boundary or direction that future work is expected to respect.

Add a new ADR when you make a scope or architecture call that future contributors (or
agents) should be able to discover and measure proposals against. Number them
sequentially and keep them short.

- [0001 — Core scope boundary](0001-core-scope-boundary.md): what belongs in core
  taskmd vs. optional/adjacent; homes for `sync` and the source-TODO scanner.
- [0002 — Lite plugin prose conformance](0002-lite-plugin-prose-conformance.md): the
  lite plugin's prose reimplements CLI algorithms; the CLI is authoritative and
  conformance tests fail the build on drift.
- [0003 — Plugin versioning policy](0003-plugin-versioning-policy.md): the three
  marketplace plugins version independently, not in lockstep with the repo;
  `plugin.json` is the sole source of truth and `release.sh` enforces the bump.
