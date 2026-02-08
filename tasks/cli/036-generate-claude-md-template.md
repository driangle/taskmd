---
id: "036"
title: "Generate CLAUDE.md template for taskmd users"
status: completed
priority: medium
effort: small
dependencies: []
tags:
  - cli
  - documentation
  - dx
  - claude-integration
created: 2026-02-08
---

# Generate CLAUDE.md Template for Taskmd Users

## Objective

Create a CLAUDE.md template file that users can place in their task directory to help Claude understand the taskmd specification and use the CLI effectively when working with their tasks.

## Context

Users who work with Claude Code in their taskmd repositories would benefit from a standardized CLAUDE.md file that:
- Explains the taskmd file format and conventions
- Documents common taskmd CLI commands and usage patterns
- Provides guidance on task management workflows
- Helps Claude maintain consistency with the taskmd specification

This should be minimal but practical—focusing on the most essential information Claude needs to work effectively with taskmd files.

## Tasks

- [x] Create `docs/templates/CLAUDE.md` template file
- [x] Include taskmd file format essentials:
  - Frontmatter schema (id, title, status, priority, etc.)
  - Status values (pending, in-progress, blocked, completed)
  - Priority levels (low, medium, high, critical)
  - Dependencies format
- [x] Document essential CLI commands:
  - `taskmd list` - List and filter tasks
  - `taskmd validate` - Validate task files
  - `taskmd graph` - Visualize dependencies
  - `taskmd next` - Find next available task
  - `taskmd stats` - Show project statistics
- [x] Include task workflow guidance:
  - How to update task status
  - How to check dependencies before starting
  - When to use validation
- [x] Keep it concise (aim for ~100-150 lines)
- [x] Use clear, actionable language
- [ ] Add to README with instructions for users

## Acceptance Criteria

- CLAUDE.md template exists in `docs/templates/`
- Covers essential taskmd format specification
- Documents key CLI commands with examples
- Provides task workflow best practices
- Concise and scannable (not overwhelming)
- README updated with instructions to copy template
- Template is ready for users to copy into their task directories

## Example Structure

```markdown
# Working with Taskmd Tasks

This project uses taskmd for task management.

## Task File Format
[Essential frontmatter schema]

## Common Commands
[Key CLI commands with examples]

## Workflow
[Task management best practices]
```

## References

- `docs/TASKMD_SPEC.md` - Full specification
- Existing `CLAUDE.md` in this repo (as reference, not template)
