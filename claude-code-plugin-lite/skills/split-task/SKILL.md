---
name: split-task
description: Split a large task into smaller sub-tasks. Accepts a task ID, evaluates complexity, and creates sibling task files if warranted.
allowed-tools: Glob, Read, Write
---

# Split Task

Evaluate a task's complexity and, if warranted, split it into smaller, focused sub-tasks — no CLI required.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077`, optionally followed by `--force` to skip the complexity check).

1. **Find the task file**:
   - Read `.taskmd.yaml` for custom `dir` (default: `tasks`) and `id` config (strategy, padding, prefix, length)
   - Use `Glob` for `<task-dir>/**/*$ARGUMENTS*.md`
   - Read frontmatter to confirm the ID matches
   - If not found, list available tasks

2. **Read the task file** with the `Read` tool to get the full description, subtasks, and acceptance criteria

3. **Assess complexity** to decide whether the task should be divided. Consider:
   - **Effort field**: `large` effort tasks are good candidates; `small` tasks almost never need splitting
   - **Subtask count**: Tasks with 5+ checkbox items that span distinct concerns are candidates
   - **Scope breadth**: Tasks that touch multiple unrelated areas (e.g., backend + frontend + docs) are candidates
   - **Independence**: Subtasks that can be worked on in parallel by different people are candidates
   - A task is **NOT** a good candidate if:
     - It has `small` or `medium` effort with fewer than 5 subtasks
     - Its subtasks are tightly coupled sequential steps of a single feature
     - Splitting would create trivial tasks that aren't worth tracking individually

4. **If the task is NOT complex enough**:
   - Explain why the task doesn't warrant splitting (be specific about which criteria it fails)
   - Do NOT create any files
   - Only proceed if `$ARGUMENTS` contains `--force` or the user explicitly insists

5. **If the task IS complex enough** (or `--force` is used):

   a. **Determine available IDs** by scanning `<task-dir>/**/*.md` with `Glob`:
      - Extract IDs from filenames and frontmatter
      - Based on the ID strategy from `.taskmd.yaml`:
        - **Sequential** (default): Find highest numeric ID, allocate next N sequential IDs (zero-padded)
        - **Prefixed**: Find highest number with prefix, allocate next N
        - **Random**: Generate N random alphanumeric IDs (length from config, default 6)
        - **ULID**: Generate N ULID-style IDs

   b. **Design the split**: Group the original task's work into 2-5 focused sub-tasks where each:
      - Has a single clear responsibility
      - Can be independently verified
      - Includes relevant subtasks and acceptance criteria from the original

   c. **Create sub-task files** as siblings of the original task file (same directory), using `Write`:
      ```yaml
      ---
      id: "<new-ID>"
      title: "<focused title>"
      status: pending
      priority: <inherit from parent>
      effort: <estimated for this slice>
      tags: <inherit relevant tags>
      phase: <inherit from parent if set>
      parent: "<original task ID>"
      created_at: <today's date YYYY-MM-DD>
      ---
      ```

      Followed by a markdown body with:
      - An H1 heading matching the title
      - An `## Objective` section describing this slice's goal
      - A `## Tasks` section with checkbox items
      - An `## Acceptance Criteria` section

   d. **Update the original task** using `Write` (append to the body):
      - Add a `## Sub-tasks` section listing the created sub-task IDs and titles
      - Keep the original content intact for reference

6. **Report** the result:
   - List each created sub-task file with its ID and title
   - Summarize how the work was divided

See `SPEC_REFERENCE.md` (in the plugin root) for ID strategies, frontmatter schema, and file naming.
