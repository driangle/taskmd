---
title: "Add safe-queue skill for dependency-aware worktree scheduling"
id: "01ky5f883"
status: completed
priority: medium
effort: medium
type: feature
tags: ["skills", "worktrees", "safety", "automation"]
touches:
  - "plugin/skills"
  - "plugin/docs"
  - "skill/benchmarks"
created: "2026-07-22"
completed_at: 2026-07-22
---

# Add safe-queue skill for dependency-aware worktree scheduling

## Objective

Add a `$taskmd:safe-queue` plugin skill that evaluates whether a task can
safely start alongside current work, proposes the correct branch/worktree
strategy, and executes that proposal only after explicit approval. The skill
must combine Taskmd dependency and `touches` metadata with actual Git worktree
and changed-file state rather than assuming that separate worktrees eliminate
overlap.

The default invocation is plan-only. It reports one decision:
`START_FROM_MAIN`, `START_STACKED`, `WAIT`, or `UNSAFE`. An explicit execution
mode may create and claim work only after every applicable safety gate passes.
The skill is orchestration guidance, not a background daemon; continued
waiting requires an active agent goal or an external scheduler to invoke it
again.

## Tasks

- [x] Define the four queue decisions and the exhaustive evidence required for
      each decision.
- [x] Inspect the requested task, dependencies, all in-progress tasks,
      `touches`, `taskmd tracks`, Git worktrees, branch ancestry, working-tree
      cleanliness, commits, and actual changed files.
- [x] Treat missing, broad, stale, or ambiguous `touches` as potentially
      overlapping instead of optimistically starting work.
- [x] Add a deterministic, non-mutating assessment script under
      `claude-code-plugin/skills/safe-queue/scripts/` that emits a
      machine-readable decision and supporting evidence.
- [x] Make plan-only assessment the default and show the proposed base
      branch/commit, worktree name, detected overlaps, dependency state, merge
      order, validation implications, and reason for waiting or refusal.
- [x] Add an explicit execution mode that creates a Worktrunk worktree and
      marks the task in progress only after worktree creation succeeds.
- [x] Permit `START_FROM_MAIN` only for an independent task whose required
      base is present in the selected main branch and whose scopes do not
      overlap mutable work.
- [x] Permit `START_STACKED` only with explicit user authorization after the
      base task is committed, validated, clean, and explicitly frozen. Record
      the exact base branch and commit and require the base task to merge
      first.
- [x] Require stacked work to rebase onto updated main and rerun its required
      validation after the base branch merges.
- [x] Leave waiting or unsafe tasks pending and avoid creating worktrees,
      assigning owners, or changing status.
- [x] Ensure automatic mode, if provided, is limited to unambiguous
      `START_FROM_MAIN` cases. Stacked, dirty, ambiguous, or operationally
      resource-conflicting cases must require confirmation.
- [x] Preserve existing task files that are ignored by Git when creating a
      worktree, without copying unrelated runtime state.
- [x] Add fixtures or skill benchmarks covering independent tasks, declared
      and actual file overlap, dirty dependencies, frozen stacked branches,
      completed-but-unmerged dependencies, merged dependencies, ambiguous
      scopes, and failed worktree creation.
- [x] Update the Taskmd plugin skill documentation and validate the completed
      skill using repository conventions.

## Acceptance Criteria

- `$taskmd:safe-queue <task>` produces a plan without mutating task or Git
  state by default.
- Every assessment reports exactly one of `START_FROM_MAIN`,
  `START_STACKED`, `WAIT`, or `UNSAFE`, with enough evidence for a user to
  validate the decision.
- The skill checks both declared scopes and actual Git changes across active
  worktrees.
- A separate worktree never causes overlapping mutable work to be classified
  as safe by itself.
- Execution requires explicit approval, except for an explicitly selected
  automatic mode restricted to unambiguous independent work.
- Task ownership and `in-progress` status are applied only after successful
  worktree creation.
- Stacked work records an exact base commit, preserves dependency-first merge
  order, and requires post-rebase validation.
- Waiting and unsafe assessments make no worktree, branch, ownership, or task
  status changes.
- The skill clearly explains that it cannot wake itself and identifies active
  goal polling or external scheduling as the mechanisms for deferred starts.
- Tests or skill benchmarks cover all four decisions and the principal race,
  ambiguity, and failure cases.
