---
name: do-task
description: Look up a task by ID or name and start working on it. Use when the user wants to pick up and execute a task.
allowed-tools: Bash, Read, Glob, Grep, Write, Edit, Task, EnterPlanMode
---

# Do Task

Look up a task and start working on it.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077` or a task name/keyword).

1. **Look up the task**: Run `taskmd get $ARGUMENTS` to find the task
   - If not found, run `taskmd list` to show available tasks and ask the user which one they meant
2. **Read the task file** with the `Read` tool to get the full description, subtasks, and acceptance criteria
3. **Mark the task as in-progress**: Run `taskmd set --task-id <ID> --status in-progress`
4. **Do the task**: Follow the task description and complete the work described
   - Use `EnterPlanMode` for non-trivial implementation tasks
   - Check off subtasks (`- [x]`) in the task file as you complete them
5. **Mark the task as completed** when done: Run `taskmd set --task-id <ID> --status completed --verify`
   - The `--verify` flag will run any verification checks defined in the task before applying the status change
   - If verification fails, fix the issues and try again
