---
id: "100"
title: "Validate .taskmd.yaml config file"
status: completed
priority: medium
effort: small
dependencies: []
tags:
  - cli
  - go
  - mvp
touches:
  - cli/validate
  - cli/config
created: 2026-02-14
---

# Validate .taskmd.yaml Config File

## Objective

Extend `taskmd validate` to also validate the `.taskmd.yaml` configuration file. Currently validation only covers task files; misconfigurations in the project config (invalid scope definitions, unknown keys, bad types) go undetected.

## Tasks

- [x] Validate `scopes` section: each scope should have a `paths` array of strings
- [x] Warn on unknown top-level keys in `.taskmd.yaml`
- [x] Validate that `touches` values in task files reference scopes defined in config (when scopes are configured)
- [x] Report config validation results alongside task validation results
- [x] Add tests for config validation (valid config, missing paths, unknown keys, orphan touches)

## Acceptance Criteria

- `taskmd validate` checks `.taskmd.yaml` if present
- Invalid `scopes` entries (e.g., missing `paths`, wrong type) are reported as errors
- Task `touches` referencing undefined scopes are reported as warnings
- No errors when config file is absent or has no `scopes` section
- Existing task validation behaviour is unchanged
