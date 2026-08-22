---
id: "01m0nv68v"
title: "Document worktree support and sync spec surfaces"
status: pending
priority: low
type: docs
effort: small
phase: worktree-support
dependencies: ["01m0nk00d"]
tags: ["worktrees", "docs"]
created_at: 2026-08-22
---

# Document worktree support and sync spec surfaces

## Objective

Document worktree support across the user-facing surfaces and keep the spec-sync
artifacts honest. Spec: `docs/specs/worktree-support.md` (Implementation order step 6).

## Tasks

- [ ] Add `worktrees:` key to `docs/.taskmd.yaml.example` and the docs-site config reference
- [ ] Document `next --for`, the `--worktrees` global flag, and the claim convention in the docs site and command help
- [ ] If `owner` semantics wording changes in `docs/taskmd_specification.md`, run `cd apps/cli && make sync-spec` and check `SPEC_REFERENCE.md` drift tests in the same commit
- [ ] Update skill prose (`do-task`, `divide-and-conquer`, tasks/CLAUDE.md agent template) to mention the CLI-enforced sibling-write guard and the claim convention
- [ ] Mark `docs/specs/worktree-support.md` status Accepted/Implemented and cross-link shipped behavior

## Acceptance Criteria

- `taskmd validate`-style audit of docs: every new flag/config key appears in the docs site and example config
- `TestSpecTemplate_MatchesCanonicalSpec` and the `SPEC_REFERENCE` drift tests pass
- Skills no longer rely solely on prose warnings for worktree write locality
