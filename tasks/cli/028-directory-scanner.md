---
id: "028"
title: "Directory scanner (multi-file support)"
status: completed
priority: medium
effort: medium
dependencies: ["016"]
tags:
  - cli
  - go
  - core
created: 2026-02-08
---

# Directory Scanner (Multi-File Support)

## Objective

Implement recursive directory scanning to find and parse all `.md` task files in a directory tree for use in TUI and directory-based commands.

## Tasks

- [x] Create `internal/scanner/` package for directory scanning
- [x] Implement recursive directory traversal
- [x] Filter for `.md` files only
- [x] Skip common non-task directories:
  - `.git`, `.github`
  - `node_modules`, `vendor`
  - `.next`, `.nuxt`, `dist`, `build`
  - Hidden directories (`.*)` by default
- [x] Parse each discovered file using the markdown parser
- [x] Derive task group from directory structure (e.g., `tasks/cli/*.md` → group: "cli")
- [x] Frontmatter `group` field takes precedence over derived group
- [x] Build task collection (map or slice)
- [x] Track file path for each task
- [x] Handle parsing errors gracefully (log warnings, continue scanning)
- [x] Write unit tests with temporary test directory

## Acceptance Criteria

- Scanner finds all `.md` files in directory tree
- Common non-task directories are skipped
- Tasks are parsed and grouped correctly
- Directory-derived groups work (e.g., `tasks/cli/` → "cli")
- Frontmatter `group` overrides directory group
- Parse errors don't crash scanner
- File paths are tracked for each task
- Unit tests pass

## Notes

This is for multi-file/directory mode, primarily used by TUI. Most commands will work with single files or stdin.
