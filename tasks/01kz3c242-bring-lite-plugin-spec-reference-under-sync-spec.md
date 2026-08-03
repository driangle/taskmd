---
id: "01kz3c242"
title: "Bring lite plugin SPEC_REFERENCE under sync-spec"
status: pending
priority: medium
effort: small
type: chore
phase: critical-feedback
dependencies: []
tags: [docs, plugins, spec-sync]
created_at: 2026-08-03
---

# Bring lite plugin SPEC_REFERENCE under sync-spec

## Objective

The project has excellent spec-sync machinery: `make sync-spec` copies the canonical
`docs/taskmd_specification.md` to the embedded CLI template and the docs site, and
`TestSpecTemplate_MatchesCanonicalSpec` fails on drift — all three copies are byte-identical.
But there is a **fourth** spec variant, `claude-code-plugin-lite/SPEC_REFERENCE.md` (a
condensed ~191-line subset), that sits entirely outside this machinery. It is hand-maintained
and can silently drift from the 424-line canonical spec — exactly the failure mode the sync
system was built to prevent, reintroduced outside the system.

## Tasks

- [ ] Decide the model: either (a) generate `SPEC_REFERENCE.md` from the canonical spec, or
      (b) drift-check it against the canonical spec in a test
- [ ] If generated: add a `sync-spec` step that derives the condensed reference deterministically
- [ ] If checked: add a test asserting the condensed reference is a faithful subset (no
      contradictory field definitions/enums)
- [ ] Document in `AGENTS.md` that `SPEC_REFERENCE.md` is derived/checked, not hand-edited
- [ ] Fix the `AGENTS.md` claim that the spec has "two copies" — it now governs more

## Acceptance Criteria

- `SPEC_REFERENCE.md` cannot silently diverge from the canonical spec (test or generation guards it)
- `make sync-spec` (or an equivalent) keeps it current, or a test fails when it drifts
- `AGENTS.md` accurately states how many spec artifacts exist and how they stay in sync
