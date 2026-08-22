---
title: "Add gitmeta package: repo identity and worktree discovery"
id: "01m0n8cpm"
status: completed
priority: critical
type: feature
tags: ["git", "worktrees"]
created: "2026-08-22"
effort: medium
phase: worktree-support
completed_at: 2026-08-22
---

# Add gitmeta package: repo identity and worktree discovery

## Objective

Create `apps/cli/internal/gitmeta`, the single helper package through which core learns about
the git repository containing the task files (per ADR 0004). It provides repo identity
(the git common dir) and worktree discovery, and is the foundation every other
worktree-support task builds on. Spec: `docs/specs/worktree-support.md` §1–2.

## Tasks

- [x] Add `Identity{CommonDir, WorktreeRoot, IsLinked}` and `Resolve(dir)` backed by a single
      `git -C <dir> rev-parse --path-format=absolute --git-common-dir --git-dir --show-toplevel` invocation
- [x] `Resolve` returns `(nil, nil)` when git is missing from PATH, dir is not a repo, or git errors; log under `--debug`
- [x] Add `Worktree{Root, Branch, IsLocal}` and `ListWorktrees(id)` backed by `git worktree list --porcelain`
- [x] Filter discovery results: skip prunable worktrees, roots missing on disk, and worktrees without a `.taskmd.yaml` (warn under `--verbose`)
- [x] Resolve each sibling worktree's tasks dir by reading its `.taskmd.yaml` directly (raw yaml, no viper), resolving relative `dir`/`task-dir` against the worktree root
- [x] Unit tests for porcelain parsing and filtering (fixture output, no git needed)
- [x] E2E tests (`-tags e2e`): real repo with `git worktree add`; assert identity/discovery, linked-vs-primary, and no-repo/no-git degradation

## Acceptance Criteria

- `Resolve` on a non-git directory and with git absent returns nil identity and no error; no taskmd command changes behavior in that case
- `IsLinked` is true inside a linked worktree, false in the primary
- Prunable/deleted worktrees and worktrees without `.taskmd.yaml` never appear in `ListWorktrees` results
- Exactly one `rev-parse` invocation per `Resolve`; no global viper state is touched
- Package has unit + e2e coverage per the CLI testing policy (≥90% as a core package)
