---
name: add-task
description: Create a new task file following the taskmd specification. Use when the user wants to add a new task to the project.
allowed-tools: Read, Glob, Write
---

# Add Task

Create a new task file under `./tasks/` following the taskmd specification.

## Instructions

The user's task description is in `$ARGUMENTS`.

1. **Read the specification** at `docs/taskmd_specification.md` for the correct format
2. **Determine the next task ID** by scanning existing files:
   - Run `Glob` for `tasks/**/*.md` to find all task files
   - Extract numeric IDs from filenames (pattern: `NNN-description.md`)
   - Pick the next sequential ID, zero-padded to 3 digits
3. **Choose the subdirectory** based on the task's domain:
   - `tasks/cli/` — CLI commands, Go backend, terminal features
   - `tasks/web/` — Web frontend, UI, React components
   - `tasks/` (root) — Cross-cutting, infrastructure, documentation, or unclear domain
4. **Determine `touches` scopes**:
   - Read `.taskmd.yaml` to see the existing `scopes` definitions
   - Based on the task description, identify which code areas the task will modify
   - Assign matching scope identifiers **only from the existing scopes list** (e.g., `cli/graph`, `web/board`, `sync/jira`)
   - Do NOT create new scopes or modify `.taskmd.yaml` — only use scopes already defined there
   - If no existing scope matches, omit `touches`
5. **Create the task file** named `<NNN>-<slug>.md` with:

```yaml
---
id: "<NNN>"
title: "<title from user>"
status: pending
priority: medium
effort: medium
tags: []
touches: []  # scope identifiers from .taskmd.yaml (omit if not applicable)
created: <today's date YYYY-MM-DD>
---
```

Followed by a markdown body with:
- An H1 heading matching the title
- An `## Objective` section describing the goal
- A `## Tasks` section with a checkbox list of subtasks
- An `## Acceptance Criteria` section

6. **Confirm** the created file path and ID to the user
