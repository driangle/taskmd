---
title: "Apply worktree overlay per-repo under --all-projects"
id: "01m10rq61"
status: completed
priority: high
type: feature
tags: ["worktrees", "all-projects"]
created: "2026-08-27"
completed_at: 2026-08-27
---

# Apply worktree overlay per-repo under --all-projects

## Objective

Close the last functional gap in the worktree-support spec (§7 of
`docs/specs/worktree-support.md`): "the overlay applies per-repo when active."

Today `--all-projects` reads (`list`, `next`, `stats`, ...) scan each registered
project's chosen root raw — `scanProjectTasks` in
`apps/cli/internal/cli/all_projects.go` never calls the overlay builder. A task
that is `in-progress` in a sibling worktree of repo X shows its stale local
status in the cross-project view, even though `taskmd list` run inside repo X
would show the merged effective status. That is the "stale reads" problem the
spec exists to fix, resurfacing through the all-projects path.

For each registered project that is a multi-worktree git repo, the all-projects
loop should build the same overlay the single-project path builds: merge that
repo's sibling worktrees, serve effective status/owner, and carry provenance.

## Tasks

- [x] Wire the overlay builder (`apps/cli/internal/worktree`) into
      `scanProjectTasks` / the all-projects loop, using the already-chosen
      scanned root (current worktree when inside the repo, else registered
      primary) as the local base
- [x] Resolve activation per repo: read each project's own `.taskmd.yaml`
      `worktree_scope` with a raw `os.ReadFile` + `yaml.Unmarshal` (the §2
      viper-bypassing pattern already used for sibling task dirs and
      `resolveProjectScanDir`) — do NOT use the current process's viper state
      for sibling projects
- [x] Precedence for sibling projects: the project's own config decides;
      decide and document whether the `--worktree-scope` flag / env override
      applies globally or only to the current project
- [x] Surface provenance in all-projects output consistently with the
      single-project views (effective status in filters/columns, `worktree` /
      `remote_only` fields in JSON/YAML)
- [x] Ensure repos with `worktree_scope: isolated`, single-worktree repos, and
      non-git projects keep exactly today's behavior
- [x] Update the e2e test that currently asserts local-only behavior
      (`apps/cli/internal/e2e/projects_worktree_test.go`) — kept
      `TestAllProjects_CountsEachRepoOnce` as the dedupe test (its assertion is
      still correct: the current worktree's own claim) and added
      `TestAllProjects_OverlayAppliesPerRepo` for the effective-status case
- [x] Unit tests: activation matrix per project (unified/isolated ×
      multi/single-worktree × non-repo) with injected discovery, per CLAUDE.md
      testing policy
- [x] E2E test: task `in-progress` in a sibling worktree of a registered repo
      is reflected in `list --all-projects` and excluded by
      `next --all-projects`
- [x] Update the spec's §7 status note and the docs site if wording changes
      (run `make sync-spec` if `docs/taskmd_specification.md` is touched)

## Acceptance Criteria

- `taskmd list --all-projects` shows the same effective status for a repo's
  tasks as `taskmd list` run inside that repo
- `taskmd next --all-projects` never recommends a task that is
  in-progress/in-review/completed in any worktree of any registered repo with
  unified scope
- A project with `worktree_scope: isolated` in its own `.taskmd.yaml` is served
  local-only under `--all-projects`, regardless of the invoking project's
  config
- Single-worktree repos and non-git projects produce byte-identical output to
  today under `--all-projects`
- New unit and e2e tests pass; `make test`, `make e2e`, and `make lint` are
  green
