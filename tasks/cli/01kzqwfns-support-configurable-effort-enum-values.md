---
title: "Support configurable effort enum values"
id: "01kzqwfns"
status: completed
priority: medium
type: feature
tags: ["config", "validator", "effort"]
created: "2026-08-11"
completed_at: 2026-08-11
---

# Support configurable effort enum values

## Objective

The `effort` field is currently hard-coded to `small | medium | large` in several
places (`sdk/go/validator/validator.go`, `sdk/go/model/task.go`,
`sdk/go/filter/filter.go`, `apps/cli/internal/cli/suggest.go`,
`apps/cli/internal/cli/colors.go`, flag help text in `add.go`/`set.go`, and the MCP
tool schemas). Teams that use their own sizing vocabulary (e.g. `xs, s, m, l, xl`
or Fibonacci points) cannot express it.

Make the effort enum configurable from `.taskmd.yaml`, defaulting to
`small, medium, large` when unset so existing projects are unaffected.

## Decisions

**Config shape — a bare ordered list, lowest to highest:**

```yaml
# .taskmd.yaml
effort: [xs, s, m, l, xl]
```

Not the `effort: values: [...]` map originally sketched above. `effort` is an
ordinal field, so the config must preserve order, which rules out a map keyed by
value. If per-value metadata is added later (e.g. the "typical duration" table in
the spec), the natural shape is an ordered list of objects mirroring `phases:`
(`- id: xs` + fields) — and a plain list of strings is exactly the shorthand item
form of that list, so the growth path is additive. A `values:` wrapper would only
have bought extensibility the list already provides. A non-list `effort:` value
produces a structural config error, the way `scopes` does.

**SDK plumbing — change signatures, no globals:** the vocabulary is passed
explicitly as an `effort.Scale` parameter to `filter.Apply`,
`taskfile.ValidateUpdateRequest`, and `board.GroupTasks`, and as a field on
`next.Options`. Rejected: additive `…WithEfforts` twins (leaves the old
default-vocabulary entry points callable, which is the exact bug class this task
fixes) and a settable process-global in the SDK (hidden mutable state in an
otherwise pure SDK). `Scale`'s zero value means the default vocabulary, so
`effort.Scale{}` is always a safe argument.

**`next` scoring — proportional to position:** `points = 5 * (N-1-rank) / (N-1)`
with integer division, rather than matching the literal names `small`/`medium`.
This reproduces today's `5 / 2 / 0` exactly for the default three values and
spreads `5 / 3 / 2 / 1 / 0` over a five-value scale, so long vocabularies are not
flattened into "lowest two score, everything else is zero". A single-value
vocabulary scores its one value at full points. `--quick-wins` means the lowest
configured value.

**Out of scope:** `apps/web`'s hard-coded `EFFORTS` constant
(`src/components/tasks/TaskTable/constants.ts`), used for the filter checkboxes
and the edit-form dropdown. Serving the configured vocabulary there needs a new
API field; tracked as follow-up task `01kzrmtj4`. Custom values still round-trip
correctly, they just don't appear in those two pickers.

## Tasks

- [x] Design the config surface (`effort: [a, b, c]` — see Decisions) and
      record the decision in the canonical spec
- [x] Add the parsed effort values to the CLI config extraction layer and to
      `validator.ConfigData` (alongside `Phases`/`Scopes`/`IDConfig`)
- [x] Replace the hard-coded `validEfforts` map in `sdk/go/validator/validator.go`
      with the configured set, falling back to the default when absent
- [x] Validate the config itself: reject empty lists, duplicate values, and
      non-string entries with a clear error
- [x] Update `sdk/go/filter/filter.go` so `--filter effort=<value>` accepts custom values.
      Note: `effort` is an **ordinal** field there (`ordinalFields`, supports
      `effort>small`), so the configured values must be an **ordered list**
      (lowest → highest), not an unordered set
- [x] Update the second copy of the enum sets in `sdk/go/taskfile/taskfile.go`
      (`validEfforts`, used by `ValidateUpdateRequest`) — the config must reach both
      validation paths, not just the validator package
- [x] Update `apps/cli/internal/cli/suggest.go` (`validEffortValues`) so `add`/`set`
      suggestions reflect the configured set
- [x] Make `apps/cli/internal/cli/colors.go` degrade gracefully for unknown values
      (stable default color rather than no styling)
- [x] Make effort-dependent behavior configuration-aware, notably `next --quick-wins`
      (currently hard-coded to `small` — use the first configured value)
- [x] Update flag help text in `add.go` and `set.go` to show the configured values
- [x] Update MCP tool schemas (`internal/mcp/set.go`, `next.go`) so descriptions
      aren't misleading under a custom set
- [x] Check the web surface (`internal/web/export.go`, `handlers.go`) for hard-coded
      effort assumptions
- [x] Update `docs/taskmd_specification.md`, run `make sync-spec`, and update
      `claude-code-plugin-lite/SPEC_REFERENCE.md` if enum facts changed
- [x] Add unit tests (validator, filter, config parsing) and an e2e test covering a
      project with a custom effort set

## Acceptance Criteria

- A project with no `effort` config validates exactly as today: `small`, `medium`,
  `large` accepted, anything else is an error
- A project configuring custom effort values (e.g. `[xs, s, m, l, xl]`) validates
  those values and rejects `medium` unless it is in the list
- `taskmd list --filter effort=<custom>`, `taskmd add --effort <custom>`, and
  `taskmd set <id> --effort <custom>` all work with the configured set
- Invalid effort config (empty list, duplicates, non-strings) produces a clear
  validation error rather than a panic or silent fallback
- `taskmd next --quick-wins` behaves sensibly under a custom set
- `taskmd validate` passes on this repo and `make test`, `make e2e`, `make lint` pass
- Spec and docs describe the new config key and its default
