---
name: complete-task
description: Mark a task as completed. Use when the user wants to mark a task as done or complete.
allowed-tools: Glob, Read, Edit, Write
---

# Complete Task

Mark a task as completed — no CLI required.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077`). If `$ARGUMENTS` is empty or does not contain a task ID, infer the task from conversation context (e.g., the task currently being worked on). If the task cannot be determined, ask the user which task to complete.

1. **Find the task file**:
   - Read `.taskmd.yaml` for custom `dir` (default: `tasks`) and `workflow` mode (default: `solo`)
   - Use `Glob` for `<task-dir>/**/*$ARGUMENTS*.md`
   - Read frontmatter to confirm the ID matches
   - If not found, list available tasks

2. **Read the task file** to understand the full task scope:
   - Get current status and verify fields
   - Identify all **subtask checklists** (`- [ ]` / `- [x]` items) in the task body
   - Identify any **acceptance criteria** section

3. **Verify subtasks and acceptance criteria are met**:
   - Review each subtask checklist item — confirm the work has been done
   - Review each acceptance criterion — confirm it is satisfied
   - **Check off** (`- [x]`) any items that are complete but not yet checked off by editing the task file
   - If any items are genuinely incomplete, report them to the user and ask how to proceed — do NOT mark the task as completed

4. **Add a final worklog entry** (if worklogs are enabled):
   - Check `.taskmd.yaml` for `worklogs: true` — worklogs are opt-in, so skip this step unless the key is explicitly set to `true`
   - If enabled, find or create the worklog file at `<task-dir>/<group>/.worklogs/<ID>.md` (or `<task-dir>/.worklogs/<ID>.md` for root tasks)
   - Append a timestamped completion summary

5. **Check the workflow mode** from `.taskmd.yaml`:

   ### Solo mode (default)
   - If the task has `verify` checks in frontmatter:
     - For `bash` type: Run each `run` command via Bash (in the specified `dir` or project root) and check exit code
     - For `assert` type: Evaluate each `check` by inspecting the codebase
     - If any check fails, report failures and do NOT mark as completed
   - If all checks pass (or no verify checks): Edit the frontmatter to set `status: completed`

   ### PR-review mode
   - Edit the frontmatter to set `status: in-review` instead of `completed`
   - Note to the user that in pr-review mode, the task completes when the PR is merged

6. **Confirm** the status change to the user

See `SPEC_REFERENCE.md` (in the plugin root) for valid field values, workflow modes, and verify check format.
