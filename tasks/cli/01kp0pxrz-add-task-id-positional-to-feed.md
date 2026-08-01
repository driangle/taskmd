---
title: "Add optional task-id positional to feed for per-task history"
id: "01kp0pxrz"
status: completed
priority: medium
type: feature
tags: ["cli"]
created: "2026-04-12"
completed_at: 2026-08-01
---

# Add optional task-id positional to feed for per-task history

## Objective

Let `taskmd feed` accept an optional `<task-id>` positional that scopes the
activity feed to a single task, giving a chronological timeline of that task's
status transitions. This reuses the existing git-log/diff machinery in
`sdk/go/feed` rather than adding a separate `history` command that would
duplicate it.

- `taskmd feed` — repo-wide activity feed (unchanged).
- `taskmd feed cli-049` — timeline of changes to that one task.
- `taskmd feed cli-049 --field priority` — pivot the timeline onto another field.

The feed SDK already extracts per-commit frontmatter `FieldChange` records
(field, old, new) with author and timestamp (`sdk/go/feed/diff.go`,
`sdk/go/feed/feed.go`), so no new git parsing is required — this is a filter
plus an optional field pivot over data feed already produces.

Example output:
```
$ taskmd feed cli-049
2026-04-12  pending ← cancelled     (by driangle)
2026-04-10  cancelled ← in-progress (by driangle)
2026-04-08  in-progress ← pending   (by driangle)
2026-04-05  created                  (by driangle)
```

## Tasks

- [x] Change `feedCmd` args from `cobra.NoArgs` to `cobra.MaximumNArgs(1)`
- [x] When a task-id positional is given, resolve it to a task file and scope `feed.Query` to that file (added `Options.TaskFile`, `--follow` single-file git log, and a task-id guard against `--follow` false positives)
- [x] Add a `--field` flag (default `status`) to select which frontmatter field's transitions to surface for the single-task view
- [x] Ensure the initial "created" event is shown for the single-task view (parser now treats git copy events, `C<score>`, as creations)
- [x] Keep repo-wide behavior (no positional) unchanged, including `--limit`, `--since`, `--scope`, `--source`
- [x] Reuse existing output helpers/formats (text, json); no new git-log parsing
- [x] Handle edge cases: unknown/ambiguous task id, task file not tracked by git, no changes for the chosen field
- [x] Update `feed` help text and any docs to describe the optional positional and `--field`
- [x] Add tests in `internal/cli/feed_test.go` covering the single-task path (plus SDK unit tests and e2e git-history tests)

## Acceptance Criteria

- `taskmd feed <task-id>` shows that task's status transitions with dates and authors
- Output includes the initial "created" event
- Reopening events (completed→pending, cancelled→pending) are clearly visible
- `taskmd feed` with no argument behaves exactly as before
- `--field priority` tracks priority changes instead of status
- `--format json` produces machine-readable output for the single-task view
- Graceful error when the id is unknown, run outside a git repo, or on an untracked file
- No duplication of git-log/diff logic — the single-task view consumes `sdk/go/feed`
- Tests cover: happy path, multiple transitions, task with no changes, JSON output
