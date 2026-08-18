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
   - Otherwise infer it from the task's domain:
     - `cli` — CLI commands, terminal features, backend/server code
     - `web` — web frontend, UI, components
     - no group — cross-cutting, infrastructure, documentation, or genuinely unclear domain
   - `Glob` for `<task-dir>/*/` to see which groups the project already uses, and prefer an
     existing name over a synonym (`web` over `frontend`). A project with no group directories
     yet is not a reason to skip grouping — it just means yours is the first.
   - The group is a **subdirectory**, so it changes where you write the file:
     `tasks/cli/007-fix-search-crash.md`, not `tasks/007-fix-search-crash.md`. Create the
     directory if it doesn't exist, and don't also add a `group:` frontmatter field — the
     directory name is what defines the group.

4. **Pick the task body shape**. Different kinds of task need different sections; a bug filed
   with a generic "Objective" body loses the information that makes it actionable.

   - Look for project templates with `Glob` on `.taskmd/templates/*.md`. If one matches the
     kind of task being filed (a bug report → `bug.md`, a feature → `feature.md`, routine work
     → `chore.md`), `Read` it and use it as the basis for the new file:
     - Drop the `_template:` block from the frontmatter — it describes the template, not the task
     - Substitute `{{title}}`, `{{id}}`, and `{{date}}`
     - Keep the template's field defaults (e.g. `type: bug`, `priority: high`) unless the user
       specified otherwise — explicit user values always win
     - Keep the template's section headings, and **fill every one of them in**
   - If no template matches, choose the sections yourself:
     - **Bug**: `## Steps to Reproduce` (numbered, concrete), `## Expected Behavior`,
       `## Actual Behavior`, `## Environment`
     - **Anything else**: `## Objective`, `## Tasks`, `## Acceptance Criteria`

   Never leave placeholder content behind — no `<!-- ... -->` comments, no bare `TODO`, no
   `1. ...`. Infer concrete content from the user's description; if a detail is genuinely
   unknown (an exact version number, say), write what is known rather than a stub.

5. **Generate the task ID**:
   - Read the ID strategy from `.taskmd.yaml` (default: `sequential`)
   - Scan existing files with `Glob` for `<task-dir>/**/*.md` to determine used IDs
   - **Sequential** (default): Find the highest numeric ID, add 1, zero-pad to `padding` width (default 3). E.g., if highest is `042`, next is `043`
   - **Prefixed**: Find highest number with the configured prefix. E.g., `dr-001`, `dr-002`
   - **Random**: Generate a random alphanumeric string (lowercase letters and digits, i.e. base36) of configured `length` (default 6)
   - **ULID**: Generate a ULID-like ID — use current timestamp in Crockford Base32 + random chars

6. **Create the task file** using `Write`:
   - Path: `<task-dir>/<group>/<ID>-<slug-title>.md` — use the group decided in step 3, creating
     the directory if needed. Write to `<task-dir>/<ID>-<slug-title>.md` only when that decision
     was "no group"
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
   created_at: <today's date YYYY-MM-DD>
   ---

   # <Title>

   <the sections chosen in step 4, each filled in>
   ```

   The default body, when no template applies and the task isn't a bug report:

   ```markdown
   ## Objective

   <Description derived from user's input>

   ## Tasks

   - [ ] <Subtask 1>
   - [ ] <Subtask 2>

   ## Acceptance Criteria

   - <Criterion derived from the task>
   ```

   Only include optional frontmatter fields that were specified or can be inferred. Don't include empty fields.

7. **Confirm** the created file path and ID to the user

See `SPEC_REFERENCE.md` (in the plugin root) for valid field values, ID strategies, and frontmatter schema.
