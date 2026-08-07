---
title: "Pin taskmd CLI to the current worktree in do-task and complete-task skills"
id: "01kzdpvr1"
status: pending
priority: high
type: chore
tags: ["plugins", "skills", "worktree"]
created: "2026-08-07"
phase: critical-feedback
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

- [ ] In `do-task` step 3, change `Run taskmd set <ID> --status in-progress` so it
      runs against the current checkout — pin the task dir explicitly, e.g.
      `taskmd -d "$(git rev-parse --show-toplevel)/<dir-from-.taskmd.yaml>" set <ID> --status in-progress`
      (`git rev-parse --show-toplevel` returns the **worktree** root when inside a
      worktree). `-d` removes the cwd dependency.
- [ ] Apply the same pinning to `complete-task` step 4
      (`taskmd set … --status completed --verify` and the `in-review` branch),
      preserving `--verify`.
- [ ] Add an explicit ⚠️ guardrail line to both skills: "Never `cd` to an absolute
      repo path or the 'primary working directory' before running taskmd — in a
      git worktree that is a different checkout on a different branch, and the
      write will corrupt the wrong copy."
- [ ] Handle the non-git case gracefully (fall back to running taskmd from the
      current directory when `git rev-parse --show-toplevel` fails).
- [ ] Mirror any equivalent guidance into the lite variants if they shell out to
      taskmd anywhere; otherwise note they are already immune (they use `Edit`).
- [ ] Update/extend the plugin↔CLI conformance check (`[[01kz3c244]]`) if it
      asserts the skills' documented commands.

## Acceptance Criteria

- Running `do-task` / `complete-task` from within a git worktree writes the
  status change to that worktree's task file, never to the primary/main checkout.
- The `taskmd` CLI (and `--verify`) is still used — no switch to `Edit`-only.
- Both full-plugin skills carry the explicit "do not cd to another checkout"
  guardrail.
- Non-git usage still works.
