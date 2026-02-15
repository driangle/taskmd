---
id: "106"
title: "Add search command for full-text task search"
status: completed
priority: medium
effort: medium
tags:
  - search
  - dx
  - mvp
touches:
  - cli
created: 2026-02-14
---

# Add Search Command for Full-Text Task Search

## Objective

Add a new `taskmd search <query>` command that performs full-text search across all task titles and markdown bodies. This enables quickly finding tasks by keyword without manually browsing files or relying on external tools like grep.

## Tasks

- [x] Create `internal/cli/search.go` with the `search` command
- [x] Accept a positional query argument (required)
- [x] Scan all task files using the existing scanner
- [x] Search across both frontmatter `title` and markdown body content
- [x] Implement case-insensitive matching
- [x] Display matching tasks with ID, title, and a snippet showing the match in context
- [x] Support `--format` flag (table, json, yaml) consistent with other commands
- [x] Support `--task-dir` flag for custom task directories
- [x] Highlight or indicate match location in output
- [x] Add comprehensive tests in `internal/cli/search_test.go`
- [x] Register command with `rootCmd`

## Acceptance Criteria

- `taskmd search "authentication"` returns all tasks mentioning "authentication" in title or body
- Search is case-insensitive
- Output shows task ID, title, and a context snippet around the match
- Supports standard `--format` and `--task-dir` flags
- Returns a clear message when no results are found
- Tests cover matching in titles, bodies, no-match case, and format options
