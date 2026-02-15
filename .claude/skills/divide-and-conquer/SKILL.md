---
name: divide-and-conquer
description: Split a large task into smaller sub-tasks. Accepts a task ID, evaluates complexity, and creates sibling task files if warranted.
allowed-tools: Bash, Read, Glob, Write
---

# Divide and Conquer

Evaluate a task's complexity and, if warranted, split it into smaller, focused sub-tasks.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077`, optionally followed by `--force` to skip the complexity check).

1. **Look up the task**: Run `taskmd get $ARGUMENTS` to find the task
   - If not found, run `taskmd list` to show available tasks and ask the user which one they meant
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
   a. **Read the specification** at `docs/taskmd_specification.md` for the correct format
   b. **Determine available IDs** by scanning `tasks/**/*.md` with `Glob`:
      - Extract numeric IDs from filenames (pattern: `NNN-description.md`)
      - Allocate the next N sequential IDs for the sub-tasks
   c. **Determine `touches` scopes**:
      - Read `.taskmd.yaml` to see the existing `scopes` definitions
      - Assign matching scope identifiers from the existing scopes list
   d. **Design the split**: Group the original task's work into 2-5 focused sub-tasks where each:
      - Has a single clear responsibility
      - Can be independently verified
      - Includes relevant subtasks and acceptance criteria from the original
   e. **Create sub-task files** as siblings of the original task file (same directory), each with:

      ```yaml
      ---
      id: "<NNN>"
      title: "<focused title>"
      status: pending
      priority: <inherit from parent>
      effort: <estimated for this slice>
      tags: <inherit relevant tags from parent>
      parent: "<original task ID>"
      touches: []  # scope identifiers from .taskmd.yaml (omit if not applicable)
      created: <today's date YYYY-MM-DD>
      ---
      ```

      Followed by a markdown body with:
      - An H1 heading matching the title
      - An `## Objective` section describing this slice's goal
      - A `## Tasks` section with checkbox items (pulled or refined from the original)
      - An `## Acceptance Criteria` section

   f. **Update the original task** (only the markdown body, not the frontmatter status):
      - Add a `## Sub-tasks` section listing the created sub-task IDs and titles
      - Keep the original content intact for reference

6. **Report** the result:
   - List each created sub-task file with its ID and title
   - Summarize how the work was divided
