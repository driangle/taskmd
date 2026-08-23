---
title: "Repo-identity resolution in projects registry and all-projects dedupe"
id: "01m0n1nxx"
status: completed
priority: high
type: feature
tags: ["worktrees", "projects", "registry"]
created: "2026-08-22"
dependencies: ["01m0n8cpm"]
effort: medium
phase: worktree-support
completed_at: 2026-08-23
---

# Repo-identity resolution in projects registry and all-projects dedupe

## Objective

Make project identity mean repository identity in the projects registry: all worktrees
of one repo resolve to a single registered project, `--project` works from any
worktree, and `--all-projects` never counts a repo twice (ADR 0005 §1).
Spec: `docs/specs/worktree-support.md` §7.

## Tasks

- [x] `projects register` resolves repo identity via `gitmeta` and stores the primary worktree's root as `path`, even when run from a linked worktree
- [x] Registering a second worktree of an already-registered repo is a friendly no-op ("already registered as `<id>`")
- [x] cwd → project matching (`resolveTaskDir`, `--project` no-op detection) compares repo identity, not path prefixes, when cwd is inside a git repo
- [x] `--project <id>` from any worktree scopes to the *current worktree's* task dir (local base + overlay), never the primary's
- [x] `--all-projects` scans each repo once: the current worktree when inside that repo, else the registered primary
- [x] Fix the stale `~/.taskmd/config.yaml` hint in `projects.go` while touching the file
- [x] Registry tests: register-from-linked-worktree, duplicate-registration no-op, identity-based cwd matching; e2e with `git worktree add`

## Acceptance Criteria

- Registering from a linked worktree stores the primary path; `taskmd projects` lists the repo once
- `taskmd list --project <id>` run inside any worktree of that repo scans that worktree's files
- `--all-projects` output contains each repo's tasks exactly once regardless of worktree count
- `~/.taskmd.yaml` schema unchanged (`{id, name, path}` entries only); non-git registered projects behave exactly as today
