---
id: "01kz6013n"
title: "Extract sync integrations into a separate module behind a stable task-source API"
status: pending
priority: high
effort: large
phase: external-integrations
dependencies: []
tags: ["scope", "architecture", "sync", "refactor"]
created_at: 2026-08-04
---

# Extract sync integrations into a separate module behind a stable task-source API

## Objective

Execute the move decided in [ADR 0001 — Core scope boundary](../../docs/adr/0001-core-scope-boundary.md):
lift `internal/sync` (Jira/Linear/Trello/GitHub) — and the provider-specific parts of
`import` — out of the core CLI and into a **separate module/plugin** behind a small,
documented "task source" contract. The default `taskmd` binary should ship core-only;
foreign-system integrations live in (and are maintained in) the extracted module.

Until this task lands, `sync` stays in-tree and supported — the ADR states the target,
it does not orphan working code.

## Tasks

- [ ] Define the stable task-source integration API (the contract providers implement:
      fetch, map to task model, conflict/state hooks). Document it.
- [ ] Move `internal/sync/{jira,linear,trello,github}` and provider glue into the new
      module; keep only the core-side contract in `apps/cli`.
- [ ] Rework `sync` / `import` command wiring so core builds don't depend on provider
      packages; decide the distribution/loading mechanism for the module.
- [ ] Migrate `.taskmd.yaml` `sync/*` scopes to the new module's paths.
- [ ] Re-sequence pending provider tasks (`085`, `086`, `128`, `129`) against the new
      module.
- [ ] Update docs (README, docs site) to describe integrations as an optional add-on.

## Acceptance Criteria

- The default `taskmd` binary/`go.mod` no longer depends on Jira/Linear/Trello/GitHub
  provider code.
- A documented, stable task-source API exists that the extracted module implements.
- `sync` and `import` still work when the module is present; core still builds and
  tests pass without it.
- The scope boundary in ADR 0001 is satisfied: core = markdown task files only.
