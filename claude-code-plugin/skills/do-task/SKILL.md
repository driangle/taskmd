---
name: do-task
description: Look up a task by ID or name and start working on it. Use when the user wants to pick up and execute a task.
allowed-tools: Bash, Read, Glob, Grep, Write, Edit, Task, EnterPlanMode
---

# Do Task

Look up a task and start working on it.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077` or a task name/keyword).

> ⚠️ **Worktree safety — never `cd` before a `taskmd` write.**
> Run every `taskmd set …` command from your current working directory. Do **not**
> prepend `cd /path/to/repo` or `cd` into a "primary working directory" first: inside
> a git worktree that path is a *different checkout on a different branch*. Unlike the
> `Edit`/`Write` tools (which the harness's worktree isolation guards), a shell
> `taskmd` write is **not** isolated — after such a `cd` it will silently flip the
> status in the wrong checkout. taskmd resolves its task directory from the current
> directory, so running in place is what keeps the write in *this* worktree. (Outside
> a git repo, running in place is already correct.)

1. **Look up the task**: Run `taskmd get $ARGUMENTS` to find the task
   - If not found, run `taskmd list` to show available tasks and ask the user which one they meant
2. **Read the task file** with the `Read` tool to get the full description, subtasks, and acceptance criteria
3. **Mark the task as in-progress**: Run `taskmd set <ID> --status in-progress` from the current working directory (do **not** `cd` elsewhere first — see the ⚠️ note above)
4. **Start a worklog entry** (only if worklogs are enabled):
   - Check `.taskmd.yaml` for `worklogs: true`. Worklogs are opt-in — an absent key means disabled. If it is not explicitly enabled, skip all worklog steps silently — do not mention worklogs, do not tell the user they are disabled, just move on.
   - If enabled, find or create the worklog file at `tasks/<group>/.worklogs/<ID>.md` (or `tasks/.worklogs/<ID>.md` for root tasks)
   - Append a timestamped entry noting your approach and initial findings
5. **Do the task**: Follow the task description and complete the work described
   - Use `EnterPlanMode` for non-trivial implementation tasks
   - Check off subtasks (`- [x]`) in the task file as you complete them
   - Append worklog entries when you make key decisions, hit blockers, or complete significant subtasks
   - In the Plan, include a reference to the original task ID, and task file path.
6. **Write a final worklog entry** (only if worklogs are enabled — otherwise skip silently) summarizing what was done, decisions made, and any open items
7. **Mark the task as done**: Use the `/complete-task` skill (invoke it with the task ID) to complete the task. It handles verification and status changes automatically.

## Worklog Format

Each worklog entry uses a timestamp heading followed by free-form notes:

```markdown
## 2026-02-15T10:30:00Z

Started implementation of the search feature.

**Approach:** Using full-text search with the existing SQLite database
rather than adding Elasticsearch -- simpler and sufficient for our scale.

**Completed:**

- [x] Added search query parser
- [x] Created search index

**Next:** Add result ranking and write tests.
```
