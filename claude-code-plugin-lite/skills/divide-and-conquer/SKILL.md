---
name: divide-and-conquer
description: Pick up a task and execute it using subagents to parallelize independent workstreams. Use when the user wants to work on a task with maximum concurrency.
allowed-tools: Glob, Read, Write, Edit, Grep, Bash, Task, Agent, EnterPlanMode
---

# Divide and Conquer

Pick up a task and execute it by splitting the work into independent workstreams — no CLI required. Each workstream runs in its **own subagent, in its own git worktree, on its own branch** — subagents never touch the primary repo checkout. Parallelize only workstreams whose scopes don't overlap; serialize the rest.

## Instructions

The user's query is in `$ARGUMENTS` (a task ID like `077` or a task name/keyword).

1. **Find the task file**:
   - Read `.taskmd.yaml` for custom `dir` (default: `tasks`), `workflow` mode, and the `scopes` map
   - Use `Glob` for `<task-dir>/**/*$ARGUMENTS*.md`
   - If multiple matches, read frontmatter to find the exact ID match
   - If not found, list available tasks

2. **Read the task file** with the `Read` tool to get the full description, subtasks, `touches` scopes, and acceptance criteria

3. **Mark the task as in-progress**:
   - Use `Edit` to change the status to `in-progress` in the frontmatter

4. **Start a worklog entry** (if worklogs are enabled):
   - Check `.taskmd.yaml` for `worklogs: true` — only create worklogs if explicitly enabled
   - If enabled, find or create worklog at `<task-dir>/<group>/.worklogs/<ID>.md`
   - Append a timestamped entry noting your approach

5. **Determine the base branch**:
   - Default to the **current local branch**: `git rev-parse --abbrev-ref HEAD`
   - Use a different base **only** if the user explicitly asked for one
   - All subagent branches are created off this base

6. **Plan workstreams and partition by scope**:
   - Use `EnterPlanMode` to design the overall approach
   - In the plan, include a reference to the original task ID and task file path
   - Break the task into candidate **workstreams** — pieces of work that could proceed independently. Examples:
     - Implementation code vs. tests vs. documentation
     - Changes to separate packages or modules
     - Backend changes vs. frontend changes
   - **Assign scopes to each workstream**: list the abstract scopes (from the task's `touches` field and the `scopes` map in `.taskmd.yaml`) plus the concrete file areas each workstream will modify.
   - **Detect overlap**: build a scope→workstream map. Two workstreams that share a scope (or otherwise write the same files) **must not** run in parallel — they would produce merge conflicts.
   - **Group into batches**:
     - Workstreams with **disjoint** scopes go in the same parallel batch.
     - Overlapping workstreams are **serialized**: pick an order, and have the later one branch off the **earlier one's completed branch** (not the base) so it builds on that work conflict-free.
   - If the task is simple enough that a single workstream covers it, just do it directly (skip to step 9)

7. **Launch subagents — one worktree + branch each**:
   - Use the `Agent` tool with `isolation: "worktree"` for **every** subagent that modifies files, so each gets an isolated worktree
   - Launch all subagents in a parallel batch in a **single message** so they run concurrently; run serialized (overlapping) workstreams in later messages, after their predecessor finishes
   - Give each subagent a self-contained prompt that includes:
     - Exactly what to do, with relevant file paths and context
     - Its **assigned branch name** (e.g. `dnc/<task-id>/<workstream-slug>`) and its **base branch** (the task base, or a predecessor's branch for serialized work)
     - These non-negotiable rules for the subagent:
       - "**Work only inside your own worktree** (your cwd). Do **not** read-modify-write, `cd` into, or otherwise touch the primary repo checkout at any other path."
       - "First run `git checkout -b <assigned-branch> <base-branch>` so all your commits land on your branch."
       - "When done, **commit your work** to your branch with a clear message, then **run all verification steps** relevant to your changes (build, tests, lint) and fix anything that fails before committing the fix."
       - "Report back: your branch name, final commit SHA, a summary of what changed, and the verification results (pass/fail with details)."

8. **Wait, then collect results**:
   - Wait for **all** subagents in the batch to complete before moving on
   - Review each subagent's reported branch, commits, and verification status for correctness
   - If a subagent failed or its verification did not pass, handle it directly (inspect its branch, fix, re-verify) rather than blindly re-launching
   - Check off subtasks (`- [x]`) in the task file using `Edit` as they are completed
   - Append worklog entries for key decisions, blockers, and completed workstreams

9. **Write a final worklog entry** summarizing what was done, which workstreams ran in parallel vs. serialized, the branches produced, decisions made, and any open items

10. **Ask the user how to finish — do not auto-commit to the base branch**:
    - Present a short summary: each workstream, its branch, commit SHA, and verification status
    - **Ask the user** whether they want you to integrate the work now (merge the workstream branches into the base branch and mark the task done) **or** leave the branches in place for them to review and commit themselves
    - Only if the user asks you to integrate:
      - Merge each workstream branch into the base branch, resolving any conflicts, and run the full verification suite on the integrated result
      - Then mark the task done per `.taskmd.yaml` `workflow`:
        - **Solo mode** (default): if the task has `verify` checks, run them (bash via `Bash`, assert via code inspection); if all pass, use `Edit` to set `status: completed`; if any fail, fix and retry
        - **PR-review mode**: open a PR, then use `Edit` to set `status: in-review` and add the PR URL to the `pr` array; stop
    - If the user prefers to handle it themselves, leave the branches untouched and stop — do not merge or change task status

## Worklog Format

Each worklog entry uses a timestamp heading followed by free-form notes:

```markdown
## 2026-02-15T10:30:00Z

Started divide-and-conquer execution of the search feature task.
Base branch: `main`.

**Workstreams (scope-partitioned):**

Parallel batch (disjoint scopes):
1. Core search implementation — scope `cli/search` — branch `dnc/077/core` (worktree)
2. Documentation updates — scope `docs` — branch `dnc/077/docs` (worktree)

Serialized (shares `cli/search` with #1):
3. Search test suite — branch `dnc/077/tests`, based off `dnc/077/core`

**Results:**

- [x] All subagents committed to their branches; verification passed on each
- Branches ready for integration; awaiting user decision on merge

**Decisions:** Used full-text search with SQLite rather than Elasticsearch.
```

See `SPEC_REFERENCE.md` (in the plugin root) for frontmatter schema, workflow modes, and verify checks.
