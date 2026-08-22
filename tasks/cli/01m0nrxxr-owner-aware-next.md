---
title: "Owner-aware next: respect owner on in-progress tasks and add --for flag"
id: "01m0nrxxr"
status: pending
priority: high
type: feature
tags: ["next", "sdk", "coordination"]
created: "2026-08-22"
effort: medium
phase: worktree-support
---

# Owner-aware next: respect owner on in-progress tasks and add --for flag

## Objective

Make the `next` recommender respect the existing `owner` frontmatter field so multiple
agents coordinate even within a single checkout: an in-progress task owned by someone
else must not be recommended to you. Independently shippable — no worktree machinery
involved. Spec: `docs/specs/worktree-support.md` §6.

## Tasks

- [ ] `sdk/go/next`: a task that is `in-progress` with a non-empty `owner` is actionable only for that owner
- [ ] Add `taskmd next --for <owner>`: tasks owned by others are excluded; tasks owned by `<owner>` sort first (resume-your-own-work)
- [ ] Default behavior (no `--for`): owned in-progress tasks are excluded; unowned tasks behave exactly as today
- [ ] `--explain` output names the owning user when a task is excluded by ownership
- [ ] Unit tests in `sdk/go/next` for the ownership rules; CLI tests for `--for`; e2e coverage
- [ ] Document the claim convention (`taskmd set <id> --status in-progress --owner <name>` as one call) in the next/set help text

## Acceptance Criteria

- Two agents running `next --for a` / `next --for b` against the same task dir are never recommended the same claimed task
- A task with `status: in-progress` and `owner: alice` is not returned by `taskmd next` without `--for alice`
- Unowned in-progress tasks keep today's resume semantics — no behavior change without owners in play
- SDK change is additive only (patch bump per the sdk-pin workflow; no `sdk-bump` marker needed)
