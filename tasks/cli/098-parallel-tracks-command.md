---
id: "098"
title: "Parallel tracks command with spatial overlap detection"
status: pending
priority: medium
effort: large
tags:
  - cli
  - go
  - mvp
created: 2026-02-14
---

# Parallel Tracks Command with Spatial Overlap Detection

## Objective

Add a new `tracks` command that shows which actionable tasks can be worked on in parallel, taking into account both dependency ordering and spatial overlaps (tasks that touch the same code areas). This helps teams and multi-agent workflows avoid merge conflicts by assigning non-overlapping tasks to separate parallel tracks.

## Concepts

### `touches` frontmatter field

A new optional frontmatter field on tasks that declares which abstract code scopes a task will modify:

```yaml
id: "096"
title: "Add colors to graph command output"
touches:
  - "cli/graph"
  - "cli/output"
```

Scopes are user-defined abstract identifiers (not file paths). Two tasks that share a scope are considered spatially overlapping and should not be worked on simultaneously.

### Scope definitions in `.taskmd.yaml`

Users define what each abstract scope maps to concretely in their project config:

```yaml
# .taskmd.yaml
scopes:
  cli/graph:
    paths:
      - "apps/cli/internal/graph/"
      - "apps/cli/internal/cli/graph.go"
  cli/output:
    paths:
      - "apps/cli/internal/cli/output.go"
      - "apps/cli/internal/cli/format.go"
  cli/scanner:
    paths:
      - "apps/cli/internal/scanner/"
  web/board:
    paths:
      - "apps/web/src/components/board/"
  core/model:
    paths:
      - "apps/cli/internal/model/"
```

The `paths` entries can be files or directories. This mapping serves as documentation and can be used by tooling to validate or suggest scopes.

### Parallel tracks algorithm

A "track" is a sequence of tasks that can be executed one after another. Multiple tracks can run simultaneously if they don't overlap spatially.

**Algorithm outline:**

1. Gather all actionable tasks (same filter as `next`: status is `pending` or `in-progress`, all dependencies met).
2. Build a set of overlap edges: two tasks overlap if their `touches` arrays share at least one scope.
3. Assign tasks to tracks greedily:
   - Sort actionable tasks by the `next` scoring (priority, critical path, downstream impact).
   - For each task in order, try to assign it to an existing track where it doesn't overlap with any task already in that track.
   - If no compatible track exists, create a new track.
4. Within each track, order tasks by dependency chain and then by score.

Tasks without a `touches` field are assumed to have no overlaps and can be placed in any track (or shown separately as "flexible" tasks).

## Tasks

- [ ] Add `touches` field support to the task model (`internal/model/task.go`)
- [ ] Update the task parser to read `touches` from frontmatter
- [ ] Add `scopes` config support to `.taskmd.yaml` loading
- [ ] Implement the parallel tracks algorithm (`internal/tracks/`)
  - [ ] Overlap detection: given two tasks, check if `touches` arrays intersect
  - [ ] Track assignment: greedy allocation of tasks to non-overlapping tracks
  - [ ] Track ordering: sort tasks within each track by dependency + score
- [ ] Implement the `tracks` CLI command (`internal/cli/tracks.go`)
  - [ ] ASCII output: show tracks side by side or as labeled lists
  - [ ] JSON output: structured tracks data
  - [ ] Support `--filter` flag for scoping (e.g., `group=cli`)
  - [ ] Support `--limit` flag to cap number of tracks shown
- [ ] Add `touches` field to the taskmd specification (`docs/taskmd_specification.md`)
- [ ] Add `scopes` config to the configuration documentation
- [ ] Add comprehensive tests
  - [ ] Track algorithm with various overlap scenarios
  - [ ] Tasks with no `touches` (flexible placement)
  - [ ] Tasks with identical `touches` (forced sequential)
  - [ ] Integration test with real task files

## Example Output

```
$ taskmd tracks

Track 1 (cli/graph, cli/output):
  1. [096] Add colors to graph command output
  2. [099] Graph export to PNG

Track 2 (web/board):
  1. [090] Add drag-and-drop to board view
  2. [093] Board view filters

Track 3 (cli/scanner, core/model):
  1. [097] Add ignore directories option

Flexible (no declared overlaps):
  1. [088] Next ID command
  2. [091] Display version in header
```

## Acceptance Criteria

- New `touches` optional frontmatter field is supported in the task model and parser
- Scope definitions can be configured in `.taskmd.yaml` under a `scopes` key
- `taskmd tracks` command outputs parallel work tracks
- Tasks sharing a `touches` scope are never placed in the same track
- Tasks with no `touches` are shown as flexible / assignable to any track
- Dependency ordering is respected within tracks
- JSON and ASCII output formats are supported
- Unknown scopes in `touches` (not defined in config) are allowed but produce a warning
- Comprehensive test coverage for the algorithm and command
