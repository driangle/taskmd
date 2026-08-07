---
title: "Resolve taskmd task directory from the git worktree root"
id: "01kzdpvr2"
status: cancelled
priority: medium
type: improvement
tags: ["cli", "worktree", "correctness"]
created: "2026-08-07"
phase: critical-feedback
cancelled_at: 2026-08-07
---

# Resolve taskmd task directory from the git worktree root

> **Cancelled (2026-08-07).** Decision: `taskmd` should **always resolve from the
> current working directory** — that predictable, cwd-anchored contract is a feature,
> not a bug. This task proposed overriding cwd with the git worktree root, which
> directly contradicts that contract. It was also found to be redundant and risky:
> `initConfig` *already* walks up from cwd and stops at the `.git` boundary, so it
> already anchors to the worktree root without shelling out to `git`; the proposed
> auto-anchor buys nothing over that for the motivating bug (see caveat below) while
> risking regressions to nested/monorepo `.taskmd.yaml`, `--project`, and
> `default_project` (all intentional "operate outside cwd" paths), plus a brand-new
> runtime dependency on the `git` binary. The `cd`-away failure it targeted is handled
> at the skill layer by the guardrail in `[[01kzdpvr1]]` (never `cd` before a taskmd
> write). If worktree-safety ever needs a CLI-level cure, do it via an explicit
> harness-provided signal, not by second-guessing cwd.

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

> ⚠️ **Caveat — this does NOT fix the `cd`-away failure that motivated `[[01kzdpvr1]]`.**
> `git rev-parse --show-toplevel` (and every other signal here — cwd walk-up, the
> project registry) is evaluated from **taskmd's own process cwd**. The observed bug
> is `cd /path/to/primary-repo && taskmd set …`: by the time taskmd runs, its cwd is
> the *primary* checkout, so `--show-toplevel` resolves to the primary root and the
> write still lands there. The safety check below has the same blind spot — the write
> to `primary/tasks/` is *inside* the primary worktree, so it is not "outside the
> current worktree" and is allowed. Once a process has `cd`'d into another checkout,
> git cannot recover which worktree the caller *meant*; that information is gone.
> Net: this task hardens resolution for the *milder* cases (cwd in a subdir of the
> worktree, or an explicit `-d` pointing at a foreign tree) but the `cd`-away case
> is only cured by the behavioral guardrail in `[[01kzdpvr1]]` ("never `cd` away
> before running taskmd"). The **only** way to make the CLI itself survive a rogue
> `cd` is to key resolution off an external worktree signal the harness provides
> (e.g. an env var set by worktree isolation) rather than cwd-relative git — a
> larger design, out of scope unless such a signal exists. Do not implement the
> auto-anchor expecting it to close the reported bug.

Design considerations:
- `git rev-parse --show-toplevel` returns the current **worktree** root (not the
  primary checkout) **as seen from taskmd's cwd** — correct only when the caller has
  not `cd`'d out of the worktree first (see caveat above).
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
