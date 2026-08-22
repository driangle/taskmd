# Technical Specifications

Implementation contracts for accepted features: precise behavior, data model,
command semantics, edge cases, and test requirements. A spec sits between a
`docs/design/` brainstorm (exploratory, may never ship) and a `docs/adr/` record
(the durable decision): the ADR says *what was decided and why*, the spec says
*exactly what to build*.

- [worktree-support.md](worktree-support.md) — git worktree awareness: one project
  identity per repo, cross-worktree read overlay, owner-aware `next`
  (ADRs [0004](../adr/0004-local-git-metadata-is-core.md),
  [0005](../adr/0005-worktrees-are-facets-of-one-project.md))
