---
id: "115"
title: "Add lightweight task metadata command (CLI, skill, MCP tool)"
status: completed
priority: medium
effort: medium
tags:
  - mvp
  - feature
  - cli
  - plugin
  - mcp
created: 2026-02-15
---

# Add lightweight task metadata command

## Objective

Create a lightweight alternative to `get` that returns only the basic frontmatter metadata of a task (no body content, no resolved dependency details, no context files, no worklog info). This is useful when you just need to quickly check a task's status, priority, or other metadata without the overhead of the full `get` output.

## Tasks

- [ ] Add CLI command `taskmd status <query>` that outputs only frontmatter fields (id, title, status, priority, effort, tags, owner, parent, created, dependencies)
  - Support `--format text|json|yaml` output formats
  - Reuse the same fuzzy-matching logic from `get`
- [ ] Add MCP tool `status` in `internal/mcp/` that returns only frontmatter metadata as JSON
- [ ] Add Claude Code plugin skill `get-task-status` in `claude-code-plugin/skills/`
- [ ] Add tests for CLI command (happy path, formats, flags, error handling)
- [ ] Add tests for MCP tool

## Acceptance Criteria

- `taskmd status <id>` returns only frontmatter fields, no body/content
- Output is noticeably smaller and faster than `get` for the same task
- All three surfaces (CLI, MCP, skill) return consistent data
- Existing `get` command is unchanged
