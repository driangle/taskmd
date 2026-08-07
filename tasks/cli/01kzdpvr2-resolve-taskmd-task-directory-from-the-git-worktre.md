---
title: "Resolve taskmd task directory from the git worktree root"
id: "01kzdpvr2"
status: pending
priority: medium
type: improvement
tags: ["cli", "worktree", "correctness"]
created: "2026-08-07"
phase: critical-feedback
---

# Resolve taskmd task directory from the git worktree root

## Description

`taskmd` resolves its task directory from the current working directory (local
`.taskmd.yaml` → `dir`, or `-d/--task-dir`, default `.`). Inside a **git
worktree**, if the caller's cwd is a *different* checkout (e.g. an agent that
`cd`s to the primary repo before running `taskmd set`), taskmd writes to that
other checkout's `tasks/` — silently mutating the wrong branch's task files. See
the skill-level mitigation in `[[01kzdpvr1]]`; this task is the durable,
CLI-level cure so no caller discipline is required.

Goal: make `taskmd` **worktree-aware**. When invoked inside a git working tree,
anchor task-directory resolution to the worktree root (`git rev-parse
--show-toplevel`) rather than relying solely on cwd walk-up / the global project
registry — or, at minimum, refuse to write to a task directory that lies outside
the current worktree.

Design considerations:
- `git rev-parse --show-toplevel` returns the current **worktree** root (not the
  primary checkout), which is exactly the desired anchor.
- Precedence must stay predictable and backward-compatible: an explicit
  `-d/--task-dir`, `--config`, or `--project` should still win over auto-anchoring.
- Auto-anchor should apply to the implicit/default case (no explicit dir/project),
  where a local `.taskmd.yaml` at the worktree root exists.
- Preserve non-git behavior (current cwd walk-up) when not in a git repo.
- Decide the guardrail for the write path: either (a) always resolve relative to
  the worktree root, or (b) keep current resolution but hard-error if a write
  would target a task dir outside `git rev-parse --show-toplevel`.

## Tasks

- [ ] Add a resolution step: when in a git working tree and no explicit
      `-d/--config/--project` was given, resolve the task dir relative to
      `git rev-parse --show-toplevel` (using that worktree's `.taskmd.yaml`).
- [ ] Keep explicit `-d/--task-dir`, `--config`, and `--project` as higher-precedence
      overrides; document the precedence order.
- [ ] Add a safety check on the write path: refuse (clear error) to write a task
      file outside the current git worktree unless explicitly overridden.
- [ ] Preserve existing behavior outside git repos (cwd-based resolution).
- [ ] Tests: cover (a) worktree + cwd elsewhere resolves to the worktree's tasks,
      (b) explicit `-d`/`--project` still overrides, (c) non-git fallback,
      (d) write-outside-worktree is refused.
- [ ] Document the worktree-aware resolution in the README / relevant ADR
      (this touches the core-scope boundary — see docs/adr/0001).

## Acceptance Criteria

- From a git worktree, `taskmd set <ID> --status …` (with no explicit dir/project)
  writes to that worktree's task file even if the process cwd is a different
  checkout.
- Explicit `-d/--task-dir`, `--config`, and `--project` overrides are unchanged.
- A write that would land outside the current worktree is refused with a clear
  error (unless explicitly overridden).
- Non-git usage is unchanged.
- Once shipped, the skill guardrail in `[[01kzdpvr1]]` becomes belt-and-suspenders
  rather than load-bearing.
