---
title: "Add sort direction flag to list command"
id: "01kyydnea"
status: completed
priority: medium
effort: small
type: feature
tags: ["cli", "list", "sort"]
created: "2026-08-01"
completed_at: 2026-08-01
---

# Add sort direction flag to list command

## Objective

The `list` command supports sorting by field via `--sort` (`id`, `title`, `status`, `priority`, `effort`, `created_at`) but every sort is hardcoded ascending. Add a `--reverse` / `-r` boolean flag to reverse the sort order, following the Unix convention (`sort -r`, `ls -r`).

A boolean flag is preferred over `--order asc|desc` or a `field:desc` suffix because `list` sorts by a single field, so a simple flip of the final order works uniformly for every sort type (including the custom priority/effort orderings) and reads naturally: `--sort created_at --reverse` = newest first.

## Tasks

- [x] Add `listReverse` package-level flag var and register `--reverse` / `-r` on `listCmd` in `apps/cli/internal/cli/list.go`
- [x] In `applyListFiltersAndSort`, reverse the task slice after `sortTasks` succeeds and before applying `--limit` (so `--limit` keeps the top N of the reversed order)
- [x] Update the command `Long` help text and examples to document `--reverse`
- [x] Add tests in `list_test.go` covering `--reverse` with multiple sort fields and combined with `--limit`
- [x] Update relevant docs (`docs/taskmd_specification.md` if applicable, CLI reference)

## Acceptance Criteria

- `taskmd list --sort priority --reverse` lists tasks from low to critical (reverse of default)
- `taskmd list --sort created_at --reverse` lists newest tasks first
- `-r` works as a shorthand for `--reverse`
- `--reverse` with no `--sort` reverses the default file-scan order (documented behavior)
- `--reverse` combined with `--limit N` returns the top N of the reversed ordering
- New tests pass; `make lint` and `make test` are clean
