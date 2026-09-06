---
id: "01m1wrg50"
title: "Show commit hash in feed text output"
status: completed
priority: low
effort: small
dependencies: []
tags: ["feed", "cli", "output"]
created_at: 2026-09-06
completed_at: 2026-09-06
---

# Show commit hash in feed text output

## Objective

`taskmd feed` already carries the commit SHA on every git-sourced entry
(`feed.FeedEntry.Hash`, `sdk/go/feed/feed.go:18`) and emits it in `--format json`,
but the text renderer drops it. `writeFeedText` in
`apps/cli/internal/cli/feed.go` formats each git entry as date + author +
message only:

```go
date := formatDim(entry.Timestamp.Format("2006-01-02 15:04"), r)
author := formatLabel(entry.Author, r)
fmt.Printf("%s %s: %s\n", date, author, entry.Message)
```

So the default view answers *when* and *who* but not *which commit*:

```
$ taskmd feed 01kkhc0y5
2026-03-12 17:29 German Greiner: chore: rename milestone feature to phase (task 01kkhc0y5)
  [Completed] tasks/01kkhc0y5-rename-milestone-feature-to-phase.md (01kkhc0y5)
```

Answering "which commit created / completed this task" — the whole point of a
per-task feed — currently forces a detour through `--format json | jq`. Print the
abbreviated hash inline so the text output is self-sufficient.

## Approach

Render the short hash (7–8 chars) dimmed, between the timestamp and the author,
for git-sourced entries:

```
2026-03-12 17:29 76acc1a0 German Greiner: chore: rename milestone feature to phase
```

Unconditional, not behind a flag — it is one short token and its absence is the
gap being fixed. Worklog entries have no hash and keep their current shape;
`entry.Hash` may be empty, so guard the prefix rather than slicing blindly.

## Tasks

- [x] Print the abbreviated `entry.Hash` in `writeFeedText` for git entries, dimmed, guarding against an empty hash
- [x] Leave worklog entries (`writeWorklogEntryText`) unchanged
- [x] Add a `writeFeedText` test asserting the hash appears for a git entry and that a hash-less entry renders without a stray separator
- [x] Update the `feed` examples in `apps/docs/guide/cli.md` if they show text output

## Acceptance Criteria

- `taskmd feed <id>` and `taskmd feed` show an abbreviated commit hash on every git-sourced line
- Worklog-sourced entries are visually unchanged
- An entry with an empty `Hash` renders cleanly (no double space, no empty column)
- `--format json` output is byte-for-byte unchanged
- `make test` and `make lint` pass
