---
name: list-tasks
description: List tasks from the project's taskmd files, with optional filters (status, priority, phase, group, owner). Use whenever the user asks what is pending, in progress, done, outstanding, assigned, high priority, or "on the plate" — including phrasings like "my tasks", "my todos" or "what's left", which refer to the project's task files and not to conversation memory.
allowed-tools: Glob, Read
---

# List Tasks

List tasks by scanning task files directly — no CLI required.

## Instructions

The user's arguments are in `$ARGUMENTS` (e.g. `--status pending`, `--priority high`, a directory path).

1. **Find the task directory**:
   - Read `.taskmd.yaml` if it exists to check for a custom `dir` field
   - Default to `tasks` if not configured or file doesn't exist

2. **Scan for task files**: Use `Glob` with pattern `<task-dir>/**/*.md`
   - Exclude files in `.worklogs/` directories
   - Exclude files that don't have YAML frontmatter

3. **Read and parse each task file**:
   - Read each file and extract YAML frontmatter (between first `---` and second `---`)
   - Parse fields: id, title, status, priority, effort, type, tags, group, owner, phase, dependencies, created

4. **Apply filters** from `$ARGUMENTS`:
   - `--status <value>`: Show only tasks matching this status
   - `--priority <value>`: Show only tasks matching this priority
   - `--filter effort=<value>`: Show only tasks matching this effort (supports >=, >, <=, <)
   - `--filter type=<value>`: Show only tasks matching this type
   - `--filter tags=<value>`: Show only tasks containing this tag
   - `--filter owner=<value>`: Show only tasks matching this owner
   - `--phase <value>`: Show only tasks matching this phase
   - `--scope <value>`: Show only tasks in this scope (supports wildcards)
   - A directory path: Only scan that directory instead of the full task dir

   Status and priority are **exact enum values, not descriptions**. `--status pending` means
   `status: pending` only — it does not include `in-progress`. `--priority high` does not
   include `critical`. When the user's wording is broader than the enum ("what's not done",
   "what's urgent"), say which statuses or priorities you included.

5. **Display results** as a formatted table:
   ```
   ID    | Status      | Priority | Title
   ------|-------------|----------|-------------------------------
   001   | pending     | high     | Implement user auth
   002   | in-progress | medium   | Fix login bug
   ```

   If no tasks match the filters, inform the user.

See `SPEC_REFERENCE.md` (in the plugin root) for valid field values and frontmatter schema.
