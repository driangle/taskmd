---
name: complete-task
description: Mark a task as completed. Use when the user wants to mark a task as done or complete.
allowed-tools: Bash, Read, Edit
---

# Complete Task

Mark a task as completed using the `taskmd` CLI.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077`). If `$ARGUMENTS` is empty or does not contain a task ID, infer the task from conversation context (e.g., the task currently being worked on). If the task cannot be determined, ask the user which task to complete.

> ⚠️ **Worktree safety — never `cd` before a `taskmd` write.**
> Run every `taskmd set …` command (including `--verify` and the `in-review` branch)
> from your current working directory. Do **not** prepend `cd /path/to/repo` or `cd`
> into a "primary working directory" first: inside a git worktree that path is a
> *different checkout on a different branch*. Unlike the `Edit`/`Write` tools (which
> the harness's worktree isolation guards), a shell `taskmd` write is **not** isolated
> — after such a `cd` it will silently flip the status in the wrong checkout. taskmd
> resolves its task directory from the current directory, so running in place is what
> keeps the write in *this* worktree. (Outside a git repo, running in place is already
> correct.)

1. **Read the task file** to understand the full task scope:
   - Run `taskmd get <ID>` to get the task contents
   - Identify all **subtask checklists** (`- [ ]` / `- [x]` items) in the task body
   - Identify any **acceptance criteria** section

2. **Verify subtasks and acceptance criteria are met**:
   - Review each subtask checklist item — confirm the work has been done
   - Review each acceptance criterion — confirm it is satisfied
   - **Check off** (`- [x]`) any items that are complete but not yet checked off by editing the task file
   - If any items are genuinely incomplete, report them to the user and ask how to proceed — do NOT mark the task as completed

3. **Add a final worklog entry** (if worklogs are enabled):
   - Check `.taskmd.yaml` for `worklogs: true` -- worklogs are opt-in, so skip this step unless the key is explicitly set to `true`
   - Otherwise, find the worklog file at `tasks/<group>/.worklogs/<ID>.md` (or `tasks/.worklogs/<ID>.md`)
   - If a worklog exists, append a timestamped completion summary

4. **Check the workflow mode** in `.taskmd.yaml` (run these `taskmd set` commands from the current working directory — do **not** `cd` elsewhere first, see the ⚠️ note above):
   - If `workflow: pr-review` is set, use `taskmd set $ARGUMENTS --status in-review` instead of `completed` (note: in pr-review mode, tasks are completed by merging the PR, not by setting status directly)
   - Otherwise (default `solo` mode), run `taskmd set $ARGUMENTS --status completed --verify`
   - The `--verify` flag runs any verification checks defined in the task before applying the status change
   - If verification fails, report the failures to the user

5. Confirm the status change to the user
