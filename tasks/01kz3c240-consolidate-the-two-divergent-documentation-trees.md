---
id: "01kz3c240"
title: "Consolidate the two divergent documentation trees"
status: completed
priority: medium
effort: medium
type: docs
phase: critical-feedback
dependencies: []
tags: [docs, cleanup]
created_at: 2026-08-03
completed_at: 2026-08-03
---

# Consolidate the two divergent documentation trees

## Objective

The repo carries two parallel documentation trees that have drifted apart: the old flat
`docs/guides/` and the newer VitePress site under `apps/docs/`. They cover the same topics
with different content — e.g. `docs/guides/cli-guide.md` (2,539 lines) vs
`apps/docs/guide/cli.md` (1,726 lines); `docs/FAQ.md` vs `apps/docs/faq.md`;
`docs/why-taskmd.md` vs `apps/docs/guide/why.md`. README still links "Documentation" at the
stale `docs/guides/` while the published site is the intended source of truth. Loose design
docs (`docs/brainstorm/`, `docs/specs/`, `docs/USER_STORIES.md`) add further clutter, and
`USER_STORIES.md` still calls the project "md-task-tracker".

## Tasks

- [x] Pick one source of truth for user docs (the VitePress `apps/docs/` site)
- [x] Diff `docs/guides/*` against `apps/docs/*`, port any unique/current content into the site
- [x] Delete or clearly archive the superseded `docs/guides/` tree
- [x] Update README "Documentation" links to point only at the canonical site
- [x] Rename remaining "md-task-tracker" references to "taskmd" (start with `USER_STORIES.md`)
- [x] Decide whether `docs/brainstorm/` and `docs/specs/` belong in-repo or in a wiki/branch

## Acceptance Criteria

- Exactly one user-facing docs tree is live-referenced from the README
- No two docs describe the same feature with conflicting content
- No committed file refers to the project as "md-task-tracker"
