---
id: "016"
title: "Task model & markdown parsing"
status: completed
priority: high
effort: medium
dependencies: ["015"]
tags:
  - cli
  - go
  - core
  - mvp
created: 2026-02-08
---

# Task Model & Markdown Parsing

## Objective

Define the core Task data model in Go and implement markdown parsing that extracts YAML frontmatter and body content from `.md` task files. This is the foundation that all other CLI features build on.

## Tasks

- [x] Define `Task` struct in `internal/model/` matching the frontmatter schema (id, title, status, priority, effort, dependencies, tags, group, created)
- [x] Define supporting types/constants for Status, Priority, Effort enums
- [x] Include a `Group` field on `Task` — populated from frontmatter `group` field if present, otherwise derived from parent directory name by the scanner
- [x] Implement frontmatter+body parser using goldmark and a YAML frontmatter library (e.g. `go-yaml`)
- [x] Parse YAML frontmatter into `Task` struct
- [x] Preserve the markdown body as a string field on the Task
- [x] Handle edge cases: missing frontmatter, malformed YAML, empty files
- [x] Write unit tests for parsing valid, invalid, and edge-case markdown files

## Acceptance Criteria

- `group` frontmatter field is parsed when present
- A `.md` file with YAML frontmatter can be parsed into a `Task` struct
- Missing or malformed frontmatter returns a clear error (not a panic)
- Markdown body is preserved for later rendering
- Unit tests pass
