---
id: "050"
title: "Add command alias suggestions for 'Did you mean?' hints"
status: completed
priority: medium
effort: small
dependencies: []
tags:
  - cli
  - go
  - ux
  - mvp
created: 2026-02-12
---

# Add Command Alias Suggestions for "Did you mean?" Hints

## Objective

Improve the "Did you mean this?" suggestions that appear when a user mistypes a command by adding a hardcoded list of common aliases/synonyms for each command via cobra's `SuggestFor` field.

## Problem

When a user runs an unrecognized command like `taskmd set`, cobra's built-in fuzzy matching suggests commands that look textually similar (e.g., `next`, `web`). However, "set" is a semantic alias for "update" -- a user who types "set" almost certainly wants `update`, not `next`. Currently, `update` may not appear in the suggestion list because it isn't textually similar enough.

Each command should declare a list of known aliases/synonyms so that cobra includes them in "Did you mean?" suggestions alongside the existing fuzzy matches.

## Tasks

- [x] Add `SuggestFor` field to each cobra command definition with a curated list of aliases
- [x] Cover at minimum the following aliases:
  - `update`: `set`, `edit`, `modify`, `change`
  - `show`: `view`, `info`, `detail`, `details`, `describe`, `get`
  - `list`: `ls`, `tasks`, `all`
  - `graph`: `deps`, `dependencies`, `tree`
  - `stats`: `summary`, `status`, `overview`
  - `validate`: `check`, `verify`, `lint`
  - `board`: `kanban`, `columns`
  - `next`: `pick`, `suggest`, `what`
  - `snapshot`: `save`, `backup`, `export`
  - `init`: `setup`, `create`, `new`
  - `tui`: `ui`, `interactive`, `dashboard`
  - `web`: `serve`, `server`, `http`
- [x] Add tests to verify that cobra suggests the aliased commands
- [x] Run `make lint` and `make test` to verify

## Acceptance Criteria

- `taskmd set` shows `update` in the "Did you mean?" list
- `taskmd ls` shows `list` in the "Did you mean?" list
- `taskmd view` shows `show` in the "Did you mean?" list
- `taskmd deps` shows `graph` in the "Did you mean?" list
- Existing fuzzy suggestions still work alongside the new alias-based suggestions
- All tests pass, lint passes

## Implementation Notes

Cobra's `Command` struct has a `SuggestFor` field (`[]string`) designed for this exact purpose. When an unknown command is entered, cobra checks both Levenshtein distance AND the `SuggestFor` lists of all registered commands:

```go
var updateCmd = &cobra.Command{
    Use:        "update",
    SuggestFor: []string{"set", "edit", "modify", "change"},
    // ...
}
```

This is a minimal, non-breaking change -- just add the `SuggestFor` slice to each command definition. No new logic or packages needed.

## References

- [cobra SuggestFor documentation](https://pkg.go.dev/github.com/spf13/cobra#Command)
- `apps/cli/internal/cli/update.go` -- example command to modify
- `apps/cli/internal/cli/root.go` -- root command configuration
