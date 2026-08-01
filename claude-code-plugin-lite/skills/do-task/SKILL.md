---
name: do-task
description: Look up a task by ID or name and start working on it. Use when the user wants to pick up and execute a task.
allowed-tools: Glob, Read, Write, Edit, Grep, Bash, Task, EnterPlanMode
---

# Do Task

Look up a task and start working on it — no CLI required.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077` or a task name/keyword).

1. **Find the task file**:
   - Read `.taskmd.yaml` for custom `dir` (default: `tasks`)
   - Use `Glob` for `<task-dir>/**/*$ARGUMENTS*.md`
   - If multiple matches, read frontmatter to find the exact ID match
   - If not found, scan all task files and list available tasks for the user to choose

2. **Read the task file** with the `Read` tool to get the full description, subtasks, and acceptance criteria

3. **Mark the task as in-progress**:
   - Use `Edit` to change `status: pending` (or current status) to `status: in-progress` in the frontmatter

4. **Start a worklog entry** (only if worklogs are enabled):
   - Check `.taskmd.yaml` for `worklogs: true`. If it is not explicitly enabled, skip all worklog steps silently — do not mention worklogs, do not tell the user they are disabled, just move on.
   - If enabled, find or create the worklog file at `<task-dir>/<group>/.worklogs/<ID>.md` (or `<task-dir>/.worklogs/<ID>.md` for root tasks)
   - Append a timestamped entry noting your approach and initial findings

5. **Do the task**: Follow the task description and complete the work described
   - Use `EnterPlanMode` for non-trivial implementation tasks
   - Check off subtasks (`- [x]`) in the task file using `Edit` as you complete them
   - Append worklog entries when you make key decisions, hit blockers, or complete significant subtasks
   - In the Plan, include a reference to the original task ID and task file path

6. **Write a final worklog entry** (only if worklogs are enabled — otherwise skip silently) summarizing what was done, decisions made, and any open items

7. **Mark the task as done**: Use the `/complete-task` skill (invoke it with the task ID) to complete the task. It handles verification and status changes automatically.

## Worklog Format

Each worklog entry uses a timestamp heading followed by free-form notes:

```markdown
## 2026-02-15T10:30:00Z

Started implementation of the search feature.

**Approach:** Using full-text search with the existing SQLite database
rather than adding Elasticsearch — simpler and sufficient for our scale.

**Completed:**

- [x] Added search query parser
- [x] Created search index

**Next:** Add result ranking and write tests.
```

See `SPEC_REFERENCE.md` (in the plugin root) for frontmatter schema, field values, and workflow modes.
