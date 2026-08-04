---
name: add-task
description: Create a new task file following the taskmd specification. Use when the user wants to add a new task to the project.
allowed-tools: Glob, Read, Write
---

# Add Task

Create a new task file — no CLI required.

## Instructions

The user's task description is in `$ARGUMENTS`.

1. **Parse the user's input** from `$ARGUMENTS` to extract:
   - The task **title** (required)
   - Any optional metadata: priority, effort, type, tags, group, dependencies, parent, owner, phase

2. **Read configuration**:
   - Read `.taskmd.yaml` if it exists for: task `dir` (default: `tasks`), `id` config (strategy, prefix, padding, length), and `phases`
   - If the user mentions a phase/milestone/sprint, check `.taskmd.yaml` for configured `phases` and use the matching `id`. If the phase doesn't exist yet, add it to the `phases` array in `.taskmd.yaml` (see `SPEC_REFERENCE.md` for the phases config format)

3. **Determine the group** based on the task's domain:
   - If the user specified `--group`, use that
   - Otherwise infer from context (e.g., CLI/backend → `cli`, web/frontend → `web`, or root for cross-cutting)

4. **Generate the task ID**:
   - Read the ID strategy from `.taskmd.yaml` (default: `sequential`)
   - Scan existing files with `Glob` for `<task-dir>/**/*.md` to determine used IDs
   - **Sequential** (default): Find the highest numeric ID, add 1, zero-pad to `padding` width (default 3). E.g., if highest is `042`, next is `043`
   - **Prefixed**: Find highest number with the configured prefix. E.g., `dr-001`, `dr-002`
   - **Random**: Generate a random alphanumeric string (lowercase letters and digits, i.e. base36) of configured `length` (default 6)
   - **ULID**: Generate a ULID-like ID — use current timestamp in Crockford Base32 + random chars

5. **Create the task file** using `Write`:
   - Path: `<task-dir>/<group>/<ID>-<slug-title>.md` (or `<task-dir>/<ID>-<slug-title>.md` if no group)
   - Slug: lowercase, hyphenated version of the title (max ~50 chars)
   - Content:

   ```markdown
   ---
   id: "<ID>"
   title: "<title>"
   status: pending
   priority: <priority if provided>
   effort: <effort if provided>
   type: <type if provided>
   tags: [<tags if provided>]
   dependencies: [<deps if provided>]
   parent: "<parent if provided>"
   owner: "<owner if provided>"
   phase: "<phase if provided>"
   created: <today's date YYYY-MM-DD>
   ---

   # <Title>

   ## Objective

   <Description derived from user's input>

   ## Tasks

   - [ ] <Subtask 1>
   - [ ] <Subtask 2>

   ## Acceptance Criteria

   - <Criterion derived from the task>
   ```

   Only include optional frontmatter fields that were specified or can be inferred. Don't include empty fields.

6. **Confirm** the created file path and ID to the user

See `SPEC_REFERENCE.md` (in the plugin root) for valid field values, ID strategies, and frontmatter schema.
