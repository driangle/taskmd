---
id: "01m6jjkg0"
title: "Add --group flag to list command"
status: completed
priority: medium
effort: small
type: feature
dependencies: []
tags: ["cli", "list"]
touches: ["cli"]
created_at: 2026-11-04
completed_at: 2026-11-04
---

# Add --group flag to list command

## Objective

Add a `--group` flag to `taskmd list` as a shortcut for `--filter group=<value>`,
matching the existing `--status`, `--priority`, and `--phase` shortcuts.

`--filter group=cli` already works (including wildcards, see
`sdk/go/filter/filter.go`), but `group` is the primary structural axis and
`taskmd add --group` already writes it, so reading it back should use the same flag.
Passing a directory (`taskmd list tasks/cli`) is not equivalent: an explicit `group:`
in frontmatter overrides the directory-derived group.

## Reference

- `apps/cli/internal/cli/list.go` — flag registration and how `--status`/`--priority`/`--phase` fold into filters
- `sdk/go/filter/filter.go` — `group` field matching and wildcard support
- `apps/cli/internal/cli/add.go` — existing `add --group`

## Tasks

- [x] Register `--group` on `listCmd`; help text must say it *filters* by group (to avoid confusion with `--group-by` on `board`/`report`)
- [x] Route it through the same path as `--status` (i.e. `group=<value>`), so wildcards and `none`/`any` work
- [x] Add examples to the command's `Long` help (`--group cli`, `--group "web*"`, `--group none`)
- [x] Tests in `list_test.go`: exact match, wildcard, `none` (root-level tasks), frontmatter-overridden group, combined with `--filter`/`--status`, no matches
- [x] Reset the new flag var wherever list tests reset flag state
- [x] Add an e2e test covering `taskmd list --group <name>`
- [x] Update CLI docs for `list` under `apps/docs/` if flags are documented there

## Acceptance Criteria

- `taskmd list --group cli` returns exactly the tasks `taskmd list --filter group=cli` returns
- `--group "web*"` matches by wildcard; `--group none` matches tasks with no group
- `--group` combines with other filters using AND logic
- Without `--group`, behavior is unchanged
- `make test`, `make e2e`, and `make lint` pass
