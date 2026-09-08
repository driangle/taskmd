---
title: "Add --group filter flag to list"
id: "01m1zfhxc"
status: pending
priority: medium
type: feature
tags: ["list", "filter", "cli", "ux"]
created: "2026-09-08"
effort: small
---

# Add --group filter flag to list

## Objective

`group` is a first-class task field — set in frontmatter, otherwise derived from
the containing directory (`deriveGroupFromPath`, `sdk/go/scanner/scanner.go:122`),
and named by `add --group`. But `list` exposes no shortcut flag for it. It has
shortcuts for `--status`, `--priority`, and `--phase`, each a thin wrapper over
the equivalent `--filter X=` expression; group is the only common field left out,
so filtering by it requires the long form:

```bash
taskmd list --filter group=cli
taskmd list --filter 'group=cl*'    # wildcards already work
```

The gap is worse than a missing convenience because `--scope` is a name trap.
On `list`, `graph`, `next`, and `tracks`, `--scope` filters by the `touches`
array — abstract *code areas*. On `feed` the same flag name means "tasks
subdirectory". So the example in the `list` help text and in
`apps/docs/guide/cli.md` silently returns nothing in this repo:

```
$ taskmd list --scope cli
No tasks found
$ taskmd list --filter group=cli
01kjm6k2f  Update status command to include children task IDs ...   (~100 more)
```

`tasks/cli/` holds ~100 tasks, but no task declares `touches: [cli]`. A user
asking for "the cli group" reaches for `--scope` first and concludes the group is
empty. Adding `--group` gives that intent an unambiguous name and puts the two
concepts side by side in `--help`, where the distinction is visible.

The positional path form (`taskmd list tasks/cli`) is not a substitute: it
changes the scan directory rather than filtering, so it needs a real path and
does not compose with `--all-projects`, where "group cli across every project" is
exactly the useful query.

## Scope

`list` only. `board`, `stats`, `archive`, `report`, and `phases` have neither
`--group` nor `--scope`; extending group filtering across the read views is a
reasonable follow-up but is deliberately **not** part of this task.

Also out of scope: resolving the underlying `--scope` collision (renaming
`feed --scope` to `--group` with a compatibility alias, or reconciling the two
meanings). This task mitigates the confusion by giving the directory/group
concept its own name on `list`; the rename is a separate, compat-affecting
change.

## Approach

Delegate to the existing filter layer rather than adding new matching logic:

- Add a `listGroup` flag variable and registration in
  `apps/cli/internal/cli/list.go` (alongside `listScope` at line 86).
- Pass it through the `FilterShortcuts` struct in
  `apps/cli/internal/cli/filter.go` (add a `Group` field next to `Scope` at
  line 24) so it becomes a `group=<value>` filter.
- `matchesEquality` already routes `group` with a `*` through `MatchScope`
  (`sdk/go/filter/filter.go:151`), so wildcards need no new code.
- `--group` and `--filter group=` must not conflict — combining them should
  behave as a conjunction like the other shortcuts do, not clobber one another.

## Tasks

- [ ] Add the `--group` flag to `list` with help text that distinguishes it from `--scope` (directory/group vs `touches` code areas)
- [ ] Thread it through `FilterShortcuts` so it resolves to the existing `group=` filter
- [ ] Confirm wildcards (`--group 'cl*'`) work via the existing `MatchScope` path
- [ ] Sharpen the `--scope` help text on `list` to say it matches `touches`, so the two flags read as distinct
- [ ] Fix the misleading `taskmd list --scope cli` examples in `list.go` and `apps/docs/guide/cli.md` (either make them `--group` or use a real `touches` value)
- [ ] Document `--group` in the `list` flags table in `apps/docs/guide/cli.md`
- [ ] Tests: exact match, wildcard match, no-match returns empty, combined with `--status`, combined with `--filter group=`, and `--group` alongside `--all-projects`

## Acceptance Criteria

- `taskmd list --group cli` returns the same tasks as `taskmd list --filter group=cli`
- `taskmd list --group 'cl*'` matches by wildcard
- `--group` composes with `--status`, `--priority`, `--phase`, and `--filter` as a conjunction
- `--group` works under `--all-projects`
- An unmatched group prints the normal empty-result output, not an error
- `--scope` behavior is unchanged; its help text no longer reads as if it filters by directory
- No `--scope cli` example remains in `list` help or the docs that returns nothing in this repo
- `make test`, `make lint`, and `make e2e` pass
