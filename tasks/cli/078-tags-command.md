---
id: "078"
title: "tags command - List all tags with task counts"
status: completed
priority: medium
effort: small
tags:
  - mvp
  - cli
  - commands
created: 2026-02-14
---

# Tags Command - List All Tags with Task Counts

## Objective

Implement a `tags` CLI command that outputs all tags used across task files along with the number of tasks per tag, sorted from most to least used. Support the same filtering flags available in other commands (e.g., `--status`, `--priority`, `--dir`).

## Tasks

- [ ] Create `internal/cli/tags.go` for the tags command
- [ ] Scan all task files and collect tags from frontmatter
- [ ] Aggregate tag counts across all matching tasks
- [ ] Sort output by count descending (most used first)
- [ ] Support standard filtering flags (`--status`, `--priority`, `--tag`, `--dir`)
- [ ] Support output formats: `table` (default), `json`
- [ ] Create `internal/cli/tags_test.go` with comprehensive tests
  - [ ] Happy path tests
  - [ ] Format tests (table, json)
  - [ ] Flag/filter tests
  - [ ] Edge cases (no tags, single tag, ties in count)

## Acceptance Criteria

- `taskmd tags` displays all tags with their task counts, sorted by count descending
- Filtering flags work consistently with other commands (e.g., `--status pending` only counts tags from pending tasks)
- `--format json` outputs structured tag data
- Comprehensive test coverage following project testing conventions
- Passes `make lint` and `make test`

## Examples

```bash
taskmd tags
taskmd tags --status pending
taskmd tags --format json
taskmd tags --dir tasks/cli
```

### Expected table output

```
TAG         COUNT
cli           12
mvp            8
commands       5
go             3
```
