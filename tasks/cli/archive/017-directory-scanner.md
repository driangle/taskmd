---
id: "archive-017"
title: "Directory scanner"
status: pending
priority: high
effort: medium
dependencies: ["016"]
tags:
  - cli
  - go
  - core
  - archived
created: 2026-02-08
---

# Directory Scanner

## Objective

Implement recursive directory scanning that finds all `.md` task files in the current directory (and subdirectories), parses them, and builds an in-memory task collection.

## Tasks

- [ ] Implement `Scanner` in `internal/scanner/` that walks a directory tree
- [ ] Filter for `.md` files only
- [ ] Skip common non-task directories (`.git`, `node_modules`, `.next`, `vendor`, etc.)
- [ ] Parse each discovered file using the markdown parser from task 016
- [ ] Derive task group from directory structure: tasks in `tasks/cli/*.md` get group `cli`, tasks in `tasks/*.md` (root) get no group (or a default)
- [ ] If a task has an explicit `group` frontmatter field, that takes precedence over the directory-derived group
- [ ] Build a `TaskStore` (map or slice) as the in-memory collection
- [ ] Track file path → task mapping for later updates
- [ ] Handle errors gracefully: log warnings for unparseable files, continue scanning
- [ ] Write unit tests with a temporary directory of test `.md` files

## Acceptance Criteria

- Running the scanner on a directory returns all valid tasks found
- Unparseable files are skipped with a warning, not a crash
- Common non-task directories are ignored
- File paths are tracked alongside tasks
- Tasks in subdirectories are assigned the subdirectory name as their group
- Explicit `group` frontmatter overrides the directory-derived group
- Unit tests pass
