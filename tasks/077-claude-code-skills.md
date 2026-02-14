---
id: "077"
title: "Add Claude Code skills for next-task and add-task"
status: completed
priority: high
effort: small
tags:
  - dx
  - claude-code
  - mvp
created: 2026-02-14
---

# Add Claude Code Skills for next-task and add-task

## Objective

Create two Claude Code skills (`.claude/skills/`) in this repo so that Claude Code can quickly find the next task to work on and create new well-formed task files.

## Skills to Create

### 1. `next-task`

**Path:** `.claude/skills/next-task/SKILL.md`

Retrieves the next recommended task using the `taskmd` CLI.

- Runs `taskmd next --filter tag=mvp` if arguments include "mvp", otherwise runs `taskmd next`
- Supports passing arbitrary filter flags via `$ARGUMENTS` (e.g. `/next-task --filter tag=cli`)
- Displays the task details so Claude can start working on it

```yaml
---
name: next-task
description: Get the next recommended task to work on. Use when the user asks what to work on next or needs a task assignment.
allowed-tools: Bash, Read
---
```

### 2. `add-task`

**Path:** `.claude/skills/add-task/SKILL.md`

Creates a new task file under `./tasks/` following the taskmd specification.

- Reads `docs/taskmd_specification.md` for the correct frontmatter schema and format
- Determines the next available task ID by scanning existing task files
- Creates the task file with proper frontmatter (id, title, status, priority, effort, tags, created) and markdown body
- Places the file in the appropriate subdirectory (e.g. `tasks/cli/`, `tasks/web/`, or `tasks/`) based on the task's domain
- Uses `$ARGUMENTS` for the task description/title

```yaml
---
name: add-task
description: Create a new task file following the taskmd specification. Use when the user wants to add a new task to the project.
allowed-tools: Read, Glob, Write
---
```

## Tasks

- [ ] Create `.claude/skills/next-task/SKILL.md` with instructions for running `taskmd next`
- [ ] Create `.claude/skills/add-task/SKILL.md` with instructions for creating task files per the spec
- [ ] Verify both skills appear in Claude Code's `/` menu
- [ ] Test `/next-task` returns a task
- [ ] Test `/add-task` creates a valid task file

## Acceptance Criteria

- `/next-task` invokes `taskmd next` and presents the result
- `/next-task --filter tag=mvp` filters to MVP tasks
- `/add-task <description>` creates a correctly formatted task file with the next sequential ID
- New task files pass `taskmd validate`
- Both skills follow the SKILL.md format with proper frontmatter
