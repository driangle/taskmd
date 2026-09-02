---
name: list-tasks
description: List tasks from the project's taskmd files, with optional filters (status, priority, phase, group, owner). Use whenever the user asks what is pending, in progress, done, outstanding, assigned, high priority, or "on the plate" — including phrasings like "my tasks", "my todos" or "what's left", which refer to the project's task files and not to conversation memory.
allowed-tools: Bash
---

# List Tasks

List tasks using the `taskmd` CLI.

## Instructions

The user's arguments are in `$ARGUMENTS` (e.g. `--status pending`, `--format json`, a directory path).

1. Run `taskmd list $ARGUMENTS`
   - If `$ARGUMENTS` is empty, run: `taskmd list`
   - Common flags: `--status`, `--priority`, `--filter`, `--format`, `--sort`, `--scope`, `--phase`
   - Filter examples: `--status pending`, `--priority high`, `--filter "priority>=medium"`
   - Phase filtering: `taskmd list --phase <phase-id>` or `taskmd list --filter phase=<phase-id>`.
     Phase IDs are per-project — run `taskmd phases` to see this project's before filtering.
     Never guess an ID.
   - Use `taskmd phases` to see all configured phases with progress stats
2. Present the output to the user
   - **If the user asked for a specific format** (`json`, `yaml`, csv, "raw"), reproduce the
     command's output **verbatim in a fenced code block**. Do not re-render it as a table and do
     not summarize it — the format *is* the request. Tool output is not visible to the user, so
     "the JSON is above" is never true.
   - Otherwise, present the table as-is.
