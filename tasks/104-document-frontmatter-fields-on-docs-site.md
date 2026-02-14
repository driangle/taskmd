---
id: "104"
title: "Document all frontmatter fields on docs site with tool usage"
status: pending
priority: medium
effort: medium
tags:
  - documentation
  - docs-site
created: 2026-02-14
---

# Document All Frontmatter Fields on Docs Site with Tool Usage

## Objective

Update the docs site specification page (`apps/docs/reference/specification.md`) and concepts page (`apps/docs/getting-started/concepts.md`) so that every frontmatter field is properly documented, including which CLI commands use each field and how.

Currently the docs site is missing the `touches`, `owner`, and `parent` fields that exist in the internal spec (`docs/taskmd_specification.md`). Additionally, none of the documented fields explain how they affect specific CLI commands or web views.

## Tasks

- [ ] Add `owner` field documentation to `apps/docs/reference/specification.md`
- [ ] Add `touches` field documentation to `apps/docs/reference/specification.md`
- [ ] Add `parent` field documentation to `apps/docs/reference/specification.md`
- [ ] Update the field summary table to include all fields (`owner`, `touches`, `parent`)
- [ ] For each field, add a "Used by" section explaining which commands consume it:
  - `id` — used by all commands for task identification; `get`, `set` use it for lookup
  - `title` — displayed in `list`, `board`, `next`, `graph`, web views
  - `status` — used by `list` (filtering), `board` (column assignment), `next` (excludes completed), `graph` (exclude-status), `stats` (status breakdown), `set` (can update)
  - `priority` — used by `next` (scoring), `list` (filtering/sorting), web board filters
  - `effort` — used by `next` (scoring), `list` (filtering/sorting), `stats`
  - `dependencies` — used by `graph` (edge drawing), `next` (blocks recommendations until satisfied), `validate` (cycle detection, missing refs)
  - `tags` — used by `tags` command, `list` (filtering), web filters
  - `group` — used by `list` (filtering), web filters, derived from directory if omitted
  - `owner` — used by `list` (filtering), web display
  - `touches` — used by `tracks` command to assign tasks to parallel work tracks; tasks sharing a scope are placed in separate tracks
  - `parent` — used for hierarchical grouping; children computed dynamically
  - `created` — used by `list` (sorting), display purposes
- [ ] Document the `scopes` configuration key in `apps/docs/reference/configuration.md` and its relationship to `touches`
- [ ] Update `apps/docs/getting-started/concepts.md` to mention `owner`, `touches`, and `parent` in the optional fields list

## Acceptance Criteria

- Every field from `docs/taskmd_specification.md` is present on the docs site specification page
- Each field documents which CLI commands and web views consume it
- The `scopes` config key is documented in the configuration reference
- The concepts page lists all optional fields
- Content is consistent with the internal specification
