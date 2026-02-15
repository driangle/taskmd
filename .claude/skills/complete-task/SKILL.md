---
name: complete-task
description: Mark a task as completed. Use when the user wants to mark a task as done or complete.
allowed-tools: Bash
---

# Complete Task

Mark a task as completed using the `taskmd` CLI.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077`).

1. Run `taskmd set --task-id $ARGUMENTS --status completed --verify`
   - The `--verify` flag runs any verification checks defined in the task before applying the status change
   - If verification fails, report the failures to the user
2. Confirm the status change to the user
