---
name: next-task
description: Get the next recommended task to work on. Use when the user asks what to work on next or needs a task assignment.
allowed-tools: Bash, Read
---

# Next Task

Find the next recommended task to work on using the `taskmd` CLI.

## Instructions

1. Run `taskmd next` with any arguments the user provided via `$ARGUMENTS`
   - If `$ARGUMENTS` is empty, run: `taskmd next`
   - If `$ARGUMENTS` contains flags, pass them through: `taskmd next $ARGUMENTS`
   - Common usage: `/next-task --filter tag=mvp` to find the next MVP task
2. Read the recommended task file to get full details
3. Present the task summary including: ID, title, status, priority, and description
