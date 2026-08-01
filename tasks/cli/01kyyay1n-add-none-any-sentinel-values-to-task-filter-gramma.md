---
title: "Add none/any sentinel values to task filter grammar"
id: "01kyyay1n"
status: completed
priority: medium
type: feature
tags: ["cli", "filter", "list", "graph"]
created: "2026-08-01"
completed_at: 2026-08-01
---

# Add none/any sentinel values to task filter grammar

## Objective

There is currently no first-class, discoverable way to filter tasks by the **absence** (or presence) of an optional field. You can filter `--filter phase=phase-support`, but not "tasks with no phase". Absence-filtering half-works by accident today (`--filter phase=` with an empty RHS already matches empty-phase tasks via `matchesEquality` in `sdk/go/filter/filter.go`), but it's undocumented, awkward in shells, and unavailable through the shortcut flags (`--phase`, `--status`, `--priority`), whose empty-string value means "flag not passed".

Rather than bolt a `--no-phase` flag onto one command, add a **general sentinel-value convention to the shared filter grammar** so it applies uniformly to every field and every command that consumes filters (`list` and `graph`):

- `field=none` → matches tasks where the field is **unset** (empty)
- `field=any` → matches tasks where the field is **set** (present)

This gives, for free, `--filter phase=none`, `--filter owner=none`, `--filter group=none`, `--filter priority=none`, and their `=any` counterparts, plus the shortcut-flag ergonomics (`taskmd list --phase none`).

## Design Decisions

- **Put the logic in `sdk/go/filter`**, not in `list` — one change reaches both `list` and `graph` (both route through `applyShortcutFilters` / `filter.Apply`).
- **`none` = unset, `any` = present**, recognized for every equality field in `getFieldValue` (status, priority, effort, type, group, owner, phase, and free-text fields).
- **Keep `field=` (empty RHS) working** as an unambiguous alias for `field=none`, so nothing that relies on the current behavior breaks.
- **Unify existing presence conventions**: `parent=true/false` and `blocked=true/false` are today's ad-hoc presence checks. Keep them working as aliases (`true`→`any`, `false`→`none`) so `none`/`any` become the single general vocabulary rather than a third convention.
- **Route shortcut flags through the grammar**: `--phase` currently bypasses the filter package (`filterTasksByPhase`, `filter.go:51-53`) and can't express empty phase because `""` means "not passed". Translate `--phase none` / `--status none` / `--priority none` into the sentinel filter expression so the flags inherit the behavior.
- **Reserved-value tradeoff**: `none`/`any` become reserved, so a field can't have a literal value of `none`/`any`. This is a non-issue for taskmd's enums; for free-text fields the empty-RHS form (`field=`) remains the unambiguous escape hatch.

## Tasks

- [x] In `sdk/go/filter/filter.go`, recognize `none` and `any` as reserved values in equality matching, applied uniformly across all fields returned by `getFieldValue`
- [x] Preserve `field=` (empty value) as an alias for `field=none`
- [x] Fold the existing `parent=true/false` and `blocked=true/false` presence checks into the `none`/`any` model (retain true/false as aliases)
- [x] Route the `--phase` shortcut through the filter grammar (via `applyShortcutFilters`) so `--phase none`/`--phase any` work; ensure `--status none` and `--priority none` also work
- [x] Update `list` and `graph` command help/`Long` text with `none`/`any` examples
- [x] Update filter documentation in the spec/docs to describe the sentinel convention
- [x] Add tests in `sdk/go/filter` covering `none`/`any` for multiple field types, the `field=` alias, and the true/false compatibility aliases
- [x] Add CLI-level tests in `list_test.go` (and graph if applicable) for `--filter phase=none`, `--phase none`, and combined filters (e.g. `--filter phase=none --status pending`)

## Acceptance Criteria

- `taskmd list --filter phase=none` returns only tasks with no phase; `--filter phase=any` returns only tasks that have a phase
- The `none`/`any` sentinels work identically for other equality fields (e.g. `owner`, `group`, `priority`, `type`)
- `taskmd list --phase none` works via the shortcut flag (and `--status none`, `--priority none`)
- `--filter field=` (empty value) continues to behave as `field=none`
- Existing `parent=true/false` and `blocked=true/false` filters continue to work unchanged
- Both `list` and `graph` support the new sentinels (shared grammar)
- Help text and documentation describe the `none`/`any` convention with examples
- New tests cover sentinel matching across field types, the empty-value alias, the true/false aliases, and shortcut-flag routing
- `make test` and `make lint` pass
