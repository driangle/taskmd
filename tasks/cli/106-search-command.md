---
id: "106"
title: "Add search command for full-text task search"
status: pending
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

- [ ] Create `internal/cli/search.go` with the `search` command
- [ ] Accept a positional query argument (required)
- [ ] Scan all task files using the existing scanner
- [ ] Search across both frontmatter `title` and markdown body content
- [ ] Implement case-insensitive matching
- [ ] Display matching tasks with ID, title, and a snippet showing the match in context
- [ ] Support `--format` flag (table, json, yaml) consistent with other commands
- [ ] Support `--task-dir` flag for custom task directories
- [ ] Highlight or indicate match location in output
- [ ] Add comprehensive tests in `internal/cli/search_test.go`
- [ ] Register command with `rootCmd`

## Acceptance Criteria

- `taskmd search "authentication"` returns all tasks mentioning "authentication" in title or body
- Search is case-insensitive
- Output shows task ID, title, and a context snippet around the match
- Supports standard `--format` and `--task-dir` flags
- Returns a clear message when no results are found
- Tests cover matching in titles, bodies, no-match case, and format options
