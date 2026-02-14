---
id: "076"
title: "Allow passing filepath or filename to get command"
status: pending
priority: medium
effort: small
tags:
  - cli
  - go
  - commands
  - dx
created: 2026-02-14
---

# Allow Passing Filepath or Filename to `get` Command

## Objective

Extend the `get` command to accept a file path or file name as the query argument, in addition to task IDs and titles. This makes it easy to look up a task when you already have the file open or can see the filename.

## Context

Currently `taskmd get <query>` matches against task IDs and titles. When working in an editor or terminal, the most readily available identifier is often the file path or filename (e.g. `tasks/cli/042-claude-code-plugin.md` or `042-claude-code-plugin.md`). Users should be able to pass these directly to `get` without having to extract the task ID first.

## Tasks

- [ ] Add filepath/filename matching as a new resolution step in `get`
- [ ] Match against full relative path (e.g. `tasks/cli/042-claude-code-plugin.md`)
- [ ] Match against just the filename (e.g. `042-claude-code-plugin.md`)
- [ ] Match against filename without extension (e.g. `042-claude-code-plugin`)
- [ ] Ensure uniqueness — if the filename is not unique across directories, prompt or error clearly
- [ ] Filepath matching should take priority over fuzzy title matching but after exact ID/title match
- [ ] Add tests for all new matching paths
- [ ] Test ambiguous filename case (same filename in different directories)

## Matching Priority (updated)

1. Exact match by task ID (existing)
2. Exact match by task title (existing)
3. **Match by file path or filename (new)**
4. Fuzzy match across IDs and titles (existing)

## Acceptance Criteria

- `taskmd get tasks/cli/042-claude-code-plugin.md` returns the task
- `taskmd get 042-claude-code-plugin.md` returns the task
- `taskmd get 042-claude-code-plugin` returns the task
- If a filename matches multiple tasks (in different directories), the user is prompted to choose or an error is shown
- Existing ID and title matching behavior is unchanged
- All new code paths have test coverage
