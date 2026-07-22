---
name: safe-queue
description: Assess whether a task can safely start in parallel, then optionally create a dependency-aware Worktrunk worktree after approval. Use when scheduling work alongside active tasks or stacking work on an unmerged dependency.
allowed-tools: Bash, Read, AskUserQuestion
---

# Safe Queue

Assess `$ARGUMENTS` using task dependencies, declared `touches`, Taskmd tracks,
Git worktrees, branch ancestry, cleanliness, commits, and actual changed files.
A separate worktree is never evidence that overlapping mutable work is safe.
When tasks declare an optional custom `resources` list (for example
`["port:3000", "database:test"]`), shared operational resources also block
parallel execution. Use `resources: []` to explicitly assert that a task has no
exclusive operational resources. Automatic execution remains disabled when
this evidence is absent.

## Decisions

- **`START_FROM_MAIN`** — the task is pending; its scopes are specific and
  resolvable; every dependency is present on the selected main commit; and no
  declared or actual mutable overlap exists.
- **`START_STACKED`** — exactly one unmerged dependency is available on a
  clean, committed branch whose HEAD is both explicitly frozen and recorded as
  validated. This decision is emitted only after explicit stacking
  authorization.
- **`WAIT`** — evidence is reliable, but work must finish or merge first:
  mutable overlap, unmet/unmerged dependencies, a dirty/unfrozen/unvalidated
  stack base, or missing stacking authorization.
- **`UNSAFE`** — reliable scheduling is impossible: task state is wrong,
  `touches` is missing/broad/stale/ambiguous, dependencies are missing, task to
  worktree mapping is ambiguous, or required evidence cannot be inspected.

`WAIT` and `UNSAFE` never create a branch/worktree or change task ownership or
status.

## Default: plan only

1. Parse the task query and optional flags from `$ARGUMENTS`.
2. Run:

   ```bash
   python3 "${CLAUDE_PLUGIN_ROOT}/skills/safe-queue/scripts/safe_queue.py" \
     --project-root . --pretty <task>
   ```

3. For a requested stacked plan, rerun only after the user explicitly
   authorizes stacking:

   ```bash
   python3 "${CLAUDE_PLUGIN_ROOT}/skills/safe-queue/scripts/safe_queue.py" \
     --project-root . --allow-stacked --pretty <task>
   ```

4. Present the single decision and all returned evidence: exact base
   branch/commit, proposed `task-<ID>` worktree, dependency state, Taskmd
   track, declared and actual overlaps, merge order, post-merge validation,
   and the reason for waiting/refusal.
5. Do not mutate anything in plan mode.

The assessment is conservative. Missing scopes, unknown scope paths, wildcard
or repository-wide mappings, active tasks with stale declarations, and actual
changed files outside optimistic assumptions are conflicts, not permission to
start.

## Freezing a stack base

Freezing is an explicit promise that no more commits or working-tree changes
will be added while stacked work uses that exact HEAD. After validating the
base task, record both facts on its branch:

```bash
git config branch.<base-branch>.taskmd-frozen true
git config branch.<base-branch>.taskmd-validated-commit \
  "$(git rev-parse <base-branch>)"
git config branch.<base-branch>.taskmd-task-id <base-task-id>
git config taskmd.task.<base-task-id>.validated-commit \
  "$(git rev-parse <base-branch>)"
```

Never set these values on the user's behalf without explicit approval and
successful validation. If HEAD changes, validation and freezing are stale.

## Explicit execution

Ask for confirmation after showing the plan. Then run:

```bash
python3 "${CLAUDE_PLUGIN_ROOT}/skills/safe-queue/scripts/execute_safe_queue.py" \
  --project-root . --approved \
  --expected-decision <DECISION> \
  --expected-base-branch <BASE_BRANCH> \
  --expected-base-commit <BASE_COMMIT> \
  --expected-approval-token <APPROVAL_TOKEN> \
  <task>
```

Copy all four expected values verbatim from the plan the user approved.
Execution refuses if a fresh assessment produces a different decision, base
branch, base commit, task, worktree name, merge order, or validation
implications. Add `--allow-stacked` only when the user approved the exact
stacked base branch and commit shown in the plan. Add `--owner <name>` when
requested.

Execution reassesses from scratch, creates the Worktrunk worktree first, copies
only the requested task file when it is ignored/untracked, and only then runs
`taskmd set <ID> --status in-progress` in the new worktree. Failed worktree
creation leaves task status and ownership unchanged.

`--automatic` may be used only when the user explicitly selects automatic
execution. The script rejects automatic execution unless the fresh decision is
an unambiguous `START_FROM_MAIN`. Stacked, dirty, ambiguous, overlapping, or
operationally conflicting cases require confirmation.

For stacked work, preserve dependency-first merge order. After the base branch
merges, rebase the stacked branch onto updated main and rerun every validation
required by the stacked task. Preserve proof that the dependency reached main:
keep its task-to-branch Git config until the mapped branch HEAD is an ancestor
of main, or record an exact merged commit after branch cleanup:

```bash
git config taskmd.task.<base-task-id>.merged-commit \
  "$(git config --get taskmd.task.<base-task-id>.validated-commit)"
```

The merged commit must equal the previously recorded validated commit and be an
ancestor of main. A `completed` task file alone is not merge evidence. This
conservative rule also works when task files are ignored by Git.

## Deferred work

This skill is not a daemon and cannot wake itself. A waiting task can be
reassessed by polling from an active agent goal or by an external scheduler
invoking the skill again.
