---
title: "Config loader and validate disagree on env task-dir and known keys"
id: "01m3z0re3"
status: pending
priority: medium
type: bug
tags: ["config", "validate"]
created: "2026-10-02"
effort: small
---

# Config loader and validate disagree on env task-dir and known keys

Two related inconsistencies in `.taskmd.yaml` handling, both found while
checking the configuration reference against the binary (2026-10-02).

## Steps to Reproduce

**A. `TASKMD_TASK_DIR` is gated on the config file**

1. In a project whose `.taskmd.yaml` has no `task-dir` key (or no config at all),
   run `TASKMD_TASK_DIR=other taskmd list`.
2. Add `task-dir: tasks` to the file and run the same command.

**B. `validate` rejects keys the loader honours**

1. Put `verbose: true`, `quiet: false`, `projects: []`, `default_project: x` in
   `.taskmd.yaml`.
2. Run `taskmd validate`.

## Expected Behavior

A. The env var overrides the task directory regardless of whether the file has
the key; that is what "flags > env > file > defaults" promises.

B. Either `validate` accepts the keys the loader reads, or the loader ignores
them. One source of truth for "known keys".

## Actual Behavior

A. Step 1 ignores the variable and scans the default directory. Step 2 honours
it. Cause: `resolveTaskDir()` in `apps/cli/internal/cli/root.go` only consults
viper when `viper.InConfig("task-dir")` (or `"dir"`) is true, so the env value
is never reached unless the file has the key. (Same for `TASKMD_DIR` / `dir`.)

B. `validate` reports all four as `unknown config key`, while `verbose` and
`quiet` are viper-bound (`TASKMD_QUIET=true` suppresses output) and
`projects` / `default_project` are read by the registry loader
(`apps/cli/internal/cli/registry.go`). `knownConfigKeys` in
`sdk/go/validator/validator.go` lists 14 keys and omits these four.

The configuration reference now documents both as known limitations; remove
those notes when this lands.

## Tasks

- [ ] A: resolve the task dir as flag > env > config key > default_project > ".", without the `InConfig` gate (keep relative-to-config-file resolution for values that came from the file)
- [ ] A: test `TASKMD_TASK_DIR` with no config file and with a config file lacking the key
- [ ] B: decide the policy for `verbose`/`quiet` (accept or stop binding) and add `projects`/`default_project` to `knownConfigKeys`, or scope the unknown-key check to project-level keys
- [ ] B: test that a global config with `projects` validates clean
- [ ] Update the two "known limitation" notes in `apps/docs/reference/configuration.md`

## Acceptance Criteria

- [ ] `TASKMD_TASK_DIR=x taskmd list` scans `x` with no `.taskmd.yaml` present
- [ ] `taskmd validate` does not warn on keys the CLI actually reads
- [ ] Docs no longer need the limitation notes

## Environment

- OS: macOS
- Version: taskmd 0.8.0
