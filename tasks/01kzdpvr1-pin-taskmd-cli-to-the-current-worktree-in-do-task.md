---
title: "Pin taskmd CLI to the current worktree in do-task and complete-task skills"
id: "01kzdpvr1"
status: completed
priority: high
type: chore
tags: ["plugins", "skills", "worktree"]
created: "2026-08-07"
phase: critical-feedback
completed_at: 2026-08-07
---

# Pin taskmd CLI to the current worktree in do-task and complete-task skills

## Description

When `do-task` / `complete-task` run inside a **git worktree** (e.g. subagents
each working a task in their own worktree), their `taskmd set … --status …`
shell commands can write the status change to the **wrong checkout**.

Observed failure: an agent isolated in `.../.claude/worktrees/agent-XXX` ran
`cd /path/to/primary-repo && taskmd set <ID> --status in-progress`. `taskmd set`
resolves its task directory from the current working directory (local
`.taskmd.yaml` → `dir: ./tasks`), so the status flip landed on the **primary/main
checkout** (a different branch), showing up as an uncommitted `status:` edit on
`main` instead of in the worktree. The real code edits were unaffected because
they go through the `Edit`/`Write` tools, which the harness's worktree isolation
guards — but a plain shell `taskmd` write bypasses that guard entirely.

Two compounding causes:
1. Agents habitually prepend `cd <primary-working-dir>` to shell commands; in a
   worktree that primary dir is a different checkout on a different branch.
2. `taskmd set` is a shell write, not covered by the Edit/Write worktree
   isolation guard.

This task hardens the **full (CLI) plugin** skills so the CLI path stays correct
in worktrees, while keeping the `taskmd` CLI (and features like
`complete-task --verify`). The durable, CLI-level fix is tracked separately in
`[[01kzdpvr2]]`; this skill guardrail is the immediate mitigation and remains
useful as defense-in-depth.

Scope: `claude-code-plugin/skills/{do-task,complete-task}/SKILL.md`. Keep the
`claude-code-plugin-lite` variants (which already use `Edit`, and are immune) and
the prose/CLI conformance check (`[[01kz3c244]]`) in sync.

## Tasks

> **Implementation note (2026-08-07):** the originally-proposed `-d "$(git rev-parse
> --show-toplevel)/…"` pinning was **dropped** as it does not do what it claims. The
> subshell inherits the taskmd process's cwd, so after a rogue `cd` into the primary
> checkout it resolves to the *primary* root — the exact wrong tree — giving false
> confidence; it also hardcodes `tasks`, breaking projects whose configured `dir`
> differs. taskmd's resolution is entirely cwd-anchored (see `resolveTaskDir` in
> `apps/cli/internal/cli/root.go`), so the only reliable fix at the skill layer is the
> behavioral guardrail: **run taskmd in place, never `cd` away first.** The durable
> CLI cure is tracked in `[[01kzdpvr2]]` (which carries the same caveat).

- [x] ~~Pin `do-task` step 3 via `-d "$(git rev-parse --show-toplevel)/…"`~~ →
      Instead: annotated step 3 to run `taskmd set <ID> --status in-progress` **from the
      current working directory** and added the ⚠️ guardrail (see note above for why the
      `-d` subshell was rejected).
- [x] ~~Apply the same pinning to `complete-task` step 4~~ → Instead: annotated step 4
      (both the `--status completed --verify` and `in-review` branches) to run in place;
      `--verify` preserved.
- [x] Add an explicit ⚠️ guardrail block to both full-plugin skills: never `cd` to an
      absolute repo path or the 'primary working directory' before a `taskmd` write — in
      a git worktree that is a different checkout on a different branch, and a shell
      `taskmd` write (unlike `Edit`/`Write`) is not worktree-isolated, so it corrupts the
      wrong copy.
- [x] Handle the non-git case gracefully — the guardrail notes running in place is
      already correct outside a git repo (no `git rev-parse` invoked at all).
- [x] Lite variants need no change: they mutate frontmatter via `Edit` (lite
      `complete-task` doesn't even allow `Bash`), so they're immune. Documented as such.
- [x] Conformance check (`[[01kz3c244]]` / `lite_conformance_test.go`) does **not** assert
      these skills' command strings (it guards ID/slug/frontmatter/enum prose) — no change
      required.

## Acceptance Criteria

- Running `do-task` / `complete-task` from within a git worktree writes the
  status change to that worktree's task file, never to the primary/main checkout.
- The `taskmd` CLI (and `--verify`) is still used — no switch to `Edit`-only.
- Both full-plugin skills carry the explicit "do not cd to another checkout"
  guardrail.
- Non-git usage still works.
