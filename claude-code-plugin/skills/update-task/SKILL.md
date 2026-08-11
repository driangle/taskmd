---
name: update-task
description: Update an existing task's fields (status, priority, title, tags, dependencies, etc.). Use when the user wants to modify a task's properties.
allowed-tools: Bash, Read, Edit
---

# Update Task

Update fields of an existing task.

## Instructions

The user's query is in `$ARGUMENTS` (e.g. "set task 042 to high priority and in-progress", "rename task 15 to Fix auth bug", "add tag backend to 042").

1. **Parse the user's input** from `$ARGUMENTS` to extract:
   - The **task ID** (required)
   - The **fields to update** and their new values

2. **Look up the task**: Run `taskmd get <ID>` to confirm the task exists
   - If not found, run `taskmd list` to show available tasks and ask the user which one they meant

3. **Determine how to apply each update**:

   ### Fields supported by `taskmd set` (use CLI):
   - `--title` — new title; also updates the matching `#` heading in the task body. Add `--rename` to move the file to `<id>-<new-slug>.md` (the file otherwise keeps its old name, which is harmless — tasks resolve by ID).
   - `--status` — pending, in-progress, completed, in-review, blocked, cancelled
   - `--priority` — low, medium, high, critical
   - `--effort` — small, medium, large
   - `--type` — feature, bug, improvement, chore, docs
   - `--owner` — assignee name
   - `--parent` — parent task ID (empty string to clear)
   - `--phase` — phase ID (must match a phase id in `.taskmd.yaml`; empty string to clear). If the phase doesn't exist in `.taskmd.yaml`, add it there first (see `add-task` skill for the phases config format).
   - `--add-tag` / `--remove-tag` — add or remove tags (repeatable)
   - `--add-pr` / `--remove-pr` — add or remove PR URLs (repeatable)
   - `--depends-on` — replace dependencies (comma-separated IDs, e.g. `010,015`; empty string to clear)

   Build a single `taskmd set <ID>` command with all applicable flags:
   ```bash
   taskmd set 042 --priority high --status in-progress --add-tag backend
   ```

   ### Fields NOT supported by `taskmd set` (edit file directly):
   - **custom frontmatter fields** — edit the frontmatter directly
   - **description / subtasks / acceptance criteria** — edit the markdown body

   For direct file edits:
   1. Read the task file with the `Read` tool
   2. Use the `Edit` tool to make the change
   3. Run `taskmd validate` to ensure the file is still valid

4. **Handle errors**:
   - If the task ID is invalid or not found, tell the user and suggest running `taskmd list`
   - If a field value is invalid (e.g. `--priority mega`), explain the valid options
   - If `taskmd validate` fails after a direct edit, fix the issue before confirming

5. **Confirm** the changes to the user, showing what was updated
